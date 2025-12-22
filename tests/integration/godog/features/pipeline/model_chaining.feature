@ModelModelChaining @Functional @Pipelines
Feature: Pipeline model chaining
  In order to compose models that rely on each other's outputs
  As a model user
  I need a pipeline that maps specific output tensors from an upstream model into the inputs of a downstream model

  Scenario: Deploy a chaining pipeline, run inference, and verify the output
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
    When I send HTTP inference request with timeout "20s" to pipeline "model-chain-tfsimples-iuw3" with payload:
    """
    {
      "inputs": [
        {
          "name": "INPUT0",
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
          "datatype": "INT32",
          "shape": [
            1,
            16
          ]
        },
        {
          "name": "INPUT1",
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
          "datatype": "INT32",
          "shape": [
            1,
            16
          ]
        }
      ]
    }
    """
    And expect http response status code "200"
    Then I send gRPC inference request with timeout "20s" to pipeline "model-chain-tfsimples-iuw3" with payload:
    """
    {
      "model_name": "simple",
      "inputs": [
        {
          "name": "INPUT0",
          "contents": {
            "int_contents": [
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
            ]
          },
          "datatype": "INT32",
          "shape": [
            1,
            16
          ]
        },
        {
          "name": "INPUT1",
          "contents": {
            "int_contents": [
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
            ]
          },
          "datatype": "INT32",
          "shape": [
            1,
            16
          ]
        }
      ]
    }

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
        "AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==",
        "AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA=="
      ]
    }
    """
