@PipelineDeployment @Functional @Pipelines @Conditional
Feature: Conditional pipeline with branching models
  This pipeline uses a conditional model to route data to either add10 or mul10.

  Scenario: Deploy tfsimple-conditional pipeline and wait for readiness
    Given I deploy model spec with timeout "30s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: conditional-nbsl
    spec:
      storageUri: "gs://seldon-models/scv2/samples/triton_23-03/conditional"
      requirements:
      - triton
      - python
    """
    And I deploy model spec with timeout "30s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: add10-nbsl
    spec:
      storageUri: "gs://seldon-models/scv2/samples/triton_23-03/add10"
      requirements:
      - triton
      - python
    """
    And I deploy model spec with timeout "30s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: mul10-nbsl
    spec:
      storageUri: "gs://seldon-models/scv2/samples/triton_23-03/mul10"
      requirements:
      - triton
      - python
    """
    Then the model "conditional-nbsl" should eventually become Ready with timeout "20s"
    And the model "add10-nbsl" should eventually become Ready with timeout "20s"
    And the model "mul10-nbsl" should eventually become Ready with timeout "20s"

    And I deploy pipeline spec with timeout "30s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: tfsimple-conditional-nbsl
    spec:
      steps:
      - name: conditional-nbsl
      - name: mul10-nbsl
        inputs:
        - conditional-nbsl.outputs.OUTPUT0
        tensorMap:
          conditional-nbsl.outputs.OUTPUT0: INPUT
      - name: add10-nbsl
        inputs:
        - conditional-nbsl.outputs.OUTPUT1
        tensorMap:
          conditional-nbsl.outputs.OUTPUT1: INPUT
      output:
        steps:
        - mul10-nbsl
        - add10-nbsl
        stepsJoin: any
    """
    Then the pipeline "tfsimple-conditional-nbsl" should eventually become Ready with timeout "40s"
