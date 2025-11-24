---
description: >-
 This guide walks you through setting up Jaeger Tracing for Seldon Core v2 on Kubernetes. By the end of this guide, you will be able to visualize inference traces through your Core 2 components.
---

## Prerequisites

* Set up and connect to a Kubernetes cluster running version 1.27 or later. For instructions on connecting to your Kubernetes cluster, refer to the documentation provided by your cloud provider.
* Install [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl), the Kubernetes command-line tool.
* Install [Helm](https://helm.sh/docs/intro/install/), the package manager for Kubernetes.
* Install [Seldon Core 2](../installation/installation)
* Install [cert-manager](https://cert-manager.io/docs/installation/kubectl/) in the namespace `cert-manager`.

To set up Jaeger Tracing for Seldon Core 2 on Kubernetes and visualize inference traces of the Seldon Core 2 components. You need to do the following:
1. [Craete a namespace](#create-a-namespace)
2. [Install Jaeger Operator](#install-jaeger-operator)
3. [Deploy a Jaeger instance](#deploy-a-minimal-jaeger-instance)
4. [Configure Core 2](#configure-seldon-core-2)
5. [Generate traffic](#generate-traffic)
6. [Visualize the traces](#cccess-the-jaeger-ui)

## Create a namespace
Create a dedicated namespace to install the Jaeger Operator and tracing resources:
```bash
kubectl create namespace tracing
```
## Install Jaeger Operator

The Jaeger Operator manages Jaeger instances in the Kubernetes cluster. Use the [Helm chart](https://github.com/jaegertracing/helm-charts/tree/v2) for Jaeger v2.

1. Add the Jaeger to the Helm repository:
```bash
helm repo add jaegertracing https://jaegertracing.github.io/helm-charts
helm repo update
```
2. Create a minimal `tracing-values.yaml`:
```bash
rbac:
  clusterRole: true
  create: true
  pspEnabled: false
```
3. Install or upgrade the Jaeger Operator in the tracing namespace:
```bash
helm upgrade tracing jaegertracing/jaeger-operator \
  --version 2.57.0 \
  -f tracing-values.yaml \
  -n tracing \
  --install
```
4. Validate that the Jaeger Operator Pod is running:
```bash
kubectl get pods -n tracing
```
Output is similar to:
```bash
NAME                                       READY   STATUS    RESTARTS   AGE
tracing-jaeger-operator-549b79b848-h4p4d   1/1     Running   0          96s
```
## Deploy a minimal Jaeger instance
Install a simple Jaeger custom resource in the namespace `seldon-mesh`, where Seldon Core 2 is running . **Note**:  This CR is suitable for local development, demos, and quick-start scenarios. It is not recommended for production because all components and trace data are ephemeral.

1. Create a manifest file named `jaeger-simplest.yaml` with these contents:
```bash
apiVersion: jaegertracing.io/v1
kind: Jaeger
metadata:
  name: simplest
  namespace: seldon-mesh
```
2. Apply the manifest:
```bash
kubectl apply -f jaeger-simplest.yaml
```
3. Verify that the Jaeger all-in-one pod is running:
```bash
kubectl get pods -n seldon-mesh | grep simplest
```
Output is similar to:
```bash
NAME                       READY  STATUS    RESTARTS   AGE
simplest-8686f5d96-4ptb4   1/1    Running   0          45s
```
This `simplest` Jaeger CR does the following:

- **All-in-one pod**: Deploys a single pod running the collector, agent, query service, and UI, using in-memory storage.

- **Core 2 integration**: receives spans from Seldon Core 2 components and exposes a UI for viewing traces.

## Configure Seldon Core 2 

To enable tracing, configure the OpenTelemetry exporter endpoint in the [SeldonRuntime](../installation/advanced-configurations/seldonconfig) resource so that traces are sent to the Jaeger collector service created by the simplest Jaeger Custom Resource.

1. Edit your `SeldonRuntime` Custom Resource to include `tracingConfig` under `spec.config`:
```bash
spec:
  config:
    agentConfig:
      rclone: {}
    kafkaConfig:
      bootstrap.servers: seldon-kafka-bootstrap.seldon-mesh:9092
      consumer:
        auto.offset.reset: earliest
      topics:
        numPartitions: 4
    scalingConfig:
      servers: {}
    serviceConfig: {}
    tracingConfig:
      otelExporterEndpoint: simplest-collector.seldon-mesh:4317
```
2. Save the updated above ```runtime``` configuration.
   
3. Restart the following Core 2 component Pods so they pick up the new tracing configuration from the `seldon-tracing` ConfigMap in the `seldon-mesh` namespace.

- seldon-dataflow-engine

- seldon-pipeline-gateway

- seldon-model-gateway

- seldon-scheduler

- Servers

After restart, these components reads the updated tracing config and start emitting traces to Jaeger.

## Generate traffic 

To visualize traces, send requests to your models or pipelines deployed in Seldon Core 2. Each inference request should produce a trace that shows the path through the Core 2 components such as gateways, dataflow engine, server agents in the Jaeger UI.

## Access the Jaeger UI

1. Port-forward the Jaeger query service to your local machine:
```bash
kubectl port-forward svc/simplest-query -n seldon-mesh 16686:16686
```
2. Open the Jaeger UI in your browser:
```bash
http://localhost:16686
```
You can now explore traces emitted by Seldon Core 2 components.

An example Jaeger trace is shown below:

![trace](<../images/jaeger-trace (5).png>)
