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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/seldonio/seldon-core/executor/api"
	"github.com/seldonio/seldon-core/executor/api/client"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"github.com/seldonio/seldon-core/executor/api/metric"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/api/util"
	"github.com/seldonio/seldon-core/executor/k8s"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strconv"
	"strings"
	"time"
)

const (
	ContentTypeJSON = "application/json"
)

type JSONRestClient struct {
	httpClient     *http.Client
	Log            logr.Logger
	Protocol       string
	DeploymentName string
	predictor      *v1.PredictorSpec
	metrics        *metric.ClientMetrics
}

func (smc *JSONRestClient) CreateErrorPayload(err error) payload.SeldonPayload {
	respFailed := proto.SeldonMessage{Status: &proto.Status{Code: http.StatusInternalServerError, Info: err.Error()}}
	m := jsonpb.Marshaler{}
	jStr, _ := m.MarshalToString(&respFailed)
	res := payload.BytesPayload{Msg: []byte(jStr)}
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

func getRestTimeoutFromAnnotations(annotations map[string]string) (int, error) {
	val := annotations[k8s.ANNOTATION_REST_TIMEOUT]
	if val != "" {
		converted, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			return 0, err
		} else {
			return int(converted), nil
		}
	} else {
		return 0, nil
	}
}

func NewJSONRestClient(protocol string, deploymentName string, predictor *v1.PredictorSpec, annotations map[string]string, options ...BytesRestClientOption) (client.SeldonApiClient, error) {

	httpClient := http.DefaultClient
	if annotations != nil {
		restTimeout, err := getRestTimeoutFromAnnotations(annotations)
		if err != nil {
			return nil, err
		}
		if restTimeout > 0 {
			httpClient = &http.Client{
				Timeout: time.Duration(restTimeout) * time.Millisecond,
			}
		}
	}

	client := JSONRestClient{
		httpClient,
		logf.Log.WithName("JSONRestClient"),
		protocol,
		deploymentName,
		predictor,
		metric.NewClientMetrics(predictor, deploymentName, ""),
	}
	for i := range options {
		options[i](&client)
	}

	return &client, nil
}

func (smc *JSONRestClient) getMetricsRoundTripper(modelName string, service string) http.RoundTripper {
	container := v1.GetContainerForPredictiveUnit(smc.predictor, modelName)
	imageName := ""
	imageVersion := ""
	if container != nil {
		imageParts := strings.Split(container.Image, ":")
		imageName = imageParts[0]
		if len(imageParts) == 2 {
			imageVersion = imageParts[1]
		}
	}
	return promhttp.InstrumentRoundTripperDuration(smc.metrics.ClientHandledHistogram.MustCurryWith(prometheus.Labels{
		metric.DeploymentNameMetric:   smc.DeploymentName,
		metric.PredictorNameMetric:    smc.predictor.Name,
		metric.PredictorVersionMetric: smc.predictor.Annotations["version"],
		metric.ServiceMetric:          service,
		metric.ModelNameMetric:        modelName,
		metric.ModelImageMetric:       imageName,
		metric.ModelVersionMetric:     imageVersion,
	}), http.DefaultTransport)
}

func (smc *JSONRestClient) addHeaders(req *http.Request, m map[string][]string) {
	for k, vv := range m {
		for _, v := range vv {
			req.Header.Set(k, v)
		}
	}
}

func (smc *JSONRestClient) doHttp(ctx context.Context, modelName string, method string, url *url.URL, msg []byte, meta map[string][]string) ([]byte, string, error) {
	smc.Log.Info("Calling HTTP", "URL", url)

	var req *http.Request
	var err error
	if msg != nil {
		smc.Log.Info("Building message")
		req, err = http.NewRequest("POST", url.String(), bytes.NewBuffer(msg))
		if err != nil {
			return nil, "", err
		}
		req.Header.Set("Content-Type", ContentTypeJSON)
	} else {
		req, err = http.NewRequest("GET", url.String(), nil)
		if err != nil {
			return nil, "", err
		}
	}

	// Add metadata passed in
	smc.addHeaders(req, meta)

	if opentracing.IsGlobalTracerRegistered() {
		tracer := opentracing.GlobalTracer()

		parentSpan := opentracing.SpanFromContext(ctx)
		clientSpan := opentracing.StartSpan(
			method,
			opentracing.ChildOf(parentSpan.Context()))
		defer clientSpan.Finish()
		tracer.Inject(clientSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	}

	client := smc.httpClient
	client.Transport = smc.getMetricsRoundTripper(modelName, method)

	response, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}

	//Read response
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, "", err
	}
	defer response.Body.Close()

	contentType := response.Header.Get("Content-Type")

	if response.StatusCode != http.StatusOK {
		smc.Log.Info("httpPost failed", "response code", response.StatusCode)
		err = &httpStatusError{StatusCode: response.StatusCode, Url: url}
	}

	return b, contentType, err
}

func (smc *JSONRestClient) modifyMethod(method string, modelName string) string {
	switch smc.Protocol {
	case api.ProtocolTensorflow:
		switch method {
		case client.SeldonPredictPath, client.SeldonTransformInputPath, client.SeldonTransformOutputPath:
			return "/v1/models/" + modelName + ":predict"
		case client.SeldonCombinePath:
			return "/v1/models/" + modelName + ":aggregate"
		case client.SeldonRoutePath:
			return "/v1/models/" + modelName + ":route"
		case client.SeldonFeedbackPath:
			return "/v1/models/" + modelName + ":feedback"
		case client.SeldonStatusPath:
			return "/v1/models/" + modelName
		case client.SeldonMetadataPath:
			return "/v1/models/" + modelName + "/metadata"
		}
	case api.ProtocolKfserving:
		switch method {
		case client.SeldonPredictPath, client.SeldonTransformInputPath, client.SeldonTransformOutputPath:
			return "/v2/models/" + modelName + "/infer"
		case client.SeldonCombinePath:
			return "/v2/models/" + modelName + "/aggregate"
		case client.SeldonRoutePath:
			return "/v2/models/" + modelName + "/route"
		case client.SeldonFeedbackPath:
			return "/v2/models/" + modelName + "/feedback"
		case client.SeldonStatusPath:
			return "/v2/models/" + modelName + "/ready"
		case client.SeldonMetadataPath:
			return "/v2/models/" + modelName
		}
	default:
		return method
	}
	return method
}

func (smc *JSONRestClient) call(ctx context.Context, modelName string, method string, host string, port int32, req payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	url := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host, strconv.Itoa(int(port))),
		Path:   method,
	}
	var bytes []byte
	if req != nil {
		bytes = req.GetPayload().([]byte)
	}
	sm, contentType, err := smc.doHttp(ctx, modelName, method, &url, bytes, meta)
	res := payload.BytesPayload{Msg: sm, ContentType: contentType}
	return &res, err
}

func (smc *JSONRestClient) Status(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	return smc.call(ctx, modelName, smc.modifyMethod(client.SeldonStatusPath, modelName), host, port, msg, meta)
}

func (smc *JSONRestClient) Metadata(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	return smc.call(ctx, modelName, smc.modifyMethod(client.SeldonMetadataPath, modelName), host, port, msg, meta)
}

func (smc *JSONRestClient) Chain(ctx context.Context, modelName string, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	switch smc.Protocol {
	case api.ProtocolSeldon: // Seldon Messages can always be chained together
		return msg, nil
	case api.ProtocolTensorflow: // Attempt to chain tensorflow payload
		return ChainTensorflow(msg)
	case api.ProtocolKfserving:
		return ChainKFserving(msg)
	}
	return nil, errors.Errorf("Unknown protocol %s", smc.Protocol)
}

func (smc *JSONRestClient) Predict(ctx context.Context, modelName string, host string, port int32, req payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	return smc.call(ctx, modelName, smc.modifyMethod(client.SeldonPredictPath, modelName), host, port, req, meta)
}

func (smc *JSONRestClient) TransformInput(ctx context.Context, modelName string, host string, port int32, req payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	return smc.call(ctx, modelName, smc.modifyMethod(client.SeldonTransformInputPath, modelName), host, port, req, meta)
}

// Try to extract from SeldonMessage otherwise fall back to extract from Json Array
func (smc *JSONRestClient) Route(ctx context.Context, modelName string, host string, port int32, req payload.SeldonPayload, meta map[string][]string) (int, error) {
	sp, err := smc.call(ctx, modelName, smc.modifyMethod(client.SeldonRoutePath, modelName), host, port, req, meta)
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
			routes = util.ExtractRouteFromSeldonMessage(&sm)
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

func (smc *JSONRestClient) Combine(ctx context.Context, modelName string, host string, port int32, msgs []payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
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
	return smc.call(ctx, modelName, smc.modifyMethod(client.SeldonCombinePath, modelName), host, port, &req, meta)
}

func (smc *JSONRestClient) TransformOutput(ctx context.Context, modelName string, host string, port int32, req payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	return smc.call(ctx, modelName, smc.modifyMethod(client.SeldonTransformOutputPath, modelName), host, port, req, meta)
}

func (smc *JSONRestClient) Feedback(ctx context.Context, modelName string, host string, port int32, req payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	return smc.call(ctx, modelName, smc.modifyMethod(client.SeldonFeedbackPath, modelName), host, port, req, meta)
}
