@PipelineDeployment @Functional @Pipelines @ModelChainingFromInputs
Feature: Pipeline model chaining using inputs and outputs
  This pipeline chains tfsimple1 into tfsimple2 using both inputs and outputs.

  Scenario: Deploy tfsimples-input pipeline and wait for readiness
    Given I deploy model spec with timeout "30s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: chain-from-input-tfsimple1-yhjo
    spec:
      storageUri: "gs://seldon-models/triton/simple"
      requirements:
      - tensorflow
      memory: 100Ki

    """
    And I deploy model spec with timeout "20s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: chain-from-input-tfsimple2-yhjo
    spec:
      storageUri: "gs://seldon-models/triton/simple"
      requirements:
      - tensorflow
      memory: 100Ki
    """
    Then the model "chain-from-input-tfsimple1-yhjo" should eventually become Ready with timeout "20s"
    Then the model "chain-from-input-tfsimple2-yhjo" should eventually become Ready with timeout "20s"

    And I deploy pipeline spec with timeout "30s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: chain-from-input-tfsimples-input-yhjo
    spec:
      steps:
        - name: chain-from-input-tfsimple1-yhjo
        - name: chain-from-input-tfsimple2-yhjo
          inputs:
          - chain-from-input-tfsimple1-yhjo.inputs.INPUT0
          - chain-from-input-tfsimple1-yhjo.outputs.OUTPUT1
          tensorMap:
            chain-from-input-tfsimple1-yhjo.outputs.OUTPUT1: INPUT1
      output:
        steps:
        - chain-from-input-tfsimple2-yhjo
    """
    Then the pipeline "chain-from-input-tfsimples-input-yhjo" should eventually become Ready with timeout "40s"
