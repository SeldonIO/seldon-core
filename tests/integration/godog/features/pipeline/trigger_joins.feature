@PipelineDeployment @Functional @Pipelines @TriggerJoins
Feature: Pipeline using trigger joins
  This pipeline uses trigger joins to decide whether mul10 or add10 should run.

  Scenario: Deploy trigger-joins pipeline and wait for readiness
    Given I deploy model spec with timeout "30s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: mul10-99lo
    spec:
      storageUri: "gs://seldon-models/scv2/samples/triton_23-03/mul10"
      requirements:
      - triton
      - python
    """
    And I deploy model spec with timeout "30s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: add10-99lo
    spec:
      storageUri: "gs://seldon-models/scv2/samples/triton_23-03/add10"
      requirements:
      - triton
      - python
    """
    Then the model "mul10-99lo" should eventually become Ready with timeout "20s"
    And the model "add10-99lo" should eventually become Ready with timeout "20s"
    And I deploy a pipeline spec with timeout "30s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: trigger-joins-99lo
    spec:
      steps:
      - name: mul10-99lo
        inputs:
        - trigger-joins-99lo.inputs.INPUT
        triggers:
        - trigger-joins-99lo.inputs.ok1
        - trigger-joins-99lo.inputs.ok2
        triggersJoinType: any
      - name: add10-99lo
        inputs:
        - trigger-joins-99lo.inputs.INPUT
        triggers:
        - trigger-joins-99lo.inputs.ok3
      output:
        steps:
        - mul10-99lo
        - add10-99lo
        stepsJoin: any
    """
    Then the pipeline "trigger-joins-99lo" should eventually become Ready with timeout "20s"
    When I send gRPC inference request with timeout "20s" to pipeline "tfsimple-conditional-nbsl" with payload:
    """
    {"model_name":"pipeline","inputs":[{"name":"ok1","contents":{"fp32_contents":[1]},"datatype":"FP32","shape":[1]},{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}
    """
    And expect gRPC response error to contain "Unimplemented"
