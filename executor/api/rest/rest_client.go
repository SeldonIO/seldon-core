package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/golang/protobuf/jsonpb"
	"github.com/opentracing/opentracing-go"
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
	reqPayload := payload.BytesPayload{Msg: msg, ContentType: ContentTypeJSON}
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

func (smc *JSONRestClient) PostHttp(ctx context.Context, method string, url *url.URL, msg []byte) ([]byte, string, error) {
	smc.Log.Info("Calling HTTP", "URL", url)

	req, err := http.NewRequest("POST", url.String(), bytes.NewBuffer(msg))
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Content-Type", ContentTypeJSON)

	if opentracing.IsGlobalTracerRegistered() {
		tracer := opentracing.GlobalTracer()

		parentSpan := opentracing.SpanFromContext(ctx)
		clientSpan := opentracing.StartSpan(
			method,
			opentracing.ChildOf(parentSpan.Context()))
		defer clientSpan.Finish()
		tracer.Inject(clientSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	}

	response, err := smc.httpClient.Do(req)
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

func (smc *JSONRestClient) call(ctx context.Context, method string, host string, port int32, req payload.SeldonPayload) (payload.SeldonPayload, error) {
	url := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host, strconv.Itoa(int(port))),
		Path:   method,
	}
	sm, contentType, err := smc.PostHttp(ctx, method, &url, req.GetPayload().([]byte))
	if err != nil {
		return nil, err
	}
	res := payload.BytesPayload{Msg: sm, ContentType: contentType}
	return &res, nil
}

func (smc *JSONRestClient) Predict(ctx context.Context, host string, port int32, req payload.SeldonPayload) (payload.SeldonPayload, error) {
	return smc.call(ctx, "/predict", host, port, req)
}

func (smc *JSONRestClient) TransformInput(ctx context.Context, host string, port int32, req payload.SeldonPayload) (payload.SeldonPayload, error) {
	return smc.call(ctx, "/transform-input", host, port, req)
}

// Try to extract from SeldonMessage otherwise fall back to extract from Json Array
func (smc *JSONRestClient) Route(ctx context.Context, host string, port int32, req payload.SeldonPayload) (int, error) {
	sp, err := smc.call(ctx, "/route", host, port, req)
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

func (smc *JSONRestClient) Combine(ctx context.Context, host string, port int32, msgs []payload.SeldonPayload) (payload.SeldonPayload, error) {
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
	return smc.call(ctx, "/aggregate", host, port, &req)
}

func (smc *JSONRestClient) TransformOutput(ctx context.Context, host string, port int32, req payload.SeldonPayload) (payload.SeldonPayload, error) {
	return smc.call(ctx, "/transform-output", host, port, req)
}
