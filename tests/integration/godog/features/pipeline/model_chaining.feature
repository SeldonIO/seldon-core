@ModelChaining @Functional @Pipelines
Feature: Pipeline model chaining
  This pipeline chains tfsimple1 into tfsimple2 using tensorMap.

  Scenario: Deploy tfsimples pipeline and wait for readiness
    Given I deploy model spec with timeout "30s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: tfsimple1
    spec:
      storageUri: "gs://seldon-models/triton/simple"
      requirements:
      - tensorflow
      memory: 100Ki

    """
    Then the model "tfsimple1" should eventually become Ready with timeout "20s"
    And I deploy model spec with timeout "20s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: tfsimple2
    spec:
      storageUri: "gs://seldon-models/triton/simple"
      requirements:
      - tensorflow
      memory: 100Ki
    """
    Then the model "tfsimple2" should eventually become Ready with timeout "20s"
    When I deploy pipeline spec with timeout "20s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: tfsimples
    spec:
      steps:
        - name: tfsimple1
        - name: tfsimple2
          inputs:
          - tfsimple1
          tensorMap:
            tfsimple1.outputs.OUTPUT0: INPUT0
            tfsimple1.outputs.OUTPUT1: INPUT1
      output:
        steps:
        - tfsimple2
    """
    Then the pipeline should eventually become Ready with timeout "20s"
