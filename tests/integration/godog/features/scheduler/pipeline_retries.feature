@SchedulerPipelineRetries @Functional @Scheduler
Feature: Scheduler retries failed pipelines
  In order to ensure pipelines recover from transient failures
  As a platform operator
  I need the scheduler to retry creating and terminating pipelines that have previously failed

  Scenario: Retry creating a pipeline that failed while Kafka was unavailable
    Given Kafka is unavailable for Core 2
    And I deploy a pipeline spec with timeout "30s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: retry-pipeline
    spec:
      steps:
        - name: retry-model
      output:
        steps:
          - retry-model
    """
    Then the pipeline "retry-pipeline" should eventually not become Ready with timeout "60s"
    And the pipeline "retry-pipeline" state should be "PipelineFailed"
    When Kafka becomes available again
    Then the pipeline "retry-pipeline" should eventually become Ready with timeout "120s"

  Scenario: Retry terminating a pipeline that previously failed to terminate
    Given I deploy a pipeline spec with timeout "30s":
    """
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: retry-pipeline
    spec:
      steps:
        - name: retry-model
      output:
        steps:
          - retry-model
    """
    Then the pipeline "retry-pipeline" should eventually become Ready with timeout "60s"
    When Kafka is unavailable for Core 2
    And I delete the pipeline "retry-pipeline" with timeout "30s"
    Then the pipeline "retry-pipeline" state should be "PipelineFailedTerminating"
    When Kafka becomes available again
    Then the pipeline "retry-pipeline" should eventually not exist with timeout "120s"

