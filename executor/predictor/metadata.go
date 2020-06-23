package predictor

import (
	"encoding/json"
	"errors"
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
	Name     string      `json:"name,omitempty"`
	Platform string      `json:"platform,omitempty"`
	Versions []string    `json:"versions,omitempty"`
	Inputs   interface{} `json:"inputs,omitempty"`
	Outputs  interface{} `json:"outputs,omitempty"`
}

func (m *ModelMetadata) ToProto() *proto.SeldonModelMetadata {
	return &proto.SeldonModelMetadata{
		Name:     m.Name,
		Versions: m.Versions,
		Platform: m.Platform,
		Inputs:   m.Inputs.([]*proto.SeldonMessageMetadata),
		Outputs:  m.Outputs.([]*proto.SeldonMessageMetadata),
	}
}

type GraphMetadata struct {
	Name         string                   `json:"name"`
	Models       map[string]ModelMetadata `json:"models"`
	GraphInputs  interface{}              `json:"graphinputs"`
	GraphOutputs interface{}              `json:"graphoutputs"`
}

func (gm *GraphMetadata) ToProto() *proto.SeldonGraphMetadata {
	output := &proto.SeldonGraphMetadata{
		Name:    gm.Name,
		Inputs:  gm.GraphInputs.([]*proto.SeldonMessageMetadata),
		Outputs: gm.GraphOutputs.([]*proto.SeldonMessageMetadata),
	}
	output.Models = map[string]*proto.SeldonModelMetadata{}
	for name, modelMetadata := range gm.Models {
		output.Models[name] = modelMetadata.ToProto()
	}
	return output
}

func protoToModelMetadata(p payload.SeldonPayload) (*ModelMetadata, error) {
	meta, ok := p.GetPayload().(*proto.SeldonModelMetadata)
	if !ok {
		return nil, errors.New("Wrong Payload")
	}
	output := &ModelMetadata{
		Name:     meta.GetName(),
		Platform: meta.GetPlatform(),
		Versions: meta.GetVersions(),
		Inputs:   meta.GetInputs(),
		Outputs:  meta.GetOutputs(),
	}
	return output, nil
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

	inputNodeMeta, outputNodeMeta := output.getEdgeNodes(spec.Graph)
	output.GraphInputs = inputNodeMeta.Inputs
	output.GraphOutputs = outputNodeMeta.Outputs

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
