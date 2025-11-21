Feature: Model scaling ensures mlserver alignment
  We want to prove that a model asking for five replicas makes the backing mlserver scale before the model stabilises.

  Background:
    Given the namespace "seldon-mesh" is used
    And the server manifest file "manifests/server-mlserver.yaml" is used
    And the model manifest file "manifests/model-iris.yaml" is used

  Scenario: Model requests five replicas
    When I apply the server resource
    And I apply the model resource
    Then the server "mlserver" eventually reports 5 replicas
    And the model "iris" eventually reports 5 replicas
