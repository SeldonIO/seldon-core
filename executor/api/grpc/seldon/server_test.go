package seldon

import (
	"context"
	"github.com/golang/protobuf/jsonpb"
	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"github.com/seldonio/seldon-core/executor/api/test"
	"github.com/seldonio/seldon-core/operator/apis/machinelearning/v1"
	"net/url"
	"testing"
)

func TestPredict(t *testing.T) {
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
