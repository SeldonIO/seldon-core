package rest

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/seldonio/seldon-core/executor/api/client"
	"github.com/seldonio/seldon-core/executor/api/metric"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/logger"
	"github.com/seldonio/seldon-core/executor/predictor"
	"github.com/seldonio/seldon-core/operator/apis/machinelearning/v1"
	"io/ioutil"
	"net/http"
	"net/url"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type SeldonRestApi struct {
	Router         *mux.Router
	Client         client.SeldonApiClient
	predictor      *v1.PredictorSpec
	Log            logr.Logger
	ProbesOnly     bool
	ServerUrl      *url.URL
	Namespace      string
	Protocol       string
	DeploymentName string
	metrics        *metric.ServerMetrics
}

func NewSeldonRestApi(predictor *v1.PredictorSpec, client client.SeldonApiClient, probesOnly bool, serverUrl *url.URL, namespace string, protocol string, deploymentName string) *SeldonRestApi {
	var serverMetrics *metric.ServerMetrics
	if !probesOnly {
		serverMetrics = metric.NewServerMetrics(predictor, deploymentName)
	}
	return &SeldonRestApi{
		mux.NewRouter(),
		client,
		predictor,
		logf.Log.WithName("SeldonRestApi"),
		probesOnly,
		serverUrl,
		namespace,
		protocol,
		deploymentName,
		serverMetrics,
	}
}

func (r *SeldonRestApi) respondWithSuccess(w http.ResponseWriter, code int, payload payload.SeldonPayload) {
	w.Header().Set("Content-Type", payload.GetContentType())
	w.WriteHeader(code)

	err := r.Client.Marshall(w, payload)
	if err != nil {
		r.Log.Error(err, "Failed to write response")
	}
}

func (r *SeldonRestApi) respondWithError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)

	errPayload := r.Client.CreateErrorPayload(err)
	err = r.Client.Marshall(w, errPayload)
	if err != nil {
		r.Log.Error(err, "Failed to write error payload")
	}
}

func (r *SeldonRestApi) wrapMetrics(service string, baseHandler http.HandlerFunc) http.HandlerFunc {

	handler := promhttp.InstrumentHandlerDuration(
		r.metrics.ServerHandledHistogram.MustCurryWith(prometheus.Labels{
			metric.DeploymentNameMetric:   r.DeploymentName,
			metric.PredictorNameMetric:    r.predictor.Name,
			metric.PredictorVersionMetric: r.predictor.Annotations["version"],
			metric.ServiceMetric:          service}),
		baseHandler,
	)
	return handler
}

func (r *SeldonRestApi) Initialise() {
	r.Router.HandleFunc("/ready", r.checkReady)
	r.Router.HandleFunc("/live", r.alive)
	r.Router.Handle("/metrics", promhttp.Handler())
	if !r.ProbesOnly {
		//predictionsHandler := r.wrapMetrics(metric.PredictionHttpServiceName, r.predictions)
		switch r.Protocol {
		case ProtocolSeldon:
			api01 := r.Router.PathPrefix("/api/v0.1").Methods("POST").Subrouter()
			api01.Handle("/predictions", r.wrapMetrics("/api/v0.1/predictions", r.predictions))
			api1 := r.Router.PathPrefix("/api/v1.0").Methods("POST").Subrouter()
			api1.Handle("/predictions", r.wrapMetrics("/api/v1/predictions", r.predictions))
		case ProtocolTensorflow:
			r.Router.NewRoute().Path("/v1/models:predict").Methods("POST").HandlerFunc(r.wrapMetrics("/v1/models:predict", r.predictions))
		}
	}
}

func (r *SeldonRestApi) checkReady(w http.ResponseWriter, req *http.Request) {
	err := predictor.Ready(r.predictor.Graph)
	if err != nil {
		r.Log.Error(err, "Ready check failed")
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func (r *SeldonRestApi) alive(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func getEventId(req *http.Request) string {
	return req.Header.Get(logger.CloudEventsIdHeader)
}

func (r *SeldonRestApi) predictions(w http.ResponseWriter, req *http.Request) {
	r.Log.Info("Predictions called")

	ctx := context.Background()
	// Apply tracing if active
	if opentracing.IsGlobalTracerRegistered() {
		tracer := opentracing.GlobalTracer()
		spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
		serverSpan := tracer.StartSpan("predictions_rest", ext.RPCServerOption(spanCtx))
		ctx = opentracing.ContextWithSpan(ctx, serverSpan)
		defer serverSpan.Finish()
	}

	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("Failed to get body", err)
		r.respondWithError(w, err)
		return
	}

	seldonPredictorProcess := predictor.NewPredictorProcess(ctx, r.Client, logf.Log.WithName("SeldonMessageRestClient"), getEventId(req), r.ServerUrl, r.Namespace)

	reqPayload, err := seldonPredictorProcess.Client.Unmarshall(bodyBytes)
	if err != nil {
		log.Error("Failed to get body", err)
		r.respondWithError(w, err)
		return
	}

	resPayload, err := seldonPredictorProcess.Execute(r.predictor.Graph, reqPayload)
	if err != nil {
		log.Error("Failed to get predictions", err)
		r.respondWithError(w, err)
		return
	}
	r.respondWithSuccess(w, http.StatusOK, resPayload)
}
