package pipeline

import (
	"context"
	"fmt"
	"net/http"
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

type fakePipelineMetricsHandler struct{}

func (f fakePipelineMetricsHandler) AddPipelineHistogramMetricsHandler(baseHandler http.HandlerFunc) http.HandlerFunc {
	return baseHandler
}

func (f fakePipelineMetricsHandler) AddPipelineInferMetrics(pipelineName string, method string, elapsedTime float64, code string) {
}

func (f fakePipelineMetricsHandler) PipelineUnaryServerInterceptor() func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		return handler(ctx, req)
	}
}

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
			error: true,
		},
	}

	testRequestId := "test-id"
	port, err := getFreePort()
	g.Expect(err).To(BeNil())
	mockInferer := &fakePipelineInferer{
		err:  nil,
		data: []byte("result"),
		key:  testRequestId,
	}
	grpcServer := NewGatewayGrpcServer(port, logrus.New(), mockInferer, fakePipelineMetricsHandler{})
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
				key:  testRequestId,
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
			var header, trailer metadata.MD
			res, err := client.ModelInfer(ctx, test.req, grpc.Header(&header), grpc.Trailer(&trailer))
			if test.error {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(proto.Equal(res, test.res)).To(BeTrue())
				g.Expect(header.Get(RequestIdHeader)).ToNot(BeNil())
				g.Expect(header.Get(RequestIdHeader)[0]).To(Equal(testRequestId))
			}
		})
	}
	grpcServer.Stop()
}
