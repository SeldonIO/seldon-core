package rest

import (
	"context"
	"encoding/json"
	"fmt"
	http2 "github.com/cloudevents/sdk-go/pkg/bindings/http"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/go-logr/logr"
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
	"time"
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
		Handler:     r.Router,
		Addr:        address,
		IdleTimeout: 65 * time.Second,
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
		r.Router.Use(xssMiddleware)
		r.Router.Use(mux.CORSMethodMiddleware(r.Router))
		r.Router.Use(handleCORSRequests)

		switch r.Protocol {
		case api.ProtocolSeldon:
			//v0.1 API
			api01 := r.Router.PathPrefix("/api/v0.1").Methods("OPTIONS", "POST").Subrouter()
			api01.Handle("/predictions", r.wrapMetrics(metric.PredictionHttpServiceName, r.predictions))
			api01.Handle("/feedback", r.wrapMetrics(metric.FeedbackHttpServiceName, r.feedback))
			r.Router.NewRoute().Path("/api/v0.1/status/{"+ModelHttpPathVariable+"}").Methods("GET", "OPTIONS").HandlerFunc(r.wrapMetrics(metric.StatusHttpServiceName, r.status))
			r.Router.NewRoute().Path("/api/v0.1/metadata/{"+ModelHttpPathVariable+"}").Methods("GET", "OPTIONS").HandlerFunc(r.wrapMetrics(metric.MetadataHttpServiceName, r.metadata))

			r.Router.NewRoute().PathPrefix("/api/v0.1/doc/").Handler(http.StripPrefix("/api/v0.1/doc/", http.FileServer(http.Dir("./openapi/"))))
			//v1.0 API
			api10 := r.Router.PathPrefix("/api/v1.0").Methods("OPTIONS", "POST").Subrouter()
			api10.Handle("/predictions", r.wrapMetrics(metric.PredictionHttpServiceName, r.predictions))
			api10.Handle("/feedback", r.wrapMetrics(metric.FeedbackHttpServiceName, r.feedback))
			r.Router.NewRoute().Path("/api/v1.0/status/{"+ModelHttpPathVariable+"}").Methods("GET", "OPTIONS").HandlerFunc(r.wrapMetrics(metric.StatusHttpServiceName, r.status))
			r.Router.NewRoute().Path("/api/v1.0/metadata").Methods("GET", "OPTIONS").HandlerFunc(r.wrapMetrics(metric.MetadataHttpServiceName, r.graphMetadata))
			r.Router.NewRoute().Path("/api/v1.0/metadata/{"+ModelHttpPathVariable+"}").Methods("GET", "OPTIONS").HandlerFunc(r.wrapMetrics(metric.MetadataHttpServiceName, r.metadata))
			r.Router.NewRoute().PathPrefix("/api/v1.0/doc/").Handler(http.StripPrefix("/api/v1.0/doc/", http.FileServer(http.Dir("./openapi/"))))
		case api.ProtocolTensorflow:
			r.Router.NewRoute().Path("/v1/models/{"+ModelHttpPathVariable+"}/:predict").Methods("OPTIONS", "POST").HandlerFunc(r.wrapMetrics(metric.PredictionHttpServiceName, r.predictions))
			r.Router.NewRoute().Path("/v1/models/{"+ModelHttpPathVariable+"}:predict").Methods("OPTIONS", "POST").HandlerFunc(r.wrapMetrics(metric.PredictionHttpServiceName, r.predictions))
			// Allow both :predict before and after final / in path.
			r.Router.NewRoute().Path("/v1/models/:predict").Methods("OPTIONS", "POST").HandlerFunc(r.wrapMetrics(metric.PredictionHttpServiceName, r.predictions)) // Nonstandard path - Seldon extension
			r.Router.NewRoute().Path("/v1/models:predict").Methods("OPTIONS", "POST").HandlerFunc(r.wrapMetrics(metric.PredictionHttpServiceName, r.predictions))  // Nonstandard path - Seldon extension
			r.Router.NewRoute().Path("/v1/models/{"+ModelHttpPathVariable+"}").Methods("GET", "OPTIONS").HandlerFunc(r.wrapMetrics(metric.StatusHttpServiceName, r.status))
			r.Router.NewRoute().Path("/v1/models/{"+ModelHttpPathVariable+"}/metadata").Methods("GET", "OPTIONS").HandlerFunc(r.wrapMetrics(metric.MetadataHttpServiceName, r.metadata))
			// Enabling for standard seldon core feedback API endpoint with standard schema
			r.Router.NewRoute().Path("/api/v1.0/feedback").Methods("OPTIONS", "POST").HandlerFunc(r.wrapMetrics(metric.FeedbackHttpServiceName, r.feedback))
		case api.ProtocolKFServing:
			r.Router.NewRoute().Path("/v2/models/{"+ModelHttpPathVariable+"}/infer").Methods("OPTIONS", "POST").HandlerFunc(r.wrapMetrics(metric.PredictionHttpServiceName, r.predictions))
			r.Router.NewRoute().Path("/v2/models/infer").Methods("OPTIONS", "POST").HandlerFunc(r.wrapMetrics(metric.PredictionHttpServiceName, r.predictions)) // Nonstandard path - Seldon extension
			r.Router.NewRoute().Path("/v2/models/{"+ModelHttpPathVariable+"}/ready").Methods("GET", "OPTIONS").HandlerFunc(r.wrapMetrics(metric.StatusHttpServiceName, r.status))
			r.Router.NewRoute().Path("/v2/models/{"+ModelHttpPathVariable+"}").Methods("GET", "OPTIONS").HandlerFunc(r.wrapMetrics(metric.MetadataHttpServiceName, r.metadata))

		}
	}
}

func (r *SeldonRestApi) checkReady(w http.ResponseWriter, req *http.Request) {
	err := predictor.Ready(&r.predictor.Graph)
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
	resPayload, err := seldonPredictorProcess.Metadata(&r.predictor.Graph, modelName, nil)
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
	resPayload, err := seldonPredictorProcess.Status(&r.predictor.Graph, modelName, nil)
	if err != nil {
		r.respondWithError(w, resPayload, err)
		return
	}
	r.respondWithSuccess(w, http.StatusOK, resPayload)
}

func (r *SeldonRestApi) feedback(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	ctx = context.WithValue(ctx, payload.SeldonPUIDHeader, req.Header.Get(payload.SeldonPUIDHeader))

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
	reqPayload, err := seldonPredictorProcess.Client.Unmarshall(bodyBytes, req.Header.Get(http2.ContentType))
	if err != nil {
		r.respondWithError(w, nil, err)
		return
	}

	resPayload, err := seldonPredictorProcess.Feedback(&r.predictor.Graph, reqPayload)
	if err != nil {
		r.respondWithError(w, resPayload, err)
		return
	}
	r.respondWithSuccess(w, http.StatusOK, resPayload)
}

func (r *SeldonRestApi) predictions(w http.ResponseWriter, req *http.Request) {
	r.Log.V(1).Info("Predictions called")

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

	reqPayload, err := seldonPredictorProcess.Client.Unmarshall(bodyBytes, req.Header.Get(http2.ContentType))
	if err != nil {
		r.respondWithError(w, nil, err)
		return
	}

	var graphNode *v1.PredictiveUnit
	if r.Protocol == api.ProtocolTensorflow {
		vars := mux.Vars(req)
		modelName := vars[ModelHttpPathVariable]
		if modelName != "" {
			if graphNode = v1.GetPredictiveUnit(&r.predictor.Graph, modelName); graphNode == nil {
				r.respondWithError(w, nil, fmt.Errorf("Failed to find model %s", modelName))
				return
			}
		} else {
			graphNode = &r.predictor.Graph
		}
	} else {
		graphNode = &r.predictor.Graph
	}
	resPayload, err := seldonPredictorProcess.Predict(graphNode, reqPayload)
	if err != nil {
		r.respondWithError(w, resPayload, err)
		return
	}
	r.respondWithSuccess(w, http.StatusOK, resPayload)
}

func (r *SeldonRestApi) graphMetadata(w http.ResponseWriter, req *http.Request) {
	r.Log.V(1).Info("Graph Metadata called.")

	ctx := req.Context()

	// Apply tracing if active
	if opentracing.IsGlobalTracerRegistered() {
		var serverSpan opentracing.Span
		ctx, serverSpan = setupTracing(ctx, req, TracingMetadataName)
		defer serverSpan.Finish()
	}

	seldonPredictorProcess := predictor.NewPredictorProcess(ctx, r.Client, logf.Log.WithName(LoggingRestClientName), r.ServerUrl, r.Namespace, req.Header)

	graphMetadata, err := seldonPredictorProcess.GraphMetadata(r.predictor)

	if err != nil {
		r.respondWithError(w, nil, err)
		return
	}

	msg, _ := json.Marshal(graphMetadata)
	resPayload := payload.BytesPayload{Msg: msg, ContentType: ContentTypeJSON}

	r.respondWithSuccess(w, http.StatusOK, &resPayload)
}
