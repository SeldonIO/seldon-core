package predictor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/golang/protobuf/jsonpb"
	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/executor/api/grpc"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/api/test"
	"github.com/seldonio/seldon-core/executor/logger"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const (
	testSourceUrl         = "http://localhost"
	testSeldonPuid        = "1"
	testCustomMetaKey     = "key"
	testCustomMetaValue   = "foo"
	modelIdHeaderName     = "Ce-Modelid"
	contentTypeHeaderName = "Content-Type"
	requestIdHeaderName   = "Ce-Requestid"
)

func createPredictorProcess(t *testing.T) *PredictorProcess {
	url, _ := url.Parse(testSourceUrl)
	ctx := context.WithValue(context.TODO(), payload.SeldonPUIDHeader, testSeldonPuid)
	pp := NewPredictorProcess(ctx, &test.SeldonMessageTestClient{}, logf.Log.WithName("SeldonMessageRestClient"), url, "default", map[string][]string{testCustomMetaKey: []string{testCustomMetaValue}}, "")
	return &pp
}

func createPredictorProcessWithModel(t *testing.T, modelName string) *PredictorProcess {
	url, _ := url.Parse(testSourceUrl)
	ctx := context.WithValue(context.TODO(), payload.SeldonPUIDHeader, testSeldonPuid)
	pp := NewPredictorProcess(ctx, &test.SeldonMessageTestClient{}, logf.Log.WithName("SeldonMessageRestClient"), url, "default", map[string][]string{testCustomMetaKey: []string{testCustomMetaValue}}, modelName)
	return &pp
}

func createPredictorProcessWithMetadata(t *testing.T, metadataResponse payload.SeldonPayload, modelMetadataMap map[string]payload.ModelMetadata) *PredictorProcess {
	url, _ := url.Parse(testSourceUrl)
	ctx := context.WithValue(context.TODO(), payload.SeldonPUIDHeader, testSeldonPuid)
	pp := NewPredictorProcess(ctx, &test.SeldonMessageTestClient{MetadataResponse: metadataResponse, ModelMetadataMap: modelMetadataMap}, logf.Log.WithName("SeldonMessageRestClient"), url, "default", map[string][]string{testCustomMetaKey: []string{testCustomMetaValue}}, "")
	return &pp
}

func createPredictorProcessWithRoute(t *testing.T, chosenRoute int) *PredictorProcess {
	url, _ := url.Parse(testSourceUrl)
	ctx := context.WithValue(context.TODO(), payload.SeldonPUIDHeader, testSeldonPuid)
	pp := NewPredictorProcess(ctx, &test.SeldonMessageTestClient{ChosenRoute: chosenRoute}, logf.Log.WithName("SeldonMessageRestClient"), url, "default", map[string][]string{}, "")
	return &pp
}

func createPredictorProcessWithError(t *testing.T, errMethod *v1.PredictiveUnitMethod, err error, errPayload payload.SeldonPayload) *PredictorProcess {
	url, _ := url.Parse(testSourceUrl)
	ctx := context.WithValue(context.TODO(), payload.SeldonPUIDHeader, testSeldonPuid)
	pp := NewPredictorProcess(ctx, &test.SeldonMessageTestClient{ErrMethod: errMethod, Err: err, ErrPayload: errPayload}, logf.Log.WithName("SeldonMessageRestClient"), url, "default", map[string][]string{}, "")
	return &pp
}

func createPredictorProcessWithoutPUIDInContext(t *testing.T) *PredictorProcess {
	url, _ := url.Parse(testSourceUrl)
	ctx := context.TODO()
	pp := NewPredictorProcess(ctx, &test.SeldonMessageTestClient{}, logf.Log.WithName("SeldonMessageRestClient"), url, "default", map[string][]string{testCustomMetaKey: []string{testCustomMetaValue}}, "")
	return &pp
}

func createPredictPayload(g *GomegaWithT) payload.SeldonPayload {
	var sm proto.SeldonMessage
	var data = ` {"data":{"ndarray":[1.1,2.0]}}`
	err := jsonpb.UnmarshalString(data, &sm)
	g.Expect(err).Should(BeNil())
	return &payload.ProtoPayload{Msg: &sm}
}

func createMetadataPayload(g *GomegaWithT) payload.SeldonPayload {
	var sm proto.SeldonModelMetadata
	var data = `{"name": "mymodel"}`
	err := jsonpb.UnmarshalString(data, &sm)
	g.Expect(err).Should(BeNil())
	return &payload.ProtoPayload{Msg: &sm}
}

func createFeedbackPayload(g *GomegaWithT) payload.SeldonPayload {
	var sm proto.Feedback
	var data = ` {"request":{"data":{"ndarray":[1.1,2.0]}}}`
	err := jsonpb.UnmarshalString(data, &sm)
	g.Expect(err).Should(BeNil())
	return &payload.ProtoPayload{Msg: &sm}
}

func TestModel(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	model := v1.MODEL
	graph := &v1.PredictiveUnit{
		Type: &model,
		Endpoint: &v1.Endpoint{
			ServiceHost: "foo",
			ServicePort: 9000,
			Type:        v1.REST,
		},
	}

	pResp, err := createPredictorProcess(t).Predict(graph, createPredictPayload(g))
	g.Expect(err).Should(BeNil())
	smRes := pResp.GetPayload().(*proto.SeldonMessage)
	g.Expect(smRes.GetData().GetNdarray().Values[0].GetNumberValue()).Should(Equal(1.1))
	g.Expect(smRes.GetData().GetNdarray().Values[1].GetNumberValue()).Should(Equal(2.0))
}

func TestModelOverride(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	model := v1.MODEL
	graph := &v1.PredictiveUnit{
		Type: &model,
		Endpoint: &v1.Endpoint{
			ServiceHost: "foo",
			ServicePort: 9000,
			Type:        v1.REST,
		},
	}

	pResp, err := createPredictorProcessWithModel(t, "cifar10").Predict(graph, createPredictPayload(g))
	g.Expect(err).Should(BeNil())
	smRes := pResp.GetPayload().(*proto.SeldonMessage)
	g.Expect(smRes.GetData().GetNdarray().Values[0].GetNumberValue()).Should(Equal(1.1))
	g.Expect(smRes.GetData().GetNdarray().Values[1].GetNumberValue()).Should(Equal(2.0))
}

func TestStatus(t *testing.T) {
	t.Logf("Started")
	modelName := "mymodel"
	g := NewGomegaWithT(t)
	model := v1.MODEL
	graph := &v1.PredictiveUnit{
		Name: modelName,
		Type: &model,
		Endpoint: &v1.Endpoint{
			ServiceHost: "foo",
			ServicePort: 9000,
			Type:        v1.REST,
		},
	}

	pResp, err := createPredictorProcess(t).Status(graph, modelName, nil)
	g.Expect(err).Should(BeNil())
	smRes := string(pResp.GetPayload().([]byte))
	g.Expect(smRes).To(Equal(test.TestClientStatusResponse))

}

func TestMetadata(t *testing.T) {
	t.Logf("Started")
	modelName := "mymodel"
	g := NewGomegaWithT(t)
	model := v1.MODEL
	graph := &v1.PredictiveUnit{
		Name: modelName,
		Type: &model,
		Endpoint: &v1.Endpoint{
			ServiceHost: "foo",
			ServicePort: 9000,
			Type:        v1.REST,
		},
	}

	data := `{"metadata":{"name":"mymodel"}}`
	metadataResponse := payload.BytesPayload{Msg: []byte(data)}

	pResp, err := createPredictorProcessWithMetadata(t, &metadataResponse, nil).Metadata(graph, modelName, createMetadataPayload(g))
	g.Expect(err).Should(BeNil())
	smRes := string(pResp.GetPayload().([]byte))
	g.Expect(smRes).To(Equal(data))
}

func TestTwoLevelModel(t *testing.T) {
	g := NewGomegaWithT(t)
	model := v1.MODEL
	graph := &v1.PredictiveUnit{
		Type: &model,
		Endpoint: &v1.Endpoint{
			ServiceHost: "foo",
			ServicePort: 9000,
			Type:        v1.REST,
		},
		Children: []v1.PredictiveUnit{
			{
				Type: &model,
				Endpoint: &v1.Endpoint{
					ServiceHost: "foo2",
					ServicePort: 9001,
					Type:        v1.REST,
				},
			},
		},
	}

	pResp, err := createPredictorProcess(t).Predict(graph, createPredictPayload(g))
	g.Expect(err).Should(BeNil())
	smRes := pResp.GetPayload().(*proto.SeldonMessage)
	g.Expect(smRes.GetData().GetNdarray().Values[0].GetNumberValue()).Should(Equal(1.1))
	g.Expect(smRes.GetData().GetNdarray().Values[1].GetNumberValue()).Should(Equal(2.0))
}

func TestCombiner(t *testing.T) {
	g := NewGomegaWithT(t)
	model := v1.MODEL
	combiner := v1.COMBINER
	graph := &v1.PredictiveUnit{
		Type: &combiner,
		Endpoint: &v1.Endpoint{
			ServiceHost: "foo",
			ServicePort: 9000,
			Type:        v1.REST,
		},
		Children: []v1.PredictiveUnit{
			{
				Type: &model,
				Endpoint: &v1.Endpoint{
					ServiceHost: "foo2",
					ServicePort: 9001,
					Type:        v1.REST,
				},
			},
			{
				Type: &model,
				Endpoint: &v1.Endpoint{
					ServiceHost: "foo3",
					ServicePort: 9002,
					Type:        v1.REST,
				},
			},
		},
	}

	pResp, err := createPredictorProcess(t).Predict(graph, createPredictPayload(g))
	g.Expect(err).Should(BeNil())
	smRes := pResp.GetPayload().(*proto.SeldonMessage)
	g.Expect(smRes.GetData().GetNdarray().Values[0].GetNumberValue()).Should(Equal(1.1))
	g.Expect(smRes.GetData().GetNdarray().Values[1].GetNumberValue()).Should(Equal(2.0))
}

func TestMethods(t *testing.T) {
	g := NewGomegaWithT(t)
	//model := v1.UNKNOWN_TYPE
	graph := &v1.PredictiveUnit{
		Methods: &[]v1.PredictiveUnitMethod{v1.TRANSFORM_INPUT, v1.TRANSFORM_OUTPUT, v1.ROUTE, v1.AGGREGATE},
		Endpoint: &v1.Endpoint{
			ServiceHost: "foo",
			ServicePort: 9000,
			Type:        v1.REST,
		},
		Children: []v1.PredictiveUnit{
			{
				Methods: &[]v1.PredictiveUnitMethod{v1.TRANSFORM_INPUT, v1.TRANSFORM_OUTPUT},
				Endpoint: &v1.Endpoint{
					ServiceHost: "foo2",
					ServicePort: 9001,
					Type:        v1.REST,
				},
			},
		},
	}

	pResp, err := createPredictorProcess(t).Predict(graph, createPredictPayload(g))
	g.Expect(err).Should(BeNil())
	smRes := pResp.GetPayload().(*proto.SeldonMessage)
	g.Expect(smRes.GetData().GetNdarray().Values[0].GetNumberValue()).Should(Equal(1.1))
	g.Expect(smRes.GetData().GetNdarray().Values[1].GetNumberValue()).Should(Equal(2.0))
}

func TestFeedback(t *testing.T) {
	g := NewGomegaWithT(t)
	//model := v1.UNKNOWN_TYPE
	graph := &v1.PredictiveUnit{
		Methods: &[]v1.PredictiveUnitMethod{v1.SEND_FEEDBACK},
		Endpoint: &v1.Endpoint{
			ServiceHost: "foo",
			ServicePort: 9000,
			Type:        v1.REST,
		},
		Children: []v1.PredictiveUnit{
			{
				Methods: &[]v1.PredictiveUnitMethod{v1.SEND_FEEDBACK},
				Endpoint: &v1.Endpoint{
					ServiceHost: "foo2",
					ServicePort: 9001,
					Type:        v1.REST,
				},
			},
		},
	}

	pResp, err := createPredictorProcess(t).Feedback(graph, createFeedbackPayload(g))
	g.Expect(err).Should(BeNil())
	smRes := pResp.GetPayload().(*proto.SeldonMessage)
	g.Expect(smRes.GetData().GetNdarray().Values[0].GetNumberValue()).Should(Equal(1.1))
	g.Expect(smRes.GetData().GetNdarray().Values[1].GetNumberValue()).Should(Equal(2.0))
}

func TestRouter(t *testing.T) {
	g := NewGomegaWithT(t)
	model := v1.MODEL
	router := v1.ROUTER
	graph := &v1.PredictiveUnit{
		Type: &router,
		Endpoint: &v1.Endpoint{
			ServiceHost: "foo",
			ServicePort: 9000,
			Type:        v1.REST,
		},
		Children: []v1.PredictiveUnit{
			{
				Type: &model,
				Endpoint: &v1.Endpoint{
					ServiceHost: "foo2",
					ServicePort: 9001,
					Type:        v1.REST,
				},
			},
			{
				Type: &model,
				Endpoint: &v1.Endpoint{
					ServiceHost: "foo3",
					ServicePort: 9002,
					Type:        v1.REST,
				},
			},
		},
	}

	pResp, err := createPredictorProcess(t).Predict(graph, createPredictPayload(g))
	g.Expect(err).Should(BeNil())
	smRes := pResp.GetPayload().(*proto.SeldonMessage)
	g.Expect(smRes.GetData().GetNdarray().Values[0].GetNumberValue()).Should(Equal(1.1))
	g.Expect(smRes.GetData().GetNdarray().Values[1].GetNumberValue()).Should(Equal(2.0))

	pResp, err = createPredictorProcessWithRoute(t, 0).Predict(graph, createPredictPayload(g))
	g.Expect(err).Should(BeNil())
	smRes = pResp.GetPayload().(*proto.SeldonMessage)
	g.Expect(smRes.GetData().GetNdarray().Values[0].GetNumberValue()).Should(Equal(1.1))
	g.Expect(smRes.GetData().GetNdarray().Values[1].GetNumberValue()).Should(Equal(2.0))

	pResp, err = createPredictorProcessWithRoute(t, 1).Predict(graph, createPredictPayload(g))
	g.Expect(err).Should(BeNil())
	smRes = pResp.GetPayload().(*proto.SeldonMessage)
	g.Expect(smRes.GetData().GetNdarray().Values[0].GetNumberValue()).Should(Equal(1.1))
	g.Expect(smRes.GetData().GetNdarray().Values[1].GetNumberValue()).Should(Equal(2.0))

	pResp, err = createPredictorProcessWithRoute(t, -2).Predict(graph, createPredictPayload(g))
	g.Expect(err).Should(BeNil())
	smRes = pResp.GetPayload().(*proto.SeldonMessage)
	g.Expect(smRes.GetData().GetNdarray().Values[0].GetNumberValue()).Should(Equal(1.1))
	g.Expect(smRes.GetData().GetNdarray().Values[1].GetNumberValue()).Should(Equal(2.0))
}

func TestModelError(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	model := v1.MODEL
	graph := &v1.PredictiveUnit{
		Type: &model,
		Endpoint: &v1.Endpoint{
			ServiceHost: "foo",
			ServicePort: 9000,
			Type:        v1.REST,
		},
	}

	errMethod := v1.TRANSFORM_INPUT
	chosenErr := errors.New("something bad happened")
	errBytes := "{\"status\":\"failed\"}"
	errPayload := payload.BytesPayload{
		Msg:         []byte(errBytes),
		ContentType: "xyz",
	}
	pResp, err := createPredictorProcessWithError(t, &errMethod, chosenErr, &errPayload).Predict(graph, createPredictPayload(g))
	g.Expect(err).ShouldNot(BeNil())
	g.Expect(pResp).To(Equal(&errPayload))
	g.Expect(err.Error()).Should(Equal("something bad happened"))
}

func TestABTest(t *testing.T) {
	g := NewGomegaWithT(t)
	model := v1.MODEL
	abtest := v1.RANDOM_ABTEST
	graph := &v1.PredictiveUnit{
		Implementation: &abtest,
		Children: []v1.PredictiveUnit{
			{
				Type: &model,
				Endpoint: &v1.Endpoint{
					ServiceHost: "foo2",
					ServicePort: 9001,
					Type:        v1.REST,
				},
			},
			{
				Type: &model,
				Endpoint: &v1.Endpoint{
					ServiceHost: "foo3",
					ServicePort: 9002,
					Type:        v1.REST,
				},
			},
		},
	}

	pResp, err := createPredictorProcess(t).Predict(graph, createPredictPayload(g))
	g.Expect(err).Should(BeNil())
	smRes := pResp.GetPayload().(*proto.SeldonMessage)
	g.Expect(smRes.GetData().GetNdarray().Values[0].GetNumberValue()).Should(Equal(1.1))
	g.Expect(smRes.GetData().GetNdarray().Values[1].GetNumberValue()).Should(Equal(2.0))
}

func TestModelWithLogRequests(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	modelName := "foo"
	logged := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		g.Expect(r.Header.Get(logger.CloudEventsTypeHeader)).To(Equal(logger.CEInferenceRequest))
		g.Expect(r.Header.Get(logger.CloudEventsTypeSource)).To(Equal(testSourceUrl))
		g.Expect(r.Header.Get(modelIdHeaderName)).To(Equal(modelName))
		g.Expect(r.Header.Get(contentTypeHeaderName)).To(Equal(grpc.ProtobufContentType))
		g.Expect(r.Header.Get(requestIdHeaderName)).To(Equal(testSeldonPuid))
		w.Write([]byte(""))
		logged = true
		fmt.Printf("%+v\n", r.Header)
		fmt.Printf("%+v\n", r.Body)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	logf.SetLogger(logf.ZapLogger(false))
	log := logf.Log.WithName("entrypoint")
	logger.StartDispatcher(1, log, "", "", "")

	model := v1.MODEL
	graph := &v1.PredictiveUnit{
		Name: modelName,
		Type: &model,
		Endpoint: &v1.Endpoint{
			ServiceHost: "foo",
			ServicePort: 9000,
			Type:        v1.REST,
		},
		Logger: &v1.Logger{
			Mode: v1.LogRequest,
			Url:  &server.URL,
		},
	}

	pResp, err := createPredictorProcess(t).Predict(graph, createPredictPayload(g))
	g.Expect(err).Should(BeNil())
	smRes := pResp.GetPayload().(*proto.SeldonMessage)
	g.Expect(smRes.GetData().GetNdarray().Values[0].GetNumberValue()).Should(Equal(1.1))
	g.Expect(smRes.GetData().GetNdarray().Values[1].GetNumberValue()).Should(Equal(2.0))
	g.Eventually(func() bool { return logged }).Should(Equal(true))
}

func TestModelWithLogRequestsAtDefaultedUrl(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	modelName := "foo"
	logged := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		g.Expect(r.Header.Get(logger.CloudEventsTypeHeader)).To(Equal(logger.CEInferenceRequest))
		g.Expect(r.Header.Get(logger.CloudEventsTypeSource)).To(Equal(testSourceUrl))
		g.Expect(r.Header.Get(modelIdHeaderName)).To(Equal(modelName))
		g.Expect(r.Header.Get(contentTypeHeaderName)).To(Equal(grpc.ProtobufContentType))
		g.Expect(r.Header.Get(requestIdHeaderName)).To(Equal(testSeldonPuid))
		w.Write([]byte(""))
		logged = true
		fmt.Printf("%+v\n", r.Header)
		fmt.Printf("%+v\n", r.Body)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	envRequestLoggerDefaultEndpoint = server.URL

	logf.SetLogger(logf.ZapLogger(false))
	log := logf.Log.WithName("entrypoint")
	logger.StartDispatcher(1, log, "", "", "")

	model := v1.MODEL
	graph := &v1.PredictiveUnit{
		Name: modelName,
		Type: &model,
		Endpoint: &v1.Endpoint{
			ServiceHost: "foo",
			ServicePort: 9000,
			Type:        v1.REST,
		},
		Logger: &v1.Logger{
			Mode: v1.LogRequest,
		},
	}

	pResp, err := createPredictorProcess(t).Predict(graph, createPredictPayload(g))
	g.Expect(err).Should(BeNil())
	smRes := pResp.GetPayload().(*proto.SeldonMessage)
	g.Expect(smRes.GetData().GetNdarray().Values[0].GetNumberValue()).Should(Equal(1.1))
	g.Expect(smRes.GetData().GetNdarray().Values[1].GetNumberValue()).Should(Equal(2.0))
	g.Eventually(func() bool { return logged }).Should(Equal(true))
}

func TestModelWithLogResponses(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	modelName := "foo"
	logged := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		g.Expect(r.Header.Get(logger.CloudEventsTypeHeader)).To(Equal(logger.CEInferenceResponse))
		g.Expect(r.Header.Get(logger.CloudEventsTypeSource)).To(Equal(testSourceUrl))
		g.Expect(r.Header.Get(modelIdHeaderName)).To(Equal(modelName))
		g.Expect(r.Header.Get(contentTypeHeaderName)).To(Equal(grpc.ProtobufContentType))
		g.Expect(r.Header.Get(requestIdHeaderName)).To(Equal(testSeldonPuid))
		w.Write([]byte(""))
		logged = true
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	logf.SetLogger(logf.ZapLogger(false))
	log := logf.Log.WithName("entrypoint")
	logger.StartDispatcher(1, log, "", "", "")

	model := v1.MODEL
	graph := &v1.PredictiveUnit{
		Name: modelName,
		Type: &model,
		Endpoint: &v1.Endpoint{
			ServiceHost: "foo",
			ServicePort: 9000,
			Type:        v1.REST,
		},
		Logger: &v1.Logger{
			Mode: v1.LogResponse,
			Url:  &server.URL,
		},
	}

	pResp, err := createPredictorProcess(t).Predict(graph, createPredictPayload(g))
	g.Expect(err).Should(BeNil())
	smRes := pResp.GetPayload().(*proto.SeldonMessage)
	g.Expect(smRes.GetData().GetNdarray().Values[0].GetNumberValue()).Should(Equal(1.1))
	g.Expect(smRes.GetData().GetNdarray().Values[1].GetNumberValue()).Should(Equal(2.0))
	g.Eventually(func() bool { return logged }).Should(Equal(true))
}

func TestPredictNilPUIDError(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})
	server := httptest.NewServer(handler)
	defer server.Close()
	graph := &v1.PredictiveUnit{}
	_, err := createPredictorProcessWithoutPUIDInContext(t).Predict(graph, createPredictPayload(g))
	g.Expect(err).NotTo(BeNil())
	g.Expect(err.Error()).Should(Equal(NilPUIDError))
}
