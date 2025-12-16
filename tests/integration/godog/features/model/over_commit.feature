@ModelDeployment @Functional @Models @CustomModelSpec @OverCommit
Feature: Explicit Model deployment
  I deploy 3 iris models, each requiring 334MB. The server memory capacity is 1GB,
  with a 10% allowance for over-commit. The third model should be evicted to disk. Send
  inference requests to all models, expected them all to pass, as the agent will
  automatically load the evicted model on-the-fly when req received

  Scenario: Deploy 3 identical models and send inference
    Given I deploy model spec with timeout "10s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: "overcommit-1"
    spec:
      replicas: 1
      requirements:
      - sklearn
      - mlserver
      memory: 334000000
      storageUri: gs://seldon-models/scv2/samples/mlserver_1.3.5/iris-sklearn
    """
    When the model "overcommit-1" should eventually become Ready with timeout "20s"
    Given I deploy model spec with timeout "10s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: "overcommit-2"
    spec:
      replicas: 1
      requirements:
      - sklearn
      - mlserver
      memory: 334000000
      storageUri: gs://seldon-models/scv2/samples/mlserver_1.3.5/iris-sklearn
    """
    When the model "overcommit-2" should eventually become Ready with timeout "20s"
    Given I deploy model spec with timeout "10s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: "overcommit-3"
    spec:
      replicas: 1
      requirements:
      - sklearn
      - mlserver
      memory: 334000000
      storageUri: gs://seldon-models/scv2/samples/mlserver_1.3.5/iris-sklearn
    """
    When the model "overcommit-3" should eventually become Ready with timeout "20s"
    Then send HTTP inference request with timeout "20s" to model "overcommit-1" with payload:
    """
    {
        "inputs": [
          {
            "name": "predict",
            "shape": [1, 4],
            "datatype": "FP32",
            "data": [[1, 2, 3, 4]]
          }
        ]
    }
    """
    And expect http response status code "200"
    Then send HTTP inference request with timeout "20s" to model "overcommit-2" with payload:
    """
    {
        "inputs": [
          {
            "name": "predict",
            "shape": [1, 4],
            "datatype": "FP32",
            "data": [[1, 2, 3, 4]]
          }
        ]
    }
    """
    And expect http response status code "200"
    Then send HTTP inference request with timeout "20s" to model "overcommit-3" with payload:
    """
    {
        "inputs": [
          {
            "name": "predict",
            "shape": [1, 4],
            "datatype": "FP32",
            "data": [[1, 2, 3, 4]]
          }
        ]
    }
    """
    And expect http response status code "200"