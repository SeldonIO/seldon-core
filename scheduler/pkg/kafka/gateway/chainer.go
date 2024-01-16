/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
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
