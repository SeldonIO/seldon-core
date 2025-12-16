@ModelChaining @Functional @Pipelines
Feature: Pipeline model chaining
  This pipeline chains tfsimple1 into tfsimple2 using tensorMap.

  Scenario: Deploy tfsimples pipeline and wait for readiness
    Given I deploy model spec with timeout "30s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: model-chain-tfsimple1-iuw3
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
      name: model-chain-tfsimple2-iuw3
    spec:
      storageUri: "gs://seldon-models/triton/simple"
      requirements:
      - tensorflow
      memory: 100Ki
    """
    Then the model "model-chain-tfsimple1-iuw3" should eventually become Ready with timeout "20s"
    Then the model "model-chain-tfsimple2-iuw3" should eventually become Ready with timeout "20s"
    When I deploy a pipeline spec with timeout "20s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: model-chain-tfsimples-iuw3
    spec:
      steps:
        - name: model-chain-tfsimple1-iuw3
        - name: model-chain-tfsimple2-iuw3
          inputs:
          - model-chain-tfsimple1-iuw3
          tensorMap:
            model-chain-tfsimple1-iuw3.outputs.OUTPUT0: INPUT0
            model-chain-tfsimple1-iuw3.outputs.OUTPUT1: INPUT1
      output:
        steps:
        - model-chain-tfsimple2-iuw3
    """
    Then the pipeline "model-chain-tfsimples-iuw3" should eventually become Ready with timeout "40s"
    When I send HTTP inference request with timeout "20s" to pipeline "tfsimple-conditional-nbsl" with payload:
    """
    {"model_name":"conditional-nbsl","inputs":[{"name":"CHOICE","contents":{"int_contents":[0]},"datatype":"INT32","shape":[1]},{"name":"INPUT0","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]},{"name":"INPUT1","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}
    """
    And expect http response status code "200"
