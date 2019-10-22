package rest

import (
	"bytes"
	"github.com/go-logr/logr"
	"github.com/golang/protobuf/jsonpb"
	"github.com/pkg/errors"
	"github.com/seldonio/seldon-core/executor/api/client"
	api "github.com/seldonio/seldon-core/executor/api/grpc"
	"github.com/seldonio/seldon-core/executor/api/grpc/proto"
	"github.com/seldonio/seldon-core/executor/api/payload"
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

func (smc *SeldonMessageRestClient) CreateErrorPayload(err error) (payload.SeldonPayload, error) {
	respFailed := proto.SeldonMessage{Status: &proto.Status{Code: http.StatusInternalServerError, Info: err.Error()}}
	res := payload.SeldonMessagePayload{Msg: &respFailed}
	return &res, nil
}

func (smc *SeldonMessageRestClient) Marshall(w io.Writer, msg payload.SeldonPayload) error {
	ma := jsonpb.Marshaler{}
	return ma.Marshal(w, msg.GetPayload().(*proto.SeldonMessage))
}

func (smc *SeldonMessageRestClient) Unmarshall(msg []byte) (payload.SeldonPayload, error) {
	var sm proto.SeldonMessage
	value := string(msg)
	if err := jsonpb.UnmarshalString(value, &sm); err != nil {
		return nil, err
	}
	reqPayload := payload.SeldonMessagePayload{Msg: &sm}
	return &reqPayload, nil
}

type Option func(client *SeldonMessageRestClient)

func NewSeldonMessageRestClient(options ...Option) client.SeldonApiClient {
	client := SeldonMessageRestClient{
		&http.Client{},
		logf.Log.WithName("SeldonMessageRestClient"),
	}

	for i := range options {
		options[i](&client)
	}
	return &client
}

func (smc *SeldonMessageRestClient) PostHttp(url *url.URL, msg payload.SeldonPayload) (*proto.SeldonMessage, error) {
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
	var sm proto.SeldonMessage
	value := string(b)
	if err := jsonpb.UnmarshalString(value, &sm); err != nil {
		return nil, err
	}
	return &sm, nil
}

func (smc *SeldonMessageRestClient) marshall(payload payload.SeldonPayload) (string, error) {
	ma := jsonpb.Marshaler{}
	var msgStr string
	var err error
	if sm, ok := payload.GetPayload().(*proto.SeldonMessage); ok {
		msgStr, err = ma.MarshalToString(sm)
	} else if sm, ok := payload.GetPayload().(*proto.SeldonMessageList); ok {
		msgStr, err = ma.MarshalToString(sm)
	} else {
		return "", errors.New("Unknown type passed")
	}
	return msgStr, err
}

func (smc *SeldonMessageRestClient) call(method string, host string, port int32, req payload.SeldonPayload) (payload.SeldonPayload, error) {
	url := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host, strconv.Itoa(int(port))),
		Path:   method,
	}
	sm, err := smc.PostHttp(&url, req)
	if err != nil {
		return nil, err
	}
	res := payload.SeldonMessagePayload{Msg: sm}
	return &res, nil
}

func (smc *SeldonMessageRestClient) Predict(host string, port int32, req payload.SeldonPayload) (payload.SeldonPayload, error) {
	return smc.call("/predict", host, port, req)
}

func (smc *SeldonMessageRestClient) TransformInput(host string, port int32, req payload.SeldonPayload) (payload.SeldonPayload, error) {
	return smc.call("/transform-input", host, port, req)
}

func (smc *SeldonMessageRestClient) Route(host string, port int32, req payload.SeldonPayload) (int, error) {
	sp, err := smc.call("/route", host, port, req)
	if err != nil {
		return 0, err
	} else {
		routes := api.ExtractRoute(sp.GetPayload().(*proto.SeldonMessage))
		//Only returning first route. API could be extended to allow multiple routes
		return routes[0], nil
	}
}

func (smc *SeldonMessageRestClient) Combine(host string, port int32, msgs []payload.SeldonPayload) (payload.SeldonPayload, error) {
	sms := make([]*proto.SeldonMessage, len(msgs))
	for i, sm := range msgs {
		sms[i] = sm.GetPayload().(*proto.SeldonMessage)
	}
	sml := proto.SeldonMessageList{SeldonMessages: sms}
	req := payload.SeldonMessageListPayload{Msg: &sml}
	return smc.call("/aggregate", host, port, &req)
}

func (smc *SeldonMessageRestClient) TransformOutput(host string, port int32, req payload.SeldonPayload) (payload.SeldonPayload, error) {
	return smc.call("/transform-output", host, port, req)
}
