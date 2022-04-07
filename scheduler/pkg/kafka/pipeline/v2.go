package pipeline

import (
	"encoding/json"
	"fmt"

	"github.com/seldonio/seldon-core/scheduler/apis/mlops/v2_dataplane"
	"google.golang.org/protobuf/proto"
)

//$inference_request =
//{
//  "id" : $string #optional,
//  "parameters" : $parameters #optional,
//  "inputs" : [ $request_input, ... ],
//  "outputs" : [ $request_output, ... ] #optional
//}

const (
	tyBool   = "BOOL"
	tyUint8  = "UINT8"
	tyUint16 = "UINT16"
	tyUint32 = "UINT32"
	tyUint64 = "UINT64"
	tyInt8   = "INT8"
	tyInt16  = "INT16"
	tyInt32  = "INT32"
	tyInt64  = "INT64"
	tyFp16   = "FP16"
	tyFp32   = "FP32"
	tyFp64   = "FP64"
	tyBytes  = "BYTES"
)

type InferenceRequest struct {
	Id         string                 `json:"id,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Inputs     []*NamedTensor         `json:"inputs"`
	Outputs    []*RequestOutput       `json:"outputs,omitempty"`
}

//$inference_response =
//{
//  "model_name" : $string,
//  "model_version" : $string #optional,
//  "id" : $string,
//  "parameters" : $parameters #optional,
//  "outputs" : [ $response_output, ... ]
//}

type InferenceResponse struct {
	ModelName         string                 `json:"model_name"`
	ModelVersion      string                 `json:"model_version,omitempty"`
	Id                string                 `json:"id,omitempty"`
	Parameters        map[string]interface{} `json:"parameters,omitempty"`
	Outputs           []*NamedTensor         `json:"outputs,omitempty"`
	RawOutputContents [][]byte               `json:"rawOutputContents,omitempty"`
}

type NamedTensor struct {
	Name       string                 `json:"name"`
	Shape      []int64                `json:"shape"`
	Datatype   string                 `json:"datatype"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Data       json.RawMessage        `json:"data"`
	tensorData *TensorData
}

func (nt *NamedTensor) MarshalJSON() ([]byte, error) {
	type Alias NamedTensor
	switch nt.Datatype {
	case tyBool:
		return json.Marshal(&struct {
			Data []bool `json:"data"`
			*Alias
		}{
			Data:  nt.tensorData.boolContents,
			Alias: (*Alias)(nt),
		})
	case tyUint8, tyUint16, tyUint32:
		return json.Marshal(&struct {
			Data []uint32 `json:"data"`
			*Alias
		}{
			Data:  nt.tensorData.uint32Contents,
			Alias: (*Alias)(nt),
		})
	case tyUint64:
		return json.Marshal(&struct {
			Data []uint64 `json:"data"`
			*Alias
		}{
			Data:  nt.tensorData.uint64Contents,
			Alias: (*Alias)(nt),
		})
	case tyInt8, tyInt16, tyInt32:
		return json.Marshal(&struct {
			Data []int32 `json:"data"`
			*Alias
		}{
			Data:  nt.tensorData.int32Contents,
			Alias: (*Alias)(nt),
		})
	case tyInt64:
		return json.Marshal(&struct {
			Data []int64 `json:"data"`
			*Alias
		}{
			Data:  nt.tensorData.int64Contents,
			Alias: (*Alias)(nt),
		})
	case tyFp16, tyFp32:
		return json.Marshal(&struct {
			Data []float32 `json:"data"`
			*Alias
		}{
			Data:  nt.tensorData.fp32Contents,
			Alias: (*Alias)(nt),
		})
	case tyFp64:
		return json.Marshal(&struct {
			Data []float64 `json:"data"`
			*Alias
		}{
			Data:  nt.tensorData.fp64Contents,
			Alias: (*Alias)(nt),
		})
	case tyBytes:
		return json.Marshal(&struct {
			Data [][]byte `json:"data"`
			*Alias
		}{
			Data:  nt.tensorData.byteContents,
			Alias: (*Alias)(nt),
		})
	default:
		return nil, fmt.Errorf("Unknown type %s", nt.Datatype)
	}
}

type RequestOutput struct {
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters"`
}

type TensorData struct {
	boolContents   []bool
	uint32Contents []uint32
	uint64Contents []uint64
	int32Contents  []int32
	int64Contents  []int64
	fp32Contents   []float32
	fp64Contents   []float64
	byteContents   [][]byte
}

func ConvertRequestToV2Bytes(data []byte, modelName string, modelVersion string) ([]byte, error) {
	res, err := convertRequestToV2(data, modelName, modelVersion)
	if err != nil {
		return nil, err
	}
	return proto.Marshal(res)
}

func convertRequestToV2(data []byte, modelName string, modelVersion string) (*v2_dataplane.ModelInferRequest, error) {
	infReq, err := convertToInferenceRequest(data)
	if err != nil {
		return nil, err
	}
	return inferenceRequestToV2Proto(infReq, modelName, modelVersion)
}

func ConvertV2ResponseBytesToJson(res []byte) ([]byte, error) {
	v2Res := &v2_dataplane.ModelInferResponse{}
	err := proto.Unmarshal(res, v2Res)
	if err != nil {
		return nil, err
	}
	return convertV2ResponseToJson(v2Res)
}

func convertV2ResponseToJson(response *v2_dataplane.ModelInferResponse) ([]byte, error) {
	infRes := convertV2toInferenceResponse(response)
	return json.Marshal(infRes)
}

func createParametersFromv2(v2Params map[string]*v2_dataplane.InferParameter) map[string]interface{} {
	params := make(map[string]interface{})
	for k, v := range v2Params {
		switch pval := v.ParameterChoice.(type) {
		case *v2_dataplane.InferParameter_BoolParam:
			params[k] = pval
		case *v2_dataplane.InferParameter_StringParam:
			params[k] = pval
		case *v2_dataplane.InferParameter_Int64Param:
			params[k] = pval
		}
	}
	return params
}

func convertV2InferOutputToNamedTensor(v2Output *v2_dataplane.ModelInferResponse_InferOutputTensor) *NamedTensor {
	td := &TensorData{
		boolContents:   v2Output.Contents.GetBoolContents(),
		uint32Contents: v2Output.Contents.GetUintContents(),
		uint64Contents: v2Output.Contents.GetUint64Contents(),
		int32Contents:  v2Output.Contents.GetIntContents(),
		int64Contents:  v2Output.Contents.GetInt64Contents(),
		fp32Contents:   v2Output.Contents.GetFp32Contents(),
		fp64Contents:   v2Output.Contents.GetFp64Contents(),
		byteContents:   v2Output.Contents.GetBytesContents(),
	}
	return &NamedTensor{
		Name:       v2Output.Name,
		Shape:      v2Output.Shape,
		Datatype:   v2Output.Datatype,
		Parameters: createParametersFromv2(v2Output.Parameters),
		tensorData: td,
	}
}

func convertV2toInferenceResponse(resV2 *v2_dataplane.ModelInferResponse) *InferenceResponse {
	var outputs []*NamedTensor
	for _, v2Out := range resV2.Outputs {
		outputs = append(outputs, convertV2InferOutputToNamedTensor(v2Out))
	}
	return &InferenceResponse{
		ModelName:         resV2.ModelName,
		ModelVersion:      resV2.ModelVersion,
		Id:                resV2.Id,
		Parameters:        createParametersFromv2(resV2.Parameters),
		Outputs:           outputs,
		RawOutputContents: resV2.RawOutputContents,
	}
}

func copyToTensor(from []interface{}, tensors *NamedTensor, idx int) int {
	for _, v := range from {
		switch val := v.(type) {
		case []interface{}:
			idx = copyToTensor(val, tensors, idx)
		default:
			switch tensors.Datatype {
			case tyBool:
				tensors.tensorData.boolContents[idx] = v.(bool)
			case tyUint8, tyUint16, tyUint32:
				tensors.tensorData.uint32Contents[idx] = uint32(v.(float64))
			case tyUint64:
				tensors.tensorData.uint64Contents[idx] = uint64(v.(float64))
			case tyInt8, tyInt16, tyInt32:
				tensors.tensorData.int32Contents[idx] = int32(v.(float64))
			case tyInt64:
				tensors.tensorData.int64Contents[idx] = int64(v.(float64))
			case tyFp16, tyFp32:
				tensors.tensorData.fp32Contents[idx] = float32(v.(float64))
			case tyFp64:
				tensors.tensorData.fp64Contents[idx] = v.(float64)
			case tyBytes:
				tensors.tensorData.byteContents[idx] = []byte(v.(string))
			}
			idx = idx + 1
		}
	}
	return idx
}

func getDataSize(shape []int64) int64 {
	tot := int64(1)
	for _, dim := range shape {
		tot = tot * dim
	}
	return tot
}

func convertTensors(req *NamedTensor) error {
	td := &TensorData{}
	req.tensorData = td
	sz := getDataSize(req.Shape)
	switch req.Datatype {
	case tyBool:
		td.boolContents = make([]bool, sz)
	case tyUint8, tyUint16, tyUint32:
		td.uint32Contents = make([]uint32, sz)
	case tyUint64:
		td.uint64Contents = make([]uint64, sz)
	case tyInt8, tyInt16, tyInt32:
		td.int32Contents = make([]int32, sz)
	case tyInt64:
		td.int64Contents = make([]int64, sz)
	case tyFp16, tyFp32:
		td.fp32Contents = make([]float32, sz)
	case tyFp64:
		td.fp64Contents = make([]float64, sz)
	case tyBytes:
		td.byteContents = make([][]byte, req.Shape[0])
	default:
		return fmt.Errorf("Unknown type %s", req.Datatype)
	}
	var data []interface{}
	err := json.Unmarshal(req.Data, &data)
	if err != nil {
		return err
	}
	copyToTensor(data, req, 0)
	return nil
}

func convertTensorsPrev(req *NamedTensor) error {
	td := &TensorData{}
	req.tensorData = td
	sz := getDataSize(req.Shape)
	switch req.Datatype {
	case tyBool:
		td.boolContents = make([]bool, sz)
		return json.Unmarshal(req.Data, &td.boolContents)
	case tyUint8, tyUint16, tyUint32:
		td.uint32Contents = make([]uint32, sz)
		return json.Unmarshal(req.Data, &td.uint32Contents)
	case tyUint64:
		td.uint64Contents = make([]uint64, sz)
		return json.Unmarshal(req.Data, &td.uint64Contents)
	case tyInt8, tyInt16, tyInt32:
		td.int32Contents = make([]int32, sz)
		return json.Unmarshal(req.Data, &td.int32Contents)
	case tyInt64:
		td.int64Contents = make([]int64, sz)
		return json.Unmarshal(req.Data, &td.int64Contents)
	case tyFp16, tyFp32:
		td.fp32Contents = make([]float32, sz)
		return json.Unmarshal(req.Data, &td.fp32Contents)
	case tyFp64:
		td.fp64Contents = make([]float64, sz)
		return json.Unmarshal(req.Data, &td.fp64Contents)
	case tyBytes:
		td.byteContents = make([][]byte, 1)
		td.byteContents[0] = make([]byte, sz)
		return json.Unmarshal(req.Data, &td.byteContents[0])
	default:
		return fmt.Errorf("Unknown type %s", req.Datatype)
	}
}

func convertToInferenceRequest(data []byte) (*InferenceRequest, error) {
	req := &InferenceRequest{}
	err := json.Unmarshal(data, req)
	if err != nil {
		return nil, err
	}
	for _, inp := range req.Inputs {
		err := convertTensors(inp)
		if err != nil {
			return nil, err
		}
		inp.Data = nil
	}
	return req, nil
}

func tensorToV2(t *TensorData, ty string) *v2_dataplane.InferTensorContents {
	switch ty {
	case tyBool:
		return &v2_dataplane.InferTensorContents{BoolContents: t.boolContents}
	case tyUint8, tyUint16, tyUint32:
		return &v2_dataplane.InferTensorContents{UintContents: t.uint32Contents}
	case tyUint64:
		return &v2_dataplane.InferTensorContents{Uint64Contents: t.uint64Contents}
	case tyInt8, tyInt16, tyInt32:
		return &v2_dataplane.InferTensorContents{IntContents: t.int32Contents}
	case tyInt64:
		return &v2_dataplane.InferTensorContents{Int64Contents: t.int64Contents}
	case tyFp16, tyFp32:
		return &v2_dataplane.InferTensorContents{Fp32Contents: t.fp32Contents}
	case tyFp64:
		return &v2_dataplane.InferTensorContents{Fp64Contents: t.fp64Contents}
	case tyBytes:
		return &v2_dataplane.InferTensorContents{BytesContents: t.byteContents}
	}
	return nil
}

func requestInputParametersToV2(val interface{}) (*v2_dataplane.InferParameter, error) {
	switch ty := val.(type) {
	case string:
		return &v2_dataplane.InferParameter{
			ParameterChoice: &v2_dataplane.InferParameter_StringParam{StringParam: val.(string)},
		}, nil
	case float64:
		return &v2_dataplane.InferParameter{
			ParameterChoice: &v2_dataplane.InferParameter_Int64Param{Int64Param: int64(val.(float64))},
		}, nil
	case bool:
		return &v2_dataplane.InferParameter{
			ParameterChoice: &v2_dataplane.InferParameter_BoolParam{BoolParam: val.(bool)},
		}, nil
	default:
		return nil, fmt.Errorf("Unknown type for parameter %v", ty)
	}
}

func parametersToV2(paramsIn map[string]interface{}) (map[string]*v2_dataplane.InferParameter, error) {
	params := make(map[string]*v2_dataplane.InferParameter)
	for k, v := range paramsIn {
		ip, err := requestInputParametersToV2(v)
		if err != nil {
			return nil, err
		}
		params[k] = ip
	}
	return params, nil
}

func requestInputToV2(req *NamedTensor) (*v2_dataplane.ModelInferRequest_InferInputTensor, error) {
	params, err := parametersToV2(req.Parameters)
	if err != nil {
		return nil, err
	}
	return &v2_dataplane.ModelInferRequest_InferInputTensor{
		Name:       req.Name,
		Datatype:   req.Datatype,
		Shape:      req.Shape,
		Parameters: params,
		Contents:   tensorToV2(req.tensorData, req.Datatype),
	}, nil
}

func requestOutputToV2(out *RequestOutput) (*v2_dataplane.ModelInferRequest_InferRequestedOutputTensor, error) {
	params, err := parametersToV2(out.Parameters)
	if err != nil {
		return nil, err
	}
	return &v2_dataplane.ModelInferRequest_InferRequestedOutputTensor{
		Name:       out.Name,
		Parameters: params,
	}, nil
}

func inferenceRequestToV2Proto(inf *InferenceRequest, modelName string, modelVersion string) (*v2_dataplane.ModelInferRequest, error) {
	params, err := parametersToV2(inf.Parameters)
	if err != nil {
		return nil, err
	}
	var inputs []*v2_dataplane.ModelInferRequest_InferInputTensor
	for _, ri := range inf.Inputs {
		it, err := requestInputToV2(ri)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, it)
	}
	var outputs []*v2_dataplane.ModelInferRequest_InferRequestedOutputTensor
	for _, out := range inf.Outputs {
		ot, err := requestOutputToV2(out)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, ot)
	}
	return &v2_dataplane.ModelInferRequest{
		ModelName:    modelName,
		ModelVersion: modelVersion,
		Parameters:   params,
		Inputs:       inputs,
		Outputs:      outputs,
	}, nil
}
