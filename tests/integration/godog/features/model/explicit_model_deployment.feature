@ModelDeployment @Functional @Models @Explicit
Feature: Explicit Model deployment
  I deploy a custom model spec, wait for model to be deployed to the servers
  and send an inference request to that model

  Scenario: Load model and send inference request to envoy
    Given I deploy model spec:
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: iris
    spec:
      replicas: 1
      requirements:
      - sklearn
      - mlserver
      storageUri: gs://seldon-models/scv2/samples/mlserver_1.3.5/iris-sklearn
    """
    When the model "iris" should eventually become Ready with timeout "20s"
    Then send HTTP inference request with timeout "20s" to model "iris" with payload:
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
    And expect http response body to contain JSON:
    """
    { "outputs": [
    {
      "name": "predict",
      "shape": [
        1,
        1
      ],
      "datatype": "INT64",
      "parameters": {
        "content_type": "np"
      },
      "data": [
        2
      ]
    }
  ] }
    """
    Then send gRPC inference request with timeout "20s" to model "iris" with payload:
    """
    {
        "inputs": [
          {
            "name": "predict",
            "shape": [1, 4],
            "datatype": "FP32",
            "contents": {
              "int64_contents" : [1, 2, 3, 4]
            }
          }
        ]
    }
    """
    And expect gRPC response body to contain JSON:
    """
    { "outputs": [
    {
      "name": "predict",
      "shape": [
        1,
        1
      ],
      "datatype": "INT64",
      "parameters": {
        "content_type": {"ParameterChoice":{"StringParam":"np"}}
      },
      "contents": {"int64_contents" : [2]}
    }
  ] }
    """