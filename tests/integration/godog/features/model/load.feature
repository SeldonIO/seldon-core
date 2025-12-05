#@ModelLoad @Models @Load
#Feature: Model performance under load
#
##  the model replicas could default to the max replicas of the available ml server
#  Background:
#    Given a clean test namespace
#    And a Ready model "load-model" with capabilities "mlserver"
#
#  Scenario: Model meets latency SLO at 200 RPS
#    When I run a load test of 200 RPS for "2m" against model "load-model"
#    Then the 95th percentile latency should be less than "150ms"
#    And the error rate should be less than "1%"
