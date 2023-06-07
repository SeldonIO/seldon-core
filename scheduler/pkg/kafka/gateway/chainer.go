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

package gateway

import (
	"encoding/json"

	"google.golang.org/protobuf/proto"

	v2 "github.com/seldonio/seldon-core/apis/go/v2/mlops/v2_dataplane"
)

func getProtoRequestAssumingResponse(data []byte) (*v2.ModelInferRequest, error) {
	iresp := v2.ModelInferResponse{}
	err := proto.Unmarshal(data, &iresp)
	if err != nil {
		return nil, err
	}
	return chainProtoResponseToRequest(&iresp), nil
}

func chainProtoResponseToRequest(response *v2.ModelInferResponse) *v2.ModelInferRequest {
	inputTensors := make([]*v2.ModelInferRequest_InferInputTensor, len(response.Outputs))
	for idx, oTensor := range response.Outputs {
		inputTensor := &v2.ModelInferRequest_InferInputTensor{
			Name:       oTensor.Name,
			Datatype:   oTensor.Datatype,
			Shape:      oTensor.Shape,
			Parameters: oTensor.Parameters,
			Contents:   oTensor.Contents,
		}
		inputTensors[idx] = inputTensor
	}

	return &v2.ModelInferRequest{
		Inputs:           inputTensors,
		RawInputContents: response.RawOutputContents,
	}
}

// Optional create a v2 infer request JSON if we find a v2 infer response JSON
func maybeChainRest(data []byte) []byte {
	var f interface{}
	err := json.Unmarshal(data, &f)
	if err != nil {
		return data
	}
	m := f.(map[string]interface{})
	if _, ok := m["inputs"]; ok {
		return data
	} else if _, ok := m["outputs"]; ok {
		m["inputs"] = m["outputs"]
		delete(m, "outputs")
		b, err := json.Marshal(m)
		if err != nil {
			return data
		}
		return b
	} else {
		return data
	}
}
