@Server
Feature: Server setup
  Deploys an mlserver with one replica. We ensure the pods
  become ready and remove any other server pods for different
  servers.

  @ServerSetup
  Scenario: Deploy mlserver Server and remove other servers
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
    When the server should eventually become Ready with timeout "30s"
    Then ensure only "1" pod(s) are deployed for server and they are Ready
    And remove any other server deployments

# TODO decide if we want to keep this, if we keep testers will need to ensure they don't run this tag when running all
#  all features in this directory, as tests will fail when server is deleted. We can not delete and it's up to the
#  feature dir server setup to ensure ONLY the required servers exist, like above.
#  @ServerTeardown
#  Scenario: Delete mlserver Server
#    Given I delete server "godog-mlserver" with timeout "10s"