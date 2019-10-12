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

func (smc *SeldonMessageRestClient) call(method string,host string, port int32, req SeldonPayload) (SeldonPayload, error) {
	smc.Log.Info("Call","Methof", method, "host", host, "port",port)
	smReq := req.GetPayload().(*api.SeldonMessage)
	url := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host,strconv.Itoa(int(port))),
		Path:   method,
	}
	sm, err := smc.PostHttp(&url,smReq)
	if err != nil {
		return nil, err
	}
	res := SeldonMessagePayload{sm}
	return &res, nil
}

func (smc *SeldonMessageRestClient) Predict(host string, port int32, req SeldonPayload) (SeldonPayload, error) {
	return smc.call("/predict", host, port, req)
}

func (smc *SeldonMessageRestClient) TransformInput(host string, port int32, req SeldonPayload) (SeldonPayload, error) {
	return smc.call("/transform-input", host, port, req)
}

func (smc *SeldonMessageRestClient) Route(host string, port int32, req SeldonPayload) (int, error) {
	sp, err := smc.call("/route", host, port, req)
	if err != nil {
		return 0, err
	} else {
		routes := ExtractRoute(sp.GetPayload().(*api.SeldonMessage))
		//Only returning first route. API could be extended to allow multiple routes
		return routes[0], nil
	}
}

func (smc *SeldonMessageRestClient) Combine(host string, port int32, msgs []SeldonPayload) (SeldonPayload, error) {
	sms := make([]*api.SeldonMessage, len(msgs))
	for i,sm := range msgs {
		sms[i] = sm.GetPayload().(*api.SeldonMessage)
	}
	sml := api.SeldonMessageList{SeldonMessages: sms}
	req := SeldonMessageListPayload{&sml}
	return smc.call("/aggregate",host, port, &req)
}

func (smc *SeldonMessageRestClient) TransformOutput(host string, port int32, req SeldonPayload) (SeldonPayload, error) {
	return smc.call("/transform-output", host, port, req)
}
