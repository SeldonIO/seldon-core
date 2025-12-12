@PipelineDeployment @Functional @Pipelines @PipelineInputTensors
Feature: Pipeline using direct input tensors
  This pipeline directly routes pipeline input tensors INPUT0 and INPUT1 into separate models.

  Scenario: Deploy pipeline-inputs pipeline and wait for readiness
    Given I deploy model spec with timeout "30s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: mul10-tw2x
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
      name: add10-tw2x
    spec:
      storageUri: "gs://seldon-models/scv2/samples/triton_23-03/add10"
      requirements:
      - triton
      - python
    """
    Then the model "mul10-tw2x" should eventually become Ready with timeout "20s"
    And the model "add10-tw2x" should eventually become Ready with timeout "20s"

    And I deploy pipeline spec with timeout "30s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: pipeline-inputs-tw2x
    spec:
      steps:
      - name: mul10-tw2x
        inputs:
        - pipeline-inputs-tw2x.inputs.INPUT0
        tensorMap:
          pipeline-inputs-tw2x.inputs.INPUT0: INPUT
      - name: add10-tw2x
        inputs:
        - pipeline-inputs-tw2x.inputs.INPUT1
        tensorMap:
          pipeline-inputs-tw2x.inputs.INPUT1: INPUT
      output:
        steps:
        - mul10-tw2x
        - add10-tw2x
    """
    Then the pipeline "pipeline-inputs-tw2x" should eventually become Ready with timeout "20s"
