@PipelineDeployment
Feature: Model deployment

  Background:
    Given a running Seldon Core 2 control plane
    And a clean namespace "could be an env var"
    And a new server with capabilties "tensor-flow" and min replicas "1" and replicas "3" and max replicas "5"


  Scenario: 
