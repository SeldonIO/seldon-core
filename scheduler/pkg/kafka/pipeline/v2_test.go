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
	"encoding/binary"
	"encoding/json"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/v2_dataplane"
)

func TestRequestInputParametersToV2(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name     string
		input    string
		expected *v2_dataplane.InferParameter
		error    bool
	}
	tests := []test{
		{
			name:  "bool",
			input: `{"parameters":{"foo":true}}`,
			expected: &v2_dataplane.InferParameter{
				ParameterChoice: &v2_dataplane.InferParameter_BoolParam{BoolParam: true},
			},
		},
		{
			name:  "int64",
			input: `{"parameters":{"foo":3}}`,
			expected: &v2_dataplane.InferParameter{
				ParameterChoice: &v2_dataplane.InferParameter_Int64Param{Int64Param: 3},
			},
		},
		{
			name:  "float64",
			input: `{"parameters":{"foo":3.3}}`,
			expected: &v2_dataplane.InferParameter{
				ParameterChoice: &v2_dataplane.InferParameter_Int64Param{Int64Param: 3},
			},
		},
		{
			name:  "string",
			input: `{"parameters":{"foo":"bar"}}`,
			expected: &v2_dataplane.InferParameter{
				ParameterChoice: &v2_dataplane.InferParameter_StringParam{StringParam: "bar"},
			},
		},
		{
			name:  "invalid",
			input: `{"parameters":{"foo":{"bar":2}}}`,
			error: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := &NamedTensor{}
			err := json.Unmarshal([]byte(test.input), req)
			g.Expect(err).To(BeNil())
			for _, val := range req.Parameters {
				param, err := requestInputParametersToV2(val)
				if test.error {
					g.Expect(err).ToNot(BeNil())
				} else {
					g.Expect(param).To(Equal(test.expected))
				}
			}
		})
	}
}

func TestConvertToInferenceRequest(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name     string
		input    string
		expected *InferenceRequest
	}
	tests := []test{
		{
			name:  "bool",
			input: `{"inputs":[{"name":"input1","datatype":"BOOL","shape":[5],"data":[true,false,true,false,true]}]}`,
			expected: &InferenceRequest{
				Inputs: []*NamedTensor{
					{
						Name:     "input1",
						Datatype: tyBool,
						Shape:    []int64{5},
						tensorData: &TensorData{
							boolContents: []bool{true, false, true, false, true},
						},
					},
				},
			},
		},
		{
			name:  "uint8",
			input: `{"inputs":[{"name":"input1","datatype":"UINT8","shape":[5],"data":[1,2,3,4,5]}]}`,
			expected: &InferenceRequest{
				Inputs: []*NamedTensor{
					{
						Name:     "input1",
						Datatype: tyUint8,
						Shape:    []int64{5},
						tensorData: &TensorData{
							uint32Contents: []uint32{1, 2, 3, 4, 5},
						},
					},
				},
			},
		},
		{
			name:  "uint16",
			input: `{"inputs":[{"name":"input1","datatype":"UINT16","shape":[5],"data":[1,2,3,4,5]}]}`,
			expected: &InferenceRequest{
				Inputs: []*NamedTensor{
					{
						Name:     "input1",
						Datatype: tyUint16,
						Shape:    []int64{5},
						tensorData: &TensorData{
							uint32Contents: []uint32{1, 2, 3, 4, 5},
						},
					},
				},
			},
		},
		{
			name:  "uint32",
			input: `{"inputs":[{"name":"input1","datatype":"UINT32","shape":[5],"data":[1,2,3,4,5]}]}`,
			expected: &InferenceRequest{
				Inputs: []*NamedTensor{
					{
						Name:     "input1",
						Datatype: tyUint32,
						Shape:    []int64{5},
						tensorData: &TensorData{
							uint32Contents: []uint32{1, 2, 3, 4, 5},
						},
					},
				},
			},
		},
		{
			name:  "uint64",
			input: `{"inputs":[{"name":"input1","datatype":"UINT64","shape":[5],"data":[1,2,3,4,5]}]}`,
			expected: &InferenceRequest{
				Inputs: []*NamedTensor{
					{
						Name:     "input1",
						Datatype: tyUint64,
						Shape:    []int64{5},
						tensorData: &TensorData{
							uint64Contents: []uint64{1, 2, 3, 4, 5},
						},
					},
				},
			},
		},
		{
			name:  "int8",
			input: `{"inputs":[{"name":"input1","datatype":"INT8","shape":[5],"data":[1,2,3,4,5]}]}`,
			expected: &InferenceRequest{
				Inputs: []*NamedTensor{
					{
						Name:     "input1",
						Datatype: tyInt8,
						Shape:    []int64{5},
						tensorData: &TensorData{
							int32Contents: []int32{1, 2, 3, 4, 5},
						},
					},
				},
			},
		},
		{
			name:  "int16",
			input: `{"inputs":[{"name":"input1","datatype":"INT16","shape":[5],"data":[1,2,3,4,5]}]}`,
			expected: &InferenceRequest{
				Inputs: []*NamedTensor{
					{
						Name:     "input1",
						Datatype: tyInt16,
						Shape:    []int64{5},
						tensorData: &TensorData{
							int32Contents: []int32{1, 2, 3, 4, 5},
						},
					},
				},
			},
		},
		{
			name:  "int32",
			input: `{"inputs":[{"name":"input1","datatype":"INT32","shape":[5],"data":[1,2,3,4,5]}]}`,
			expected: &InferenceRequest{
				Inputs: []*NamedTensor{
					{
						Name:     "input1",
						Datatype: tyInt32,
						Shape:    []int64{5},
						tensorData: &TensorData{
							int32Contents: []int32{1, 2, 3, 4, 5},
						},
					},
				},
			},
		},
		{
			name:  "int64",
			input: `{"inputs":[{"name":"input1","datatype":"INT64","shape":[5],"data":[1,2,3,4,5]}]}`,
			expected: &InferenceRequest{
				Inputs: []*NamedTensor{
					{
						Name:     "input1",
						Datatype: tyInt64,
						Shape:    []int64{5},
						tensorData: &TensorData{
							int64Contents: []int64{1, 2, 3, 4, 5},
						},
					},
				},
			},
		},
		{
			name:  "float16",
			input: `{"inputs":[{"name":"input1","datatype":"FP16","shape":[5],"data":[1.1,2.2,3.3,4.4,5.5]}]}`,
			expected: &InferenceRequest{
				Inputs: []*NamedTensor{
					{
						Name:     "input1",
						Datatype: tyFp16,
						Shape:    []int64{5},
						tensorData: &TensorData{
							fp32Contents: []float32{1.1, 2.2, 3.3, 4.4, 5.5},
						},
					},
				},
			},
		},
		{
			name:  "float32",
			input: `{"inputs":[{"name":"input1","datatype":"FP32","shape":[5],"data":[1.1,2.2,3.3,4.4,5.5]}]}`,
			expected: &InferenceRequest{
				Inputs: []*NamedTensor{
					{
						Name:     "input1",
						Datatype: tyFp32,
						Shape:    []int64{5},
						tensorData: &TensorData{
							fp32Contents: []float32{1.1, 2.2, 3.3, 4.4, 5.5},
						},
					},
				},
			},
		},
		{
			name:  "float64",
			input: `{"inputs":[{"name":"input1","datatype":"FP64","shape":[5],"data":[1.1,2.2,3.3,4.4,5.5]}]}`,
			expected: &InferenceRequest{
				Inputs: []*NamedTensor{
					{
						Name:     "input1",
						Datatype: tyFp64,
						Shape:    []int64{5},
						tensorData: &TensorData{
							fp64Contents: []float64{1.1, 2.2, 3.3, 4.4, 5.5},
						},
					},
				},
			},
		},
		{
			name:  "bytes",
			input: `{"inputs":[{"name":"input1","datatype":"BYTES","shape":[1,10],"data":["test"]}]}`,
			expected: &InferenceRequest{
				Inputs: []*NamedTensor{
					{
						Name:     "input1",
						Datatype: tyBytes,
						Shape:    []int64{1, 10},
						tensorData: &TensorData{
							byteContents: [][]byte{[]byte("test")},
						},
					},
				},
			},
		},
		{
			name: "parameters",
			input: `{"id": "53c3f354-6f29-415d-9128-1d15978318e8", "parameters": {"content_type": "str", "headers": null}, "inputs": [{"name": "predict", "shape": [1], "datatype": "BYTES", "parameters": {"content_type": "str", "headers": null}, "data": ["Hello world"]}], "outputs": []}
`,
			expected: &InferenceRequest{
				Id:         "53c3f354-6f29-415d-9128-1d15978318e8",
				Parameters: map[string]interface{}{"content_type": "str", "headers": nil},
				Inputs: []*NamedTensor{
					{
						Name:       "predict",
						Datatype:   tyBytes,
						Shape:      []int64{1},
						Parameters: map[string]interface{}{"content_type": "str", "headers": nil},
						tensorData: &TensorData{
							byteContents: [][]byte{[]byte("Hello world")},
						},
					},
				},
				Outputs: []*RequestOutput{},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := convertToInferenceRequest([]byte(test.input))
			g.Expect(err).To(BeNil())
			g.Expect(req).To(Equal(test.expected))
		})
	}
}

func TestInferenceRequestToV2Proto(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name         string
		modelName    string
		modelVersion string
		input        *InferenceRequest
		expected     *v2_dataplane.ModelInferRequest
	}
	tests := []test{
		{
			name:         "test",
			modelName:    "model",
			modelVersion: "1",
			input: &InferenceRequest{
				Parameters: map[string]interface{}{
					"foo": float64(1),
					"bar": true,
					"zoo": "something",
				},
				Inputs: []*NamedTensor{
					{
						Name:     "input1",
						Datatype: tyInt64,
						Shape:    []int64{5},
						Parameters: map[string]interface{}{
							"foo": float64(1),
							"bar": true,
							"zoo": "something",
						},
						tensorData: &TensorData{
							int64Contents: []int64{1, 2, 3, 4, 5},
						},
					},
				},
				Outputs: []*RequestOutput{
					{
						Name: "out1",
						Parameters: map[string]interface{}{
							"foo": float64(1),
							"bar": true,
							"zoo": "something",
						},
					},
				},
			},
			expected: &v2_dataplane.ModelInferRequest{
				Parameters: map[string]*v2_dataplane.InferParameter{
					"foo": {ParameterChoice: &v2_dataplane.InferParameter_Int64Param{Int64Param: 1}},
					"bar": {ParameterChoice: &v2_dataplane.InferParameter_BoolParam{BoolParam: true}},
					"zoo": {ParameterChoice: &v2_dataplane.InferParameter_StringParam{StringParam: "something"}},
				},
				Inputs: []*v2_dataplane.ModelInferRequest_InferInputTensor{
					{
						Name:     "input1",
						Datatype: tyInt64,
						Shape:    []int64{5},
						Parameters: map[string]*v2_dataplane.InferParameter{
							"foo": {ParameterChoice: &v2_dataplane.InferParameter_Int64Param{Int64Param: 1}},
							"bar": {ParameterChoice: &v2_dataplane.InferParameter_BoolParam{BoolParam: true}},
							"zoo": {ParameterChoice: &v2_dataplane.InferParameter_StringParam{StringParam: "something"}},
						},
						Contents: &v2_dataplane.InferTensorContents{Int64Contents: []int64{1, 2, 3, 4, 5}},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			proto, err := inferenceRequestToV2Proto(test.input, test.modelName, test.modelVersion)
			g.Expect(err).To(BeNil())
			for k, v := range test.expected.Parameters {
				g.Expect(proto.Parameters[k]).To(Equal(v))
			}
			for idx, inp := range test.expected.Inputs {
				g.Expect(inp.Name).To(Equal(proto.Inputs[idx].Name))
				g.Expect(inp.Shape).To(Equal(proto.Inputs[idx].Shape))
				g.Expect(inp.Datatype).To(Equal(proto.Inputs[idx].Datatype))
				g.Expect(inp.Contents).To(Equal(proto.Inputs[idx].Contents))
				for k, v := range inp.Parameters {
					g.Expect(proto.Inputs[idx].Parameters[k]).To(Equal(v))
				}
			}
		})
	}
}

func TestMarshallNamedTensor(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name  string
		nt    *NamedTensor
		out   string
		error bool
	}

	tests := []test{
		{
			name: "bool",
			nt: &NamedTensor{
				Name:       "out1",
				Datatype:   tyBool,
				Shape:      []int64{5},
				tensorData: &TensorData{boolContents: []bool{true, false, true, false, true}},
			},
			out: `{"data":[true,false,true,false,true],"name":"out1","shape":[5],"datatype":"BOOL"}`,
		},
		{
			name: "uint32",
			nt: &NamedTensor{
				Name:       "out1",
				Datatype:   tyUint32,
				Shape:      []int64{5},
				tensorData: &TensorData{uint32Contents: []uint32{1, 2, 3, 4, 5}},
			},
			out: `{"data":[1,2,3,4,5],"name":"out1","shape":[5],"datatype":"UINT32"}`,
		},
		{
			name: "uint64",
			nt: &NamedTensor{
				Name:       "out1",
				Datatype:   tyUint64,
				Shape:      []int64{5},
				tensorData: &TensorData{uint64Contents: []uint64{1, 2, 3, 4, 5}},
			},
			out: `{"data":[1,2,3,4,5],"name":"out1","shape":[5],"datatype":"UINT64"}`,
		},
		{
			name: "int32",
			nt: &NamedTensor{
				Name:       "out1",
				Datatype:   tyInt32,
				Shape:      []int64{5},
				tensorData: &TensorData{int32Contents: []int32{1, 2, 3, 4, 5}},
			},
			out: `{"data":[1,2,3,4,5],"name":"out1","shape":[5],"datatype":"INT32"}`,
		},
		{
			name: "int64",
			nt: &NamedTensor{
				Name:       "out1",
				Datatype:   tyInt64,
				Shape:      []int64{5},
				tensorData: &TensorData{int64Contents: []int64{1, 2, 3, 4, 5}},
			},
			out: `{"data":[1,2,3,4,5],"name":"out1","shape":[5],"datatype":"INT64"}`,
		},
		{
			name: "fp32",
			nt: &NamedTensor{
				Name:       "out1",
				Datatype:   tyFp32,
				Shape:      []int64{5},
				tensorData: &TensorData{fp32Contents: []float32{1.1, 2.2, 3.3, 4.4, 5.5}},
			},
			out: `{"data":[1.1,2.2,3.3,4.4,5.5],"name":"out1","shape":[5],"datatype":"FP32"}`,
		},
		{
			name: "fp64",
			nt: &NamedTensor{
				Name:       "out1",
				Datatype:   tyFp64,
				Shape:      []int64{5},
				tensorData: &TensorData{fp64Contents: []float64{1.1, 2.2, 3.3, 4.4, 5.5}},
			},
			out: `{"data":[1.1,2.2,3.3,4.4,5.5],"name":"out1","shape":[5],"datatype":"FP64"}`,
		},
		{
			name: "bytes",
			nt: &NamedTensor{
				Name:       "out1",
				Datatype:   tyBytes,
				Shape:      []int64{5},
				tensorData: &TensorData{byteContents: [][]byte{[]byte("test")}},
			},
			out: `{"data":["dGVzdA=="],"name":"out1","shape":[5],"datatype":"BYTES"}`,
		},
		{
			name: "bytes with str content type",
			nt: &NamedTensor{
				Name:       "out1",
				Datatype:   tyBytes,
				Shape:      []int64{5},
				tensorData: &TensorData{strContents: []string{"test"}},
				Parameters: map[string]interface{}{contentTypeKey: "str"},
			},
			out: `{"data":["test"],"name":"out1","shape":[5],"datatype":"BYTES","parameters":{"content_type":"str"}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out, err := json.Marshal(test.nt)
			if test.error {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(string(out)).To(Equal(test.out))
			}

		})
	}
}

func TestConvertResponseToJSON(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name string
		res  *v2_dataplane.ModelInferResponse
		out  string
	}

	tests := []test{
		{
			name: "bool",
			res: &v2_dataplane.ModelInferResponse{
				ModelName:    "model",
				ModelVersion: "1",
				Id:           "1234",
				Outputs: []*v2_dataplane.ModelInferResponse_InferOutputTensor{
					{
						Name:     "t1",
						Datatype: tyBool,
						Shape:    []int64{2},
						Contents: &v2_dataplane.InferTensorContents{
							BoolContents: []bool{true, false},
						},
					},
				},
			},
			out: `{"model_name":"model","model_version":"1","id":"1234","outputs":[{"data":[true,false],"name":"t1","shape":[2],"datatype":"BOOL"}]}`,
		},
		{
			name: "uint32",
			res: &v2_dataplane.ModelInferResponse{
				ModelName:    "model",
				ModelVersion: "1",
				Id:           "1234",
				Outputs: []*v2_dataplane.ModelInferResponse_InferOutputTensor{
					{
						Name:     "t1",
						Datatype: tyUint32,
						Shape:    []int64{2},
						Contents: &v2_dataplane.InferTensorContents{
							UintContents: []uint32{1, 2},
						},
					},
				},
			},
			out: `{"model_name":"model","model_version":"1","id":"1234","outputs":[{"data":[1,2],"name":"t1","shape":[2],"datatype":"UINT32"}]}`,
		},
		{
			name: "uint64",
			res: &v2_dataplane.ModelInferResponse{
				ModelName:    "model",
				ModelVersion: "1",
				Id:           "1234",
				Outputs: []*v2_dataplane.ModelInferResponse_InferOutputTensor{
					{
						Name:     "t1",
						Datatype: tyUint64,
						Shape:    []int64{2},
						Contents: &v2_dataplane.InferTensorContents{
							Uint64Contents: []uint64{1, 2},
						},
					},
				},
			},
			out: `{"model_name":"model","model_version":"1","id":"1234","outputs":[{"data":[1,2],"name":"t1","shape":[2],"datatype":"UINT64"}]}`,
		},
		{
			name: "int32",
			res: &v2_dataplane.ModelInferResponse{
				ModelName:    "model",
				ModelVersion: "1",
				Id:           "1234",
				Outputs: []*v2_dataplane.ModelInferResponse_InferOutputTensor{
					{
						Name:     "t1",
						Datatype: tyInt32,
						Shape:    []int64{2},
						Contents: &v2_dataplane.InferTensorContents{
							IntContents: []int32{1, 2},
						},
					},
				},
			},
			out: `{"model_name":"model","model_version":"1","id":"1234","outputs":[{"data":[1,2],"name":"t1","shape":[2],"datatype":"INT32"}]}`,
		},
		{
			name: "int64",
			res: &v2_dataplane.ModelInferResponse{
				ModelName:    "model",
				ModelVersion: "1",
				Id:           "1234",
				Outputs: []*v2_dataplane.ModelInferResponse_InferOutputTensor{
					{
						Name:     "t1",
						Datatype: tyInt64,
						Shape:    []int64{2},
						Contents: &v2_dataplane.InferTensorContents{
							Int64Contents: []int64{1, 2},
						},
					},
				},
			},
			out: `{"model_name":"model","model_version":"1","id":"1234","outputs":[{"data":[1,2],"name":"t1","shape":[2],"datatype":"INT64"}]}`,
		},
		{
			name: "fp32",
			res: &v2_dataplane.ModelInferResponse{
				ModelName:    "model",
				ModelVersion: "1",
				Id:           "1234",
				Outputs: []*v2_dataplane.ModelInferResponse_InferOutputTensor{
					{
						Name:     "t1",
						Datatype: tyFp32,
						Shape:    []int64{2},
						Contents: &v2_dataplane.InferTensorContents{
							Fp32Contents: []float32{1.1, 2.2},
						},
					},
				},
			},
			out: `{"model_name":"model","model_version":"1","id":"1234","outputs":[{"data":[1.1,2.2],"name":"t1","shape":[2],"datatype":"FP32"}]}`,
		},
		{
			name: "fp64",
			res: &v2_dataplane.ModelInferResponse{
				ModelName:    "model",
				ModelVersion: "1",
				Id:           "1234",
				Outputs: []*v2_dataplane.ModelInferResponse_InferOutputTensor{
					{
						Name:     "t1",
						Datatype: tyFp64,
						Shape:    []int64{2},
						Contents: &v2_dataplane.InferTensorContents{
							Fp64Contents: []float64{1.1, 2.2},
						},
					},
				},
			},
			out: `{"model_name":"model","model_version":"1","id":"1234","outputs":[{"data":[1.1,2.2],"name":"t1","shape":[2],"datatype":"FP64"}]}`,
		},
		{
			name: "bytes",
			res: &v2_dataplane.ModelInferResponse{
				ModelName:    "model",
				ModelVersion: "1",
				Id:           "1234",
				Outputs: []*v2_dataplane.ModelInferResponse_InferOutputTensor{
					{
						Name:     "t1",
						Datatype: tyBytes,
						Shape:    []int64{2},
						Contents: &v2_dataplane.InferTensorContents{
							BytesContents: [][]byte{[]byte("test")},
						},
					},
				},
			},
			out: `{"model_name":"model","model_version":"1","id":"1234","outputs":[{"data":["dGVzdA=="],"name":"t1","shape":[2],"datatype":"BYTES"}]}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out, err := convertV2ResponseToJson(test.res)
			g.Expect(err).To(BeNil())
			g.Expect(string(out)).To(Equal(test.out))
		})
	}
}

func TestConvertRequestToV2(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name         string
		modelName    string
		modelVersion string
		inp          string
		out          *v2_dataplane.ModelInferRequest
	}

	tests := []test{
		{
			name:         "bool",
			modelName:    "foo",
			modelVersion: "1",
			inp:          `{"inputs":[{"data":[true,false,true,false,true],"name":"out1","shape":[5],"datatype":"BOOL"}]}`,
			out: &v2_dataplane.ModelInferRequest{
				ModelName:    "foo",
				ModelVersion: "1",
				Parameters:   map[string]*v2_dataplane.InferParameter{},
				Inputs: []*v2_dataplane.ModelInferRequest_InferInputTensor{
					{
						Name:       "out1",
						Datatype:   tyBool,
						Shape:      []int64{5},
						Parameters: map[string]*v2_dataplane.InferParameter{},
						Contents:   &v2_dataplane.InferTensorContents{BoolContents: []bool{true, false, true, false, true}},
					},
				},
			},
		},
		{
			name:         "uint32",
			modelName:    "foo",
			modelVersion: "1",
			inp:          `{"inputs":[{"data":[1,2,3,4,5],"name":"out1","shape":[5],"datatype":"UINT32"}]}`,
			out: &v2_dataplane.ModelInferRequest{
				ModelName:    "foo",
				ModelVersion: "1",
				Parameters:   map[string]*v2_dataplane.InferParameter{},
				Inputs: []*v2_dataplane.ModelInferRequest_InferInputTensor{
					{
						Name:       "out1",
						Datatype:   tyUint32,
						Shape:      []int64{5},
						Parameters: map[string]*v2_dataplane.InferParameter{},
						Contents:   &v2_dataplane.InferTensorContents{UintContents: []uint32{1, 2, 3, 4, 5}},
					},
				},
			},
		},
		{
			name:         "uint64",
			modelName:    "foo",
			modelVersion: "1",
			inp:          `{"inputs":[{"data":[1,2,3,4,5],"name":"out1","shape":[5],"datatype":"UINT64"}]}`,
			out: &v2_dataplane.ModelInferRequest{
				ModelName:    "foo",
				ModelVersion: "1",
				Parameters:   map[string]*v2_dataplane.InferParameter{},
				Inputs: []*v2_dataplane.ModelInferRequest_InferInputTensor{
					{
						Name:       "out1",
						Datatype:   tyUint64,
						Shape:      []int64{5},
						Parameters: map[string]*v2_dataplane.InferParameter{},
						Contents:   &v2_dataplane.InferTensorContents{Uint64Contents: []uint64{1, 2, 3, 4, 5}},
					},
				},
			},
		},
		{
			name:         "int32",
			modelName:    "foo",
			modelVersion: "1",
			inp:          `{"inputs":[{"data":[1,2,3,4,5],"name":"out1","shape":[5],"datatype":"INT32"}]}`,
			out: &v2_dataplane.ModelInferRequest{
				ModelName:    "foo",
				ModelVersion: "1",
				Parameters:   map[string]*v2_dataplane.InferParameter{},
				Inputs: []*v2_dataplane.ModelInferRequest_InferInputTensor{
					{
						Name:       "out1",
						Datatype:   tyInt32,
						Shape:      []int64{5},
						Parameters: map[string]*v2_dataplane.InferParameter{},
						Contents:   &v2_dataplane.InferTensorContents{IntContents: []int32{1, 2, 3, 4, 5}},
					},
				},
			},
		},
		{
			name:         "int64",
			modelName:    "foo",
			modelVersion: "1",
			inp:          `{"inputs":[{"data":[1,2,3,4,5],"name":"out1","shape":[5],"datatype":"INT64"}]}`,
			out: &v2_dataplane.ModelInferRequest{
				ModelName:    "foo",
				ModelVersion: "1",
				Parameters:   map[string]*v2_dataplane.InferParameter{},
				Inputs: []*v2_dataplane.ModelInferRequest_InferInputTensor{
					{
						Name:       "out1",
						Datatype:   tyInt64,
						Shape:      []int64{5},
						Parameters: map[string]*v2_dataplane.InferParameter{},
						Contents:   &v2_dataplane.InferTensorContents{Int64Contents: []int64{1, 2, 3, 4, 5}},
					},
				},
			},
		},
		{
			name:         "fp32",
			modelName:    "foo",
			modelVersion: "1",
			inp:          `{"inputs":[{"data":[1.1,2.2,3.3,4.4,5.5],"name":"out1","shape":[5],"datatype":"FP32"}]}`,
			out: &v2_dataplane.ModelInferRequest{
				ModelName:    "foo",
				ModelVersion: "1",
				Parameters:   map[string]*v2_dataplane.InferParameter{},
				Inputs: []*v2_dataplane.ModelInferRequest_InferInputTensor{
					{
						Name:       "out1",
						Datatype:   tyFp32,
						Shape:      []int64{5},
						Parameters: map[string]*v2_dataplane.InferParameter{},
						Contents:   &v2_dataplane.InferTensorContents{Fp32Contents: []float32{1.1, 2.2, 3.3, 4.4, 5.5}},
					},
				},
			},
		},
		{
			name:         "fp32-iris",
			modelName:    "foo",
			modelVersion: "1",
			inp:          `{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [1, 2, 3, 4]}]}`,
			out: &v2_dataplane.ModelInferRequest{
				ModelName:    "foo",
				ModelVersion: "1",
				Parameters:   map[string]*v2_dataplane.InferParameter{},
				Inputs: []*v2_dataplane.ModelInferRequest_InferInputTensor{
					{
						Name:       "predict",
						Datatype:   tyFp32,
						Shape:      []int64{1, 4},
						Parameters: map[string]*v2_dataplane.InferParameter{},
						Contents:   &v2_dataplane.InferTensorContents{Fp32Contents: []float32{1, 2, 3, 4}},
					},
				},
			},
		},
		{
			name:         "fp64",
			modelName:    "foo",
			modelVersion: "1",
			inp:          `{"inputs":[{"data":[1.1,2.2,3.3,4.4,5.5],"name":"out1","shape":[5],"datatype":"FP64"}]}`,
			out: &v2_dataplane.ModelInferRequest{
				ModelName:    "foo",
				ModelVersion: "1",
				Parameters:   map[string]*v2_dataplane.InferParameter{},
				Inputs: []*v2_dataplane.ModelInferRequest_InferInputTensor{
					{
						Name:       "out1",
						Datatype:   tyFp64,
						Shape:      []int64{5},
						Parameters: map[string]*v2_dataplane.InferParameter{},
						Contents:   &v2_dataplane.InferTensorContents{Fp64Contents: []float64{1.1, 2.2, 3.3, 4.4, 5.5}},
					},
				},
			},
		},
		{
			name:         "bytes",
			modelName:    "foo",
			modelVersion: "1",
			inp:          `{"inputs":[{"data":["test"],"name":"out1","shape":[1,5],"datatype":"BYTES"}]}`,
			out: &v2_dataplane.ModelInferRequest{
				ModelName:    "foo",
				ModelVersion: "1",
				Parameters:   map[string]*v2_dataplane.InferParameter{},
				Inputs: []*v2_dataplane.ModelInferRequest_InferInputTensor{
					{
						Name:       "out1",
						Datatype:   tyBytes,
						Shape:      []int64{1, 5},
						Parameters: map[string]*v2_dataplane.InferParameter{},
						Contents:   &v2_dataplane.InferTensorContents{BytesContents: [][]byte{[]byte("test")}},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			v2, err := convertRequestToV2([]byte(test.inp), test.modelName, test.modelVersion)
			g.Expect(err).To(BeNil())
			g.Expect(v2).To(Equal(test.out))
		})
	}
}

func TestUpdateResponseFromRawContents(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name string
		res  *v2_dataplane.ModelInferResponse
		err  bool
	}

	i := 100
	bInt64 := make([]byte, 8)
	binary.LittleEndian.PutUint64(bInt64, uint64(i))
	tests := []test{
		{
			name: "no raw contents",
			res: &v2_dataplane.ModelInferResponse{
				Outputs: []*v2_dataplane.ModelInferResponse_InferOutputTensor{
					{
						Name: "t1",
					},
				},
			},
			err: false,
		},
		{
			name: "raw contents bytes",
			res: &v2_dataplane.ModelInferResponse{
				Outputs: []*v2_dataplane.ModelInferResponse_InferOutputTensor{
					{
						Name:     "t1",
						Datatype: tyBytes,
					},
				},
				RawOutputContents: [][]byte{
					createRawBytesFromStringSlice([]string{"result"}),
				},
			},
			err: false,
		},
		{
			name: "raw contents int",
			res: &v2_dataplane.ModelInferResponse{
				Outputs: []*v2_dataplane.ModelInferResponse_InferOutputTensor{
					{
						Name:     "t1",
						Datatype: tyUint64,
					},
				},
				RawOutputContents: [][]byte{
					bInt64,
				},
			},
			err: false,
		},
		{
			name: "error raw contents cant find place",
			res: &v2_dataplane.ModelInferResponse{
				Outputs: []*v2_dataplane.ModelInferResponse_InferOutputTensor{
					{
						Name:     "t1",
						Datatype: tyUint64,
						Contents: &v2_dataplane.InferTensorContents{},
					},
				},
				RawOutputContents: [][]byte{
					bInt64,
				},
			},
			err: true,
		},
		{
			name: "mixed raw and normal contents",
			res: &v2_dataplane.ModelInferResponse{
				Outputs: []*v2_dataplane.ModelInferResponse_InferOutputTensor{
					{
						Name:     "t1",
						Datatype: tyUint64,
						Contents: &v2_dataplane.InferTensorContents{},
					},
					{
						Name:     "t1",
						Datatype: tyUint64,
					},
				},
				RawOutputContents: [][]byte{
					bInt64,
				},
			},
			err: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := updateResponseFromRawContents(test.res)
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
			}
		})
	}
}

func TestResponseV2ParametersToJson(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name     string
		input    map[string]*v2_dataplane.InferParameter
		expected string
	}
	tests := []test{
		{
			name: "bool",
			input: map[string]*v2_dataplane.InferParameter{"foo": {
				ParameterChoice: &v2_dataplane.InferParameter_BoolParam{BoolParam: true},
			}},
			expected: `{"foo":true}`,
		},
		{
			name: "int64",
			input: map[string]*v2_dataplane.InferParameter{"foo": {
				ParameterChoice: &v2_dataplane.InferParameter_Int64Param{Int64Param: 3},
			}},
			expected: `{"foo":3}`,
		},
		{
			name: "string",
			input: map[string]*v2_dataplane.InferParameter{"foo": {
				ParameterChoice: &v2_dataplane.InferParameter_StringParam{StringParam: "bar"},
			}},
			expected: `{"foo":"bar"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resMap := createParametersFromv2(test.input)
			jStr, err := json.Marshal(resMap)
			g.Expect(err).To(BeNil())
			g.Expect(string(jStr)).To(Equal(test.expected))
		})
	}
}

func createRawBytesFromStringSlice(vals []string) []byte {
	var result []byte
	for _, val := range vals {
		data := []byte(val)
		strLen := len(data)
		lenB := make([]byte, 4)
		binary.LittleEndian.PutUint32(lenB, uint32(strLen))
		result = append(result, lenB...)
		result = append(result, data...)
	}
	return result
}

func TestConvertRawBytesToByteContents(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name     string
		input    []byte
		expected [][]byte
	}
	tests := []test{
		{
			name:     "single string",
			input:    createRawBytesFromStringSlice([]string{"hello"}),
			expected: [][]byte{[]byte("hello")},
		},
		{
			name:     "multiple strings",
			input:    createRawBytesFromStringSlice([]string{"hello", "there"}),
			expected: [][]byte{[]byte("hello"), []byte("there")},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := convertRawBytesToByteContents(test.input)
			g.Expect(data).To(Equal(test.expected))
		})
	}
}
