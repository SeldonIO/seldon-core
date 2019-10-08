package client

import (
	"bytes"
	"github.com/go-logr/logr"
	"github.com/golang/protobuf/jsonpb"
	"github.com/pkg/errors"
	api "github.com/seldonio/seldon-core/executor/api/grpc"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strconv"
)

const (
	ContentTypeJSON = "application/json"
)

type SeldonMessageRestClient struct {
	Log logr.Logger
}

func NewSeldonMessageRestClient() SeldonApiClient {
	client := SeldonMessageRestClient{
		logf.Log.WithName("SeldonMessageRestClient"),
	}
	return &client
}



func (smc *SeldonMessageRestClient) Hello() int {
	return 1
}

func (smc *SeldonMessageRestClient) PostHttp(url *url.URL, msg *api.SeldonMessage) (*api.SeldonMessage, error) {
	smc.Log.Info("Calling HTTP","URL",url)

	// Marshall message into JSON
	ma := jsonpb.Marshaler{}
	msgStr, err := ma.MarshalToString(msg)
	if err != nil {
		return nil, err
	}

	// Call URL
	response, err := http.Post(url.String(), ContentTypeJSON, bytes.NewBufferString(msgStr))
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		smc.Log.Info("httpPost failed","response code", response.StatusCode)
		return nil, errors.Errorf("Internal service call failed with to %s status code %d",url,response.StatusCode)
	}

	//Read response
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if err = response.Body.Close(); err != nil {
		return nil, err
	}

	// Return SeldonMessage
	var sm api.SeldonMessage
	value := string(b)
	if err := jsonpb.UnmarshalString(value, &sm); err != nil {
		return nil, err
	}
	return &sm, nil
}


func (smc *SeldonMessageRestClient) Predict(host string, port int32, req SeldonPayload) (SeldonPayload, error) {
	smc.Log.Info("Predict","port",port)

	smReq := req.GetPayload().(*api.SeldonMessage)

	url := url.URL{
	Scheme: "http",
	Host:   net.JoinHostPort(host,strconv.Itoa(int(port))),
	Path:   "/predict",
	}

	sm, err := smc.PostHttp(&url,smReq)
	if err != nil {
		return nil, err
	}

	res := SeldonMessagePayload{sm}

	return &res, nil
}

func (smc *SeldonMessageRestClient) TransformInput(host string, port int32, req SeldonPayload) (SeldonPayload, error) {
	smc.Log.Info("Predict","port",port)
	smReq := req.GetPayload().(*api.SeldonMessage)
	url := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host,strconv.Itoa(int(port))),
		Path:   "/transform-input",
	}
	sm,  err:=  smc.PostHttp(&url,smReq)
	if err != nil {
		return nil, err
	}
	res := SeldonMessagePayload{sm}

	return &res, nil
}

