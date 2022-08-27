package kfserving

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/seldonio/seldon-core/executor/api/grpc/kfserving/inference"
	"github.com/seldonio/seldon-core/executor/api/payload"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func TestChain(t *testing.T) {
	g := NewGomegaWithT(t)

	var prevModelName = "foo"
	var nextModelName = "bar"

	tests := []struct {
		name     string
		msg      payload.SeldonPayload
		expected payload.SeldonPayload
	}{
		{
			name: "request should be returned as-is",
			msg: &payload.ProtoPayload{
				Msg: &inference.ModelInferRequest{
					ModelName: prevModelName,
					Inputs: []*inference.ModelInferRequest_InferInputTensor{
						{
							Name:     "input-1",
							Datatype: "INT32",
							Shape:    []int64{1},
							Contents: &inference.InferTensorContents{IntContents: []int32{1}},
						},
					},
				},
			},
			expected: &payload.ProtoPayload{
				Msg: &inference.ModelInferRequest{
					ModelName: prevModelName,
					Inputs: []*inference.ModelInferRequest_InferInputTensor{
						{
							Name:     "input-1",
							Datatype: "INT32",
							Shape:    []int64{1},
							Contents: &inference.InferTensorContents{IntContents: []int32{1}},
						},
					},
				},
			},
		},
		{
			name: "ensure that request's model name is always set",
			msg: &payload.ProtoPayload{
				Msg: &inference.ModelInferRequest{
					Inputs: []*inference.ModelInferRequest_InferInputTensor{
						{
							Name:     "input-1",
							Datatype: "INT32",
							Shape:    []int64{1},
							Contents: &inference.InferTensorContents{IntContents: []int32{1}},
						},
					},
				},
			},
			expected: &payload.ProtoPayload{
				Msg: &inference.ModelInferRequest{
					ModelName: nextModelName,
					Inputs: []*inference.ModelInferRequest_InferInputTensor{
						{
							Name:     "input-1",
							Datatype: "INT32",
							Shape:    []int64{1},
							Contents: &inference.InferTensorContents{IntContents: []int32{1}},
						},
					},
				},
			},
		},
		{
			name: "response should be chained",
			msg: &payload.ProtoPayload{
				Msg: &inference.ModelInferResponse{
					ModelName: prevModelName,
					Outputs: []*inference.ModelInferResponse_InferOutputTensor{
						{
							Name:     "input-1",
							Datatype: "INT32",
							Shape:    []int64{1},
							Contents: &inference.InferTensorContents{IntContents: []int32{1}},
						},
					},
					Parameters: map[string]*inference.InferParameter{
						"param-1-bool":   {ParameterChoice: &inference.InferParameter_BoolParam{BoolParam: true}},
						"param-1-int64":  {ParameterChoice: &inference.InferParameter_Int64Param{Int64Param: 42}},
						"param-1-string": {ParameterChoice: &inference.InferParameter_StringParam{StringParam: "param"}},
					},
				},
			},
			expected: &payload.ProtoPayload{
				Msg: &inference.ModelInferRequest{
					ModelName: nextModelName,
					Inputs: []*inference.ModelInferRequest_InferInputTensor{
						{
							Name:     "input-1",
							Datatype: "INT32",
							Shape:    []int64{1},
							Contents: &inference.InferTensorContents{IntContents: []int32{1}},
						},
					},
					Parameters: map[string]*inference.InferParameter{
						"param-1-bool":   {ParameterChoice: &inference.InferParameter_BoolParam{BoolParam: true}},
						"param-1-int64":  {ParameterChoice: &inference.InferParameter_Int64Param{Int64Param: 42}},
						"param-1-string": {ParameterChoice: &inference.InferParameter_StringParam{StringParam: "param"}},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &KFServingGrpcClient{Log: logf.Log.WithName("SeldonGrpcClient")}
			ctx := context.Background()

			chained, err := client.Chain(ctx, nextModelName, tt.msg)
			g.Expect(err).Should(BeNil())
			g.Expect(chained).Should(Equal(tt.expected))
		})
	}
}
