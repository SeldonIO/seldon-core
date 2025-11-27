@ModelDeployment
Feature: Model deployment

  Background:
    Given a running Seldon Core 2 control plane
    And a clean namespace "could be an env var"
    And a new server with capabilties "tensor-flow" and min replicas "1" and replicas "3" and max replicas "5"

  Scenario: Success - Load a model
    Given I have a model
    And the model has "" min replicas
    And the model has "" max replicas
    And the model has "" replicas
    When the model is loaded
    Then the model should be succesfully deployed

  Scenario: Success - Load a model with min replicas
    Given I have a model
    And the model has "" min replicas
    When the model is loaded
    Then the model should be succesfully deployed

  Scenario: Success - Load a big model
    Given I have a model with name "" and storage uri ""
    And the model has "" min replicas
    When the model is loaded
    Then the model should be succesfully deployed

  Scenario: Fail Load Model - no server capabilities in cluster
    Given I have a model
    And the model has "" capabilities
    When the model is loaded
    Then the model should fail to schedule in a server