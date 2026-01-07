@ModelVersion @Functional @Models @CustomModelSpec
Feature: Deploy different model versions
  I deploy a custom model spec, wait for model to be deployed to the servers
  and send an inference request to that model and expect a successful response.
  I then update the model with a new URI and send inference and expect a new response
  for model.

  Scenario: Load model and send inference request to envoy
    Given I create model spec with timeout "10s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: model-version-alpha-1
    spec:
      replicas: 1
      requirements:
      - sklearn
      - mlserver
      storageUri: gs://seldon-models/scv2/samples/mlserver_1.3.5/iris-sklearn
    """
    When the model "model-version-alpha-1" should eventually become Ready with timeout "20s"
    Then I send HTTP inference request with timeout "20s" to model "model-version-alpha-1" with payload:
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
    Then I update model spec with timeout "10s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: model-version-alpha-1
    spec:
      replicas: 1
      requirements:
      - xgboost
      - mlserver
      storageUri: gs://seldon-models/scv2/samples/mlserver_1.3.5/income-xgb
    """
    When the model "model-version-alpha-1" should eventually become Ready with timeout "20s"
    # FIXME we have to wait otherwise get 503, likely from envoy due to DNS not updating, will be fixed when we use IP routing (2s is the DNS refresh policy)
    And I wait for "2s"
    Then I send HTTP inference request with timeout "20s" to model "model-version-alpha-1" with payload:
    """
    { "parameters": { "content_type" : "pd" }, "inputs": [{"name": "Age", "shape": [1, 1], "datatype": "INT64", "data": [47]},{"name": "Workclass", "shape": [1, 1], "datatype": "INT64", "data": [4]},{"name": "Education", "shape": [1, 1], "datatype": "INT64", "data": [1]},{"name": "Marital Status", "shape": [1, 1], "datatype": "INT64", "data": [1]},{"name": "Occupation", "shape": [1, 1], "datatype": "INT64", "data": [1]},{"name": "Relationship", "shape": [1, 1], "datatype": "INT64", "data": [3]},{"name": "Race", "shape": [1, 1], "datatype": "INT64", "data": [4]},{"name": "Sex", "shape": [1, 1], "datatype": "INT64", "data": [1]},{"name": "Capital Gain", "shape": [1, 1], "datatype": "INT64", "data": [0]},{"name": "Capital Loss", "shape": [1, 1], "datatype": "INT64", "data": [0]},{"name": "Hours per week", "shape": [1, 1], "datatype": "INT64", "data": [40]},{"name": "Country", "shape": [1, 1], "datatype": "INT64", "data": [9]}]}
    """
    And expect http response status code "200"
    And expect http response body to contain JSON:
    """
    {"outputs":[{"name":"predict","shape":[1,1],"datatype":"FP32","parameters":{"content_type":"np"},"data":[-1.8380107879638672]}]}
    """
    Then delete the model "model-version-alpha-1" with timeout "10s"
    Then I send HTTP inference request with timeout "20s" to model "model-version-alpha-1" with payload:
    """
    { "parameters": { "content_type" : "pd" }, "inputs": [{"name": "Age", "shape": [1, 1], "datatype": "INT64", "data": [47]},{"name": "Workclass", "shape": [1, 1], "datatype": "INT64", "data": [4]},{"name": "Education", "shape": [1, 1], "datatype": "INT64", "data": [1]},{"name": "Marital Status", "shape": [1, 1], "datatype": "INT64", "data": [1]},{"name": "Occupation", "shape": [1, 1], "datatype": "INT64", "data": [1]},{"name": "Relationship", "shape": [1, 1], "datatype": "INT64", "data": [3]},{"name": "Race", "shape": [1, 1], "datatype": "INT64", "data": [4]},{"name": "Sex", "shape": [1, 1], "datatype": "INT64", "data": [1]},{"name": "Capital Gain", "shape": [1, 1], "datatype": "INT64", "data": [0]},{"name": "Capital Loss", "shape": [1, 1], "datatype": "INT64", "data": [0]},{"name": "Hours per week", "shape": [1, 1], "datatype": "INT64", "data": [40]},{"name": "Country", "shape": [1, 1], "datatype": "INT64", "data": [9]}]}
    """
    And expect http response status code "404"
    Then I send gRPC inference request with timeout "20s" to model "model-version-alpha-1" with payload:
    """
    {
  "parameters": {
    "content_type": {
      "string_param": "pd"
    }
  },
  "inputs": [
    {"name": "Age", "shape": [1, 1], "datatype": "INT64", "contents": {"int64_contents": [47]}},
    {"name": "Workclass", "shape": [1, 1], "datatype": "INT64", "contents": {"int64_contents": [4]}},
    {"name": "Education", "shape": [1, 1], "datatype": "INT64", "contents": {"int64_contents": [1]}},
    {"name": "Marital Status", "shape": [1, 1], "datatype": "INT64", "contents": {"int64_contents": [1]}},
    {"name": "Occupation", "shape": [1, 1], "datatype": "INT64", "contents": {"int64_contents": [1]}},
    {"name": "Relationship", "shape": [1, 1], "datatype": "INT64", "contents": {"int64_contents": [3]}},
    {"name": "Race", "shape": [1, 1], "datatype": "INT64", "contents": {"int64_contents": [4]}},
    {"name": "Sex", "shape": [1, 1], "datatype": "INT64", "contents": {"int64_contents": [1]}},
    {"name": "Capital Gain", "shape": [1, 1], "datatype": "INT64", "contents": {"int64_contents": [0]}},
    {"name": "Capital Loss", "shape": [1, 1], "datatype": "INT64", "contents": {"int64_contents": [0]}},
    {"name": "Hours per week", "shape": [1, 1], "datatype": "INT64", "contents": {"int64_contents": [40]}},
    {"name": "Country", "shape": [1, 1], "datatype": "INT64", "contents": {"int64_contents": [9]}}
  ]
}
    """
    And expect gRPC response error to contain "Unimplemented"