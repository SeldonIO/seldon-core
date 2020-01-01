package rest

import (
	"context"
	"crypto/tls"
	"github.com/golang/protobuf/jsonpb"
	"github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"github.com/seldonio/seldon-core/executor/api/metric"
	"github.com/seldonio/seldon-core/executor/api/payload"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning/v1"
	v12 "k8s.io/api/core/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
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
)

func testingHTTPClient(g *gomega.GomegaWithT, handler http.Handler) (string, int, *http.Client, func()) {
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
	g.Expect(err).Should(gomega.BeNil())
	port, err := strconv.Atoi(url.Port())
	g.Expect(err).Should(gomega.BeNil())

	return url.Hostname(), port, cli, s.Close
}

func SetHTTPClient(httpClient *http.Client) BytesRestClientOption {
	return func(cli *JSONRestClient) {
		cli.httpClient = httpClient
	}
}

func createPayload(g *gomega.GomegaWithT) payload.SeldonPayload {
	var data = ` {"data":{"ndarray":[1.1,2.0]}}`
	return &payload.BytesPayload{Msg: []byte(data)}
}

func TestSimpleMethods(t *testing.T) {
	t.Logf("Started")
	g := gomega.NewGomegaWithT(t)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(okPredictResponse))
	})
	host, port, httpClient, teardown := testingHTTPClient(g, h)
	defer teardown()
	predictor := v1.PredictorSpec{
		Name:        "test",
		Annotations: map[string]string{},
	}
	seldonRestClient := NewJSONRestClient(ProtocolSeldon, "test", &predictor, SetHTTPClient(httpClient))

	methods := []func(context.Context, string, string, int32, payload.SeldonPayload) (payload.SeldonPayload, error){seldonRestClient.Predict, seldonRestClient.TransformInput, seldonRestClient.TransformOutput}
	for _, method := range methods {
		resPayload, err := method(context.TODO(), "model", host, int32(port), createPayload(g))
		g.Expect(err).Should(gomega.BeNil())

		data := resPayload.GetPayload().([]byte)
		var smRes proto.SeldonMessage
		err = jsonpb.UnmarshalString(string(data), &smRes)
		g.Expect(err).Should(gomega.BeNil())
		g.Expect(smRes.GetData().GetNdarray().Values[0].GetListValue().Values[0].GetNumberValue()).Should(gomega.Equal(0.9))
		g.Expect(smRes.GetData().GetNdarray().Values[0].GetListValue().Values[1].GetNumberValue()).Should(gomega.Equal(0.1))
	}

}

func TestRouter(t *testing.T) {
	t.Logf("Started")
	g := gomega.NewGomegaWithT(t)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(okRouteResponse))
	})
	host, port, httpClient, teardown := testingHTTPClient(g, h)
	defer teardown()
	predictor := v1.PredictorSpec{
		Name:        "test",
		Annotations: map[string]string{},
	}
	seldonRestClient := NewJSONRestClient(ProtocolSeldon, "test", &predictor, SetHTTPClient(httpClient))

	route, err := seldonRestClient.Route(context.TODO(), "model", host, int32(port), createPayload(g))
	g.Expect(err).Should(gomega.BeNil())

	g.Expect(route).Should(gomega.Equal(1))
}
func createCombinerPayload(g *gomega.GomegaWithT) []payload.SeldonPayload {
	var data = ` {"data":{"ndarray":[1.1,2.0]}}`
	smp := []payload.SeldonPayload{&payload.BytesPayload{Msg: []byte(data)}, &payload.BytesPayload{Msg: []byte(data)}}
	return smp
}

func TestCombiner(t *testing.T) {
	t.Logf("Started")
	g := gomega.NewGomegaWithT(t)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(okPredictResponse))
	})
	host, port, httpClient, teardown := testingHTTPClient(g, h)
	defer teardown()
	predictor := v1.PredictorSpec{
		Name:        "test",
		Annotations: map[string]string{},
	}
	seldonRestClient := NewJSONRestClient(ProtocolSeldon, "test", &predictor, SetHTTPClient(httpClient))

	resPayload, err := seldonRestClient.Combine(context.TODO(), "model", host, int32(port), createCombinerPayload(g))
	g.Expect(err).Should(gomega.BeNil())

	data := resPayload.GetPayload().([]byte)
	var smRes proto.SeldonMessage
	err = jsonpb.UnmarshalString(string(data), &smRes)
	g.Expect(err).Should(gomega.BeNil())
	g.Expect(smRes.GetData().GetNdarray().Values[0].GetListValue().Values[0].GetNumberValue()).Should(gomega.Equal(0.9))
	g.Expect(smRes.GetData().GetNdarray().Values[0].GetListValue().Values[1].GetNumberValue()).Should(gomega.Equal(0.1))
}

func TestClientMetrics(t *testing.T) {
	t.Logf("Started")
	g := gomega.NewGomegaWithT(t)
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
	seldonRestClient := NewJSONRestClient(ProtocolSeldon, "test", &predictor, SetHTTPClient(httpClient))

	methods := []func(context.Context, string, string, int32, payload.SeldonPayload) (payload.SeldonPayload, error){seldonRestClient.Predict, seldonRestClient.TransformInput, seldonRestClient.TransformOutput}
	for _, method := range methods {
		resPayload, err := method(context.TODO(), "model", host, int32(port), createPayload(g))
		g.Expect(err).Should(gomega.BeNil())

		data := resPayload.GetPayload().([]byte)
		var smRes proto.SeldonMessage
		err = jsonpb.UnmarshalString(string(data), &smRes)
		g.Expect(err).Should(gomega.BeNil())
		g.Expect(smRes.GetData().GetNdarray().Values[0].GetListValue().Values[0].GetNumberValue()).Should(gomega.Equal(0.9))
		g.Expect(smRes.GetData().GetNdarray().Values[0].GetListValue().Values[1].GetNumberValue()).Should(gomega.Equal(0.1))

		mfs, err := prometheus.DefaultGatherer.Gather()
		g.Expect(err).Should(gomega.BeNil())
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
		g.Expect(found).Should(gomega.Equal(true))
		g.Expect(foundImage).Should(gomega.Equal(true))
		g.Expect(foundImageVersion).Should(gomega.Equal(true))
	}

}
