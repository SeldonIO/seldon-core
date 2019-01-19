# Seldon Core Analytics

This package provides an example Prometheus and Grafana install with default dashboard for Seldon Core. Its only intended for demonstration. In production you should create your own Prometheus and Grafana installation.

## Install

 * Install seldon-core as described [here](../../docs/install.md#with-ksonnet)
 * Install seldon-core-analytics. Parameters:
   * ```password``` set the Grafana dashboard password - default 'admin'
   * ```prometheusServiceType``` the Prometheus service type - default 'ClusterIP'
   * ```grafanaServiceType``` the Grafama service type - default `NodePort`

```
    ks pkg install seldon-core/seldon-core-analytics@master
    ks generate seldon-core-analytics seldon-core-analytics
```
 * Launch components onto cluster
 ```
 ks apply default
 ```
Notes

 * You can use ```--namespace``` to install seldon-core to a particular namespace
