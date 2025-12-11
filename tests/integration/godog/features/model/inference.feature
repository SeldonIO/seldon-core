@ModelInference @Models @Inference
Feature: Basic model inferencing

  Scenario Outline: Success - Inference for <model> model
    Given I have an "<model>" model
    When the model is applied
    Then the model should eventually become Ready
    When I send a valid HTTP inference request with timeout "20s"
    Then expect http response status code "200"
    When I send a valid gRPC inference request with timeout "20s"


    Examples:
      | model     |
      | tfsimple1 |
