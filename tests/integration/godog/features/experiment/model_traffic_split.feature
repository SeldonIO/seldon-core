@ExperimentTrafficSplit @Functional @Experiments
Feature: Experiment traffic splitting
  In order to perform A/B testing
  As a model user
  I need to create an Experiment resource that splits traffic 50/50 between models

  Scenario: Success - Create experiment with 50/50 traffic split between two iris models
    Given I create model spec with timeout "10s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: experiment-1
    spec:
      replicas: 1
      requirements:
      - sklearn
      - mlserver
      storageUri: gs://seldon-models/scv2/samples/mlserver_1.3.5/iris-sklearn
    """
    When the model "experiment-1" should eventually become Ready with timeout "20s"
    Given I create model spec with timeout "10s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: experiment-2
    spec:
      replicas: 1
      requirements:
      - sklearn
      - mlserver
      storageUri: gs://seldon-models/scv2/samples/mlserver_1.3.5/iris-sklearn
    """
    When the model "experiment-2" should eventually become Ready with timeout "20s"
    When I deploy experiment spec with timeout "60s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Experiment
    metadata:
      name: experiment-50-50
    spec:
      default: experiment-1
      candidates:
      - name: experiment-1
        weight: 50
      - name: experiment-2
        weight: 50
    """
    Then the experiment should eventually become Ready with timeout "60s"
    When I send "20" HTTP inference requests to the experiment and expect all models in response, with payoad:
    """
    {"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}
    """