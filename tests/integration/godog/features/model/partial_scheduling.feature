#@ModelPartialScheduling
#Feature: Partial Model Scheduling
#  In order to make a model partially available for inference
#  As a model user
#  I need to create a Model resource with min replicas less than the available server replicas and replicas bigger than the available server replicas and verify that the model deploys the minimum amount of possible model replicas and becomes ready
#
#  Scenario: Success - Load a model with partial replicas
#    Given I have an "iris" model
#    And the model has "1" min replicas
#    And the model has "5" max replicas
#    And the model has "1" replicas
#    When the model is applied
#    Then the model should eventually becomes Ready