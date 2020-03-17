package rest

import (
	guuid "github.com/google/uuid"
	. "github.com/onsi/gomega"
	"github.com/prometheus/common/expfmt"
	"github.com/seldonio/seldon-core/executor/api"
	"github.com/seldonio/seldon-core/executor/api/metric"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/api/test"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning/v1"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

const (
	TestSeldonPuid = "1"
)

func TestAliveEndpoint(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	url, _ := url.Parse("http://localhost")
	r := NewServerRestApi(nil, nil, true, url, "default", api.ProtocolSeldon, "test", "/metrics")
	r.Initialise()

	req, _ := http.NewRequest("GET", "/live", nil)
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)

	g.Expect(res.Code).To(Equal(200))
}

func TestSimpleModel(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	model := v1.MODEL
	p := v1.PredictorSpec{
		Name: "p",
		Graph: &v1.PredictiveUnit{
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: "foo",
				ServicePort: 9000,
				Type:        v1.REST,
			},
		},
	}

	url, _ := url.Parse("http://localhost")
	r := NewServerRestApi(&p, test.NewSeldonMessageTestClient(t, 0, nil, nil), false, url, "default", api.ProtocolSeldon, "test", "/metrics")
	r.Initialise()
	var data = ` {"data":{"ndarray":[1.1,2.0]}}`

	req, _ := http.NewRequest("POST", "/api/v0.1/predictions", strings.NewReader(data))
	req.Header = map[string][]string{"Content-Type": []string{"application/json"}}
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	g.Expect(res.Code).To(Equal(200))
}

func TestRequestPuuidIsSet(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	model := v1.MODEL
	p := v1.PredictorSpec{
		Name: "p",
		Graph: &v1.PredictiveUnit{
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: "foo",
				ServicePort: 9000,
				Type:        v1.REST,
			},
		},
	}

	url, _ := url.Parse("http://localhost")
	r := NewServerRestApi(&p, test.NewSeldonMessageTestClient(t, 0, nil, nil), false, url, "default", api.ProtocolSeldon, "test", "/metrics")
	r.Initialise()
	var data = ` {"data":{"ndarray":[1.1,2.0]}}`

	req, _ := http.NewRequest("POST", "/api/v0.1/predictions", strings.NewReader(data))
	req.Header = map[string][]string{"Content-Type": []string{"application/json"}}
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	g.Expect(res.Code).To(Equal(200))
	// Check that the SeldonPUUIDHeader is set
	g.Expect(len(res.Header().Get(payload.SeldonPUIDHeader))).ShouldNot(BeZero())
	g.Expect(len(res.Header().Get(payload.SeldonPUIDHeader))).To(Equal(len(guuid.New().String())))
}

func TestRequestPuuidIsSet(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	called := false

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := ioutil.ReadAll(r.Body)
		g.Expect(err).To(BeNil())
		g.Expect(len(r.Header.Get(payload.SeldonPUIDHeader))).To(Equal(len(guuid.New().String())))
		called = true
		w.Write([]byte(bodyBytes))
	})
	server := httptest.NewServer(handler)
	defer server.Close()
	url, err := url.Parse(server.URL)
	g.Expect(err).Should(BeNil())
	urlParts := strings.Split(url.Host, ":")
	port, err := strconv.Atoi(urlParts[1])
	g.Expect(err).Should(BeNil())

	model := v1.MODEL
	p := v1.PredictorSpec{
		Name: "p",
		Graph: &v1.PredictiveUnit{
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: urlParts[0],
				ServicePort: int32(port),
				Type:        v1.REST,
			},
		},
	}

	client, err := NewJSONRestClient(api.ProtocolSeldon, "dep", &p, nil)
	g.Expect(err).To(BeNil())
	r := NewServerRestApi(&p, client, false, url, "default", api.ProtocolSeldon, "test", "/metrics")
	r.Initialise()
	var data = ` {"data":{"ndarray":[1.1,2.0]}}`

	req, _ := http.NewRequest("POST", "/api/v0.1/predictions", strings.NewReader(data))
	req.Header = map[string][]string{"Content-Type": []string{"application/json"}}
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	g.Expect(res.Code).To(Equal(200))
	g.Expect(called).To(Equal(true))
}

func TestModelWithServer(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	called := false

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := ioutil.ReadAll(r.Body)
		g.Expect(err).To(BeNil())
		g.Expect(r.Header.Get(payload.SeldonPUIDHeader)).To(Equal(TestSeldonPuid))
		called = true
		w.Write([]byte(bodyBytes))
	})
	server := httptest.NewServer(handler)
	defer server.Close()
	url, err := url.Parse(server.URL)
	g.Expect(err).Should(BeNil())
	urlParts := strings.Split(url.Host, ":")
	port, err := strconv.Atoi(urlParts[1])
	g.Expect(err).Should(BeNil())

	model := v1.MODEL
	p := v1.PredictorSpec{
		Name: "p",
		Graph: &v1.PredictiveUnit{
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: urlParts[0],
				ServicePort: int32(port),
				Type:        v1.REST,
			},
		},
	}

	client, err := NewJSONRestClient(api.ProtocolSeldon, "dep", &p, nil)
	g.Expect(err).To(BeNil())
	r := NewServerRestApi(&p, client, false, url, "default", api.ProtocolSeldon, "test", "/metrics")
	r.Initialise()
	var data = ` {"data":{"ndarray":[1.1,2.0]}}`

	req, _ := http.NewRequest("POST", "/api/v0.1/predictions", strings.NewReader(data))
	req.Header = map[string][]string{"Content-Type": []string{"application/json"}, payload.SeldonPUIDHeader: []string{TestSeldonPuid}}
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	g.Expect(res.Code).To(Equal(200))
	g.Expect(called).To(Equal(true))

}

func TestServerMetrics(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	model := v1.MODEL
	p := v1.PredictorSpec{
		Name: "p",
		Graph: &v1.PredictiveUnit{
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: "foo",
				ServicePort: 9000,
				Type:        v1.REST,
			},
		},
	}

	url, _ := url.Parse("http://localhost")
	r := NewServerRestApi(&p, test.NewSeldonMessageTestClient(t, 0, nil, nil), false, url, "default", api.ProtocolSeldon, "test", "/metrics")
	r.Initialise()

	var data = ` {"data":{"ndarray":[1.1,2.0]}}`

	req, _ := http.NewRequest("POST", "/api/v0.1/predictions", strings.NewReader(data))
	req.Header = map[string][]string{"Content-Type": []string{"application/json"}}
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	g.Expect(res.Code).To(Equal(200))

	req, _ = http.NewRequest("GET", "/metrics", nil)
	res = httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	g.Expect(res.Code).To(Equal(200))
	tp := expfmt.TextParser{}
	metrics, err := tp.TextToMetricFamilies(res.Body)
	g.Expect(err).Should(BeNil())
	g.Expect(metrics[metric.ServerRequestsMetricName]).ShouldNot(BeNil())

}

func TestTensorflowStatus(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	model := v1.MODEL
	p := v1.PredictorSpec{
		Name: "p",
		Graph: &v1.PredictiveUnit{
			Name: "mymodel",
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: "foo",
				ServicePort: 9000,
				Type:        v1.REST,
			},
		},
	}

	url, _ := url.Parse("http://localhost")
	r := NewServerRestApi(&p, test.NewSeldonMessageTestClient(t, 0, nil, nil), false, url, "default", api.ProtocolTensorflow, "test", "/metrics")
	r.Initialise()

	req, _ := http.NewRequest("GET", "/v1/models/mymodel", nil)
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	g.Expect(res.Code).To(Equal(200))
	g.Expect(res.Body.String()).To(Equal(test.TestClientStatusResponse))
}

func TestSeldonStatus(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	model := v1.MODEL
	p := v1.PredictorSpec{
		Name: "p",
		Graph: &v1.PredictiveUnit{
			Name: "mymodel",
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: "foo",
				ServicePort: 9000,
				Type:        v1.REST,
			},
		},
	}

	url, _ := url.Parse("http://localhost")
	r := NewServerRestApi(&p, test.NewSeldonMessageTestClient(t, 0, nil, nil), false, url, "default", api.ProtocolSeldon, "test", "/metrics")
	r.Initialise()

	req, _ := http.NewRequest("GET", "/api/v1.0/status/mymodel", nil)
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	g.Expect(res.Code).To(Equal(200))
	g.Expect(res.Body.String()).To(Equal(test.TestClientStatusResponse))
}

func TestSeldonMetadata(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	model := v1.MODEL
	p := v1.PredictorSpec{
		Name: "p",
		Graph: &v1.PredictiveUnit{
			Name: "mymodel",
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: "foo",
				ServicePort: 9000,
				Type:        v1.REST,
			},
		},
	}

	url, _ := url.Parse("http://localhost")
	r := NewServerRestApi(&p, test.NewSeldonMessageTestClient(t, 0, nil, nil), false, url, "default", api.ProtocolSeldon, "test", "/metrics")
	r.Initialise()

	req, _ := http.NewRequest("GET", "/api/v1.0/metadata/mymodel", nil)
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	g.Expect(res.Code).To(Equal(200))
	g.Expect(res.Body.String()).To(Equal(test.TestClientMetadataResponse))
}

func TestTensorflowMetadata(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	model := v1.MODEL
	p := v1.PredictorSpec{
		Name: "p",
		Graph: &v1.PredictiveUnit{
			Name: "mymodel",
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: "foo",
				ServicePort: 9000,
				Type:        v1.REST,
			},
		},
	}

	url, _ := url.Parse("http://localhost")
	r := NewServerRestApi(&p, test.NewSeldonMessageTestClient(t, 0, nil, nil), false, url, "default", api.ProtocolTensorflow, "test", "/metrics")
	r.Initialise()

	req, _ := http.NewRequest("GET", "/v1/models/mymodel/metadata", nil)
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	g.Expect(res.Code).To(Equal(200))
	g.Expect(res.Body.String()).To(Equal(test.TestClientMetadataResponse))
}

func TestPredictErrorWithServer(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	called := false

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := ioutil.ReadAll(r.Body)
		g.Expect(err).To(BeNil())
		g.Expect(r.Header.Get(payload.SeldonPUIDHeader)).To(Equal(TestSeldonPuid))
		called = true
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errorPredictResponse))
	})
	server := httptest.NewServer(handler)
	defer server.Close()
	url, err := url.Parse(server.URL)
	g.Expect(err).Should(BeNil())
	urlParts := strings.Split(url.Host, ":")
	port, err := strconv.Atoi(urlParts[1])
	g.Expect(err).Should(BeNil())

	model := v1.MODEL
	p := v1.PredictorSpec{
		Name: "p",
		Graph: &v1.PredictiveUnit{
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: urlParts[0],
				ServicePort: int32(port),
				Type:        v1.REST,
			},
		},
	}
	client, err := NewJSONRestClient(api.ProtocolSeldon, "dep", &p, nil)
	g.Expect(err).Should(BeNil())
	r := NewServerRestApi(&p, client, false, url, "default", api.ProtocolSeldon, "test", "/metrics")
	r.Initialise()
	var data = ` {"data":{"ndarray":[1.1,2.0]}}`

	req, _ := http.NewRequest("POST", "/api/v0.1/predictions", strings.NewReader(data))
	req.Header = map[string][]string{"Content-Type": []string{"application/json"}, payload.SeldonPUIDHeader: []string{TestSeldonPuid}}
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	g.Expect(res.Code).To(Equal(http.StatusInternalServerError))
	g.Expect(called).To(Equal(true))
	b, err := ioutil.ReadAll(res.Body)
	g.Expect(err).Should(BeNil())
	g.Expect(string(b)).To(Equal(errorPredictResponse))
}
