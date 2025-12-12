@PipelineDeployment @Functional @Pipelines @ModelJoin
Feature: Pipeline model join
  This pipeline joins outputs from tfsimple1 and tfsimple2 and feeds them into tfsimple3.

  Scenario: Deploy tfsimples-join pipeline and wait for readiness
    Given I deploy model spec with timeout "30s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: join-tfsimple1-w4e3
    spec:
      storageUri: "gs://seldon-models/triton/simple"
      requirements:
      - tensorflow
      memory: 100Ki
    """
    And I deploy model spec with timeout "30s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: join-tfsimple2-w4e3
    spec:
      storageUri: "gs://seldon-models/triton/simple"
      requirements:
      - tensorflow
      memory: 100Ki
    """
    And I deploy model spec with timeout "30s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: join-tfsimple3-w4e3
    spec:
      storageUri: "gs://seldon-models/triton/simple"
      requirements:
      - tensorflow
      memory: 100Ki
    """
    Then the model "join-tfsimple1-w4e3" should eventually become Ready with timeout "20s"
    And the model "join-tfsimple2-w4e3" should eventually become Ready with timeout "20s"
    And the model "join-tfsimple3-w4e3" should eventually become Ready with timeout "20s"

    And I deploy pipeline spec with timeout "30s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: join-pipeline-w4e3
    spec:
      steps:
        - name: join-tfsimple1-w4e3
        - name: join-tfsimple2-w4e3
        - name: join-tfsimple3-w4e3
          inputs:
          - join-tfsimple1-w4e3.outputs.OUTPUT0
          - join-tfsimple2-w4e3.outputs.OUTPUT1
          tensorMap:
            join-tfsimple1-w4e3.outputs.OUTPUT0: INPUT0
            join-tfsimple2-w4e3.outputs.OUTPUT1: INPUT1
      output:
        steps:
        - join-tfsimple3-w4e3
    """
    Then the pipeline "join-pipeline-w4e3" should eventually become Ready with timeout "40s"
