# Pipeline

Pipelines allow one to connect flows of inference data transformed by `Model` components. A directed acyclic graph (DAG) of steps can be defined to join Models together. Each Model will need to be capable of receiving a V2 inference request and respond with a V2 inference response. An example Pipeline is shown below:

```{literalinclude} ../../../../../../samples/pipelines/tfsimples-join.yaml
:language: yaml
```

The `steps` list shows three models: `tfsimple1`, `tfsimple2` and `tfsimple3`. These three models each take two tensors called `INPUT0` and `INPUT1` of integers. The models produce two outputs `OUTPUT0` (the sum of the inputs) and `OUTPUT1` (subtraction of the second input from the first).

`tfsimple1` and `tfsimple2` take as inputs the input to the Pipeline: the default assumption when no explicit inputs are defined. `tfsimple3` takes one V2 tensor input from each of the outputs of `tfsimple1` and `tfsimple2`. As the outputs of `tfsimple1` and `tfsimple2` have tensors named `OUTPUT0` and `OUTPUT1` their names need to be changed to respect the expected input tensors and this is done with a `tensorMap` component providing this tensor renaming. This is only required if your models can not be directly chained together.

The output of the Pipeline is the output from the `tfsimple3` model.

## Detailed Specification

The full GoLang specification for a Pipeline is shown below:

```{literalinclude} ../../../../../../operator/apis/mlops/v1alpha1/pipeline_types.go
:language: golang
:start-after: // PipelineSpec
:end-before: // PipelineStatus
```

