@PipelineNoModelsRecovery @Functional @Pipelines
Feature: Pipeline recovery after models created.
  I want to ensure I can create a pipeline before creating the models,
  the pipeline should remain in a Unready state, until the models are added

  Scenario: Create a pipeline before models, expect pipeline to become ready and accept inference
    When I deploy a pipeline spec with timeout "20s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: pipeline-recovery
    spec:
      steps:
        - name: pipeline-recovery-a
        - name: pipeline-recovery-b
          inputs:
          - pipeline-recovery-a
          tensorMap:
            pipeline-recovery-a.outputs.OUTPUT0: INPUT0
            pipeline-recovery-a.outputs.OUTPUT1: INPUT1
      output:
        steps:
        - pipeline-recovery-b
    """
    Given I create model spec with timeout "30s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: pipeline-recovery-a
    spec:
      storageUri: "gs://seldon-models/triton/simple"
      requirements:
      - tensorflow
      memory: 100Ki

    """
    And I create model spec with timeout "20s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: pipeline-recovery-b
    spec:
      storageUri: "gs://seldon-models/triton/simple"
      requirements:
      - tensorflow
      memory: 100Ki
    """
    Then the model "pipeline-recovery-a" should eventually become Ready with timeout "20s"
    Then the model "pipeline-recovery-b" should eventually become Ready with timeout "20s"
    Then the pipeline "pipeline-recovery" should eventually become Ready with timeout "40s"
    When I send HTTP inference request with timeout "20s" to pipeline "pipeline-recovery" with payload:
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
    Then I send gRPC inference request with timeout "20s" to pipeline "pipeline-recovery" with payload:
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
