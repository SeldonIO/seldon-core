@PipelineModelChainingFromInputs @Functional @Pipelines @ModelChainingFromInputs
Feature: Pipeline model chaining using inputs and outputs
  In order to build multi-stage inference workflows
  As a model user
  I need a pipeline that chains models together by passing outputs from one stage into the next

  Scenario: Scenario: Deploy a model-chaining pipeline, run inference, and verify the output
    Given I deploy model spec with timeout "30s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: tfsimple1-yhjo
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
      name: tfsimple2-yhjo
    spec:
      storageUri: "gs://seldon-models/triton/simple"
      requirements:
      - tensorflow
      memory: 100Ki
    """
    Then the model "tfsimple1-yhjo" should eventually become Ready with timeout "20s"
    Then the model "tfsimple2-yhjo" should eventually become Ready with timeout "20s"
    And I deploy a pipeline spec with timeout "30s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: chain-from-input-yhjo
    spec:
      steps:
        - name: tfsimple1-yhjo
        - name: tfsimple2-yhjo
          inputs:
          - tfsimple1-yhjo.inputs.INPUT0
          - tfsimple1-yhjo.outputs.OUTPUT1
          tensorMap:
            tfsimple1-yhjo.outputs.OUTPUT1: INPUT1
      output:
        steps:
        - tfsimple2-yhjo
    """
    Then the pipeline "chain-from-input-yhjo" should eventually become Ready with timeout "40s"
    When I send HTTP inference request with timeout "20s" to pipeline "chain-from-input-yhjo" with payload:
    """
    {"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}
    """
    And expect http response status code "200"
    And expect http response body to contain JSON:
    """
    {
      "model_name": "",
      "outputs": [
        {
          "data": [
            1,
            2,
            3,
            4,
            5,
            6,
            7,
            8,
            9,
            10,
            11,
            12,
            13,
            14,
            15,
            16
          ],
          "name": "OUTPUT0",
          "shape": [
            1,
            16
          ],
          "datatype": "INT32"
        },
        {
          "data": [
            1,
            2,
            3,
            4,
            5,
            6,
            7,
            8,
            9,
            10,
            11,
            12,
            13,
            14,
            15,
            16
          ],
          "name": "OUTPUT1",
          "shape": [
            1,
            16
          ],
          "datatype": "INT32"
        }
      ]
    }
    """
    Then I send gRPC inference request with timeout "20s" to pipeline "chain-from-input-yhjo" with payload:
    """
    {"model_name":"simple","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}
    """
    And expect gRPC response body to contain JSON:
    """
    {
      "outputs": [
        {
          "name": "OUTPUT0",
          "datatype": "INT32",
          "shape": [
            1,
            16
          ]
        },
        {
          "name": "OUTPUT1",
          "datatype": "INT32",
          "shape": [
            1,
            16
          ]
        }
      ],
      "raw_output_contents": [
        "AQAAAAIAAAADAAAABAAAAAUAAAAGAAAABwAAAAgAAAAJAAAACgAAAAsAAAAMAAAADQAAAA4AAAAPAAAAEAAAAA==",
        "AQAAAAIAAAADAAAABAAAAAUAAAAGAAAABwAAAAgAAAAJAAAACgAAAAsAAAAMAAAADQAAAA4AAAAPAAAAEAAAAA=="
      ]
    }
    """
