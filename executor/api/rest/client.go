package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	http2 "github.com/cloudevents/sdk-go/pkg/bindings/http"
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
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	ContentTypeJSON = "application/json"
)

var headersIgnore = map[string]bool{http2.ContentType: true}

type JSONRestClient struct {
	httpClient     *http.Client
	Log            logr.Logger
	Protocol       string
	DeploymentName string
	predictor      *v1.PredictorSpec
	metrics        *metric.ClientMetrics
}

func (smc *JSONRestClient) IsGrpc() bool {
	return false
}

func (smc *JSONRestClient) CreateErrorPayload(err error) payload.SeldonPayload {
	respFailed := proto.SeldonMessage{
		Status: &proto.Status{
			Code:   http.StatusInternalServerError,
			Info:   err.Error(),
			Status: proto.Status_FAILURE,
		},
	}
	m := jsonpb.Marshaler{}
	jStr, _ := m.MarshalToString(&respFailed)
	res := payload.BytesPayload{Msg: []byte(jStr)}
	return &res
}

func (smc *JSONRestClient) Marshall(w io.Writer, msg payload.SeldonPayload) error {
	payload, ok := msg.GetPayload().([]byte)
	if !ok {
		return invalidPayload("couldn't convert to []byte")
	}

	var err error
	// When "Content-Encoding" header is not empty it means that payload is compressed.
	// We do not want to modify it in this situation. This change was added during update of Triton
	// image from 20.08 to 21.08 as new version allowed for gzip-encoded payloads.
	// Related PR: https://github.com/SeldonIO/seldon-core/pull/3589
	// More on this header: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Encoding
	if msg.GetContentEncoding() != "" {
		_, err = w.Write(payload)
	} else {
		var escaped bytes.Buffer

		json.HTMLEscape(&escaped, payload)
		_, err = escaped.WriteTo(w)
	}

	return err
}

func (smc *JSONRestClient) Unmarshall(msg []byte, contentType string) (payload.SeldonPayload, error) {
	reqPayload := payload.BytesPayload{Msg: msg, ContentType: contentType}
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
	roundTripper := promhttp.InstrumentRoundTripperDuration(smc.metrics.ClientHandledHistogram.MustCurryWith(prometheus.Labels{
		metric.DeploymentNameMetric:   smc.DeploymentName,
		metric.PredictorNameMetric:    smc.predictor.Name,
		metric.PredictorVersionMetric: smc.predictor.Annotations["version"],
		metric.ServiceMetric:          service,
		metric.ModelNameMetric:        modelName,
		metric.ModelImageMetric:       imageName,
		metric.ModelVersionMetric:     imageVersion,
	}), http.DefaultTransport)

	return promhttp.InstrumentRoundTripperDuration(smc.metrics.ClientHandledSummary.MustCurryWith(prometheus.Labels{
		metric.DeploymentNameMetric:   smc.DeploymentName,
		metric.PredictorNameMetric:    smc.predictor.Name,
		metric.PredictorVersionMetric: smc.predictor.Annotations["version"],
		metric.ServiceMetric:          service,
		metric.ModelNameMetric:        modelName,
		metric.ModelImageMetric:       imageName,
		metric.ModelVersionMetric:     imageVersion,
	}), roundTripper)
}

func (smc *JSONRestClient) addHeaders(req *http.Request, m map[string][]string) {
	for k, vv := range m {
		if _, ok := headersIgnore[k]; !ok {
			for _, v := range vv {
				req.Header.Add(k, v)
			}
		}
	}
}

func (smc *JSONRestClient) doHttp(ctx context.Context, modelName string, method string, url *url.URL, msg []byte, meta map[string][]string, contentType string, contentEncoding string) ([]byte, string, string, error) {
	smc.Log.V(1).Info("Calling HTTP", "URL", url)

	var req *http.Request
	var err error
	if msg != nil {
		req, err = http.NewRequest("POST", url.String(), bytes.NewBuffer(msg))
		if err != nil {
			return nil, "", "", err
		}
		req.Header.Set(http2.ContentType, contentType)
		if contentEncoding != "" {
			req.Header.Set("Content-Encoding", contentEncoding)
		}
	} else {
		req, err = http.NewRequest("GET", url.String(), nil)
		if err != nil {
			return nil, "", "", err
		}
	}

	// Add metadata passed in
	smc.addHeaders(req, meta)

	if opentracing.IsGlobalTracerRegistered() {
		tracer := opentracing.GlobalTracer()

		startSpanOptions := make([]opentracing.StartSpanOption, 0)
		parentSpan := opentracing.SpanFromContext(ctx)
		if parentSpan != nil {
			startSpanOptions = append(startSpanOptions, opentracing.ChildOf(parentSpan.Context()))
		}
		clientSpan := opentracing.StartSpan(
			method,
			startSpanOptions...)
		defer clientSpan.Finish()
		tracer.Inject(clientSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	}

	client := smc.httpClient
	client.Transport = smc.getMetricsRoundTripper(modelName, method)

	response, err := client.Do(req)
	if err != nil {
		return nil, "", "", err
	}

	//Read response
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, "", "", err
	}
	defer response.Body.Close()

	contentTypeResponse := response.Header.Get(http2.ContentType)
	contentEncodingResponse := response.Header.Get("Content-Encoding")

	if response.StatusCode != http.StatusOK {
		smc.Log.Info("httpPost failed", "response code", response.StatusCode)
		err = &httpStatusError{StatusCode: response.StatusCode, Url: url}
	}

	return b, contentTypeResponse, contentEncodingResponse, err
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
	case api.ProtocolV2:
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
	var contentType = ContentTypeJSON
	var contentEncoding = ""
	if req != nil {
		bytes = req.GetPayload().([]byte)
		contentType = req.GetContentType()
		contentEncoding = req.GetContentEncoding()
	}

	sm, contentType, contentEncoding, err := smc.doHttp(ctx, modelName, method, &url, bytes, meta, contentType, contentEncoding)

	// Check if a httpStatusError was returned.
	if err != nil {
		if _, ok := err.(*httpStatusError); !ok {
			return smc.CreateErrorPayload(err), err
		}
	}

	res := payload.BytesPayload{Msg: sm, ContentType: contentType, ContentEncoding: contentEncoding}
	return &res, err
}

func (smc *JSONRestClient) Status(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	return smc.call(ctx, modelName, smc.modifyMethod(client.SeldonStatusPath, modelName), host, port, msg, meta)
}

// Return model's metadata as payload.SeldonPaylaod (to expose as received on corresponding executor endpoint)
func (smc *JSONRestClient) Metadata(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	return smc.call(ctx, modelName, smc.modifyMethod(client.SeldonMetadataPath, modelName), host, port, msg, meta)
}

// Return model's metadata decoded to payload.ModelMetadata (to build GraphMetadata)
func (smc *JSONRestClient) ModelMetadata(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.ModelMetadata, error) {
	resPayload, err := smc.Metadata(ctx, modelName, host, port, msg, meta)
	if err != nil {
		return payload.ModelMetadata{}, err
	}

	resString, err := resPayload.GetBytes()
	if err != nil {
		return payload.ModelMetadata{}, err
	}
	var modelMetadata payload.ModelMetadata
	err = json.Unmarshal(resString, &modelMetadata)
	if err != nil {
		return payload.ModelMetadata{}, err
	}
	return modelMetadata, nil
}

func (smc *JSONRestClient) Chain(ctx context.Context, modelName string, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	switch smc.Protocol {
	case api.ProtocolSeldon: // Seldon Messages can always be chained together
		return msg, nil
	case api.ProtocolTensorflow: // Attempt to chain tensorflow payload
		return ChainTensorflow(msg)
	case api.ProtocolV2:
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
		return util.ExtractRouteFromSeldonJson(sp)
	}
}

func (smc *JSONRestClient) Combine(ctx context.Context, modelName string, host string, port int32, msgs []payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	req, err := CombineSeldonMessagesToJson(msgs)
	if err != nil {
		return nil, err
	}
	return smc.call(ctx, modelName, smc.modifyMethod(client.SeldonCombinePath, modelName), host, port, req, meta)
}

func (smc *JSONRestClient) TransformOutput(ctx context.Context, modelName string, host string, port int32, req payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	return smc.call(ctx, modelName, smc.modifyMethod(client.SeldonTransformOutputPath, modelName), host, port, req, meta)
}

func (smc *JSONRestClient) Feedback(ctx context.Context, modelName string, host string, port int32, req payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	// Currently feedback is enabled across all protocols but client only works on seldon protocol
	if smc.Protocol != api.ProtocolSeldon {
		return req, nil
	}
	return smc.call(ctx, modelName, smc.modifyMethod(client.SeldonFeedbackPath, modelName), host, port, req, meta)
}
