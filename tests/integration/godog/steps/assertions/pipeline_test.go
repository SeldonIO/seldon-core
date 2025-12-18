package assertions

import (
	"testing"

	"github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

func TestPipelineReady(t *testing.T) {
	pipelineYAML := `
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: model-chain-tfsimples-iuw3
  namespace: seldon-mesh
status:
  conditions:
  - type: ModelsReady
    status: "True"
  - type: PipelineGwReady
    status: "True"
  - type: PipelineReady
    status: "True"
  - type: Ready
    status: "True"
`

	pipeline := &v1alpha1.Pipeline{}

	// Unmarshal YAML into the Pipeline struct
	err := yaml.Unmarshal([]byte(pipelineYAML), pipeline)
	require.NoError(t, err)

	// Call the function under test
	ready, err := PipelineReady(pipeline)

	// Assertions
	require.NoError(t, err)
	require.True(t, ready)
}

func TestPipelineNotReady(t *testing.T) {
	pipelineYAML := `
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: model-chain-tfsimples-iuw3
  namespace: seldon-mesh
status:
  conditions:
  - type: ModelsReady
    status: "False"
  - type: PipelineGwReady
    status: "True"
  - type: PipelineReady
    status: "True"
  - type: Ready
    status: "False"
`

	pipeline := &v1alpha1.Pipeline{}

	// Unmarshal YAML into the Pipeline struct
	err := yaml.Unmarshal([]byte(pipelineYAML), pipeline)
	require.NoError(t, err)

	// Call the function under test
	ready, err := PipelineReady(pipeline)

	// Assertions
	require.NoError(t, err)
	require.False(t, ready)
}
