package rest

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/golang/protobuf/jsonpb"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/seldonio/seldon-core/executor/api"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"github.com/seldonio/seldon-core/executor/api/metric"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/k8s"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	v12 "k8s.io/api/core/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	okPredictResponse = `{
		"data": {
           "names" : [ "a", "b" ],
           "ndarray" : [[0.9,0.1]]
       }
	}`
	okRouteResponse = `{
		"data": {
           "ndarray" : [1]
       }
	}`
	okStatusResponse = `{
        "status": "ok"
	}`
	okMetadataResponse = `{
		"name": "mymodel",
		"platform": "seldon-platform"
	}`
	errorPredictResponse = `{
       "status":"failed"
    }`
)

func testingHTTPClient(g *GomegaWithT, handler http.Handler) (string, int, *http.Client, func()) {
	s := httptest.NewServer(handler)

	cli := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, network, _ string) (net.Conn, error) {
				return net.Dial(network, s.Listener.Addr().String())
			},
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	url, err := url.Parse(s.URL)
	g.Expect(err).Should(BeNil())
	port, err := strconv.Atoi(url.Port())
	g.Expect(err).Should(BeNil())

	return url.Hostname(), port, cli, s.Close
}

func SetHTTPClient(httpClient *http.Client) BytesRestClientOption {
	return func(cli *JSONRestClient) {
		cli.httpClient = httpClient
	}
}

func createPayload(g *GomegaWithT) payload.SeldonPayload {
	var data = ` {"data":{"ndarray":[1.1,2.0]}}`
	return &payload.BytesPayload{Msg: []byte(data)}
}

func createTestContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, payload.SeldonPUIDHeader, "1")
	return ctx
}

func TestSimpleMethods(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(okPredictResponse))
	})
	host, port, httpClient, teardown := testingHTTPClient(g, h)
	defer teardown()
	predictor := v1.PredictorSpec{
		Name:        "test",
		Annotations: map[string]string{},
	}
	seldonRestClient, err := NewJSONRestClient(api.ProtocolSeldon, "test", &predictor, nil, SetHTTPClient(httpClient))
	g.Expect(err).To(BeNil())

	methods := []func(context.Context, string, string, int32, payload.SeldonPayload, map[string][]string) (payload.SeldonPayload, error){seldonRestClient.Predict, seldonRestClient.TransformInput, seldonRestClient.TransformOutput, seldonRestClient.Feedback}
	for _, method := range methods {
		resPayload, err := method(createTestContext(), "model", host, int32(port), createPayload(g), map[string][]string{})
		g.Expect(err).Should(BeNil())

		data := resPayload.GetPayload().([]byte)
		var smRes proto.SeldonMessage
		err = jsonpb.UnmarshalString(string(data), &smRes)
		g.Expect(err).Should(BeNil())
		g.Expect(smRes.GetData().GetNdarray().Values[0].GetListValue().Values[0].GetNumberValue()).Should(Equal(0.9))
		g.Expect(smRes.GetData().GetNdarray().Values[0].GetListValue().Values[1].GetNumberValue()).Should(Equal(0.1))
	}
}

func TestRouter(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(okRouteResponse))
	})
	host, port, httpClient, teardown := testingHTTPClient(g, h)
	defer teardown()
	predictor := v1.PredictorSpec{
		Name:        "test",
		Annotations: map[string]string{},
	}
	seldonRestClient, err := NewJSONRestClient(api.ProtocolSeldon, "test", &predictor, nil, SetHTTPClient(httpClient))
	g.Expect(err).To(BeNil())

	route, err := seldonRestClient.Route(createTestContext(), "model", host, int32(port), createPayload(g), map[string][]string{})
	g.Expect(err).Should(BeNil())

	g.Expect(route).Should(Equal(1))
}

func TestStatus(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(okStatusResponse))
	})
	host, port, httpClient, teardown := testingHTTPClient(g, h)
	defer teardown()
	predictor := v1.PredictorSpec{
		Name:        "test",
		Annotations: map[string]string{},
	}
	seldonRestClient, err := NewJSONRestClient(api.ProtocolSeldon, "test", &predictor, nil, SetHTTPClient(httpClient))
	g.Expect(err).To(BeNil())

	status, err := seldonRestClient.Status(createTestContext(), "model", host, int32(port), nil, map[string][]string{})
	g.Expect(err).Should(BeNil())
	data := string(status.GetPayload().([]byte))
	g.Expect(data).To(Equal(okStatusResponse))
}

func TestMetadata(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(okMetadataResponse))
	})
	host, port, httpClient, teardown := testingHTTPClient(g, h)
	defer teardown()
	predictor := v1.PredictorSpec{
		Name:        "test",
		Annotations: map[string]string{},
	}
	seldonRestClient, err := NewJSONRestClient(api.ProtocolSeldon, "test", &predictor, nil, SetHTTPClient(httpClient))
	g.Expect(err).To(BeNil())

	status, err := seldonRestClient.Metadata(createTestContext(), "model", host, int32(port), nil, map[string][]string{})
	g.Expect(err).Should(BeNil())
	data := string(status.GetPayload().([]byte))
	g.Expect(data).To(Equal(okMetadataResponse))
}

func createCombinerPayload(g *GomegaWithT) []payload.SeldonPayload {
	var data = ` {"data":{"ndarray":[1.1,2.0]}}`
	smp := []payload.SeldonPayload{&payload.BytesPayload{Msg: []byte(data)}, &payload.BytesPayload{Msg: []byte(data)}}
	return smp
}

func TestCombiner(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(okPredictResponse))
	})
	host, port, httpClient, teardown := testingHTTPClient(g, h)
	defer teardown()
	predictor := v1.PredictorSpec{
		Name:        "test",
		Annotations: map[string]string{},
	}
	seldonRestClient, err := NewJSONRestClient(api.ProtocolSeldon, "test", &predictor, nil, SetHTTPClient(httpClient))
	g.Expect(err).To(BeNil())

	resPayload, err := seldonRestClient.Combine(createTestContext(), "model", host, int32(port), createCombinerPayload(g), map[string][]string{})
	g.Expect(err).Should(BeNil())

	data := resPayload.GetPayload().([]byte)
	var smRes proto.SeldonMessage
	err = jsonpb.UnmarshalString(string(data), &smRes)
	g.Expect(err).Should(BeNil())
	g.Expect(smRes.GetData().GetNdarray().Values[0].GetListValue().Values[0].GetNumberValue()).Should(Equal(0.9))
	g.Expect(smRes.GetData().GetNdarray().Values[0].GetListValue().Values[1].GetNumberValue()).Should(Equal(0.1))
}

func TestClientMetrics(t *testing.T) {
	t.Logf("Started")
	metric.RecreateServerHistogram = true
	metric.RecreateClientHistogram = true
	g := NewGomegaWithT(t)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(okPredictResponse))
	})
	host, port, httpClient, teardown := testingHTTPClient(g, h)
	defer teardown()
	predictor := v1.PredictorSpec{
		Name:        "test",
		Annotations: map[string]string{},
		ComponentSpecs: []*v1.SeldonPodSpec{
			&v1.SeldonPodSpec{
				Metadata: v1meta.ObjectMeta{},
				Spec: v12.PodSpec{
					Containers: []v12.Container{
						v12.Container{
							Name:  "model",
							Image: "foo:0.1",
						},
					},
				},
				HpaSpec: nil,
			},
		},
	}
	seldonRestClient, err := NewJSONRestClient(api.ProtocolSeldon, "test", &predictor, nil, SetHTTPClient(httpClient))
	g.Expect(err).To(BeNil())

	methods := []func(context.Context, string, string, int32, payload.SeldonPayload, map[string][]string) (payload.SeldonPayload, error){seldonRestClient.Predict, seldonRestClient.TransformInput, seldonRestClient.TransformOutput}
	for _, method := range methods {
		resPayload, err := method(createTestContext(), "model", host, int32(port), createPayload(g), map[string][]string{})
		g.Expect(err).Should(BeNil())

		data := resPayload.GetPayload().([]byte)
		var smRes proto.SeldonMessage
		err = jsonpb.UnmarshalString(string(data), &smRes)
		g.Expect(err).Should(BeNil())
		g.Expect(smRes.GetData().GetNdarray().Values[0].GetListValue().Values[0].GetNumberValue()).Should(Equal(0.9))
		g.Expect(smRes.GetData().GetNdarray().Values[0].GetListValue().Values[1].GetNumberValue()).Should(Equal(0.1))

		mfs, err := prometheus.DefaultGatherer.Gather()
		g.Expect(err).Should(BeNil())
		found := false
		foundImage := false
		foundImageVersion := false
		for _, mf := range mfs {
			if mf.Name != nil && *mf.Name == metric.ClientRequestsMetricName {
				for _, label := range mf.Metric[0].Label {
					if *label.Name == metric.ModelImageMetric && *label.Value == "foo" {
						foundImage = true
					}
					if *label.Name == metric.ModelVersionMetric && *label.Value == "0.1" {
						foundImageVersion = true
					}
				}
				found = true
			}
		}
		g.Expect(found).Should(Equal(true))
		g.Expect(foundImage).Should(Equal(true))
		g.Expect(foundImageVersion).Should(Equal(true))
	}

}

func TestErrorResponse(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errorPredictResponse))
	})
	host, port, _, teardown := testingHTTPClient(g, h)

	defer teardown()
	predictor := v1.PredictorSpec{
		Name:        "test",
		Annotations: map[string]string{},
	}

	seldonRestClient, err := NewJSONRestClient(api.ProtocolSeldon, "test", &predictor, nil)
	g.Expect(err).To(BeNil())

	methods := []func(context.Context, string, string, int32, payload.SeldonPayload, map[string][]string) (payload.SeldonPayload, error){seldonRestClient.Predict}
	for _, method := range methods {
		resPayload, err := method(createTestContext(), "model", host, int32(port), createPayload(g), map[string][]string{})
		g.Expect(err).ToNot(BeNil())

		data := resPayload.GetPayload().([]byte)
		var objmap map[string]interface{}
		err = json.Unmarshal(data, &objmap)
		g.Expect(err).To(BeNil())
		g.Expect(string(data)).To(Equal(errorPredictResponse))
	}
}

func TestTimeout(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		w.Write([]byte(okStatusResponse))
	})
	host, port, _, teardown := testingHTTPClient(g, h)

	defer teardown()
	predictor := v1.PredictorSpec{
		Name:        "test",
		Annotations: map[string]string{},
	}
	annotations := map[string]string{k8s.ANNOTATION_REST_TIMEOUT: "1"}
	seldonRestClient, err := NewJSONRestClient(api.ProtocolSeldon, "test", &predictor, annotations)
	g.Expect(err).To(BeNil())

	_, err = seldonRestClient.Status(createTestContext(), "model", host, int32(port), nil, map[string][]string{})
	g.Expect(err).ToNot(BeNil())
}

func TestMarshall(t *testing.T) {
	g := NewGomegaWithT(t)

	tests := []struct {
		response string
		expected string
	}{
		{
			response: okPredictResponse,
			expected: okPredictResponse,
		},
		{
			response: `"<div class=\"div-class\"></div>"`,
			expected: `"\u003cdiv class=\"div-class\"\u003e\u003c/div\u003e"`,
		},
		{
			response: `{
        "strData": "<div class=\"div-class\"></div>"
      }`,
			expected: `{
        "strData": "\u003cdiv class=\"div-class\"\u003e\u003c/div\u003e"
      }`,
		},
	}

	smc := &JSONRestClient{}

	for _, test := range tests {
		res := &payload.BytesPayload{
			Msg:         []byte(test.response),
			ContentType: ContentTypeJSON,
		}

		var w bytes.Buffer
		err := smc.Marshall(&w, res)

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(w.String()).To(Equal(test.expected))
	}
}
