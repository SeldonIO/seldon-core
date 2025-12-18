@SchedulerPipelineRetries @Functional @Scheduler
Feature: Scheduler retries failed pipelines
  In order to ensure pipelines recover from transient failures
  As a platform operator
  I need the scheduler to retry creating and terminating pipelines that have previously failed
  @0
  Scenario: Retry creating a pipeline that failed while Kafka was unavailable
    Given kafka-nodepool is unavailable for Core 2 with timeout "40s"
    Given I deploy model spec with timeout "30s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: tfsimple1-hhk2
    spec:
      storageUri: "gs://seldon-models/triton/simple"
      requirements:
      - tensorflow
    apiVersion: mlops.seldon.io/v1alpha1
    """
    Then the model "tfsimple1-hhk2" should eventually become Ready with timeout "20s"
    And I deploy a pipeline spec with timeout "30s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: retry-pipe-hhk2
    spec:
      steps:
        - name: tfsimple1-hhk2
      output:
        steps:
        - tfsimple1-hhk2
    """
    Then the pipeline should eventually become NotReady with timeout "30s"
    And the pipeline status should eventually become PipelineFailed with timeout "30s"
    When Kafka is available for Core 2 with timeout "40s"
    Then the pipeline should eventually become Ready with timeout "120s"

  Scenario: Retry terminating a pipeline that previously failed to terminate
    Given I deploy model spec with timeout "30s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: tfsimple1-hfk5
    spec:
      storageUri: "gs://seldon-models/triton/simple"
      requirements:
      - tensorflow
    apiVersion: mlops.seldon.io/v1alpha1
    """
    Then the model "tfsimple1-hfk5" should eventually become Ready with timeout "20s"
    Given I deploy a pipeline spec with timeout "30s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: retry-pipe-hfk5
    spec:
      steps:
        - name: tfsimple1-hfk5
      output:
        steps:
          - tfsimple1-hfk5
    """
    Then the pipeline should eventually become Ready with timeout "60s"
    When kafka-nodepool is unavailable for Core 2 with timeout "40s"
    And I delete pipeline the with timeout "30s"
    Then the pipeline status should eventually become PipelineFailed with timeout "<string>"
    When Kafka is available for Core 2 with timeout "40s"
    Then the pipeline should eventually not exist with timeout "120s"

