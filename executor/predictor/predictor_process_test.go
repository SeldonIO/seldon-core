package predictor

import (
	"errors"
	"github.com/golang/protobuf/jsonpb"
	"github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/executor/api/grpc/proto"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/api/test"
	"github.com/seldonio/seldon-core/executor/logger"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning/v1"
	"net/http"
	"net/http/httptest"
	"net/url"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"testing"
)

const (
	testEventId   = "1"
	testSourceUrl = "http://localhost"
)

func createPredictorProcess(t *testing.T) *PredictorProcess {
	url, _ := url.Parse(testSourceUrl)
	pp := NewPredictorProcess(test.NewSeldonMessageTestClient(t, -1, nil, nil), logf.Log.WithName("SeldonMessageRestClient"), testEventId, url, "default")
	return &pp
}

func createPredictorProcessWithRoute(t *testing.T, chosenRoute int) *PredictorProcess {
	url, _ := url.Parse(testSourceUrl)
	pp := NewPredictorProcess(test.NewSeldonMessageTestClient(t, chosenRoute, nil, nil), logf.Log.WithName("SeldonMessageRestClient"), testEventId, url, "default")
	return &pp
}

func createPredictorProcessWithError(t *testing.T, errMethod *v1.PredictiveUnitMethod, err error) *PredictorProcess {
	url, _ := url.Parse(testSourceUrl)
	pp := NewPredictorProcess(test.NewSeldonMessageTestClient(t, -1, errMethod, err), logf.Log.WithName("SeldonMessageRestClient"), testEventId, url, "default")
	return &pp
}

func createPayload(g *gomega.GomegaWithT) payload.SeldonPayload {
	var sm proto.SeldonMessage
	var data = ` {"data":{"ndarray":[1.1,2.0]}}`
	err := jsonpb.UnmarshalString(data, &sm)
	g.Expect(err).Should(gomega.BeNil())
	return &payload.SeldonMessagePayload{Msg: &sm, ContentType: "application/json"}
}

func TestModel(t *testing.T) {
	t.Logf("Started")
	g := gomega.NewGomegaWithT(t)
	model := v1.MODEL
	graph := &v1.PredictiveUnit{
		Type: &model,
		Endpoint: &v1.Endpoint{
			ServiceHost: "foo",
			ServicePort: 9000,
			Type:        v1.REST,
		},
	}

	pResp, err := createPredictorProcess(t).Execute(graph, createPayload(g))
	g.Expect(err).Should(gomega.BeNil())
	smRes := pResp.GetPayload().(*proto.SeldonMessage)
	g.Expect(smRes.GetData().GetNdarray().Values[0].GetNumberValue()).Should(gomega.Equal(1.1))
	g.Expect(smRes.GetData().GetNdarray().Values[1].GetNumberValue()).Should(gomega.Equal(2.0))
}

func TestTwoLevelModel(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
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

	pResp, err := createPredictorProcess(t).Execute(graph, createPayload(g))
	g.Expect(err).Should(gomega.BeNil())
	smRes := pResp.GetPayload().(*proto.SeldonMessage)
	g.Expect(smRes.GetData().GetNdarray().Values[0].GetNumberValue()).Should(gomega.Equal(1.1))
	g.Expect(smRes.GetData().GetNdarray().Values[1].GetNumberValue()).Should(gomega.Equal(2.0))
}

func TestCombiner(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
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

	pResp, err := createPredictorProcess(t).Execute(graph, createPayload(g))
	g.Expect(err).Should(gomega.BeNil())
	smRes := pResp.GetPayload().(*proto.SeldonMessage)
	g.Expect(smRes.GetData().GetNdarray().Values[0].GetNumberValue()).Should(gomega.Equal(1.1))
	g.Expect(smRes.GetData().GetNdarray().Values[1].GetNumberValue()).Should(gomega.Equal(2.0))
}

func TestMethods(t *testing.T) {
	t.Logf("Started")
	g := gomega.NewGomegaWithT(t)
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

	pResp, err := createPredictorProcess(t).Execute(graph, createPayload(g))
	g.Expect(err).Should(gomega.BeNil())
	smRes := pResp.GetPayload().(*proto.SeldonMessage)
	g.Expect(smRes.GetData().GetNdarray().Values[0].GetNumberValue()).Should(gomega.Equal(1.1))
	g.Expect(smRes.GetData().GetNdarray().Values[1].GetNumberValue()).Should(gomega.Equal(2.0))
}

func TestRouter(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
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

	pResp, err := createPredictorProcess(t).Execute(graph, createPayload(g))
	g.Expect(err).Should(gomega.BeNil())
	smRes := pResp.GetPayload().(*proto.SeldonMessage)
	g.Expect(smRes.GetData().GetNdarray().Values[0].GetNumberValue()).Should(gomega.Equal(1.1))
	g.Expect(smRes.GetData().GetNdarray().Values[1].GetNumberValue()).Should(gomega.Equal(2.0))

	pResp, err = createPredictorProcessWithRoute(t, 0).Execute(graph, createPayload(g))
	g.Expect(err).Should(gomega.BeNil())
	smRes = pResp.GetPayload().(*proto.SeldonMessage)
	g.Expect(smRes.GetData().GetNdarray().Values[0].GetNumberValue()).Should(gomega.Equal(1.1))
	g.Expect(smRes.GetData().GetNdarray().Values[1].GetNumberValue()).Should(gomega.Equal(2.0))

	pResp, err = createPredictorProcessWithRoute(t, 1).Execute(graph, createPayload(g))
	g.Expect(err).Should(gomega.BeNil())
	smRes = pResp.GetPayload().(*proto.SeldonMessage)
	g.Expect(smRes.GetData().GetNdarray().Values[0].GetNumberValue()).Should(gomega.Equal(1.1))
	g.Expect(smRes.GetData().GetNdarray().Values[1].GetNumberValue()).Should(gomega.Equal(2.0))
}

func TestModelError(t *testing.T) {
	t.Logf("Started")
	g := gomega.NewGomegaWithT(t)
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
	pResp, err := createPredictorProcessWithError(t, &errMethod, chosenErr).Execute(graph, createPayload(g))
	g.Expect(err).ShouldNot(gomega.BeNil())
	g.Expect(pResp).Should(gomega.BeNil())
	g.Expect(err.Error()).Should(gomega.Equal("something bad happened"))
}

func TestABTest(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
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

	pResp, err := createPredictorProcess(t).Execute(graph, createPayload(g))
	g.Expect(err).Should(gomega.BeNil())
	smRes := pResp.GetPayload().(*proto.SeldonMessage)
	g.Expect(smRes.GetData().GetNdarray().Values[0].GetNumberValue()).Should(gomega.Equal(1.1))
	g.Expect(smRes.GetData().GetNdarray().Values[1].GetNumberValue()).Should(gomega.Equal(2.0))
}

func TestModelWithLogRequests(t *testing.T) {
	t.Logf("Started")
	g := gomega.NewGomegaWithT(t)
	modelName := "foo"
	logged := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		g.Expect(r.Header.Get(logger.CloudEventsIdHeader)).Should(gomega.Equal(testEventId))
		g.Expect(r.Header.Get(logger.CloudEventsTypeHeader)).Should(gomega.Equal(logger.CEInferenceRequest))
		g.Expect(r.Header.Get(logger.CloudEventsTypeSource)).Should(gomega.Equal(testSourceUrl))
		g.Expect(r.Header.Get(logger.ModelIdHeader)).Should(gomega.Equal(modelName))
		g.Expect(r.Header.Get("Content-Type")).Should(gomega.Equal("application/json"))
		w.Write([]byte(""))
		logged = true
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	logf.SetLogger(logf.ZapLogger(false))
	log := logf.Log.WithName("entrypoint")
	logger.StartDispatcher(1, log)

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

	pResp, err := createPredictorProcess(t).Execute(graph, createPayload(g))
	g.Expect(err).Should(gomega.BeNil())
	smRes := pResp.GetPayload().(*proto.SeldonMessage)
	g.Expect(smRes.GetData().GetNdarray().Values[0].GetNumberValue()).Should(gomega.Equal(1.1))
	g.Expect(smRes.GetData().GetNdarray().Values[1].GetNumberValue()).Should(gomega.Equal(2.0))
	g.Eventually(func() bool { return logged }).Should(gomega.Equal(true))
}
