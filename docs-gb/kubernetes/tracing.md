---
description: >-
 This guide walks you through setting up Jaeger Tracing for Seldon Core v2 on Kubernetes. By the end of this guide, you will be able to visualize inference traces through your Core 2 components.
---

## Prerequisites

Ensure you have the following installed:

- **Cert-Manager**
- **Jaeger Tracing v2 Kubernetes Operator**
- **Seldon Core v2**

## Install Cert-Manager
1. Install the cert-manager using below command
```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/latest/download/cert-manager.yaml
```
2. Wait until all the Cert-Manager pods are running in the `cert-manager` namespace. You can check with:
```bash
kubectl get po -n cert-manager
```

## Create tracing namespace
Create a dedicated namespace to install the Jaeger Operator and tracing resources:
```bash
kubectl create namespace tracing
```
## Install Jaeger Operator

The Jaeger Operator manages Jaeger instances in your cluster. The Helm chart for Jaeger v2 is available [here](https://github.com/jaegertracing/helm-charts/tree/v2)

1. Add the Jaeger Helm repository if not already added:
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
4. Validate that the Jaeger Operator pod is running:
```bash
kubectl get pods -n tracing
```
Example:
```bash
NAME                                       READY   STATUS    RESTARTS   AGE
tracing-jaeger-operator-549b79b848-h4p4d   1/1     Running   0          96s
```
## Deploy a minimal Jaeger instance
Install a simple Jaeger custom resource in the namespace where Seldon Core v2 is running (for example, seldon-mesh). This CR deploys an all-in-one Jaeger instance suitable for development and non-production use.

1. Save the following manifest as `jaeger-simplest.yaml`:
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
Example:
```bash
NAME                       READY  STATUS    RESTARTS   AGE
simplest-8686f5d96-4ptb4   1/1    Running   0          45s
```
What the above `simplest` Jaeger CR does:

- **All-in-one pod**: Deploys a single pod running the collector, agent, query service, and UI, using in-memory storage.

- **Use case**: Suitable for local development, demos, and quick-start scenarios. Not recommended for production because all components and trace data are ephemeral.

- **Core 2 integration**: The all-in-one pod receives spans from Seldon Core 2 components and exposes a UI for viewing traces.

## Configure Seldon Core v2 to emit traces

To enable tracing, configure the OpenTelemetry exporter endpoint in the `SeldonRuntime` resource so that traces are sent to the Jaeger collector service created by the simplest Jaeger CR.

1. Edit your `SeldonRuntime`Custom Resource to include `tracingConfig` under `spec.config`:
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
   
3. Restart the below Core 2 components pods so they pick up the new tracing configuration from the `seldon-tracing` ConfigMap in the `seldon-mesh` namespace.

- seldon-dataflow-engine

- seldon-pipeline-gateway

- seldon-model-gateway

- seldon-scheduler

- Servers

After restart, these components will read the updated tracing config and start emitting traces to Jaeger.

## Generate traffic to see traces

To visualize traces, send requests to your models or pipelines deployed in Seldon Core 2. Each inference request should produce a trace that shows the path through the Core 2 components (e.g., gateways, dataflow engine, server agents) in the Jaeger UI.

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
