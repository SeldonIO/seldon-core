package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/gorilla/mux"
	"github.com/seldonio/seldon-core/examples/wrappers/go/pkg/api"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
)

var (
	port       = flag.Int("port", 10000, "The server port")
	serverType = flag.String("server_type", "grpc", "The type of server grpc or rest")
)

// One struct for each type of Seldon Server. Here we just create one for MODELs
type ModelServer struct {
}

// Example Predict call with SeldonMessage proto
func  (s *ModelServer) Predict(ctx context.Context, m *api.SeldonMessage) (*api.SeldonMessage, error){

	test := &api.SeldonMessage{
		Status: &api.Status{
			Code:                 0,
			Info:                 "",
			Reason:               "",
			Status:               0,

		},
		Meta: &api.Meta{
			Puid:                 "",
			Tags:                 nil,
			Routing:              nil,
			RequestPath:          nil,
			Metrics:              nil,

		},
		DataOneof:            &api.SeldonMessage_Data{
			Data: &api.DefaultData{
				Names:                nil,
				DataOneof:            &api.DefaultData_Tensor{
					Tensor: &api.Tensor{
						Shape:                []int32{1,1},
						Values:               []float64{1, 3},
						XXX_NoUnkeyedLiteral: struct{}{},
						XXX_unrecognized:     nil,
						XXX_sizecache:        0,
					},
				},

			},
		},

	}
	return test, nil
}

// Feedback template
func (s *ModelServer) SendFeedback(ctx context.Context, f *api.Feedback) (*api.SeldonMessage, error) {
	return &api.SeldonMessage{}, nil
}

func handleError(w http.ResponseWriter,msg string) {
	ma := jsonpb.Marshaler{}
	errJson := &api.SeldonMessage{
		Status: &api.Status{
			Code:   400,
			Info:   "Failed",
			Reason: msg,
			Status: 1,
		},
	}
	_ = ma.Marshal(w, errJson)
}

// REST predict call. Extract parameter and send to Proto version
func RestPredict(w http.ResponseWriter, r *http.Request) {
	log.Println("REST Predict called")
	ma := jsonpb.Marshaler{}
	sm := &api.SeldonMessage{}
	err := r.ParseForm()
	if err != nil {
		handleError(w,"Failed to parse request")
	}
	value := r.FormValue("json")
	if err := jsonpb.UnmarshalString(value, sm); err != nil {
		log.Println("Error converting JSON to proto:", err)
		handleError(w,"Failed to extract json from request")
		return
	}

	log.Printf("message is %v\n",sm)
	modelServer := &ModelServer{}
	sm_resp, _ := modelServer.Predict(r.Context(),sm)
	log.Printf("message is %v\n",sm_resp)
	_ = ma.Marshal(w, sm_resp)
}

// REST SendFeedback call. Extract parameters and send to Proto version.
func RestSendFeedback(w http.ResponseWriter, r *http.Request) {
	log.Println("REST SendFeedback called")
	fe := &api.Feedback{}
	err := r.ParseForm()
	if err != nil {
		handleError(w,"Failed to parse request")
	}
	value := r.FormValue("json")
	if err := jsonpb.UnmarshalString(value, fe); err != nil {
		log.Println("Error converting JSON to proto:", err)
		handleError(w,"Failed to extract json from request")
		return
	}
	log.Printf("message is %v\n",fe)
	modelServer := &ModelServer{}
	sm_resp, _ := modelServer.SendFeedback(r.Context(),fe)
	log.Printf("message is %v\n",sm_resp)
	ma := jsonpb.Marshaler{}
	_ = ma.Marshal(w, sm_resp)
}


/*
  Start gRPC or REST server
 */
func main() {
	flag.Parse()
	var portEnv = os.Getenv("PREDICTIVE_UNIT_SERVICE_PORT")
	if portEnv != "" {
		portVal,err := strconv.Atoi(portEnv)
		if err != nil {
			log.Fatal("Bad env variable PREDICTIVE_UNIT_SERVICE_PORT")
		} else {
			port = &portVal
		}
	}
	log.Printf("Server_type: %s\n",*serverType)
	log.Printf("Port: %d\n",*port)
	if *serverType == "rest" {
		log.Println("Starting REST Server")
		router := mux.NewRouter()
		router.HandleFunc("/predict", RestPredict).Methods("GET","POST")
		router.HandleFunc("/send-feedback", RestSendFeedback).Methods("GET","POST")
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), router))
	} else {
		log.Println("Starting gRPC Server")
		flag.Parse()
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		var opts []grpc.ServerOption
		grpcServer := grpc.NewServer(opts...)
		api.RegisterModelServer(grpcServer, &ModelServer{})
		grpcServer.Serve(lis)
	}


}