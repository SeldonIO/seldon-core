package seldon

import (
	"context"
	"github.com/golang/protobuf/jsonpb"
	empty "github.com/golang/protobuf/ptypes/empty"
	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/api/test"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"net/url"
	"testing"
)

const testSeldonPuid = "1"

func TestPredict(t *testing.T) {
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
	server := NewGrpcSeldonServer(&p, &test.SeldonMessageTestClient{}, url, "default")

	var sm proto.SeldonMessage
	var data = ` {"data":{"ndarray":[[1.1,2.0]]}}`
	err := jsonpb.UnmarshalString(data, &sm)
	g.Expect(err).Should(BeNil())

	res, err := server.Predict(context.TODO(), &sm)
	g.Expect(err).To(BeNil())
	g.Expect(res.GetData().GetNdarray().Values[0].GetListValue().Values[0].GetNumberValue()).Should(Equal(1.1))
	g.Expect(res.GetData().GetNdarray().Values[0].GetListValue().Values[1].GetNumberValue()).Should(Equal(2.0))
}

func TestFeedback(t *testing.T) {
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
	server := NewGrpcSeldonServer(&p, &test.SeldonMessageTestClient{}, url, "default")

	var sm proto.Feedback
	var data = ` {"request":{"data":{"ndarray":[[1.1,2.0]]}}}`
	err := jsonpb.UnmarshalString(data, &sm)
	g.Expect(err).Should(BeNil())

	res, err := server.SendFeedback(context.TODO(), &sm)
	g.Expect(err).To(BeNil())
	g.Expect(res.GetData().GetNdarray().Values[0].GetListValue().Values[0].GetNumberValue()).Should(Equal(1.1))
	g.Expect(res.GetData().GetNdarray().Values[0].GetListValue().Values[1].GetNumberValue()).Should(Equal(2.0))
}

func TestMetadata(t *testing.T) {
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

	protoMetadata := proto.SeldonModelMetadata{Name: "mymodel"}

	metadataPayload := payload.ProtoPayload{Msg: &protoMetadata}
	server := NewGrpcSeldonServer(&p, &test.SeldonMessageTestClient{MetadataResponse: &metadataPayload}, url, "default")

	res, err := server.ModelMetadata(context.TODO(), &proto.SeldonModelMetadataRequest{})
	g.Expect(err).To(BeNil())
	g.Expect(res.GetName()).To(Equal("mymodel"))
}

func TestGraphMetadata(t *testing.T) {
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
			Inputs: []*proto.SeldonMessageMetadata{
				{Name: "input-model-1"},
			},
			Outputs: []*proto.SeldonMessageMetadata{
				{Name: "output-model-1"},
			},
		},
		"model-2": {
			Name:     "model-2",
			Platform: "platform-name",
			Versions: []string{"model-version"},
			Inputs: []*proto.SeldonMessageMetadata{
				{Name: "input-model-2"},
			},
			Outputs: []*proto.SeldonMessageMetadata{
				{Name: "output-model-2"},
			},
		},
	}

	url, _ := url.Parse("http://localhost")

	server := NewGrpcSeldonServer(&p, &test.SeldonMessageTestClient{ModelMetadataMap: metadataMap}, url, "default")

	res, err := server.GraphMetadata(context.TODO(), &empty.Empty{})
	g.Expect(err).To(BeNil())
	g.Expect(res.GetName()).To(Equal("predictor-name"))
	g.Expect(res.GetInputs()).To(Equal(metadataMap["model-1"].Inputs))
	g.Expect(res.GetOutputs()).To(Equal(metadataMap["model-2"].Outputs))

	for name, modelMetadata := range metadataMap {
		protoMetadata := &proto.SeldonModelMetadata{
			Name:     modelMetadata.Name,
			Versions: modelMetadata.Versions,
			Platform: modelMetadata.Platform,
			Inputs:   modelMetadata.Inputs.([]*proto.SeldonMessageMetadata),
			Outputs:  modelMetadata.Outputs.([]*proto.SeldonMessageMetadata),
		}
		g.Expect(res.GetModels()[name]).To(Equal(protoMetadata))
	}

}
