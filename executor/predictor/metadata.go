package predictor

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type MetadataTensor struct {
	DataType string `json:"datatype,omitempty"`
	Name     string `json:"name,omitempty"`
	Shape    []int  `json:"shape,omitempty"`
}

type ModelMetadata struct {
	ApiVersion string      `json:"apiVersion,omitempty"`
	Name       string      `json:"name,omitempty"`
	Platform   string      `json:"platform,omitempty"`
	Versions   []string    `json:"versions,omitempty"`
	Inputs     interface{} `json:"inputs,omitempty"`
	Outputs    interface{} `json:"outputs,omitempty"`
}

func (m *ModelMetadata) ToProto() *proto.SeldonModelMetadata {
	output := &proto.SeldonModelMetadata{
		ApiVersion: m.ApiVersion,
		Name:       m.Name,
		Versions:   m.Versions,
		Platform:   m.Platform,
	}
	fmt.Println("ApiVersion:", m.ApiVersion)
	switch m.ApiVersion {
	case "v1":
		output.Input = m.Inputs.(*proto.SeldonMessageMeta)
		output.Output = m.Outputs.(*proto.SeldonMessageMeta)
	case "v2":
		output.Inputs = m.Inputs.([]*proto.TensorMetadata)
		output.Outputs = m.Outputs.([]*proto.TensorMetadata)
	}
	return output
}

type GraphMetadata struct {
	Name           string                   `json:"name"`
	Models         map[string]ModelMetadata `json:"models"`
	GraphInputs    interface{}              `json:"graphinputs"`
	GraphOutputs   interface{}              `json:"graphoutputs"`
	inputNodeMeta  *ModelMetadata
	outputNodeMeta *ModelMetadata
}

func (gm *GraphMetadata) ToProto() *proto.SeldonGraphMetadata {
	output := &proto.SeldonGraphMetadata{
		Name: gm.Name,
	}
	output.Models = map[string]*proto.SeldonModelMetadata{}
	for name, modelMetadata := range gm.Models {
		output.Models[name] = modelMetadata.ToProto()
	}

	switch gm.inputNodeMeta.ApiVersion {
	case "v1":
		output.Input = gm.inputNodeMeta.ToProto().Input
	case "v2":
		output.Inputs = gm.inputNodeMeta.ToProto().Inputs
	}

	switch gm.outputNodeMeta.ApiVersion {
	case "v1":
		output.Output = gm.outputNodeMeta.ToProto().Output
	case "v2":
		output.Outputs = gm.outputNodeMeta.ToProto().Outputs
	}

	return output
}

func v1ProtoToModelMetadata(meta *proto.SeldonModelMetadata) (*ModelMetadata, error) {
	output := &ModelMetadata{
		ApiVersion: meta.GetApiVersion(),
		Name:       meta.GetName(),
		Platform:   meta.GetPlatform(),
		Versions:   meta.GetVersions(),
		Inputs:     meta.GetInput(),
		Outputs:    meta.GetOutput(),
	}
	return output, nil
}

func v2ProtoToModelMetadata(meta *proto.SeldonModelMetadata) (*ModelMetadata, error) {
	output := &ModelMetadata{
		ApiVersion: meta.GetApiVersion(),
		Name:       meta.GetName(),
		Platform:   meta.GetPlatform(),
		Versions:   meta.GetVersions(),
		Inputs:     meta.GetInputs(),
		Outputs:    meta.GetOutputs(),
	}
	return output, nil
}

func protoToModelMetadata(p payload.SeldonPayload) (*ModelMetadata, error) {
	meta, ok := p.GetPayload().(*proto.SeldonModelMetadata)
	if !ok {
		return nil, errors.New("Wrong Payload")
	}
	switch meta.GetApiVersion() {
	case "v1":
		modelMetadata, err := v1ProtoToModelMetadata(meta)
		return modelMetadata, err
	case "v2":
		modelMetadata, err := v2ProtoToModelMetadata(meta)
		return modelMetadata, err
	default:
		return nil, errors.New("Unsupported ModelMetadata protocol")
	}
}

func jsonToModelMetadata(p payload.SeldonPayload) (*ModelMetadata, error) {
	if p.GetContentType() != "application/json" {
		return nil, errors.New("Expected application/json ContentType")
	}
	resString, err := p.GetBytes()
	if err != nil {
		return nil, err
	}
	var meta ModelMetadata
	err = json.Unmarshal(resString, &meta)
	if err != nil {
		return nil, err
	}
	return &meta, nil
}

func payloadToModelMetadata(p payload.SeldonPayload) (*ModelMetadata, error) {
	fmt.Println("Works 1")

	switch p.GetContentType() {
	case "application/json":
		return jsonToModelMetadata(p)
	case "application/protobuf":
		return protoToModelMetadata(p)
	default:
		return nil, errors.New("Unknown ContentType")
	}
}

func NewGraphMetadata(p *PredictorProcess, spec *v1.PredictorSpec) (*GraphMetadata, error) {
	metadataMap, err := p.MetadataMap(spec.Graph)
	if err != nil {
		return nil, err
	}

	var models = map[string]ModelMetadata{}
	for key, payload := range metadataMap {
		modelMetadata, err := payloadToModelMetadata(payload)
		if err != nil {
			return nil, err
		}
		models[key] = *modelMetadata
	}

	output := &GraphMetadata{
		Name:   spec.Name,
		Models: models,
	}

	output.inputNodeMeta, output.outputNodeMeta = output.getEdgeNodes(spec.Graph)
	output.GraphInputs, output.GraphOutputs = output.getShapeFromGraph(spec.Graph)
	return output, nil
}

func (gm *GraphMetadata) getEdgeNodes(node *v1.PredictiveUnit) (
	input *ModelMetadata, output *ModelMetadata,
) {
	nodeMeta := gm.Models[node.Name]

	// Single node graphs: code path terminates here if this is the case
	if node.Children == nil || len(node.Children) == 0 {
		// We treat node's inputs/outputs as global despite its Type
		return &nodeMeta, &nodeMeta
	}

	// Multi nodes graphs
	if *node.Type == v1.MODEL || *node.Type == v1.TRANSFORMER {
		// Ignore all children except first one for Models and Transformers
		_, childOutput := gm.getEdgeNodes(&node.Children[0])
		return &nodeMeta, childOutput
	} else if *node.Type == v1.OUTPUT_TRANSFORMER {
		// Ignore all children except first one for Output Transformers
		// OUTPUT_TRANSFORMER first passes its input to (first) child and returns the output.
		childInput, _ := gm.getEdgeNodes(&node.Children[0])
		return childInput, &nodeMeta
	} else if *node.Type == v1.COMBINER {
		// Combiner will pass request to all of its children and combine their output.
		// We assume that all children take same type of inputs.
		childInput, _ := gm.getEdgeNodes(&node.Children[0])

		return childInput, &nodeMeta
	} else if *node.Type == v1.ROUTER {
		// ROUTER will pass request to one of its children and return child's output.
		// We assume that all children take same type of inputs.
		childInput, childOutputs := gm.getEdgeNodes(&node.Children[0])
		return childInput, childOutputs
	}

	// If we got here it means none of the cases above
	logger := log.Log.WithName("GraphMetadata")
	logger.Info("Unimplemented case: Couldn't derive graph-level inputs and outputs.")
	return nil, nil
}

func (gm *GraphMetadata) getShapeFromGraph(node *v1.PredictiveUnit) (
	input interface{}, output interface{},
) {
	nodeMeta := gm.Models[node.Name]
	nodeInputs := nodeMeta.Inputs
	nodeOutputs := nodeMeta.Outputs

	// Single node graphs: code path terminates here if this is the case
	if node.Children == nil || len(node.Children) == 0 {
		// We treat node's inputs/outputs as global despite its Type
		return nodeInputs, nodeOutputs
	}

	// Multi nodes graphs
	if *node.Type == v1.MODEL || *node.Type == v1.TRANSFORMER {
		// Ignore all children except first one for Models and Transformers
		_, childOutputs := gm.getShapeFromGraph(&node.Children[0])
		return nodeInputs, childOutputs
	} else if *node.Type == v1.OUTPUT_TRANSFORMER {
		// Ignore all children except first one for Output Transformers
		// OUTPUT_TRANSFORMER first passes its input to (first) child and returns the output.
		childInputs, _ := gm.getShapeFromGraph(&node.Children[0])
		return childInputs, nodeOutputs
	} else if *node.Type == v1.COMBINER {
		// Combiner will pass request to all of its children and combine their output.
		// We assume that all children take same type of inputs.
		childInputs, _ := gm.getShapeFromGraph(&node.Children[0])

		return childInputs, nodeOutputs
	} else if *node.Type == v1.ROUTER {
		// ROUTER will pass request to one of its children and return child's output.
		// We assume that all children take same type of inputs.
		childInputs, childOutputs := gm.getShapeFromGraph(&node.Children[0])
		return childInputs, childOutputs
	}

	// If we got here it means none of the cases above
	logger := log.Log.WithName("GraphMetadata")
	logger.Info("Unimplemented case: Couldn't derive graph-level inputs and outputs.")
	return nil, nil
}
