@ModelInference @Models @Inference @Functional
Feature: Basic model inferencing

  Scenario Outline: Success - Inference for <model> model
    Given I have an "<model>" model
    When the model is applied
    Then the model should eventually become Ready
    When I send a valid HTTP inference request with timeout "20s"
    Then expect http response status code "200"
    When I send a valid gRPC inference request with timeout "20s"

    Examples:
      | model         |
      | iris          |
#      | income-xgb | having errors with GRPC
#      | mnist-onnx |
#      | income-lgb | having errors with response
      | tfsimple1     |
      | wine          |
#      | mnist-pytorch | having errors with response
