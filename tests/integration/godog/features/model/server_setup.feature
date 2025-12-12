@ServerSetup
Feature: Server setup
  TODO

  Scenario: TODO
    Given I deploy server spec with timeout "10s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Server
    metadata:
      name: godog-mlserver
    spec:
      replicas: 1
      serverConfig: mlserver
      requirements:
      - sklearn
      - mlserver
      storageUri: gs://seldon-models/scv2/samples/mlserver_1.3.5/iris-sklearn
    """
    When the server "godog-mlserver" should eventually become Ready with timeout "20s"
    And ensure only "1" pod is deployed for server "godog-mlserver" and is Ready
    And remove any other server deployments which are not "godog-mlserver"