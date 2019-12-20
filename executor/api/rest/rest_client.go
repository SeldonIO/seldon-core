package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	"strings"
)

const (
	ContentTypeJSON = "application/json"
)

type JSONRestClient struct {
	httpClient *http.Client
	Log        logr.Logger
}

func (smc *JSONRestClient) CreateErrorPayload(err error) payload.SeldonPayload {
	respFailed := proto.SeldonMessage{Status: &proto.Status{Code: http.StatusInternalServerError, Info: err.Error()}}
	res := payload.SeldonMessagePayload{Msg: &respFailed}
	return &res
}

func (smc *JSONRestClient) Marshall(w io.Writer, msg payload.SeldonPayload) error {
	_, err := w.Write(msg.GetPayload().([]byte))
	return err
}

func (smc *JSONRestClient) Unmarshall(msg []byte) (payload.SeldonPayload, error) {
	reqPayload := payload.BytesPayload{Msg: msg}
	return &reqPayload, nil
}

type BytesRestClientOption func(client *JSONRestClient)

func NewJSONRestClient(options ...BytesRestClientOption) client.SeldonApiClient {
	client := JSONRestClient{
		&http.Client{},
		logf.Log.WithName("JSONRestClient"),
	}
	for i := range options {
		options[i](&client)
	}

	return &client
}

func (smc *JSONRestClient) PostHttp(url *url.URL, msg []byte) ([]byte, string, error) {
	smc.Log.Info("Calling HTTP", "URL", url)

	// Call URL
	response, err := smc.httpClient.Post(url.String(), ContentTypeJSON, bytes.NewBuffer(msg))
	if err != nil {
		return nil, "", err
	}

	if response.StatusCode != http.StatusOK {
		smc.Log.Info("httpPost failed", "response code", response.StatusCode)
		return nil, "", errors.Errorf("Internal service call failed with to %s status code %d", url, response.StatusCode)
	}

	//Read response
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, "", err
	}
	defer response.Body.Close()

	contentType := response.Header.Get("Content-Type")

	return b, contentType, nil
}

func (smc *JSONRestClient) call(method string, host string, port int32, req payload.SeldonPayload) (payload.SeldonPayload, error) {
	url := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host, strconv.Itoa(int(port))),
		Path:   method,
	}
	sm, contentType, err := smc.PostHttp(&url, req.GetPayload().([]byte))
	if err != nil {
		return nil, err
	}
	res := payload.BytesPayload{Msg: sm, ContentType: contentType}
	return &res, nil
}

func (smc *JSONRestClient) Predict(host string, port int32, req payload.SeldonPayload) (payload.SeldonPayload, error) {
	return smc.call("/predict", host, port, req)
}

func (smc *JSONRestClient) TransformInput(host string, port int32, req payload.SeldonPayload) (payload.SeldonPayload, error) {
	return smc.call("/transform-input", host, port, req)
}

// Try to extract from SeldonMessage otherwise fall back to extract from Json Array
func (smc *JSONRestClient) Route(host string, port int32, req payload.SeldonPayload) (int, error) {
	sp, err := smc.call("/route", host, port, req)
	if err != nil {
		return 0, err
	} else {
		var routes []int
		msg := sp.GetPayload().([]byte)

		var sm proto.SeldonMessage
		value := string(msg)
		err := jsonpb.UnmarshalString(value, &sm)
		if err == nil {
			//Remove in future
			routes = api.ExtractRouteFromSeldonMessage(&sm)
		} else {
			routes, err = ExtractRouteAsJsonArray(msg)
			if err != nil {
				return 0, err
			}
		}

		//Only returning first route. API could be extended to allow multiple routes
		return routes[0], nil
	}
}

func isJSON(data []byte) bool {
	var js json.RawMessage
	return json.Unmarshal(data, &js) == nil
}

func (smc *JSONRestClient) Combine(host string, port int32, msgs []payload.SeldonPayload) (payload.SeldonPayload, error) {
	// Extract into string array checking the data is JSON
	strData := make([]string, len(msgs))
	for i, sm := range msgs {
		if !isJSON(sm.GetPayload().([]byte)) {
			return nil, fmt.Errorf("Data is not JSON")
		} else {
			strData[i] = string(sm.GetPayload().([]byte))
		}
	}
	// Create JSON list of messages
	joined := strings.Join(strData, ",")
	jStr := "[" + joined + "]"
	req := payload.BytesPayload{Msg: []byte(jStr)}
	return smc.call("/aggregate", host, port, &req)
}

func (smc *JSONRestClient) TransformOutput(host string, port int32, req payload.SeldonPayload) (payload.SeldonPayload, error) {
	return smc.call("/transform-output", host, port, req)
}
