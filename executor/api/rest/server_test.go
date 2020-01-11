package rest

import (
	"github.com/onsi/gomega"
	"github.com/prometheus/common/expfmt"
	"github.com/seldonio/seldon-core/executor/api/metric"
	"github.com/seldonio/seldon-core/executor/api/test"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning/v1"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestAliveEndpoint(t *testing.T) {
	t.Logf("Started")
	g := gomega.NewGomegaWithT(t)

	url, _ := url.Parse("http://localhost")
	r := NewServerRestApi(nil, nil, true, url, "default", ProtocolSeldon, "test", "/metrics")
	r.Initialise()

	req, _ := http.NewRequest("GET", "/live", nil)
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)

	g.Expect(res.Code).To(gomega.Equal(200))
}

func TestSimpleModel(t *testing.T) {
	t.Logf("Started")
	g := gomega.NewGomegaWithT(t)

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
	r := NewServerRestApi(&p, test.NewSeldonMessageTestClient(t, 0, nil, nil), false, url, "default", ProtocolSeldon, "test", "/metrics")
	r.Initialise()
	var data = ` {"data":{"ndarray":[1.1,2.0]}}`

	req, _ := http.NewRequest("POST", "/api/v0.1/predictions", strings.NewReader(data))
	req.Header = map[string][]string{"Content-Type": []string{"application/json"}}
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	g.Expect(res.Code).To(gomega.Equal(200))
}

func TestServerMetrics(t *testing.T) {
	t.Logf("Started")
	g := gomega.NewGomegaWithT(t)

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
	r := NewServerRestApi(&p, test.NewSeldonMessageTestClient(t, 0, nil, nil), false, url, "default", ProtocolSeldon, "test", "/metrics")
	r.Initialise()

	var data = ` {"data":{"ndarray":[1.1,2.0]}}`

	req, _ := http.NewRequest("POST", "/api/v0.1/predictions", strings.NewReader(data))
	req.Header = map[string][]string{"Content-Type": []string{"application/json"}}
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	g.Expect(res.Code).To(gomega.Equal(200))

	req, _ = http.NewRequest("GET", "/metrics", nil)
	res = httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	g.Expect(res.Code).To(gomega.Equal(200))
	tp := expfmt.TextParser{}
	metrics, err := tp.TextToMetricFamilies(res.Body)
	g.Expect(err).Should(gomega.BeNil())
	g.Expect(metrics[metric.ServerRequestsMetricName]).ShouldNot(gomega.BeNil())

}

func TestTensorflowStatus(t *testing.T) {
	t.Logf("Started")
	g := gomega.NewGomegaWithT(t)

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
	r := NewServerRestApi(&p, test.NewSeldonMessageTestClient(t, 0, nil, nil), false, url, "default", ProtocolTensorflow, "test", "/metrics")
	r.Initialise()

	req, _ := http.NewRequest("GET", "/v1/models/mymodel", nil)
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	g.Expect(res.Code).To(gomega.Equal(200))
	g.Expect(res.Body.String()).To(gomega.Equal(test.TestClientStatusResponse))
}

func TestSeldonStatus(t *testing.T) {
	t.Logf("Started")
	g := gomega.NewGomegaWithT(t)

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
	r := NewServerRestApi(&p, test.NewSeldonMessageTestClient(t, 0, nil, nil), false, url, "default", ProtocolSeldon, "test", "/metrics")
	r.Initialise()

	req, _ := http.NewRequest("GET", "/api/v1.0/status/mymodel", nil)
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	g.Expect(res.Code).To(gomega.Equal(200))
	g.Expect(res.Body.String()).To(gomega.Equal(test.TestClientStatusResponse))
}

func TestSeldonMetadata(t *testing.T) {
	t.Logf("Started")
	g := gomega.NewGomegaWithT(t)

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
	r := NewServerRestApi(&p, test.NewSeldonMessageTestClient(t, 0, nil, nil), false, url, "default", ProtocolSeldon, "test", "/metrics")
	r.Initialise()

	req, _ := http.NewRequest("GET", "/api/v1.0/metadata/mymodel", nil)
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	g.Expect(res.Code).To(gomega.Equal(200))
	g.Expect(res.Body.String()).To(gomega.Equal(test.TestClientMetadataResponse))
}

func TestTensorflowMetadata(t *testing.T) {
	t.Logf("Started")
	g := gomega.NewGomegaWithT(t)

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
	r := NewServerRestApi(&p, test.NewSeldonMessageTestClient(t, 0, nil, nil), false, url, "default", ProtocolTensorflow, "test", "/metrics")
	r.Initialise()

	req, _ := http.NewRequest("GET", "/v1/models/mymodel/metadata", nil)
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	g.Expect(res.Code).To(gomega.Equal(200))
	g.Expect(res.Body.String()).To(gomega.Equal(test.TestClientMetadataResponse))
}
