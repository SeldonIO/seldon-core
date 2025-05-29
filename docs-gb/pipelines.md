---
description: Learn how to create and manage ML inference pipelines in Seldon Core, including model chaining, tensor mapping, and conditional logic.
---

# Pipelines

Pipelines allow models to be connected into flows of data transformations. This allows more
complex machine learning pipelines to be created with multiple models, feature transformations
and monitoring components such as drift and outlier detectors.

## Creating Pipelines

The simplest way to create Pipelines is by defining them with the
[Pipeline resource we provide for Kubernetes](./kubernetes/resources/pipeline.md). This format
is accepted by our Kubernetes implementation but also locally via our `seldon` CLI.

Internally in both cases Pipelines are created via our [Scheduler API](./apis/scheduler.md). Advanced
users could submit Pipelines directly using this gRPC service.

An example that chains two models together is shown below:

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: chain
  namespace: seldon-mesh
spec:
  steps:
    - name: model1
    - name: model2
      inputs:
      - model1
  output:
    steps:
    - model2
```

* `steps` allow you to specify the models you want to combine into a pipeline. Each step name will
correspond to a model of the same name. These models will need to have been deployed and available
for the Pipeline to function, however Pipelines can be deployed before or at the same time you deploy
the underlying models.
* `steps.inputs` allow you to specify the inputs to this step.
* `outputs.steps` allow you to specify the output of the Pipeline. A pipeline can have multiple paths
include flows of data that do not reach the output, e.g. Drift detection steps. However, if you wish
to call your Pipeline in a synchronous manner via REST/gRPC then an output must be present so the
Pipeline can be treated as a function.

## Expressing input data sources


Model step inputs are defined with a dot notation of the form:

```sh
<stepName>|<pipelineName>.<inputs|outputs>.<tensorName>
```

Inputs with just a step name will be assumed to be `step.outputs`.

The default payloads for Pipelines is the V2 protocol which requires named tensors as inputs and outputs
from a model. If you require just certain tensors from a model you can reference those in the inputs,
e.g. `mymodel.outputs.t1` will reference the tensor `t1` from the model `mymodel`.

For the specification of the [V2 protocol](apis/inference/README.md).

## Chain

The simplest Pipeline chains models together: the output of one model goes into the input of the next.
This will work out of the box if the output tensor names from a model match the input tensor names for
the one being chained to. If they do not then the `tensorMap` construct presently needs to be used to
define the mapping explicitly, e.g. see below for a simple chained pipeline of two tfsimple example models:

{% embed url="https://github.com/SeldonIO/seldon-core/blob/v2/samples/pipelines/tfsimples.yaml" %}

```mermaid
flowchart LR
    classDef pipeIO fill:#F6E083

    subgraph input
        INPUT0:::pipeIO
        INPUT1:::pipeIO
    end

    INPUT0 --> TF1(tfsimple1)
    INPUT1 --> TF1
    TF1 --->|OUTPUT0: INPUT0| TF2(tfsimple2)
    TF1 --->|OUTPUT1: INPUT1| TF2

    subgraph output
        OUTPUT0:::pipeIO
        OUTPUT1:::pipeIO
    end

    TF2 --> OUTPUT0
    TF2 --> OUTPUT1
```

In the above we rename tensor `OUTPUT0` to `INPUT0` and `OUTPUT1` to `INPUT1`. This allows these models to
be chained together. The shape and data-type of the tensors needs to match as well.

This example can be found in the [pipeline examples](examples/pipeline-examples.md#model-chaining).

## Join

Joining allows us to combine outputs from multiple steps as input to a new step.

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

```mermaid
flowchart LR
    classDef pipeIO fill:#F6E083
    classDef hidden fill:#ffffff,stroke:#ffffff

    subgraph input
        INPUT0:::pipeIO
        INPUT1:::pipeIO
    end

    INPUT0 --> TF1(tfsimple1)
    INPUT1 --> TF1
    INPUT0 --> TF2(tfsimple2)
    INPUT1 --> TF2
    TF1 -.-> |OUTPUT1| tf1( ):::hidden
    TF1 --> |OUTPUT0: INPUT0| TF3(tfsimple3)
    TF2 -.-> |OUTPUT0| tf2( ):::hidden
    TF2 --> |OUTPUT1: INPUT1| TF3

    subgraph output
        OUTPUT0:::pipeIO
        OUTPUT1:::pipeIO
    end
    TF3 --> OUTPUT0
    TF3 --> OUTPUT1
```    

Caption: "*Joining the outputs of two models into a third model. The dashed lines signify model outputs that are not captured in the output of the pipeline.*"

Here we pass the pipeline inputs to two models and then take one output tensor from each and pass to the
final model. We use the same `tensorMap` technique to rename tensors as disucssed in the previous section.

Joins can have a join type which can be specified with `inputsJoinType` and can take the values:
* `inner`: require all inputs to be available to join.
* `outer`: wait for `joinWindowMs` to join any inputs. Ignoring any inputs that have not sent any data at that
point. This will mean this step of the pipeline is guaranteed to have a latency of at least `joinWindowMs`.
* `any`: wait for any of the specified data sources.

This example can be found in the [pipeline examples](examples/pipeline-examples.md#model-join).

## Conditional Logic

Pipelines can create conditional flows via various methods. We will discuss each in turn.

### Model routing via tensors

The simplest way is to create a model that outputs different named tensors based on its decision. This way downstream
steps can be dependant on different expected tensors. An example is shown below:

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: tfsimple-conditional
spec:
  steps:
  - name: conditional
  - name: mul10
    inputs:
    - conditional.outputs.OUTPUT0
    tensorMap:
      conditional.outputs.OUTPUT0: INPUT
  - name: add10
    inputs:
    - conditional.outputs.OUTPUT1
    tensorMap:
      conditional.outputs.OUTPUT1: INPUT
  output:
    steps:
    - mul10
    - add10
    stepsJoin: any
```

```mermaid
flowchart LR
    classDef pipeIO fill:#F6E083
    classDef join fill:#CEE741;

    subgraph input
    INPUT0:::pipeIO
        CHOICE:::pipeIO
        INPUT0:::pipeIO
        INPUT1:::pipeIO
    end

    CHOICE --> conditional
    INPUT0 --> conditional
    INPUT1 --> conditional

    conditional --> |OUTPUT0: INPUT| add10
    conditional --> |OUTPUT1: INPUT| mul10

    add10 --> |OUTPUT| any(any):::join
    mul10 --> |OUTPUT| any

    subgraph output
        OUTPUT(OUTPUT):::pipeIO
    end

    any --> OUTPUT

    linkStyle 3 stroke:blue,color:blue;
    linkStyle 5 stroke:blue,color:blue;
    linkStyle 4 stroke:red,color:red;
    linkStyle 6 stroke:red,color:red;
```
Caption: "*Pipeline with a conditional output model. The model **conditional** only outputs one of the two tensors, so only one path through the graph (red or blue) is taken by a single request*"

In the above we have a step `conditional` that either outputs a tensor named `OUTPUT0` or a tensor named `OUTPUT1`.
The `mul10` step depends on an output in `OUTPUT0` while the add10 step depends on an output from `OUTPUT1`.

Note, we also have a final Pipeline output step that does an `any` join on these two models essentially outputting
fron the pipeline whichever data arrives from either model. This type of Pipeline can be used for Multi-Armed bandit
solutions where you want to route traffic dynamically.

This example can be found in the [pipeline examples](examples/pipeline-examples.md#conditional).

### Errors

Its also possible to abort pipelines when an error is produced to in effect create a condition. This is illustrated below:

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: error
spec:
  steps:
    - name: outlier-error
  output:
    steps:
    - outlier-error
```

This Pipeline runs normally or throws an error based on whether the input tensors have certain values.

### Triggers

Sometimes you want to run a step if an output is received from a previous step but not to send the data from
that step to the model. This is illustrated below:

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: joincheck
spec:
  steps:
    - name: tfsimple1
    - name: tfsimple2
    - name: check
      inputs:
      - tfsimple1.outputs.OUTPUT0
      tensorMap:
        tfsimple1.outputs.OUTPUT0: INPUT
    - name: tfsimple3
      inputs:
      - tfsimple1.outputs.OUTPUT0
      - tfsimple2.outputs.OUTPUT1
      tensorMap:
        tfsimple1.outputs.OUTPUT0: INPUT0
        tfsimple2.outputs.OUTPUT1: INPUT1
      triggers:
      - check.outputs.OUTPUT
  output:
    steps:
    - tfsimple3
```

```mermaid
flowchart LR
  classDef pipeIO fill:#F6E083
  classDef hidden fill:#ffffff,stroke:#ffffff

  subgraph input
      INPUT0:::pipeIO
      INPUT1:::pipeIO
  end

  INPUT0 --> TF1(tfsimple1)
  INPUT1 --> TF1
  INPUT0 --> TF2(tfsimple2)
  INPUT1 --> TF2
  TF1 -.-> |OUTPUT1| tf1( ):::hidden

  TF1 --> |OUTPUT0: INPUT| check
  TF1 --> |OUTPUT0: INPUT0| TF3(tfsimple3)
  TF2 --> |OUTPUT1: INPUT1| TF3
  TF2 -.-> |OUTPUT0| tf2( ):::hidden

  check --o |OUTPUT| TF3
  linkStyle 9 stroke:#CEE741,color:black;

  subgraph output
    OUTPUT0:::pipeIO
    OUTPUT1:::pipeIO
  end

  TF3 --> OUTPUT0
  TF3 --> OUTPUT1
```
Caption: "*A pipeline with a single trigger. The model **tfsimple3** only runs if the model **check** returns a tensor named `OUTPUT`. The green edge signifies that this is a trigger and not an additional input to **tfsimple3**. The dashed lines signify model outputs that are not captured in the output of the pipeline.*"

In this example the last step `tfsimple3` runs only if there are outputs from `tfsimple1` and `tfsimple2` but also
data from the `check` step. However, if the step `tfsimple3` is run it only receives the join of data from `tfsimple1` and `tfsimple2`.

This example can be found in the [pipeline examples](examples/pipeline-examples.md#model-join-with-trigger).

### Trigger Joins

You can also define multiple triggers which need to happen based on a particulr join type. For example:

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: trigger-joins
spec:
  steps:
  - name: mul10
    inputs:
    - trigger-joins.inputs.INPUT
    triggers:
    - trigger-joins.inputs.ok1
    - trigger-joins.inputs.ok2
    triggersJoinType: any
  - name: add10
    inputs:
    - trigger-joins.inputs.INPUT
    triggers:
    - trigger-joins.inputs.ok3
  output:
    steps:
    - mul10
    - add10
    stepsJoin: any
```

```mermaid
flowchart LR
    classDef pipeIO fill:#F6E083
    classDef pipeIOopt fill:#F6E083,stroke-dasharray: 5 5;
    classDef trigger fill:#CEE741;
    classDef hidden fill:#ffffff,stroke:#ffffff
    classDef join fill:#CEE741;

    subgraph input
        ok1:::pipeIOopt
        ok2:::pipeIOopt
        INPUT:::pipeIO
        ok3:::pipeIOopt
    end

    ok1 --o any
    ok2 --o any
    any((any)):::trigger --o mul10
    linkStyle 0 stroke:#CEE741,color:green;
    linkStyle 1 stroke:#CEE741,color:green;
    linkStyle 2 stroke:#CEE741,color:green;

    INPUT --> mul10
    INPUT --> add10

    ok3 --o add10
    linkStyle 5 stroke:#CEE741,color:green;

    subgraph output
      OUTPUT:::pipeIO
    end

    mul10 -->|OUTPUT| anyOut(any):::join
    add10 --> |OUTPUT| anyOut
    anyOut --> OUTPUT
```
Caption: "*A pipeline with multiple triggers and a trigger join of type `any`. The pipeline has four inputs, but three of these are optional (signified by the dashed borders).*"

Here the `mul10` step is run if data is seen on the pipeline inputs in the `ok1` or `ok2` tensors based on
the `any` join type. If data is seen on `ok3` then the `add10` step is run.

If we changed the `triggersJoinType` for `mul10` to `inner` then both `ok1` and `ok2` would need to appear
before `mul10` is run.

### Pipeline Inputs

Pipelines by default can be accessed synchronously via http/grpc or asynchronously via the Kafka topic
created for them. However, it's also possible to create a pipeline to take input from one or more other
pipelines by specifying an `input` section. If for example we already have the `tfsimple` pipeline shown below:

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: tfsimple
spec:
  steps:
    - name: tfsimple1
  output:
    steps:
    - tfsimple1
```

We can create another pipeline which takes its input from this pipeline, as shown below:

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: tfsimple-extended
spec:
  input:
    externalInputs:
      - tfsimple.outputs
    tensorMap:
      tfsimple.outputs.OUTPUT0: INPUT0
      tfsimple.outputs.OUTPUT1: INPUT1
  steps:
    - name: tfsimple2
  output:
    steps:
    - tfsimple2
```

```mermaid
flowchart LR
    classDef pipeIO fill:#F6E083

    subgraph tfsimple.inputs
        INPUT0:::pipeIO
        INPUT1:::pipeIO
    end

    INPUT0 --> TF1(tfsimple1)
    INPUT1 --> TF1

    subgraph tfsimple.outputs
        OUTPUT0:::pipeIO
        OUTPUT1:::pipeIO
    end

    TF1 --> OUTPUT0
    TF1 --> OUTPUT1

    subgraph tfsimple-extended.inputs
      INPUT10(INPUT0):::pipeIO
      INPUT11(INPUT1):::pipeIO
    end

  OUTPUT0 --> INPUT10
  OUTPUT1 --> INPUT11

  INPUT10 --> TF2(tfsimple2)
  INPUT11 --> TF2

    subgraph tfsimple-extended.outputs
        OUTPUT10(OUTPUT0):::pipeIO
        OUTPUT11(OUTPUT1):::pipeIO
    end

    TF2 --> OUTPUT10
    TF2 --> OUTPUT11
```
Caption: "*A pipeline taking as input the output of another pipeline.*"

In this way pipelines can be built to extend existing running pipelines to allow extensibility and sharing of data flows.

The spec follows the same spec for a step except that references to other pipelines are contained in
the `externalInputs` section which takes the form of pipeline or pipeline.step references:
* `<pipelineName>.(inputs|outputs).<tensorName>`
* `<pipelineName>.(step).<stepName>.<tensorName>`

Tensor names are optional and only needed if you want to take just one tensor from an input or output.

There is also an `externalTriggers` section which allows triggers from other pipelines.

Further examples can be found in the [pipeline-to-pipeline examples](examples/pipeline-to-pipeline.md).

Present caveats:
* Circular dependencies are not presently detected.
* Pipeline status is local to each pipeline.

## Data Centric Implementation

Internally Pipelines are implemented using Kafka. Each input and output to a pipeline step has an
associated Kafka topic. This has many advantages and allows auditing, replay and debugging easier as data
is preserved from every step in your pipeline.

Tracing allows you to monitor the processing latency of your pipelines.

![tracing](images/jaeger-tracing.png)

As each request to a pipelines moves through the steps its data will appear in input and output topics. This
allows a full audit of every transformation to be carried out.
