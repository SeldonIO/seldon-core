# Operational Monitoring

Seldon Core 2 provides robust tools for tracking the performance and health of machine learning models in production.

### Monitoring

* Real-Time metrics: collects and displays real-time metrics from deployed models, such as response times, error rates, and resource usage.
* Model performance tracking: monitors key performance indicators (KPIs) like accuracy, drift detection, and model degradation over time.
* Custom metrics: allows you to define and track custom metrics specific to their models and use cases.
* Visualization: Provides dashboards and visualizations to easily observe the status and performance of models.

There are two kinds of metrics present in Seldon Core 2 that you can monitor:

* [operational metrics](operational.md)
* [usage metrics](usage.md)

Operational metrics describe the performance of components in the system. Some examples of common operational\
considerations are memory consumption and CPU usage, request latency and throughput, and cache utilisation rates.\
Generally speaking, these are the metrics system administrators, operations teams, and engineers will be interested in.

Usage metrics describe the system at a higher and less dynamic level. Some examples include the number of deployed\
servers and models, and component versions. These are not typically metrics that engineers need insight into, but\
may be relevant to platform providers and operations teams.
