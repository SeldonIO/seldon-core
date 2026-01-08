# Seldon Core Helm Charts

Helm charts are published to our official repo.

## Core Chart

The core chart for installing Seldon Core is shown below.

 * [seldon-core-operator](../../helm-charts/seldon-core-operator/README.md)

For advanced configuration of the core chart, please see [Advanced Helm Chart Configuration](../install/advanced-helm-chart-configuration.md).

## Inference Graph Templates

A set of charts to provide example templates for creating particular inference graphs using Seldon Core
 * [seldon-single-model](../../helm-charts/seldon-single-model/README.md)
   * Serve a single model with attached Persistent Volume.
 * [seldon-abtest](../../helm-charts/seldon-abtest/README.md)
   * Serve an AB test between two models.
 * [seldon-mab](../../helm-charts/seldon-mab/README.md)
   * Serve a multi-armed bandit between two models.
 * [seldon-od-model](../../helm-charts/seldon-od-model/README.md) and [seldon-od-transformer](../../helm-charts/seldon-od-transformer/README.md)

[A notebook with examples of using the above charts](../notebooks/helm_examples.md) is provided.

