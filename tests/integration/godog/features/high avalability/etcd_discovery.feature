@etcdDiscoveryHighAvailability @HighAvailability @Functional
Feature:
  In order to ensure scheduler and control plane high availability
  As a Core 2 platform operator
  I want to be able to change the cluster membership of scheduler when needed

  @etcdDiscoveryRuleFromOneNodeScaleUp
  Scenario:The cluster size of scheduler is set to 1 but the Core 2 platform operator wants to increase the scheduler
  cluster size by two more nodes
    Given the control plane is deployed with 1 scheduler replicas
    And the scheduler uses a highly available data store
    When I increase the scheduler cluster size by "2"
    Then a new leader should be elected within "5s"
    And exactly 1 scheduler pod should be Ready
    And the Ready scheduler pod should be the new leader
    And the scheduler Service should route traffic to the new leader

  @etcdDiscoveryRuleFromThreeNodeScaleDownToOne
  Scenario:The cluster size of scheduler is set to 3 but the Core 2 platform operator wants to decrease the scheduler
  cluster size by two more nodes
    Given the control plane is deployed with 3 scheduler replicas
    And the scheduler uses a highly available data store
    When I decrease the scheduler cluster size by "2"
    Then exactly 1 scheduler pod should be Ready
    And the Ready scheduler pod should be the new leader
    And the scheduler Service should route traffic to the new leader

  @etcdDiscoveryRuleFromThreeNodeScaleUpToFive
  Scenario:The cluster size of scheduler is set to 3 but the Core 2 platform operator wants to increase the scheduler
  cluster size by two more nodes
    Given the control plane is deployed with 3 scheduler replicas
    And the scheduler uses a highly available data store
    When I increase the scheduler cluster size by "2"
    Then a new leader should be elected within "5s"
    And exactly 1 scheduler pod should be Ready
    And the Ready scheduler pod should be the new leader
    And the scheduler Service should route traffic to the new leader