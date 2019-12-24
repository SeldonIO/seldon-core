package api

import (
	"context"
	"github.com/golang/protobuf/jsonpb"
	"github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/executor/api/grpc/proto"
	"github.com/seldonio/seldon-core/executor/api/test"
	"github.com/seldonio/seldon-core/operator/apis/machinelearning/v1"
	"net/url"
	"testing"
)

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
	server := NewGrpcSeldonServer(&p, test.NewSeldonMessageTestClient(t, 0, nil, nil), url, "default")

	var sm proto.SeldonMessage
	var data = ` {"data":{"ndarray":[[1.1,2.0]]}}`
	err := jsonpb.UnmarshalString(data, &sm)
	g.Expect(err).Should(gomega.BeNil())

	res, err := server.Predict(context.TODO(), &sm)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(res.GetData().GetNdarray().Values[0].GetListValue().Values[0].GetNumberValue()).Should(gomega.Equal(1.1))
	g.Expect(res.GetData().GetNdarray().Values[0].GetListValue().Values[1].GetNumberValue()).Should(gomega.Equal(2.0))
}
