@ModelDeployment @Functional @Models
Feature: Model deployment
  In order to make a model available for inference
  As a model user
  I need to create a Model resource and verify it becomes Ready



  Scenario: Success - Load a model with min replicas
    Given I have an "iris" model
    And the model has 1 min replicas
    And the model has no explicit max replicas
    And the model has no explicit replicas
    When the model is applied
    Then the model should be Ready

  Scenario: Success - Load a big model
    Given I have an "large-model" model
    When the model is applied
    Then the model should be Ready

#    this would belong more to the feature of model server scheduling or capabilities
  Scenario: Fail Load Model - no server capabilities in cluster
    Given Given I have an "iris" model
    And the model has "xgboost" capabilities
    And there is no server in the cluster with capabilities "xgboost"
    When the model is applied
    Then the model should not be Ready
    And the model status message should be "ModelFailed"

