@PipelineConditional @Functional @Pipelines @Conditional
Feature: Conditional pipeline with branching models
  In order to support decision-based inference
  As a model user
  I need a conditional pipeline that directs inputs to one of multiple models based on a condition

  Scenario: Deploy a conditional pipeline, run inference, and verify the output
    Given I create model spec with timeout "30s":
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
    And I create model spec with timeout "30s":
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
    And I create model spec with timeout "30s":
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

    When I deploy a pipeline spec with timeout "30s":
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
    Then I send gRPC inference request with timeout "20s" to pipeline "tfsimple-conditional-nbsl" with payload:
    """
    {
      "model_name": "conditional-nbsl",
      "inputs": [
        {
          "name": "CHOICE",
          "contents": {
            "int_contents": [
              0
            ]
          },
          "datatype": "INT32",
          "shape": [
            1
          ]
        },
        {
          "name": "INPUT0",
          "contents": {
            "fp32_contents": [
              1,
              2,
              3,
              4
            ]
          },
          "datatype": "FP32",
          "shape": [
            4
          ]
        },
        {
          "name": "INPUT1",
          "contents": {
            "fp32_contents": [
              1,
              2,
              3,
              4
            ]
          },
          "datatype": "FP32",
          "shape": [
            4
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
          "name": "OUTPUT",
          "datatype": "FP32",
          "shape": [
            4
          ]
        }
      ],
      "raw_output_contents": [
        "AAAgQQAAoEEAAPBBAAAgQg=="
      ]
    }
    """
