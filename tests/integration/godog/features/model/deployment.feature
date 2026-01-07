@ModelDeployment @Functional @Models
Feature: Model deployment
  In order to make a model available for inference
  As a model user
  I need to create a Model resource and verify it is deployed

  Scenario Outline: Success - Load a <model> model
    Given I have an "<model>" model
    When the model is applied
    Then the model should eventually become Ready

    Examples:
      | model         |
      | iris          |
      | income-xgb    |
      | mnist-onnx    |
      | income-lgb    |
      | wine          |
      | mnist-pytorch |
      | tfsimple1     |


  Scenario: Success - Load a model and expect status model available
    Given I have an "iris" model
    When the model is applied
    And the model eventually becomes Ready
    Then the model status message should eventually be "ModelAvailable"


  Scenario: Load a specific model
    Given I create model spec with timeout "10s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: deployment-test-1
    spec:
      replicas: 1
      requirements:
      - sklearn
      - mlserver
      storageUri: gs://seldon-models/scv2/samples/mlserver_1.3.5/iris-sklearn
    """
    Then the model should eventually become Ready


