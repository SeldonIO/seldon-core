@ModelDeployment @Functional @Models
Feature: Model deployment
  In order to make a model available for inference
  As a model user
  I need to create a Model resource and verify it is deployed

  Scenario: Success - Load a model
    Given I have an "iris" model
    When the model is applied
    Then the model should eventually become Ready


  Scenario: Success - Load a model again
    Given I have an "iris" model
    When the model is applied
    Then the model should eventually become Ready

  Scenario: Load a specific model
    Given I deploy model spec with timeout "10s":
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

  Scenario: Success - Load a model and expect status model available
    Given I have an "iris" model
    When the model is applied
    And the model eventually becomes Ready
    Then the model status message should eventually be "ModelAvailable"

  Scenario: Success - Load a model with min replicas
    Given I have an "iris" model
    And the model has "1" min replicas
    When the model is applied
    Then the model should eventually become Ready

# todo: change model type
  Scenario: Success - Load a big model
    Given I have an "iris" model
    When the model is applied
    Then the model should eventually become Ready


