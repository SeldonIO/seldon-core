@ModelDeployment @Functional @Models @CustomModelSpec
Feature: Explicit Model deployment
  I deploy a custom model spec, wait for model to be deployed to the servers
  and send an inference request to that model and expect a successful response.
  I then delete the model and send inference requests and expect them to fail.

  Scenario: Load model and send inference request to envoy
    Given I deploy model spec with timeout "10s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: alpha-1
    spec:
      replicas: 1
      requirements:
      - sklearn
      - mlserver
      storageUri: gs://seldon-models/scv2/samples/mlserver_1.3.5/iris-sklearn
    """
    When the model "alpha-1" should eventually become Ready with timeout "20s"
    Then send HTTP inference request with timeout "20s" to model "alpha-1" with payload:
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
    Then send gRPC inference request with timeout "20s" to model "alpha-1" with payload:
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
    Then delete the model "alpha-1" with timeout "10s"
    Then send HTTP inference request with timeout "20s" to model "alpha-1" with payload:
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
    And expect http response status code "404"
    Then send gRPC inference request with timeout "20s" to model "alpha-1" with payload:
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
    And expect gRPC response error to contain "Unimplemented"