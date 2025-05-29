---
description: Learn how to implement A/B testing and traffic splitting in Seldon Core 2 for ML models and pipelines. This guide covers HTTP traffic distribution, model mirroring for testing, percentage-based traffic routing, and best practices for conducting ML model experiments in production environments.
---

# Experiments

An Experiment defines an http traffic split between Models or Pipelines.

Experiments also allow a mirror model or pipeline to be tested where some
percentage of the traffic to the main model is sent to the mirror but the result is not returned.

Further details are given [here](kubernetes/resources/experiment.md).
