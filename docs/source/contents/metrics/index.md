# Metrics

There are two kinds of metrics present in Seldon Core v2:
* [operational metrics](./operational.md)
* [usage metrics](./usage.md)

Operational metrics describe the performance of components in the system.
Examples of common operational considerations are memory consumption and CPU usage, request latency and throughput, and cache utilisation rates.
Generally speaking, these are the metrics system administrators, operations teams, and engineers will be interested in.

Usage metrics describe the system at a higher and less dynamic level.
Examples include the number of deployed servers and models, and component versions.
These are not typically metrics that engineers need insight into, but may be relevant to platform providers and operations teams.

```{toctree}
:maxdepth: 1
:hidden:

operational.md
usage.md
```
