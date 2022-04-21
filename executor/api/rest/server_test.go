package rest

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	guuid "github.com/google/uuid"
	. "github.com/onsi/gomega"
	"github.com/prometheus/common/expfmt"
	"github.com/seldonio/seldon-core/executor/api"
	"github.com/seldonio/seldon-core/executor/api/metric"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/api/test"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
)

const (
	TestSeldonPuid = "1"
)

func TestAliveEndpoint(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	url, _ := url.Parse("http://localhost")
	r := NewServerRestApi(nil, nil, true, url, "default", api.ProtocolSeldon, "test", "/metrics", true)
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
		Graph: v1.PredictiveUnit{
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: "foo",
				ServicePort: 9000,
				Type:        v1.REST,
			},
		},
	}

	url, _ := url.Parse("http://localhost")
	r := NewServerRestApi(&p, &test.SeldonMessageTestClient{}, false, url, "default", api.ProtocolSeldon, "test", "/metrics", true)
	r.Initialise()
	var data = ` {"data":{"ndarray":[1.1,2.0]}}`

	req, _ := http.NewRequest("POST", "/api/v0.1/predictions", strings.NewReader(data))
	req.Header = map[string][]string{"Content-Type": []string{"application/json"}}
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	g.Expect(res.Code).To(Equal(200))
}

func TestCloudeventHeaderIsSet(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	testDepName := "test-deployment"
	testPath := "/api/v0.1/predictions"
	testNamespace := "test-namespace"

	model := v1.MODEL
	p := v1.PredictorSpec{
		Name: "p",
		Graph: v1.PredictiveUnit{
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: "foo",
				ServicePort: 9000,
				Type:        v1.REST,
			},
		},
	}

	url, _ := url.Parse("http://localhost")
	r := NewServerRestApi(&p, &test.SeldonMessageTestClient{}, false, url, testNamespace, api.ProtocolSeldon, testDepName, "/metrics", true)
	r.Initialise()
	var data = ` {"data":{"ndarray":[1.1,2.0]}}`

	req, _ := http.NewRequest("POST", testPath, strings.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ce-Specversion", CLOUDEVENTS_HEADER_SPECVERSION_DEFAULT)
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	g.Expect(res.Code).To(Equal(200))
	g.Expect(len(res.Header().Get(CLOUDEVENTS_HEADER_ID_NAME))).ShouldNot(BeZero())
	g.Expect(len(res.Header().Get(CLOUDEVENTS_HEADER_ID_NAME))).To(Equal(len(guuid.New().String())))
	g.Expect(res.Header().Get(CLOUDEVENTS_HEADER_SPECVERSION_NAME)).To(Equal(CLOUDEVENTS_HEADER_SPECVERSION_DEFAULT))
	g.Expect(res.Header().Get(CLOUDEVENTS_HEADER_PATH_NAME)).To(Equal(testPath))
	g.Expect(res.Header().Get(CLOUDEVENTS_HEADER_TYPE_NAME)).To(Equal("seldon." + testDepName + "." + testNamespace + ".response"))
	g.Expect(res.Header().Get(CLOUDEVENTS_HEADER_SOURCE_NAME)).To(Equal("seldon." + testDepName))
}

func TestCloudeventHeaderIsNotSet(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	testDepName := "test-deployment"
	testPath := "/api/v0.1/predictions"
	testNamespace := "test-namespace"

	model := v1.MODEL
	p := v1.PredictorSpec{
		Name: "p",
		Graph: v1.PredictiveUnit{
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: "foo",
				ServicePort: 9000,
				Type:        v1.REST,
			},
		},
	}

	url, _ := url.Parse("http://localhost")
	r := NewServerRestApi(&p, &test.SeldonMessageTestClient{}, false, url, testNamespace, api.ProtocolSeldon, testDepName, "/metrics", true)
	r.Initialise()
	var data = ` {"data":{"ndarray":[1.1,2.0]}}`

	req, _ := http.NewRequest("POST", testPath, strings.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	g.Expect(res.Code).To(Equal(200))
	g.Expect(len(res.Header().Get(CLOUDEVENTS_HEADER_ID_NAME))).Should(BeZero())
	g.Expect(len(res.Header().Get(CLOUDEVENTS_HEADER_SPECVERSION_NAME))).Should(BeZero())
	g.Expect(len(res.Header().Get(CLOUDEVENTS_HEADER_PATH_NAME))).Should(BeZero())
	g.Expect(len(res.Header().Get(CLOUDEVENTS_HEADER_TYPE_NAME))).Should(BeZero())
	g.Expect(len(res.Header().Get(CLOUDEVENTS_HEADER_SOURCE_NAME))).Should(BeZero())
}

func TestReponsePuuidHeaderIsSet(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	model := v1.MODEL
	p := v1.PredictorSpec{
		Name: "p",
		Graph: v1.PredictiveUnit{
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: "foo",
				ServicePort: 9000,
				Type:        v1.REST,
			},
		},
	}

	url, _ := url.Parse("http://localhost")
	r := NewServerRestApi(&p, &test.SeldonMessageTestClient{}, false, url, "default", api.ProtocolSeldon, "test", "/metrics", true)
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

func TestRequestPuuidHeaderIsSet(t *testing.T) {
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
		Graph: v1.PredictiveUnit{
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: urlParts[0],
				ServicePort: int32(port),
				Type:        v1.REST,
				HttpPort:    int32(port),
			},
		},
	}

	client, err := NewJSONRestClient(api.ProtocolSeldon, "dep", &p, nil)
	g.Expect(err).To(BeNil())
	r := NewServerRestApi(&p, client, false, url, "default", api.ProtocolSeldon, "test", "/metrics", true)
	r.Initialise()
	var data = ` {"data":{"ndarray":[1.1,2.0]}}`

	req, _ := http.NewRequest("POST", "/api/v0.1/predictions", strings.NewReader(data))
	req.Header = map[string][]string{"Content-Type": []string{"application/json"}}
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	g.Expect(res.Code).To(Equal(200))
	g.Expect(called).To(Equal(true))
}

func TestXSSHeaderIsSet(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	model := v1.MODEL
	p := v1.PredictorSpec{
		Name: "p",
		Graph: v1.PredictiveUnit{
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: "foo",
				ServicePort: 9000,
				Type:        v1.REST,
			},
		},
	}

	url, _ := url.Parse("http://localhost")
	r := NewServerRestApi(&p, &test.SeldonMessageTestClient{}, false, url, "default", api.ProtocolSeldon, "test", "/metrics", true)
	r.Initialise()
	var data = ` {"data":{"ndarray":[1.1,2.0]}}`

	req, _ := http.NewRequest("POST", "/api/v0.1/predictions", strings.NewReader(data))
	req.Header = map[string][]string{"Content-Type": []string{"application/json"}}
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	g.Expect(res.Code).To(Equal(200))

	// Check that the XSS middleware is set
	headerVal := res.Header().Get(contentTypeOptsHeader)
	g.Expect(headerVal).To(Equal(contentTypeOptsValue))
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
		Graph: v1.PredictiveUnit{
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: urlParts[0],
				ServicePort: int32(port),
				Type:        v1.REST,
				HttpPort:    int32(port),
			},
		},
	}

	client, err := NewJSONRestClient(api.ProtocolSeldon, "dep", &p, nil)
	g.Expect(err).To(BeNil())
	r := NewServerRestApi(&p, client, false, url, "default", api.ProtocolSeldon, "test", "/metrics", true)
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
		Graph: v1.PredictiveUnit{
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: "foo",
				ServicePort: 9000,
				Type:        v1.REST,
			},
		},
	}

	url, _ := url.Parse("http://localhost")
	r := NewServerRestApi(&p, &test.SeldonMessageTestClient{}, false, url, "default", api.ProtocolSeldon, "test", "/metrics", true)
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
		Graph: v1.PredictiveUnit{
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
	r := NewServerRestApi(&p, &test.SeldonMessageTestClient{}, false, url, "default", api.ProtocolTensorflow, "test", "/metrics", true)
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
		Graph: v1.PredictiveUnit{
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
	r := NewServerRestApi(&p, &test.SeldonMessageTestClient{}, false, url, "default", api.ProtocolSeldon, "test", "/metrics", true)
	r.Initialise()

	req, _ := http.NewRequest("GET", "/api/v1.0/status/mymodel", nil)
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	g.Expect(res.Code).To(Equal(200))
	g.Expect(res.Body.String()).To(Equal(test.TestClientStatusResponse))
}

func TestSeldonStatusDefault(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		Graph: v1.PredictiveUnit{
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: urlParts[0],
				ServicePort: int32(port),
				Type:        v1.REST,
				HttpPort:    int32(port),
			},
		},
	}
	client, err := NewJSONRestClient(api.ProtocolSeldon, "dep", &p, nil)
	g.Expect(err).Should(BeNil())
	r := NewServerRestApi(&p, client, false, url, "default", api.ProtocolSeldon, "test", "/metrics", true)
	r.Initialise()

	req, _ := http.NewRequest("GET", "/api/v1.0/status", nil)
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	g.Expect(res.Code).To(Equal(200))
}

func TestSeldonMetadata(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	model := v1.MODEL
	p := v1.PredictorSpec{
		Name: "p",
		Graph: v1.PredictiveUnit{
			Name: "mymodel",
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: "foo",
				ServicePort: 9000,
				Type:        v1.REST,
			},
		},
	}

	data := `{"metadata":{"name":"mymodel"}}`
	metadataResponse := payload.BytesPayload{Msg: []byte(data)}

	url, _ := url.Parse("http://localhost")
	r := NewServerRestApi(&p, &test.SeldonMessageTestClient{MetadataResponse: &metadataResponse}, false, url, "default", api.ProtocolSeldon, "test", "/metrics", true)
	r.Initialise()

	req, _ := http.NewRequest("GET", "/api/v1.0/metadata/mymodel", nil)
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	g.Expect(res.Code).To(Equal(200))
	g.Expect(res.Body.String()).To(Equal(data))
}

func TestSeldonFeedback(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	model := v1.MODEL
	p := v1.PredictorSpec{
		Name: "p",
		Graph: v1.PredictiveUnit{
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: "foo",
				ServicePort: 9000,
				Type:        v1.REST,
			},
		},
	}
	url, _ := url.Parse("http://localhost")
	r := NewServerRestApi(&p, &test.SeldonMessageTestClient{}, false, url, "default", api.ProtocolSeldon, "test", "/metrics", true)
	r.Initialise()
	var data = ` {"data":{"ndarray":[1.1,2.0]}}`

	req, _ := http.NewRequest("POST", "/api/v1.0/feedback", strings.NewReader(data))
	req.Header = map[string][]string{"Content-Type": []string{"application/json"}}
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	g.Expect(res.Code).To(Equal(200))
}

func TestSeldonGraphMetadata(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	model := v1.MODEL
	p := v1.PredictorSpec{
		Name: "predictor-name",
		Graph: v1.PredictiveUnit{
			Name: "model-1",
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: "foo",
				ServicePort: 9000,
				Type:        v1.REST,
			},
			Children: []v1.PredictiveUnit{
				{
					Name: "model-2",
					Type: &model,
					Endpoint: &v1.Endpoint{
						ServiceHost: "foo",
						ServicePort: 9001,
						Type:        v1.REST,
					},
				},
			},
		},
	}

	metadataMap := map[string]payload.ModelMetadata{
		"model-1": {
			Name:     "model-1",
			Platform: "platform-name",
			Versions: []string{"model-version"},
			Inputs: []map[string]interface{}{
				{"name": "input", "datatype": "BYTES", "shape": []int{1, 5}},
			},
			Outputs: []map[string]interface{}{
				{"name": "output", "datatype": "BYTES", "shape": []int{1, 3}},
			},
		},
		"model-2": {
			Name:     "model-2",
			Platform: "platform-name",
			Versions: []string{"model-version"},
			Inputs: []map[string]interface{}{
				{"name": "input", "datatype": "BYTES", "shape": []int{1, 3}},
			},
			Outputs: []map[string]interface{}{
				{"name": "output", "datatype": "BYTES", "shape": []int{3}},
			},
		},
	}

	url, _ := url.Parse("http://localhost")
	r := NewServerRestApi(&p, &test.SeldonMessageTestClient{ModelMetadataMap: metadataMap}, false, url, "default", api.ProtocolSeldon, "test", "/metrics", true)
	r.Initialise()

	req, _ := http.NewRequest("GET", "/api/v1.0/metadata", nil)
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	g.Expect(res.Code).To(Equal(200))
	g.Expect(res.Body.String()).To(Equal(strings.Join(strings.Fields(test.TestGraphMeta), "")))
}

func TestTensorflowMetadata(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	model := v1.MODEL
	p := v1.PredictorSpec{
		Name: "p",
		Graph: v1.PredictiveUnit{
			Name: "mymodel",
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: "foo",
				ServicePort: 9000,
				Type:        v1.REST,
			},
		},
	}

	data := `{"metadata":{"name":"mymodel"}}`
	metadataResponse := payload.BytesPayload{Msg: []byte(data)}

	url, _ := url.Parse("http://localhost")
	r := NewServerRestApi(&p, &test.SeldonMessageTestClient{MetadataResponse: &metadataResponse}, false, url, "default", api.ProtocolTensorflow, "test", "/metrics", true)
	r.Initialise()

	req, _ := http.NewRequest("GET", "/v1/models/mymodel/metadata", nil)
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	g.Expect(res.Code).To(Equal(200))
	g.Expect(res.Body.String()).To(Equal(data))
}

func TestPredictErrorWithServer(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	called := false
	errorCode := http.StatusConflict

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := ioutil.ReadAll(r.Body)
		g.Expect(err).To(BeNil())
		g.Expect(r.Header.Get(payload.SeldonPUIDHeader)).To(Equal(TestSeldonPuid))
		called = true
		w.WriteHeader(errorCode)
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
		Graph: v1.PredictiveUnit{
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: urlParts[0],
				ServicePort: int32(port),
				Type:        v1.REST,
				HttpPort:    int32(port),
			},
		},
	}
	client, err := NewJSONRestClient(api.ProtocolSeldon, "dep", &p, nil)
	g.Expect(err).Should(BeNil())
	r := NewServerRestApi(&p, client, false, url, "default", api.ProtocolSeldon, "test", "/metrics", true)
	r.Initialise()
	var data = ` {"data":{"ndarray":[1.1,2.0]}}`

	req, _ := http.NewRequest("POST", "/api/v0.1/predictions", strings.NewReader(data))
	req.Header = map[string][]string{"Content-Type": []string{"application/json"}, payload.SeldonPUIDHeader: []string{TestSeldonPuid}}
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	// check error code is the one returned by client
	g.Expect(res.Code).To(Equal(errorCode))
	g.Expect(called).To(Equal(true))
	b, err := ioutil.ReadAll(res.Body)
	g.Expect(err).Should(BeNil())
	g.Expect(string(b)).To(Equal(errorPredictResponse))
}

func TestTensorflowModel(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	model := v1.MODEL
	p := v1.PredictorSpec{
		Name: "p",
		Graph: v1.PredictiveUnit{
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
	r := NewServerRestApi(&p, &test.SeldonMessageTestClient{}, false, url, "default", api.ProtocolTensorflow, "test", "/metrics", true)
	r.Initialise()

	var data = ` {"instances":[[1,2,3]]}`
	req, _ := http.NewRequest("POST", "/v1/models/:predict", strings.NewReader(data))
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	g.Expect(res.Code).To(Equal(200))
	g.Expect(res.Body.String()).To(Equal(data))
}

func TestTensorflowModelBadModelName(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	model := v1.MODEL
	p := v1.PredictorSpec{
		Name: "p",
		Graph: v1.PredictiveUnit{
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
	r := NewServerRestApi(&p, &test.SeldonMessageTestClient{}, false, url, "default", api.ProtocolTensorflow, "test", "/metrics", true)
	r.Initialise()

	var data = ` {"instances":[[1,2,3]]}`
	req, _ := http.NewRequest("POST", "/v1/models/xyz/:predict", strings.NewReader(data))
	res := httptest.NewRecorder()
	r.Router.ServeHTTP(res, req)
	g.Expect(res.Code).To(Equal(200))
}
