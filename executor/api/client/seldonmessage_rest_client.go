package client

import (
	"bytes"
	"github.com/go-logr/logr"
	"github.com/golang/protobuf/jsonpb"
	"github.com/pkg/errors"
	api "github.com/seldonio/seldon-core/executor/api/grpc"
	"io"
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
	httpClient *http.Client
	Log        logr.Logger
}

func (smc *SeldonMessageRestClient) CreateErrorPayload(err error) (SeldonPayload, error) {
	respFailed := api.SeldonMessage{Status: &api.Status{Code: http.StatusInternalServerError, Info: err.Error()}}
	res := SeldonMessagePayload{&respFailed}
	return &res, nil
}

func (smc *SeldonMessageRestClient) Marshall(w io.Writer, msg SeldonPayload) error {
	ma := jsonpb.Marshaler{}
	return ma.Marshal(w, msg.GetPayload().(*api.SeldonMessage))
}

func (smc *SeldonMessageRestClient) Unmarshall(msg []byte) (SeldonPayload, error) {
	var sm api.SeldonMessage
	value := string(msg)
	if err := jsonpb.UnmarshalString(value, &sm); err != nil {
		return nil, err
	}
	reqPayload := SeldonMessagePayload{Msg: &sm}
	return &reqPayload, nil
}

type Option func(client *SeldonMessageRestClient)

func NewSeldonMessageRestClient(options ...Option) SeldonApiClient {
	client := SeldonMessageRestClient{
		&http.Client{},
		logf.Log.WithName("SeldonMessageRestClient"),
	}

	for i := range options {
		options[i](&client)
	}
	return &client
}

func (smc *SeldonMessageRestClient) PostHttp(url *url.URL, msg SeldonPayload) (*api.SeldonMessage, error) {
	smc.Log.Info("Calling HTTP", "URL", url)

	// Marshall message into JSON
	msgStr, err := smc.marshall(msg)
	if err != nil {
		return nil, err
	}

	// Call URL
	response, err := smc.httpClient.Post(url.String(), ContentTypeJSON, bytes.NewBufferString(msgStr))
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		smc.Log.Info("httpPost failed", "response code", response.StatusCode)
		return nil, errors.Errorf("Internal service call failed with to %s status code %d", url, response.StatusCode)
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

func (smc *SeldonMessageRestClient) marshall(payload SeldonPayload) (string, error) {
	ma := jsonpb.Marshaler{}
	var msgStr string
	var err error
	if sm, ok := payload.GetPayload().(*api.SeldonMessage); ok {
		msgStr, err = ma.MarshalToString(sm)
	} else if sm, ok := payload.GetPayload().(*api.SeldonMessageList); ok {
		msgStr, err = ma.MarshalToString(sm)
	} else {
		return "", errors.New("Unknown type passed")
	}
	return msgStr, err
}

func (smc *SeldonMessageRestClient) call(method string, host string, port int32, req SeldonPayload) (SeldonPayload, error) {
	url := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host, strconv.Itoa(int(port))),
		Path:   method,
	}
	sm, err := smc.PostHttp(&url, req)
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
	for i, sm := range msgs {
		sms[i] = sm.GetPayload().(*api.SeldonMessage)
	}
	sml := api.SeldonMessageList{SeldonMessages: sms}
	req := SeldonMessageListPayload{&sml}
	return smc.call("/aggregate", host, port, &req)
}

func (smc *SeldonMessageRestClient) TransformOutput(host string, port int32, req SeldonPayload) (SeldonPayload, error) {
	return smc.call("/transform-output", host, port, req)
}
