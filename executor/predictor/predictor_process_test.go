package predictor

import (
	"github.com/golang/protobuf/jsonpb"
	"github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/executor/api/client"
	api "github.com/seldonio/seldon-core/executor/api/grpc"
	"github.com/seldonio/seldon-core/executor/api/machinelearning/v1alpha2"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"testing"
)

type SeldonMessageTestClient struct {
	t *testing.T
}

func (s SeldonMessageTestClient) Predict(host string, port int32, msg client.SeldonPayload) (client.SeldonPayload, error) {
	s.t.Logf("Predict %s %d", host,port)
	return msg, nil
}

func (s SeldonMessageTestClient) TransformInput(host string, port int32, msg client.SeldonPayload) (client.SeldonPayload, error) {
	s.t.Logf("TransformInput %s %d", host,port)
	return msg, nil
}

func (s SeldonMessageTestClient) Route(host string, port int32, msg client.SeldonPayload) (int, error) {
	s.t.Logf("Route %s %d", host,port)
	return -1, nil
}

func (s SeldonMessageTestClient) Combine(host string, port int32, msgs []client.SeldonPayload) (client.SeldonPayload, error) {
	s.t.Logf("Combine %s %d", host,port)
	return msgs[0], nil
}

func (s SeldonMessageTestClient) TransformOutput(host string, port int32, msg client.SeldonPayload) (client.SeldonPayload, error) {
	s.t.Logf("TransformOutput %s %d", host,port)
	return msg, nil
}

func NewSeldonMessageTestClient(t *testing.T) client.SeldonApiClient {
	client := SeldonMessageTestClient{
		t:t,
	}
	return &client
}

func createPredictorProcess(t *testing.T) *PredictorProcess {
	return &PredictorProcess{
		NewSeldonMessageTestClient(t),
		logf.Log.WithName("SeldonMessageRestClient"),
	}
}

func createPayload(g *gomega.GomegaWithT) client.SeldonPayload {
	var sm api.SeldonMessage
	var data = ` {"data":{"ndarray":[1.1,2.0]}}`
	err := jsonpb.UnmarshalString(data, &sm)
	g.Expect(err).Should(gomega.BeNil())
	return &client.SeldonMessagePayload{&sm}
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
	smRes := pResp.GetPayload().(*api.SeldonMessage)
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
		Children: []v1alpha2.PredictiveUnit {
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
	smRes := pResp.GetPayload().(*api.SeldonMessage)
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
		Children: []v1alpha2.PredictiveUnit {
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
	smRes := pResp.GetPayload().(*api.SeldonMessage)
	g.Expect(smRes.GetData().GetNdarray().Values[0].GetNumberValue()).Should(gomega.Equal(1.1))
	g.Expect(smRes.GetData().GetNdarray().Values[1].GetNumberValue()).Should(gomega.Equal(2.0))
}


func TestTransformersInNode(t *testing.T) {
	t.Logf("Started")
	g := gomega.NewGomegaWithT(t)
	//model := v1alpha2.UNKNOWN_TYPE
	graph := &v1alpha2.PredictiveUnit{
		Methods: &[]v1alpha2.PredictiveUnitMethod{v1alpha2.TRANSFORM_INPUT,v1alpha2.TRANSFORM_OUTPUT},
		Endpoint: &v1alpha2.Endpoint{
			ServiceHost: "foo",
			ServicePort: 9000,
			Type:        v1alpha2.REST,
		},
	}

	pResp, err := createPredictorProcess(t).Execute(graph, createPayload(g))
	g.Expect(err).Should(gomega.BeNil())
	smRes := pResp.GetPayload().(*api.SeldonMessage)
	g.Expect(smRes.GetData().GetNdarray().Values[0].GetNumberValue()).Should(gomega.Equal(1.1))
	g.Expect(smRes.GetData().GetNdarray().Values[1].GetNumberValue()).Should(gomega.Equal(2.0))
}
