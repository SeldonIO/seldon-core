package gateway

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/gomega"
	v2 "github.com/seldonio/seldon-core/scheduler/apis/mlops/v2_dataplane"
	"google.golang.org/protobuf/proto"
)

func TestGrpcChain(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name string
		res  *v2.ModelInferResponse
		req  *v2.ModelInferRequest
	}
	tests := []test{
		{
			name: "empty response test",
			res:  &v2.ModelInferResponse{},
			req:  &v2.ModelInferRequest{},
		},
		{
			name: "basic test",
			res: &v2.ModelInferResponse{
				Outputs: []*v2.ModelInferResponse_InferOutputTensor{
					{
						Name:     "out1",
						Datatype: "float",
						Shape:    []int64{1, 2, 3},
						Contents: &v2.InferTensorContents{
							IntContents: []int32{1, 2, 3},
						},
					},
				},
			},
			req: &v2.ModelInferRequest{
				Inputs: []*v2.ModelInferRequest_InferInputTensor{
					{
						Name:     "out1",
						Datatype: "float",
						Shape:    []int64{1, 2, 3},
						Contents: &v2.InferTensorContents{
							IntContents: []int32{1, 2, 3},
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := chainProtoResponseToRequest(test.res)
			g.Expect(proto.Equal(req, test.req)).To(BeTrue())
		})
	}
}

func TestJsonChain(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name string
		res  []byte
		req  []byte
	}
	tests := []test{
		{
			name: "bad json",
			res:  []byte(""),
			req:  []byte(""),
		},
		{
			name: "empty json",
			res:  []byte("{}"),
			req:  []byte("{}"),
		},
		{
			name: "no op",
			res:  []byte(`{"inputs":[{"data":[[1,2,3,4]],"datatype":"FP32","name":"predict","shape":[1,4]}]}`),
			req:  []byte(`{"inputs":[{"data":[[1,2,3,4]],"datatype":"FP32","name":"predict","shape":[1,4]}]}`),
		},
		{
			name: "v2 response",
			res:  []byte(`{"model_name":"iris_1","model_version":"1","id":"126a73e8-ba24-4681-8ec5-26282f8818fe","parameters":null,"outputs":[{"name":"predict","shape":[1],"datatype":"INT64","parameters":null,"data":[2]}]}`),
			req:  []byte(`{"model_name":"iris_1","model_version":"1","id":"126a73e8-ba24-4681-8ec5-26282f8818fe","parameters":null,"inputs":[{"name":"predict","shape":[1],"datatype":"INT64","parameters":null,"data":[2]}]}`),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := maybeChainRest(test.res)
			var f interface{}
			err := json.Unmarshal(test.req, &f)
			if err == nil {
				b, err := json.Marshal(f)
				g.Expect(err).To(BeNil())
				g.Expect(req).To(Equal(b))
			}
		})
	}
}

func TestGetProtoRequestAssumingResponse(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name  string
		data  []byte
		req   *v2.ModelInferRequest
		error bool
	}
	getProtoBytes := func(res proto.Message) []byte {
		b, _ := proto.Marshal(res)
		return b
	}
	tests := []test{
		{
			name: "empty response test",
			data: getProtoBytes(&v2.ModelInferResponse{}),
			req:  &v2.ModelInferRequest{},
		},
		{
			name: "wrong empty type will succeed",
			data: getProtoBytes(&v2.ModelInferRequest{}),
			req:  &v2.ModelInferRequest{},
		},
		{
			name: "wrong type but will succeed", //Maybe unexpected as its transferred Inputs to Outputs back to Inputs
			data: getProtoBytes(&v2.ModelInferRequest{
				Inputs: []*v2.ModelInferRequest_InferInputTensor{
					{
						Name:     "out1",
						Datatype: "float",
						Shape:    []int64{1, 2, 3},
						Contents: &v2.InferTensorContents{
							IntContents: []int32{1, 2, 3},
						},
					},
				},
			}),
			req: &v2.ModelInferRequest{
				Inputs: []*v2.ModelInferRequest_InferInputTensor{
					{
						Name:     "out1",
						Datatype: "float",
						Shape:    []int64{1, 2, 3},
						Contents: &v2.InferTensorContents{
							IntContents: []int32{1, 2, 3},
						},
					},
				},
			},
		},
		{
			name: "basic test",
			data: getProtoBytes(&v2.ModelInferResponse{
				Outputs: []*v2.ModelInferResponse_InferOutputTensor{
					{
						Name:     "out1",
						Datatype: "float",
						Shape:    []int64{1, 2, 3},
						Contents: &v2.InferTensorContents{
							IntContents: []int32{1, 2, 3},
						},
					},
				},
			}),
			req: &v2.ModelInferRequest{
				Inputs: []*v2.ModelInferRequest_InferInputTensor{
					{
						Name:     "out1",
						Datatype: "float",
						Shape:    []int64{1, 2, 3},
						Contents: &v2.InferTensorContents{
							IntContents: []int32{1, 2, 3},
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := getProtoRequestAssumingResponse(test.data)
			if test.error {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(proto.Equal(req, test.req)).To(BeTrue())
			}

		})
	}
}
