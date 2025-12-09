@SchedulerHighAvailability @HighAvailability @Functional
Feature: Scheduler High Availability
  In order to ensure reliable model scheduling and orchestration
  As a Core 2 platform operator
  I want the control plane to continue functioning even if one or more scheduler replicas fail

  Background:
    Given the control plane is deployed with at least 3 scheduler replicas

  Scenario: Scheduler elects a new leader and exposes it via the service when the current leader fails
    Given exactly 1 scheduler pod is Ready
    And the Ready scheduler pod is the leader
    When I terminate the scheduler leader pod
    Then a new leader should be elected within "5" seconds
    And exactly 1 scheduler pod should be Ready
    And the Ready scheduler pod should be the new leader
    And the scheduler Service should route traffic to the new leader

  Scenario: Followers remain unroutable and do not become leaders on restart
    Given exactly 1 scheduler pod is Ready
    And the Ready scheduler pod is the leader
    When I terminate a scheduler follower pod
    Then the scheduler cluster should remain Ready
    And a new follower pod should be running within "10" seconds
    And exactly 1 scheduler pod should still be Ready
    And the Ready scheduler pod should still be the leader
    And no follower pod should be Ready or receive traffic

  Scenario: Only the leader scheduler pod is Ready and routable
    When I inspect the scheduler pods
    Then exactly 1 scheduler pod should be Ready
    And the Ready scheduler pod should be the leader
    And all follower scheduler pods should be NotReady
    And the scheduler Service should route traffic to the leader pod

  Scenario: Followers do not accept scheduling requests directly
    Given exactly 1 scheduler pod is Ready
    And the Ready scheduler pod is the leader
    When I send a scheduling request directly to a follower pod
    Then the request should be rejected or not routable
    And the follower should not make scheduling decisions


  Scenario: There is only a leader when the scheduler cluster is restarted
    When the scheduler cluster is restarted
    Then a new leader should be elected withing "5" seconds
    And exactly 1 scheduler pod should be Ready
    And the Ready scheduler pod should be the new leader
    And the scheduler Service should route traffic to the new leader

#  Scenario: Scheduler cluster data is the same when there is a leadership change
#    THis test case might be difficult to do since data in the scheduler can easily change without intervention
#    e.g server restart while we do the operation etc