@ModelInference @Models @Inference
Feature Basic model inferencing

  Background:
    Given a clean test namespace

  Scenario: Model can serve prediction
    Given I have an "iris" model
    And the model is applied
    And the model eventually becomes Ready
    When I send a prediction request with payload:
      """
      { "inputs": [1.0, 2.0, 3.0] }
      """
    Then the response status should be 200
    And the response body should contain "predictions"