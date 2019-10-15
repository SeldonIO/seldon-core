package rest

import (
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/prometheus/common/log"
	"github.com/seldonio/seldon-core/executor/api/client"
	"github.com/seldonio/seldon-core/executor/api/machinelearning/v1alpha2"
	"github.com/seldonio/seldon-core/executor/predictor"
	"io/ioutil"
	"net/http"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type SeldonRestApi struct {
	Router    *mux.Router
	Client    client.SeldonApiClient
	predictor *v1alpha2.PredictorSpec
	Log       logr.Logger
}

func NewSeldonRestApi(predictor *v1alpha2.PredictorSpec, client client.SeldonApiClient) *SeldonRestApi {
	return &SeldonRestApi{
		mux.NewRouter(),
		client,
		predictor,
		logf.Log.WithName("SeldonRestApi"),
	}
}

func (r *SeldonRestApi) respondWithJSON(w http.ResponseWriter, code int, payload client.SeldonPayload) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	err := r.Client.Marshall(w, payload)
	if err != nil {
		r.Log.Error(err, "Failed to write response")
	}
}

func (r *SeldonRestApi) respondWithError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)

	errPayload, err2 := r.Client.CreateErrorPayload(err)
	if err2 != nil {
		r.Log.Error(err2, "Can't create error payload")
	}
	err2 = r.Client.Marshall(w, errPayload)
	if err2 != nil {
		r.Log.Error(err, "Failed to write error payload")
	}
}

func (r *SeldonRestApi) Initialise() {
	r.Router.HandleFunc("/ready", r.checkReady)
	r.Router.HandleFunc("/live", r.alive)
	s := r.Router.PathPrefix("/api/v0.1").Methods("POST").HeadersRegexp("Content-Type", "application/json").Subrouter()
	s.HandleFunc("/predictions", r.predictions)
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

func (r *SeldonRestApi) predictions(w http.ResponseWriter, req *http.Request) {
	r.Log.Info("Prediction called")

	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("Failed to get body", err)
		r.respondWithError(w, err)
		return
	}

	seldonPredictorProcess := &predictor.PredictorProcess{
		Client: client.NewSeldonMessageRestClient(),
		Log:    logf.Log.WithName("SeldonMessageRestClient"),
	}

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
	r.respondWithJSON(w, http.StatusOK, resPayload)
}
