package seldon

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/executor/api/grpc"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/test"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/k8s"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning/v1"
	"net"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"testing"
)

func createPredictPayload(g *GomegaWithT) payload.SeldonPayload {
	var sm proto.SeldonMessage
	var data = ` {"data":{"ndarray":[1.1,2.0]}}`
	err := jsonpb.UnmarshalString(data, &sm)
	g.Expect(err).Should(BeNil())
	return &payload.ProtoPayload{Msg: &sm}
}

func createTestGrpcServer(g *GomegaWithT, annotations map[string]string) (*v1.PredictorSpec, string, int32, func()) {
	const port = 9000
	const host = "0.0.0.0"
	const deploymentName = "dep"
	const predictorName = "p"
	model := v1.MODEL
	p := v1.PredictorSpec{
		Name: predictorName,
		Graph: &v1.PredictiveUnit{
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: host,
				ServicePort: port,
				Type:        v1.REST,
			},
		},
	}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	g.Expect(err).To(BeNil())

	logger := logf.Log.WithName("entrypoint")
	grpcServer, err := grpc.CreateGrpcServer(&p, deploymentName, annotations, logger)
	g.Expect(err).To(BeNil())

	testSeldonGrpcServer := test.NewSeldonTestServer(1)
	proto.RegisterModelServer(grpcServer, testSeldonGrpcServer)

	go grpcServer.Serve(lis)
	stopFunc := grpcServer.Stop

	return &p, host, port, stopFunc
}

func TestClientPredict(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	p, host, port, stopFunc := createTestGrpcServer(g, nil)
	defer stopFunc()

	client := NewSeldonGrpcClient(p, "", nil)

	req := createPredictPayload(g)
	reqSm := req.GetPayload().(*proto.SeldonMessage)
	resp, err := client.Predict(context.TODO(), "m", host, port, req, nil)
	respSm := resp.GetPayload().(*proto.SeldonMessage)
	g.Expect(err).To(BeNil())
	g.Expect(respSm.GetData().GetNdarray().Values[0].GetNumberValue()).To(Equal(reqSm.GetData().GetNdarray().Values[0].GetNumberValue()))
}

func TestClientPredictTimeout(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	p, host, port, stopFunc := createTestGrpcServer(g, nil)
	defer stopFunc()

	annotations := map[string]string{k8s.ANNOTATION_GRPC_TIMEOUT: "100"}
	client := NewSeldonGrpcClient(p, "", annotations)

	req := createPredictPayload(g)
	_, err := client.Predict(context.TODO(), "m", host, port, req, nil)
	g.Expect(err).NotTo(BeNil())
	g.Expect(err.Error()).To(Equal("rpc error: code = DeadlineExceeded desc = context deadline exceeded"))
}

func TestClientPredictMessageSize(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	annotations := map[string]string{k8s.ANNOTATION_GRPC_MAX_MESSAGE_SIZE: "1"}
	p, host, port, stopFunc := createTestGrpcServer(g, annotations)
	defer stopFunc()

	client := NewSeldonGrpcClient(p, "", annotations)

	req := createPredictPayload(g)
	_, err := client.Predict(context.TODO(), "m", host, port, req, nil)
	g.Expect(err).NotTo(BeNil())
	g.Expect(err.Error()).To(Equal("rpc error: code = ResourceExhausted desc = grpc: received message larger than max (26 vs. 1)"))
}
