@ExperimentRecoveryNoModels @Functional @Experiments
Feature: I should be able to create an experiment
  where the candidate models do not exist yet. Once the
  models are created the experiment should become ready

  Scenario: Create experiment first with 50/50 traffic split between two iris models
    When I deploy experiment spec with timeout "60s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Experiment
    metadata:
      name: experiment-recovery-50-50
    spec:
      default: experiment-1-recovery
      candidates:
      - name: experiment-1-recovery
        weight: 50
      - name: experiment-2-recovery
        weight: 50
    """
    Given I create model spec with timeout "10s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: experiment-1-recovery
    spec:
      replicas: 1
      requirements:
      - sklearn
      - mlserver
      storageUri: gs://seldon-models/scv2/samples/mlserver_1.3.5/iris-sklearn
    """
    When the model "experiment-1-recovery" should eventually become Ready with timeout "20s"
    Given I create model spec with timeout "10s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: experiment-2-recovery
    spec:
      replicas: 1
      requirements:
      - sklearn
      - mlserver
      storageUri: gs://seldon-models/scv2/samples/mlserver_1.3.5/iris-sklearn
    """
    When the model "experiment-2-recovery" should eventually become Ready with timeout "20s"
    Then the experiment should eventually become Ready with timeout "60s"
    When I send "20" HTTP inference requests to the experiment and expect all models in response, with payoad:
    """
    {"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}
    """