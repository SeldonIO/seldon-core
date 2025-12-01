@ModelDeployment @Functional @Models
Feature: Model deployment
  In order to make a model available for inference
  As a model user
  I need to create a Model resource and verify it becomes Ready

#  Background:
#    Given a running Seldon Core 2 control plane
#    And a clean namespace "<test-namespace>"
#    And a new server with capabilties "mlserver" and min replicas 1 and replicas 3 and max replicas 5
#  todo: this will be quite expensive to run per scenario might better to run at the top level when executing a test case

  Scenario: Success - Load a model
    Given I have a model "My-model"
    And the model has 1 min replicas
    And the model has 5 max replicas
    And the model has 1 replicas
    When the model is applied
    Then the model should be Ready
    And the model status message should be "ModelAvailable"

  Scenario: Success - Load a model with min replicas
    Given I have a model named "min-replicas-model"
    And the model has 1 min replicas
    And the model has no explicit max replicas
    And the model has no explicit replicas
    When the model is applied
    Then the model should be Ready
    And the model should have 1 available replicas

  Scenario: Success - Load a big model
    Given I have a model with name "big-model"
    And the model has 1 min replicas
    When the model is applied
    Then the model should be Ready

  Scenario: Fail Load Model - no server capabilities in cluster
    Given I have a model "unsupported-model" requiring capabilities "xgboost"
    And there is no server in the cluster with capabilities "xgboost"
    When the model is applied
    Then the model should not be Ready
    And the model status message should be "ModelFailed"