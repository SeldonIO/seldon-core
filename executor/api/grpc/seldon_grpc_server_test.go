package api

import (
	"github.com/golang/protobuf/jsonpb"
	"github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/executor/api/grpc/proto"
	"github.com/seldonio/seldon-core/executor/api/machinelearning/v1alpha2"
	"github.com/seldonio/seldon-core/executor/api/test"
	"testing"
)

func TestSimpleModel(t *testing.T) {
	t.Logf("Started")
	g := gomega.NewGomegaWithT(t)

	model := v1alpha2.MODEL
	p := v1alpha2.PredictorSpec{
		Name: "p",
		Graph: &v1alpha2.PredictiveUnit{
			Type: &model,
			Endpoint: &v1alpha2.Endpoint{
				ServiceHost: "foo",
				ServicePort: 9000,
				Type:        v1alpha2.REST,
			},
		},
	}
	server := NewGrpcSeldonServer(&p, test.NewSeldonMessageTestClient(t, 0, nil, nil))

	var sm proto.SeldonMessage
	var data = ` {"data":{"ndarray":[[1.1,2.0]]}}`
	err := jsonpb.UnmarshalString(data, &sm)
	g.Expect(err).Should(gomega.BeNil())

	res, err := server.Predict(nil, &sm)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(res.GetData().GetNdarray().Values[0].GetListValue().Values[0].GetNumberValue()).Should(gomega.Equal(1.1))
	g.Expect(res.GetData().GetNdarray().Values[0].GetListValue().Values[1].GetNumberValue()).Should(gomega.Equal(2.0))
}
