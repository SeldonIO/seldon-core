# Tracing

We support Open Telemetry tracing. By default all components will attempt to send OLTP events to `seldon-collector.seldon-mesh:4317` which will export to Jaeger at `simplest-collector.seldon-mesh:14250`.

The components can be installed from the `tracing/k8s` folder. In future an Ansible playbook will be created. This installs a Open Telemetry collector and a simple Jaeger install with a service that can be port forwarded to at `simplest.seldon-mesh:16686`.

An example Jaeger trace is show below:

![trace](jaeger-trace.png)

