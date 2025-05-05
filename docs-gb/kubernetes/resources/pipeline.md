# Pipeline

Pipelines allow one to connect flows of inference data transformed by `Model` components. A directed acyclic graph (DAG) of steps can be defined to join Models together. Each Model will need to be capable of receiving a V2 inference request and respond with a V2 inference response. An example Pipeline is shown below:

{% @github-files/github-code-block url="https://github.com/SeldonIO/seldon-core/blob/v2/samples/pipelines/tfsimples-join.yaml" %}

The `steps` list shows three models: `tfsimple1`, `tfsimple2` and `tfsimple3`. These three models each take two tensors called `INPUT0` and `INPUT1` of integers. The models produce two outputs `OUTPUT0` (the sum of the inputs) and `OUTPUT1` (subtraction of the second input from the first).

`tfsimple1` and `tfsimple2` take as inputs the input to the Pipeline: the default assumption when no explicit inputs are defined. `tfsimple3` takes one V2 tensor input from each of the outputs of `tfsimple1` and `tfsimple2`. As the outputs of `tfsimple1` and `tfsimple2` have tensors named `OUTPUT0` and `OUTPUT1` their names need to be changed to respect the expected input tensors and this is done with a `tensorMap` component providing this tensor renaming. This is only required if your models can not be directly chained together.

The output of the Pipeline is the output from the `tfsimple3` model.

## Support for Cyclic Pipelines

Seldon Core 2 supports cyclic pipelines, allowing feedback loops within the inference graph. However, this feature should be used with caution, as improper use may result in infinite loops and unexpected behavior.

The risk of infinite loops arises from the way Kafka Streams performs stream joins. If a feedback message enters the pipeline within the join window interval, and reaches a step that already holds messages from a previous iteration, Kafka Streams may perform a join between messages from different iterations. This can lead to unintended message propagation and potentially an unbounded flow of messages through your Kafka topics.

For more details on how Kafka Streams handles joins and the implications for feedback loops, refer to this [Confluent blog post](https://www.confluent.io/blog/crossing-streams-joins-apache-kafka/).

To enable a cyclic pipeline, set the `allowCycles` flag in your pipeline manifest:
```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: pipeline
spec:
  allowCycles: true
  ...
```

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

	// Dataflow specs
	Dataflow *DataflowSpec `json:"dataflow,omitempty"`

	// Allow cyclic pipeline
	AllowCycles bool `json:"allowCycles,omitempty"`
}

type DataflowSpec struct {
	// Flag to indicate whether the kafka input/output topics
	// should be cleaned up when the model is deleted
	// Default false
	CleanTopicsOnDelete bool `json:"cleanTopicsOnDelete,omitempty"`
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

