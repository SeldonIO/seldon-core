Feature: Model scaling ensures mlserver alignment
  We want to prove that a model asking for five replicas makes the backing mlserver scale before the model stabilises.

  Background:
    Given the namespace "seldon-mesh" is used

  Scenario: Model requests five replicas
    Given the following server resource:
      """
      apiVersion: mlops.seldon.io/v1alpha1
      kind: Server
      metadata:
        name: mlserver
      spec:
        serverConfig: mlserver
        replicas: 5
        minReplicas: 2
        maxReplicas: 5
      """
    And the following model resource:
      """
      apiVersion: mlops.seldon.io/v1alpha1
      kind: Model
      metadata:
        name: iris
      spec:
        storageUri: gs://seldon-models/scv2/samples/mlserver_1.3.5/iris-sklearn
        requirements:
        - sklearn
        - mlserver
        replicas: 5
        minReplicas: 2
        maxReplicas: 5
      """
    When I apply the server resource
    And I apply the model resource
    Then the server "mlserver" eventually reports 5 replicas
    And the model "iris" eventually reports 5 replicas
