# seldon-core-analytics

![Version: 1.6.0-dev](https://img.shields.io/static/v1?label=Version&message=1.6.0--dev&color=informational&style=flat-square)

Prometheus and Grafana installation with a basic Grafana dashboard showing
the default Prometheus metrics exposed by Seldon for each inference graph
deployed.

## Usage

To use this chart, you will first need to add the `seldonio` Helm repo:

```bash
helm repo add seldonio https://storage.googleapis.com/seldon-charts
helm repo update
```

Onca that's done, you should then be able to deploy the chart as:

```bash
kubectl create namespace seldon-system
helm install seldon-core-analytics seldonio/seldon-core-analytics --namespace seldon-system
```

**Homepage:** <https://github.com/SeldonIO/seldon-core>

## Source Code

* <https://github.com/SeldonIO/seldon-core>
* <https://github.com/SeldonIO/seldon-core/tree/master/helm-charts/seldon-core-analytics>

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://charts.helm.sh/stable | grafana | ~5.1.4 |
| https://charts.helm.sh/stable | prometheus | ~11.4.0 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| alertmanager.config.enabled | bool | `false` |  |
| grafana.adminPassword | string | `"password"` |  |
| grafana.adminUser | string | `"admin"` |  |
| grafana.datasources."datasources.yaml".apiVersion | int | `1` |  |
| grafana.datasources."datasources.yaml".datasources[0].access | string | `"proxy"` |  |
| grafana.datasources."datasources.yaml".datasources[0].name | string | `"prometheus"` |  |
| grafana.datasources."datasources.yaml".datasources[0].type | string | `"prometheus"` |  |
| grafana.datasources."datasources.yaml".datasources[0].url | string | `"http://seldon-core-analytics-prometheus-seldon"` |  |
| grafana.enabled | bool | `true` |  |
| grafana.sidecar.dashboards.enabled | bool | `true` |  |
| grafana.sidecar.dashboards.label | string | `"seldon_dashboard"` |  |
| prometheus.enabled | bool | `true` |  |
| prometheus.server.configPath | string | `"/etc/prometheus/conf/prometheus-config.yaml"` |  |
| prometheus.server.extraConfigmapMounts[0].configMap | string | `"prometheus-server-conf"` |  |
| prometheus.server.extraConfigmapMounts[0].mountPath | string | `"/etc/prometheus/conf/"` |  |
| prometheus.server.extraConfigmapMounts[0].name | string | `"prometheus-config-volume"` |  |
| prometheus.server.extraConfigmapMounts[0].readOnly | bool | `true` |  |
| prometheus.server.extraConfigmapMounts[0].subPath | string | `""` |  |
| prometheus.server.extraConfigmapMounts[1].configMap | string | `"prometheus-rules"` |  |
| prometheus.server.extraConfigmapMounts[1].mountPath | string | `"/etc/prometheus-rules"` |  |
| prometheus.server.extraConfigmapMounts[1].name | string | `"prometheus-rules-volume"` |  |
| prometheus.server.extraConfigmapMounts[1].readOnly | bool | `true` |  |
| prometheus.server.extraConfigmapMounts[1].subPath | string | `""` |  |
| prometheus.server.name | string | `"seldon"` |  |
| prometheus.server.persistentVolume.enabled | bool | `false` |  |
| prometheus.server.persistentVolume.existingClaim | string | `"seldon-claim"` |  |
| prometheus.server.persistentVolume.mountPath | string | `"/seldon-data"` |  |
| prometheus.service_type | string | `"ClusterIP"` |  |
| rbac.enabled | bool | `true` |  |
