@ModelInference @Models @Inference @Functional
Feature: Basic model inferencing

  Scenario Outline: Success - Inference for <model> model
    Given I have an "<model>" model
    When the model is applied
    Then the model should eventually become Ready
    When I send a valid HTTP inference request with timeout "20s"
    Then expect http response status code "200"
    And expect http response body to contain valid JSON
    When I send a valid gRPC inference request with timeout "20s"
    And expect gRPC response to not return an error

    Examples:
      | model      |
      | mnist-pytorch |
      | wine       |
      | tfsimple1  |
      | iris       |
      | income-xgb |
      | income-lgb |
      | mnist-onnx |
