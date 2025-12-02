@ModelDisruption @Models @Disruption
Feature: Model resilence to Core 2 disruption

  Background:
    Given a clean test namespace
    And a Ready model "resilient-model" with capabilities "mlserver"

  Scenario: Model keeps serving during a control plane restart
    Given a load test of 100 RPS is running against model "resilient-model"
    When I restart the Seldon Core 2 control plane
    Then at least 99% of requests should succeed during "2m"
    And the 95th percentile latency during "2m" should be less than "250ms"
    And no outage should last longer than "10s"