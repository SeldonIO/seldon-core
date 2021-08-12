package predictor

import (
	// "github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type GraphMetadata struct {
	Name         string                           `json:"name"`
	Models       map[string]payload.ModelMetadata `json:"models"`
	GraphInputs  interface{}                      `json:"graphinputs"`
	GraphOutputs interface{}                      `json:"graphoutputs"`
}

type MetadataTensor struct {
	DataType string `json:"datatype,omitempty"`
	Name     string `json:"name,omitempty"`
	Shape    []int  `json:"shape,omitempty"`
}

func (gm *GraphMetadata) getEdgeNodes(node *v1.PredictiveUnit) (
	input *payload.ModelMetadata, output *payload.ModelMetadata,
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
