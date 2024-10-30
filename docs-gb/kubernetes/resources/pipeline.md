# Pipeline

Pipelines allow one to connect flows of inference data transformed by `Model` components. A directed acyclic graph (DAG) of steps can be defined to join Models together. Each Model will need to be capable of receiving a V2 inference request and respond with a V2 inference response. An example Pipeline is shown below:

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: join
spec:
  steps:
    - name: tfsimple1
    - name: tfsimple2
    - name: tfsimple3      
      inputs:
      - tfsimple1.outputs.OUTPUT0
      - tfsimple2.outputs.OUTPUT1
      tensorMap:
        tfsimple1.outputs.OUTPUT0: INPUT0
        tfsimple2.outputs.OUTPUT1: INPUT1
  output:
    steps:
    - tfsimple3
```

The `steps` list shows three models: `tfsimple1`, `tfsimple2` and `tfsimple3`. These three models each take two tensors called `INPUT0` and `INPUT1` of integers. The models produce two outputs `OUTPUT0` (the sum of the inputs) and `OUTPUT1` (subtraction of the second input from the first).

`tfsimple1` and `tfsimple2` take as inputs the input to the Pipeline: the default assumption when no explicit inputs are defined. `tfsimple3` takes one V2 tensor input from each of the outputs of `tfsimple1` and `tfsimple2`. As the outputs of `tfsimple1` and `tfsimple2` have tensors named `OUTPUT0` and `OUTPUT1` their names need to be changed to respect the expected input tensors and this is done with a `tensorMap` component providing this tensor renaming. This is only required if your models can not be directly chained together.

The output of the Pipeline is the output from the `tfsimple3` model.

## Detailed Specification

The full GoLang specification for a Pipeline is shown below:

```go
type PipelineSpec struct {
	// External inputs to this pipeline, optional
	Input *PipelineInput `json:"input,omitempty"`

	// The steps of this inference graph pipeline
	Steps []PipelineStep `json:"steps"`

	// Synchronous output from this pipeline, optional
	Output *PipelineOutput `json:"output,omitempty"`
}

// +kubebuilder:validation:Enum=inner;outer;any
type JoinType string

const (
	// data must be available from all inputs
	JoinTypeInner JoinType = "inner"
	// data will include any data from any inputs at end of window
	JoinTypeOuter JoinType = "outer"
	// first data input that arrives will be forwarded
	JoinTypeAny JoinType = "any"
)

type PipelineStep struct {
	// Name of the step
	Name string `json:"name"`

	// Previous step to receive data from
	Inputs []string `json:"inputs,omitempty"`

	// msecs to wait for messages from multiple inputs to arrive before joining the inputs
	JoinWindowMs *uint32 `json:"joinWindowMs,omitempty"`

	// Map of tensor name conversions to use e.g. output1 -> input1
	TensorMap map[string]string `json:"tensorMap,omitempty"`

	// Triggers required to activate step
	Triggers []string `json:"triggers,omitempty"`

	// +kubebuilder:default=inner
	InputsJoinType *JoinType `json:"inputsJoinType,omitempty"`

	TriggersJoinType *JoinType `json:"triggersJoinType,omitempty"`

	// Batch size of request required before data will be sent to this step
	Batch *PipelineBatch `json:"batch,omitempty"`
}

type PipelineBatch struct {
	Size     *uint32 `json:"size,omitempty"`
	WindowMs *uint32 `json:"windowMs,omitempty"`
	Rolling  bool    `json:"rolling,omitempty"`
}

type PipelineInput struct {
	// Previous external pipeline steps to receive data from
	ExternalInputs []string `json:"externalInputs,omitempty"`

	// Triggers required to activate inputs
	ExternalTriggers []string `json:"externalTriggers,omitempty"`

	// msecs to wait for messages from multiple inputs to arrive before joining the inputs
	JoinWindowMs *uint32 `json:"joinWindowMs,omitempty"`

	// +kubebuilder:default=inner
	JoinType *JoinType `json:"joinType,omitempty"`

	// +kubebuilder:default=inner
	TriggersJoinType *JoinType `json:"triggersJoinType,omitempty"`

	// Map of tensor name conversions to use e.g. output1 -> input1
	TensorMap map[string]string `json:"tensorMap,omitempty"`
}

type PipelineOutput struct {
	// Previous step to receive data from
	Steps []string `json:"steps,omitempty"`

	// msecs to wait for messages from multiple inputs to arrive before joining the inputs
	JoinWindowMs uint32 `json:"joinWindowMs,omitempty"`

	// +kubebuilder:default=inner
	StepsJoin *JoinType `json:"stepsJoin,omitempty"`

	// Map of tensor name conversions to use e.g. output1 -> input1
	TensorMap map[string]string `json:"tensorMap,omitempty"`
}
```

