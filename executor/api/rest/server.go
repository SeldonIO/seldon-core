package rest

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	guuid "github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/seldonio/seldon-core/executor/api"
	"github.com/seldonio/seldon-core/executor/api/client"
	"github.com/seldonio/seldon-core/executor/api/metric"
	"github.com/seldonio/seldon-core/executor/api/payload"
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
	prometheusPath string
}

func NewServerRestApi(predictor *v1.PredictorSpec, client client.SeldonApiClient, probesOnly bool, serverUrl *url.URL, namespace string, protocol string, deploymentName string, prometheusPath string) *SeldonRestApi {
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
		prometheusPath,
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
	r.Router.Handle(r.prometheusPath, promhttp.Handler())
	if !r.ProbesOnly {
		r.Router.Use(puidHeader)
		switch r.Protocol {
		case api.ProtocolSeldon:
			//v0.1 API
			api01 := r.Router.PathPrefix("/api/v0.1").Methods("POST").Subrouter()
			api01.Handle("/predictions", r.wrapMetrics(metric.PredictionHttpServiceName, r.predictions))
			r.Router.NewRoute().Path("/api/v0.1/status/{" + ModelHttpPathVariable + "}").Methods("GET").HandlerFunc(r.wrapMetrics(metric.StatusHttpServiceName, r.status))
			r.Router.NewRoute().Path("/api/v0.1/metadata/{" + ModelHttpPathVariable + "}").Methods("GET").HandlerFunc(r.wrapMetrics(metric.StatusHttpServiceName, r.metadata))
			//v1.0 API
			api1 := r.Router.PathPrefix("/api/v1.0").Methods("POST").Subrouter()
			api1.Handle("/predictions", r.wrapMetrics(metric.PredictionServiceMetricName, r.predictions))
			r.Router.NewRoute().Path("/api/v1.0/status/{" + ModelHttpPathVariable + "}").Methods("GET").HandlerFunc(r.wrapMetrics(metric.StatusHttpServiceName, r.status))
			r.Router.NewRoute().Path("/api/v1.0/metadata/{" + ModelHttpPathVariable + "}").Methods("GET").HandlerFunc(r.wrapMetrics(metric.StatusHttpServiceName, r.metadata))

		case api.ProtocolTensorflow:
			r.Router.NewRoute().Path("/v1/models/{" + ModelHttpPathVariable + "}/:predict").Methods("POST").HandlerFunc(r.wrapMetrics(metric.PredictionHttpServiceName, r.predictions))
			r.Router.NewRoute().Path("/v1/models/{" + ModelHttpPathVariable + "}").Methods("GET").HandlerFunc(r.wrapMetrics(metric.StatusHttpServiceName, r.status))
			r.Router.NewRoute().Path("/v1/models/{" + ModelHttpPathVariable + "}/metadata").Methods("GET").HandlerFunc(r.wrapMetrics(metric.MetadataHttpServiceName, r.metadata))
		}
	}
}

func puidHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if puid := r.Header.Get(payload.SeldonPUIDHeader); puid == "" {
			r.Header.Set(payload.SeldonPUIDHeader, guuid.New().String())
		}
		next.ServeHTTP(w, r)
	})
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

func (r *SeldonRestApi) failWithError(w http.ResponseWriter, err error) {
	r.Log.Error(err, "Failed")
	r.respondWithError(w, err)
}

func getGraphNodeForModelName(req *http.Request, graph *v1.PredictiveUnit) (*v1.PredictiveUnit, error) {
	vars := mux.Vars(req)
	modelName := vars[ModelHttpPathVariable]
	if graphNode := v1.GetPredictiveUnit(graph, modelName); graphNode == nil {
		return nil, fmt.Errorf("Failed to find model %s", modelName)
	} else {
		return graphNode, nil
	}
}

func setupTracing(ctx context.Context, req *http.Request, spanName string) (context.Context, opentracing.Span) {
	tracer := opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	serverSpan := tracer.StartSpan(spanName, ext.RPCServerOption(spanCtx))
	ctx = opentracing.ContextWithSpan(ctx, serverSpan)
	return ctx, serverSpan
}

func (r *SeldonRestApi) metadata(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	// Apply tracing if active
	if opentracing.IsGlobalTracerRegistered() {
		var serverSpan opentracing.Span
		ctx, serverSpan = setupTracing(ctx, req, TracingMetadataName)
		defer serverSpan.Finish()
	}

	vars := mux.Vars(req)
	modelName := vars[ModelHttpPathVariable]

	seldonPredictorProcess := predictor.NewPredictorProcess(ctx, r.Client, logf.Log.WithName(LoggingRestClientName), r.ServerUrl, r.Namespace, req.Header)
	resPayload, err := seldonPredictorProcess.Metadata(r.predictor.Graph, modelName, nil)
	if err != nil {
		r.failWithError(w, err)
		return
	}
	r.respondWithSuccess(w, http.StatusOK, resPayload)
}

func (r *SeldonRestApi) status(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	// Apply tracing if active
	if opentracing.IsGlobalTracerRegistered() {
		var serverSpan opentracing.Span
		ctx, serverSpan = setupTracing(ctx, req, TracingStatusName)
		defer serverSpan.Finish()
	}

	vars := mux.Vars(req)
	modelName := vars[ModelHttpPathVariable]

	seldonPredictorProcess := predictor.NewPredictorProcess(ctx, r.Client, logf.Log.WithName(LoggingRestClientName), r.ServerUrl, r.Namespace, req.Header)
	resPayload, err := seldonPredictorProcess.Status(r.predictor.Graph, modelName, nil)
	if err != nil {
		r.failWithError(w, err)
		return
	}
	r.respondWithSuccess(w, http.StatusOK, resPayload)
}

func (r *SeldonRestApi) predictions(w http.ResponseWriter, req *http.Request) {
	r.Log.Info("Predictions called")

	ctx := req.Context()
	// Add Seldon Puid to Context
	ctx = context.WithValue(ctx, payload.SeldonPUIDHeader, req.Header.Get(payload.SeldonPUIDHeader))

	// Apply tracing if active
	if opentracing.IsGlobalTracerRegistered() {
		var serverSpan opentracing.Span
		ctx, serverSpan = setupTracing(ctx, req, TracingPredictionsName)
		defer serverSpan.Finish()
	}

	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		r.failWithError(w, err)
		return
	}

	seldonPredictorProcess := predictor.NewPredictorProcess(ctx, r.Client, logf.Log.WithName(LoggingRestClientName), r.ServerUrl, r.Namespace, req.Header)

	reqPayload, err := seldonPredictorProcess.Client.Unmarshall(bodyBytes)
	if err != nil {
		r.failWithError(w, err)
		return
	}

	var graphNode *v1.PredictiveUnit
	if r.Protocol == api.ProtocolTensorflow {
		graphNode, err = getGraphNodeForModelName(req, r.predictor.Graph)
		if err != nil {
			r.failWithError(w, err)
			return
		}
	} else {
		graphNode = r.predictor.Graph
	}
	resPayload, err := seldonPredictorProcess.Predict(graphNode, reqPayload)
	if err != nil {
		r.failWithError(w, err)
		return
	}
	r.respondWithSuccess(w, http.StatusOK, resPayload)
}
