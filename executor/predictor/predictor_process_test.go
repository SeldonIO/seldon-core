package predictor

import (
	"errors"
	"github.com/golang/protobuf/jsonpb"
	"github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/executor/api/grpc/proto"
	"github.com/seldonio/seldon-core/executor/api/machinelearning/v1alpha2"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/api/test"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"testing"
)

func createPredictorProcess(t *testing.T) *PredictorProcess {
	return &PredictorProcess{
		Client: test.NewSeldonMessageTestClient(t, -1, nil, nil),
		Log:    logf.Log.WithName("SeldonMessageRestClient"),
	}
}

func createPredictorProcessWithRoute(t *testing.T, chosenRoute int) *PredictorProcess {
	return &PredictorProcess{
		Client: test.NewSeldonMessageTestClient(t, chosenRoute, nil, nil),
		Log:    logf.Log.WithName("SeldonMessageRestClient"),
	}
}

func createPredictorProcessWithError(t *testing.T, errMethod *v1alpha2.PredictiveUnitMethod, err error) *PredictorProcess {
	return &PredictorProcess{
		Client: test.NewSeldonMessageTestClient(t, -1, errMethod, err),
		Log:    logf.Log.WithName("SeldonMessageRestClient"),
	}
}

func createPayload(g *gomega.GomegaWithT) payload.SeldonPayload {
	var sm proto.SeldonMessage
	var data = ` {"data":{"ndarray":[1.1,2.0]}}`
	err := jsonpb.UnmarshalString(data, &sm)
	g.Expect(err).Should(gomega.BeNil())
	return &payload.SeldonMessagePayload{Msg: &sm}
}

func TestModel(t *testing.T) {
	t.Logf("Started")
	g := gomega.NewGomegaWithT(t)
	model := v1alpha2.MODEL
	graph := &v1alpha2.PredictiveUnit{
		Type: &model,
		Endpoint: &v1alpha2.Endpoint{
			ServiceHost: "foo",
			ServicePort: 9000,
			Type:        v1alpha2.REST,
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
	model := v1alpha2.MODEL
	graph := &v1alpha2.PredictiveUnit{
		Type: &model,
		Endpoint: &v1alpha2.Endpoint{
			ServiceHost: "foo",
			ServicePort: 9000,
			Type:        v1alpha2.REST,
		},
		Children: []v1alpha2.PredictiveUnit{
			{
				Type: &model,
				Endpoint: &v1alpha2.Endpoint{
					ServiceHost: "foo2",
					ServicePort: 9001,
					Type:        v1alpha2.REST,
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
	model := v1alpha2.MODEL
	combiner := v1alpha2.COMBINER
	graph := &v1alpha2.PredictiveUnit{
		Type: &combiner,
		Endpoint: &v1alpha2.Endpoint{
			ServiceHost: "foo",
			ServicePort: 9000,
			Type:        v1alpha2.REST,
		},
		Children: []v1alpha2.PredictiveUnit{
			{
				Type: &model,
				Endpoint: &v1alpha2.Endpoint{
					ServiceHost: "foo2",
					ServicePort: 9001,
					Type:        v1alpha2.REST,
				},
			},
			{
				Type: &model,
				Endpoint: &v1alpha2.Endpoint{
					ServiceHost: "foo3",
					ServicePort: 9002,
					Type:        v1alpha2.REST,
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
	//model := v1alpha2.UNKNOWN_TYPE
	graph := &v1alpha2.PredictiveUnit{
		Methods: &[]v1alpha2.PredictiveUnitMethod{v1alpha2.TRANSFORM_INPUT, v1alpha2.TRANSFORM_OUTPUT, v1alpha2.ROUTE, v1alpha2.AGGREGATE},
		Endpoint: &v1alpha2.Endpoint{
			ServiceHost: "foo",
			ServicePort: 9000,
			Type:        v1alpha2.REST,
		},
		Children: []v1alpha2.PredictiveUnit{
			{
				Methods: &[]v1alpha2.PredictiveUnitMethod{v1alpha2.TRANSFORM_INPUT, v1alpha2.TRANSFORM_OUTPUT},
				Endpoint: &v1alpha2.Endpoint{
					ServiceHost: "foo2",
					ServicePort: 9001,
					Type:        v1alpha2.REST,
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
	model := v1alpha2.MODEL
	router := v1alpha2.ROUTER
	graph := &v1alpha2.PredictiveUnit{
		Type: &router,
		Endpoint: &v1alpha2.Endpoint{
			ServiceHost: "foo",
			ServicePort: 9000,
			Type:        v1alpha2.REST,
		},
		Children: []v1alpha2.PredictiveUnit{
			{
				Type: &model,
				Endpoint: &v1alpha2.Endpoint{
					ServiceHost: "foo2",
					ServicePort: 9001,
					Type:        v1alpha2.REST,
				},
			},
			{
				Type: &model,
				Endpoint: &v1alpha2.Endpoint{
					ServiceHost: "foo3",
					ServicePort: 9002,
					Type:        v1alpha2.REST,
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
	model := v1alpha2.MODEL
	graph := &v1alpha2.PredictiveUnit{
		Type: &model,
		Endpoint: &v1alpha2.Endpoint{
			ServiceHost: "foo",
			ServicePort: 9000,
			Type:        v1alpha2.REST,
		},
	}

	errMethod := v1alpha2.TRANSFORM_INPUT
	chosenErr := errors.New("something bad happened")
	pResp, err := createPredictorProcessWithError(t, &errMethod, chosenErr).Execute(graph, createPayload(g))
	g.Expect(err).ShouldNot(gomega.BeNil())
	g.Expect(pResp).Should(gomega.BeNil())
	g.Expect(err.Error()).Should(gomega.Equal("something bad happened"))
}

func TestABTest(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	model := v1alpha2.MODEL
	abtest := v1alpha2.RANDOM_ABTEST
	graph := &v1alpha2.PredictiveUnit{
		Implementation: &abtest,
		Children: []v1alpha2.PredictiveUnit{
			{
				Type: &model,
				Endpoint: &v1alpha2.Endpoint{
					ServiceHost: "foo2",
					ServicePort: 9001,
					Type:        v1alpha2.REST,
				},
			},
			{
				Type: &model,
				Endpoint: &v1alpha2.Endpoint{
					ServiceHost: "foo3",
					ServicePort: 9002,
					Type:        v1alpha2.REST,
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
