package rest

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/common/log"
	"github.com/seldonio/seldon-core/executor/api/client"
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
	Router     *mux.Router
	Client     client.SeldonApiClient
	predictor  *v1.PredictorSpec
	Log        logr.Logger
	ProbesOnly bool
	ServerUrl  *url.URL
	Namespace  string
}

func NewSeldonRestApi(predictor *v1.PredictorSpec, client client.SeldonApiClient, probesOnly bool, serverUrl *url.URL, namespace string) *SeldonRestApi {
	return &SeldonRestApi{
		mux.NewRouter(),
		client,
		predictor,
		logf.Log.WithName("SeldonRestApi"),
		probesOnly,
		serverUrl,
		namespace,
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

func (r *SeldonRestApi) Initialise() {
	r.Router.HandleFunc("/ready", r.checkReady)
	r.Router.HandleFunc("/live", r.alive)
	if !r.ProbesOnly {
		api01 := r.Router.PathPrefix("/api/v0.1").Methods("POST").Subrouter()
		api01.HandleFunc("/predictions", r.predictions)
		api1 := r.Router.PathPrefix("/api/v1").Methods("POST").Subrouter()
		api1.HandleFunc("/predictions", r.predictions)
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
