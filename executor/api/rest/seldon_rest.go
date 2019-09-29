package rest

import (
	"github.com/go-logr/logr"
	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/mux"
	"github.com/prometheus/common/log"
	"io/ioutil"
	"net/http"
	"github.com/seldonio/seldon-core/executor/api/grpc"
	"github.com/seldonio/seldon-core/executor/api/client"
	"github.com/golang/protobuf/jsonpb"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type SeldonRestApi struct {
	Router *mux.Router
	Client *client.SeldonDeploymentClient
	Log logr.Logger
}



func NewSeldonRestApi(client *client.SeldonDeploymentClient) *SeldonRestApi {
	return &SeldonRestApi{
		mux.NewRouter(),
		client,
		logf.Log.WithName("SeldonRestApi"),
	}
}

func respondWithJSON(w http.ResponseWriter, code int, payload proto.Message) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	ma := jsonpb.Marshaler{}
	ma.Marshal(w, payload)
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
	s := r.Router.PathPrefix("/api/v0.1").Methods("POST").HeadersRegexp("Content-Type", "application/json").Subrouter()
	s.HandleFunc("/predictions", r.predictions)
}


func (r *SeldonRestApi) predictions(w http.ResponseWriter, req *http.Request) {
	r.Log.Info("Prediction called")

	sm, err := r.getSeldonMessage(req)
	if err != nil {
		log.Error("Failed to parse request:",err)
	}

	respondWithJSON(w,200, sm)
}