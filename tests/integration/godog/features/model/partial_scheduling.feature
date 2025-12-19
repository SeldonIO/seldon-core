@ModelPartialScheduling @Functional @Models
Feature: Partial scheduling of models
  In order to keep serving traffic while servers scale
  As a model owner
  I want a model to be Ready once it reaches its minimum replicas

  Background:
    Given I have an "iris" model

  # SCENARIOS WHERE THE MODEL SHOULD BE READY
  Scenario Outline: Model is Ready when at least minReplicas are available
    And there is a server with capabilities "mlserver" and replicas "<available>"
    And the model has "<min>" min replicas
    And the model has "<desired>" replicas
    When the model is applied
    Then the model should be Ready
    And the model should have "<desired>" replicas
    And the model should have "<available>" available replicas

    Examples:
      | min | desired | available | note                          |
      | 1   | 3       | 2         | Ready with partial scheduling |
      | 2   | 2       | 2         | Fully scheduled               |

  # SCENARIOS WHERE THE MODEL SHOULD NOT BE READY
  Scenario Outline: Model is not Ready when requirements cannot be met
    And there is a server with capabilities "mlserver" and replicas "<serverReplicas>"
    And the model has "<min>" min replicas
    And the model has "<desired>" replicas
    When the model is applied
    Then the model should not be Ready
    And the model status message should be "ScheduleFailed"

    Examples:
      | min | desired | serverReplicas | note                        |
      | 3   | 3       | 2              | available below minReplicas |

  Scenario: Model is not Ready when minReplicas is unset and desired exceeds available
    And there is a server with capabilities "mlserver" and replicas "1"
    And the model has "3" replicas
    When the model is applied
    Then the model should not be Ready
    And the model status message should be "ScheduleFailed"

  # SCENARIO 5: SCALE UP TO FULLY SCHEDULED WITHOUT REAPPLYING
  Scenario: Model moves from partial to full scheduling as server scales up
    And there is a server with capabilities "mlserver" and replicas "2"
    And the model has "1" min replicas
    And the model has "4" replicas
    When the model is applied
    Then the model should be Ready
    And the model should have "4" replicas
    And the model should have "2" available replicas
    When the server scales to "4" replicas
    And I wait for the model to be reconciled
    Then the model should be Ready
    And the model should have "4" replicas
    And the model should have "4" available replicas

  # SCENARIO 6: SERVER RESTART RECOVERY (PR 6885)
  Scenario: Model remains fully scheduled after server pods restart
    And there is a server with capabilities "mlserver" and replicas "3"
    And the model has "3" min replicas
    And the model has "3" replicas
    When the model is applied
    Then the model should be Ready
    And the model should have "3" replicas
    And the model should have "3" available replicas
    When all server pods are restarted
    And I wait for the server to have "3" ready replicas again
    And I wait for the model to be reconciled
    Then the model should be Ready
    And the model should have "3" replicas
    And the model should have "3" available replicas

  # SCENARIO 7: OPERATOR / REPLICA MISMATCH (INFRA-1629)
  Scenario: Scheduler sees reduced capacity after failed StatefulSet update
    And there is a server with capabilities "mlserver" and replicas "3"
    And the model has "2" min replicas
    And the model has "3" replicas
    When the model is applied
    Then the model should be Ready
    And the model should have "3" replicas
    And the model should have "3" available replicas
    When the server StatefulSet update fails with an immutable field change
    Then the scheduler should see the server replicas as "1"
    And I wait for the model to be reconciled
    And the model should not be Ready
    And the model status message should be "ScheduleFailed"
