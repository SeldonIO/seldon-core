package pipeline

import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/grpc/credentials/insecure"

	. "github.com/onsi/gomega"
	v2 "github.com/seldonio/seldon-core/scheduler/apis/mlops/v2_dataplane"
	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/resources"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

func TestGrpcServer(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name   string
		header string
		req    *v2.ModelInferRequest
		res    *v2.ModelInferResponse
		error  bool
	}

	tests := []test{
		{
			name:   "ok",
			header: "foo",
			req: &v2.ModelInferRequest{
				Inputs: []*v2.ModelInferRequest_InferInputTensor{
					{
						Name:     "input1",
						Datatype: tyInt64,
						Shape:    []int64{5},
						Contents: &v2.InferTensorContents{Int64Contents: []int64{1, 2, 3, 4, 5}},
					},
				},
			},
			res: &v2.ModelInferResponse{
				ModelName:    "model",
				ModelVersion: "1",
				Id:           "1234",
				Outputs: []*v2.ModelInferResponse_InferOutputTensor{
					{
						Name:     "t1",
						Datatype: tyBool,
						Shape:    []int64{2},
						Contents: &v2.InferTensorContents{
							BoolContents: []bool{true, false},
						},
					},
				},
			},
		},
		{
			name:   "bad header",
			header: "",
			req: &v2.ModelInferRequest{
				Inputs: []*v2.ModelInferRequest_InferInputTensor{
					{
						Name:     "input1",
						Datatype: tyInt64,
						Shape:    []int64{5},
						Contents: &v2.InferTensorContents{Int64Contents: []int64{1, 2, 3, 4, 5}},
					},
				},
			},
		},
	}

	port, err := getFreePort()
	g.Expect(err).To(BeNil())
	mockInferer := &fakePipelineInferer{
		err:  nil,
		data: []byte("result"),
	}
	grpcServer := NewGatewayGrpcServer(port, logrus.New(), mockInferer)
	go func() {
		err := grpcServer.Start()
		g.Expect(err).To(BeNil())
	}()
	waitForServer(port)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b, err := proto.Marshal(test.res)
			g.Expect(err).To(BeNil())
			mockInferer := &fakePipelineInferer{
				err:  nil,
				data: b,
			}
			grpcServer.gateway = mockInferer
			opts := []grpc.DialOption{
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			}
			conn, err := grpc.Dial(fmt.Sprintf("0.0.0.0:%d", port), opts...)
			g.Expect(err).To(BeNil())
			client := v2.NewGRPCInferenceServiceClient(conn)
			ctx := context.TODO()
			ctx = metadata.AppendToOutgoingContext(ctx, resources.SeldonModelHeader, test.header)
			res, err := client.ModelInfer(ctx, test.req)
			if test.error {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(proto.Equal(res, test.res)).To(BeTrue())
			}
		})
	}
	grpcServer.Stop()
}
