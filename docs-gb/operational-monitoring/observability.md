---
description: >-
  Installing kube-prometheus-stack in the same Kubernetes cluster that hosts the
  Seldon Core 2.
---

# Monitoring

`kube-prometheus`, also known as Prometheus Operator, is a popular open-source project that provides complete monitoring and alerting solutions for Kubernetes clusters. It combines tools and components to create a monitoring stack for Kubernetes environments.

{% hint style="info" %}
**Note**: In this example Prometheus is installed within the same Kubernetes cluster as the Seldon Core 2. However, Seldoc Core 2 exposes metrics to any of the managed Prometheus endpoints as well.
{% endhint %}

The Seldon Core 2, along with any deployed models, automatically exposes metrics to Prometheus. By default, certain alerting rules are pre-configured, and an alertmanager instance is included.

You can install `kube-prometheus` to monitor Seldon components, and ensure that the appropriate `ServiceMonitors` are in place for Seldon deployments. The analytics component is configured with the Prometheus integration. The monitoring for Seldon Core 2 is based on the Prometheus Operator and the related `PodMonitor` and `PrometheusRule` resources.

Monitoring the model deployments in Seldon Core 2 involves:

1. [Installing kube-prometheus](observability.md#installing-kube-prometheus)
2. [Configuring monitoring](observability.md#configuring-monitoring-for-seldon-core-2)

## Prerequisites

1. Install [Seldon Core 2](../installation/production-environment/).
2. Install [Ingress Controller](../installation/production-environment/ingress-controller/).
3. Install [Grafana](https://grafana.com/docs/grafana/latest/setup-grafana/installation/helm/) in the namespace `seldon-monitoring`.

## Installing kube-prometheus

1.  Create a namespace for the monitoring components of Seldon Core 2.

    ```
    kubectl create ns seldon-monitoring || echo "Namespace seldon-monitoring already exists"
    ```
4.  Create a YAML file to specify the initial configuration. For example, create the `prometheus-values.yaml` file. Use your preferred text editor to create and save the file with the following content:

    ```yaml
    fullnameOverride: seldon-monitoring
    kube-state-metrics:
      extraArgs:
        metric-labels-allowlist: pods=[*]
    ```

    **Note**: Make sure to include `metric-labels-allowlist: pods=[*]` in the Helm values file. If you are using your own Prometheus Operator installation, ensure that the pods labels, particularly `app.kubernetes.io/managed-by=seldon-core`, are part of the collected metrics. These labels are essential for calculating deployment usage rules.
5.  Change to the directory that contains the `prometheus-values` file and run the following command to install version `9.5.12` of `kube-prometheus`.

    ```
    helm upgrade --install prometheus kube-prometheus \
     --version 9.5.12 \
     --namespace seldon-monitoring \
     --values prometheus-values.yaml \
     --repo https://charts.bitnami.com/bitnami
    ```

    When the installation is complete, you should see this:

    ```
    WARNING: There are "resources" sections in the chart not set. Using "resourcesPreset" is not recommended for production. For production installations, please set the following values according to your workload needs:
      - alertmanager.resources
      - blackboxExporter.resources
      - operator.resources
      - prometheus.resources
      - prometheus.thanos.resources
    +info https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

    ```
6.  Check the status of the installation.

    ```
    kubectl rollout status -n seldon-monitoring deployment/seldon-monitoring-operator
    ```

    When the installation is complete, you should see this:

    ```
    Waiting for deployment "seldon-monitoring-operator" rollout to finish: 0 of 1 updated replicas are available...
    deployment "seldon-monitoring-operator" successfully rolled out
    ```

## Configuring monitoring for Seldon Core 2

1.  You can access Prometheus from outside the cluster by running the following commands:

    ```
    echo "Prometheus URL: http://127.0.0.1:9090/"
    kubectl port-forward --namespace seldon-monitoring svc/seldon-monitoring-prometheus 9090:9090
    ```
2.  You can access Alertmanager from outside the cluster by running the following commands:

    ```
    echo "Alertmanager URL: http://127.0.0.1:9093/"
    kubectl port-forward --namespace seldon-monitoring svc/seldon-monitoring-alertmanager 9093:9093
    ```
3.  Apply the Custom RBAC Configuration settings for kube-prometheus.
    ```bash
    CUSTOM_RBAC=https://raw.githubusercontent.com/SeldonIO/seldon-core/v2.8.2/prometheus/rbac

    kubectl apply -f ${CUSTOM_RBAC}/cr.yaml
    ```
4.  Configure metrics collection by createing the following `PodMonitor` resources.
    ```bash
    PODMONITOR_RESOURCE_LOCATION=https://raw.githubusercontent.com/SeldonIO/seldon-core/v2.8.2/prometheus/monitors

    kubectl apply -f ${PODMONITOR_RESOURCE_LOCATION}/agent-podmonitor.yaml
    kubectl apply -f ${PODMONITOR_RESOURCE_LOCATION}/envoy-servicemonitor.yaml
    kubectl apply -f ${PODMONITOR_RESOURCE_LOCATION}/pipelinegateway-podmonitor.yaml
    kubectl apply -f ${PODMONITOR_RESOURCE_LOCATION}/server-podmonitor.yaml
    ```
    When the resources are created, you should see this:
    ```bash
    podmonitor.monitoring.coreos.com/agent created
    servicemonitor.monitoring.coreos.com/envoy created
    podmonitor.monitoring.coreos.com/pipelinegateway created
    podmonitor.monitoring.coreos.com/server created
    ```  
## Next

### Prometheus User Interface
You may now be able to check the status of Seldon components in Prometheus:
1. Open your browser and navigate to `http://127.0.0.1:9090/` to access Prometheus UI from outside the cluster.
1. Go to **Status** and select **Targets**.

The status of all the endpoints and the scrape details are displayed.

### Grafana
You can view the metrics in Grafana Dashboard after you set Prometheus as the Data Source, and import `seldon.json` dashboard located at `seldon-core/v2.8.2/prometheus/dashboards` in [GitHub repository](https://github.com/SeldonIO/seldon-core/tree/v2/prometheus/dashboards).
