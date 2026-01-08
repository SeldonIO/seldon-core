@Autoscaling
Feature: Deploy model and change the desired replica count
  I deploy a custom model spec, wait for model to be deployed to the server.
  I then increate the desired replica count, wait for new server replicas
  to spin up and for model to become ready.


  Scenario: Load model and send inference request to envoy
    Given I create model spec with timeout "10s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: auto-scaling-alpha-1
    spec:
      replicas: 1
      requirements:
      - sklearn
      - mlserver
      storageUri: gs://seldon-models/scv2/samples/mlserver_1.3.5/iris-sklearn
    """
    When the model "auto-scaling-alpha-1" should eventually become Ready with timeout "20s"
    Then I send HTTP inference request with timeout "20s" to model "auto-scaling-alpha-1" with payload:
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
    Then I update model spec with timeout "10s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: auto-scaling-alpha-1
    spec:
      replicas: 3
      requirements:
      - sklearn
      - mlserver
      storageUri: gs://seldon-models/scv2/samples/mlserver_1.3.5/iris-sklearn
    """
    # FIXME we have to wait otherwise model appears ready immediately, potentially due to watcher not refreshing in time
    And I wait for "2s"
    When the model "auto-scaling-alpha-1" should eventually become Ready with timeout "20s"
    Then I send HTTP inference request with timeout "20s" to model "auto-scaling-alpha-1" with payload:
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
    And eventually only "3" pod(s) are deployed for server name "godog-mlserver" and they are Ready with timeout "10s"
    Then I update model spec with timeout "10s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: auto-scaling-alpha-1
    spec:
      replicas: 1
      requirements:
      - sklearn
      - mlserver
      storageUri: gs://seldon-models/scv2/samples/mlserver_1.3.5/iris-sklearn
    """
    # FIXME we have to wait otherwise model appears ready immediately, potentially due to watcher not refreshing in time
    And I wait for "2s"
    When the model "auto-scaling-alpha-1" should eventually become Ready with timeout "20s"
    Then I send HTTP inference request with timeout "20s" to model "auto-scaling-alpha-1" with payload:
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
    And eventually only "1" pod(s) are deployed for server name "godog-mlserver" and they are Ready with timeout "10s"
