@ModelInference @Models @Inference
Feature Basic model inferencing

  Background:
    Given a clean test namespace

  Scenario: Model can serve prediction
    Given a Ready model "basic-model" with capabilities "mlserver"
    When I send a prediction request with payload:
      """
      { "inputs": [1.0, 2.0, 3.0] }
      """
    Then the response status should be 200
    And the response body should contain "predictions"