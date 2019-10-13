package rest

import (
	"github.com/go-logr/logr"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/jsonpb"
	"github.com/gorilla/mux"
	"github.com/prometheus/common/log"
	"github.com/seldonio/seldon-core/executor/api/client"
	api "github.com/seldonio/seldon-core/executor/api/grpc"
	"github.com/seldonio/seldon-core/executor/api/machinelearning/v1alpha2"
	"github.com/seldonio/seldon-core/executor/predictor"
	"io/ioutil"
	"net/http"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type SeldonRestApi struct {
	Router    *mux.Router
	predictor *v1alpha2.PredictorSpec
	Log       logr.Logger
}

func NewSeldonRestApi(predictor *v1alpha2.PredictorSpec) *SeldonRestApi {
	return &SeldonRestApi{
		mux.NewRouter(),
		predictor,
		logf.Log.WithName("SeldonRestApi"),
	}
}

func (r *SeldonRestApi) respondWithJSON(w http.ResponseWriter, code int, payload proto.Message) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	ma := jsonpb.Marshaler{}
	err := ma.Marshal(w, payload)
	if err != nil {
		r.Log.Error(err, "Failed to write response")
	}
}

// Extract a SeldonMessage proto from the REST request
func (r *SeldonRestApi) getSeldonMessage(req *http.Request) (*api.SeldonMessage, error) {
	var sm api.SeldonMessage
	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	value := string(bodyBytes)
	if err := jsonpb.UnmarshalString(value, &sm); err != nil {
		return nil, err
	}
	return &sm, nil
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

	sm, err := r.getSeldonMessage(req)
	if err != nil {
		log.Error("Failed to parse request:", err)
	}

	seldonPredictorProcess := &predictor.PredictorProcess{
		client.NewSeldonMessageRestClient(),
		logf.Log.WithName("SeldonMessageRestClient"),
	}

	reqPayload := client.SeldonMessagePayload{sm}
	resPayload, err := seldonPredictorProcess.Execute(r.predictor.Graph, &reqPayload)
	if err != nil {
		respFailed := api.SeldonMessage{Status: &api.Status{Code: http.StatusInternalServerError, Info: err.Error()}}
		r.respondWithJSON(w, http.StatusInternalServerError, &respFailed)
	} else {
		smResp := resPayload.GetPayload().(*api.SeldonMessage)
		r.respondWithJSON(w, http.StatusOK, smResp)
	}

}
