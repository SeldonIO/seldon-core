package rest

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

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
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const (
	CLOUDEVENTS_HEADER_ID_NAME             = "Ce-Id"
	CLOUDEVENTS_HEADER_SPECVERSION_NAME    = "Ce-Specversion"
	CLOUDEVENTS_HEADER_SOURCE_NAME         = "Ce-Source"
	CLOUDEVENTS_HEADER_TYPE_NAME           = "Ce-Type"
	CLOUDEVENTS_HEADER_PATH_NAME           = "Ce-Path"
	CLOUDEVENTS_HEADER_SPECVERSION_DEFAULT = "0.3"
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

func (r *SeldonRestApi) CreateHttpServer(port int) *http.Server {

	address := fmt.Sprintf("0.0.0.0:%d", port)
	r.Log.Info("Listening", "Address", address)

	// Note that we leave the Server timeouts to 0. This means that the server
	// will apply no timeout at all. Instead, we control that through the
	// http.Client instance making requests to the underlying node graph servers.
	return &http.Server{
		Handler: r.Router,
		Addr:    address,
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

func (r *SeldonRestApi) respondWithError(w http.ResponseWriter, payload payload.SeldonPayload, err error) {

	if serr, ok := err.(*httpStatusError); ok {
		w.WriteHeader(serr.StatusCode)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}

	if payload != nil && payload.GetPayload() != nil {
		w.Header().Set("Content-Type", payload.GetContentType())
		err := r.Client.Marshall(w, payload)
		if err != nil {
			r.Log.Error(err, "Failed to write response")
		}
	} else {
		errPayload := r.Client.CreateErrorPayload(err)
		w.Header().Set("Content-Type", errPayload.GetContentType())
		err = r.Client.Marshall(w, errPayload)
		if err != nil {
			r.Log.Error(err, "Failed to write error payload")
		}
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
		cloudeventHeaderMiddleware := CloudeventHeaderMiddleware{deploymentName: r.DeploymentName, namespace: r.Namespace}
		r.Router.Use(puidHeader)
		r.Router.Use(cloudeventHeaderMiddleware.Middleware)
		switch r.Protocol {
		case api.ProtocolSeldon:
			//v0.1 API
			api01 := r.Router.PathPrefix("/api/v0.1").Methods("POST").Subrouter()
			api01.Handle("/predictions", r.wrapMetrics(metric.PredictionHttpServiceName, r.predictions))
			api01.Handle("/feedback", r.wrapMetrics(metric.FeedbackHttpServiceName, r.feedback))
			r.Router.NewRoute().Path("/api/v0.1/status/{" + ModelHttpPathVariable + "}").Methods("GET").HandlerFunc(r.wrapMetrics(metric.StatusHttpServiceName, r.status))
			r.Router.NewRoute().Path("/api/v0.1/metadata/{" + ModelHttpPathVariable + "}").Methods("GET").HandlerFunc(r.wrapMetrics(metric.MetadataHttpServiceName, r.metadata))
			r.Router.NewRoute().PathPrefix("/api/v0.1/doc/").Handler(http.StripPrefix("/api/v0.1/doc/", http.FileServer(http.Dir("./openapi/"))))
			//v1.0 API
			api10 := r.Router.PathPrefix("/api/v1.0").Methods("POST").Subrouter()
			api10.Handle("/predictions", r.wrapMetrics(metric.PredictionHttpServiceName, r.predictions))
			api10.Handle("/feedback", r.wrapMetrics(metric.FeedbackHttpServiceName, r.feedback))
			r.Router.NewRoute().Path("/api/v1.0/status/{" + ModelHttpPathVariable + "}").Methods("GET").HandlerFunc(r.wrapMetrics(metric.StatusHttpServiceName, r.status))
			r.Router.NewRoute().Path("/api/v1.0/metadata/{" + ModelHttpPathVariable + "}").Methods("GET").HandlerFunc(r.wrapMetrics(metric.MetadataHttpServiceName, r.metadata))
			r.Router.NewRoute().PathPrefix("/api/v1.0/doc/").Handler(http.StripPrefix("/api/v1.0/doc/", http.FileServer(http.Dir("./openapi/"))))

		case api.ProtocolTensorflow:
			r.Router.NewRoute().Path("/v1/models/{" + ModelHttpPathVariable + "}/:predict").Methods("POST").HandlerFunc(r.wrapMetrics(metric.PredictionHttpServiceName, r.predictions))
			r.Router.NewRoute().Path("/v1/models/:predict").Methods("POST").HandlerFunc(r.wrapMetrics(metric.PredictionHttpServiceName, r.predictions)) // Nonstandard path - Seldon extension
			r.Router.NewRoute().Path("/v1/models/{" + ModelHttpPathVariable + "}").Methods("GET").HandlerFunc(r.wrapMetrics(metric.StatusHttpServiceName, r.status))
			r.Router.NewRoute().Path("/v1/models/{" + ModelHttpPathVariable + "}/metadata").Methods("GET").HandlerFunc(r.wrapMetrics(metric.MetadataHttpServiceName, r.metadata))
		case api.ProtocolKfserving:
			r.Router.NewRoute().Path("/v2/models/{" + ModelHttpPathVariable + "}/infer").Methods("POST").HandlerFunc(r.wrapMetrics(metric.PredictionHttpServiceName, r.predictions))
			r.Router.NewRoute().Path("/v1/models/infer").Methods("POST").HandlerFunc(r.wrapMetrics(metric.PredictionHttpServiceName, r.predictions)) // Nonstandard path - Seldon extension
			r.Router.NewRoute().Path("/v2/models/{" + ModelHttpPathVariable + "}/ready").Methods("GET").HandlerFunc(r.wrapMetrics(metric.StatusHttpServiceName, r.status))
			r.Router.NewRoute().Path("/v2/models/{" + ModelHttpPathVariable + "}").Methods("GET").HandlerFunc(r.wrapMetrics(metric.MetadataHttpServiceName, r.metadata))
		}
	}
}

type CloudeventHeaderMiddleware struct {
	deploymentName string
	namespace      string
}

func (h *CloudeventHeaderMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Checking if request is cloudevent based on specname being present
		if _, ok := r.Header[CLOUDEVENTS_HEADER_SPECVERSION_NAME]; ok {
			puid := r.Header.Get(payload.SeldonPUIDHeader)
			w.Header().Set(CLOUDEVENTS_HEADER_ID_NAME, puid)
			w.Header().Set(CLOUDEVENTS_HEADER_SPECVERSION_NAME, CLOUDEVENTS_HEADER_SPECVERSION_DEFAULT)
			w.Header().Set(CLOUDEVENTS_HEADER_PATH_NAME, r.URL.Path)
			w.Header().Set(CLOUDEVENTS_HEADER_TYPE_NAME, "seldon."+h.deploymentName+"."+h.namespace+".response")
			w.Header().Set(CLOUDEVENTS_HEADER_SOURCE_NAME, "seldon."+h.deploymentName)
		}

		next.ServeHTTP(w, r)
	})
}

func puidHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		puid := r.Header.Get(payload.SeldonPUIDHeader)
		if len(puid) == 0 {
			puid = guuid.New().String()
			r.Header.Set(payload.SeldonPUIDHeader, puid)
		}
		if res_puid := w.Header().Get(payload.SeldonPUIDHeader); len(res_puid) == 0 {
			w.Header().Set(payload.SeldonPUIDHeader, puid)
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
		r.respondWithError(w, resPayload, err)
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
		r.respondWithError(w, resPayload, err)
		return
	}
	r.respondWithSuccess(w, http.StatusOK, resPayload)
}

func (r *SeldonRestApi) feedback(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	// Apply tracing if active
	if opentracing.IsGlobalTracerRegistered() {
		var serverSpan opentracing.Span
		ctx, serverSpan = setupTracing(ctx, req, TracingStatusName)
		defer serverSpan.Finish()
	}

	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		r.respondWithError(w, nil, err)
		return
	}

	seldonPredictorProcess := predictor.NewPredictorProcess(ctx, r.Client, logf.Log.WithName(LoggingRestClientName), r.ServerUrl, r.Namespace, req.Header)
	reqPayload, err := seldonPredictorProcess.Client.Unmarshall(bodyBytes)
	if err != nil {
		r.respondWithError(w, nil, err)
		return
	}

	resPayload, err := seldonPredictorProcess.Feedback(r.predictor.Graph, reqPayload)
	if err != nil {
		r.respondWithError(w, resPayload, err)
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
		r.respondWithError(w, nil, err)
		return
	}

	seldonPredictorProcess := predictor.NewPredictorProcess(ctx, r.Client, logf.Log.WithName(LoggingRestClientName), r.ServerUrl, r.Namespace, req.Header)

	reqPayload, err := seldonPredictorProcess.Client.Unmarshall(bodyBytes)
	if err != nil {
		r.respondWithError(w, nil, err)
		return
	}

	var graphNode *v1.PredictiveUnit
	if r.Protocol == api.ProtocolTensorflow {
		vars := mux.Vars(req)
		modelName := vars[ModelHttpPathVariable]
		if modelName != "" {
			if graphNode = v1.GetPredictiveUnit(r.predictor.Graph, modelName); graphNode == nil {
				r.respondWithError(w, nil, fmt.Errorf("Failed to find model %s", modelName))
				return
			}
		} else {
			graphNode = r.predictor.Graph
		}
	} else {
		graphNode = r.predictor.Graph
	}
	resPayload, err := seldonPredictorProcess.Predict(graphNode, reqPayload)
	if err != nil {
		r.respondWithError(w, resPayload, err)
		return
	}
	r.respondWithSuccess(w, http.StatusOK, resPayload)
}
