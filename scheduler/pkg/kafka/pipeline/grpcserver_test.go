/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pipeline

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"

	v2 "github.com/seldonio/seldon-core/apis/go/v2/mlops/v2_dataplane"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/envoy/resources"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/internal/testing_utils"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
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
	port, err := testing_utils.GetFreePortForTest()
	g.Expect(err).To(BeNil())
	mockInferer := &fakePipelineInferer{
		err:  nil,
		data: []byte("result"),
		key:  testRequestId,
	}
	grpcServer := NewGatewayGrpcServer(port, logrus.New(), mockInferer, fakePipelineMetricsHandler{}, &util.TLSOptions{}, nil)
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
				g.Expect(header.Get(util.RequestIdHeader)).ToNot(BeNil())
				g.Expect(header.Get(util.RequestIdHeader)[0]).To(Equal(testRequestId))
			}
		})
	}
	grpcServer.Stop()
}
