# Helm Configuration Options

## Envoy

### Prestop

| Key | Chart | Description | Default
| --- | --- | --- | --- |
| `envoy.preStopSleepPeriodSeconds` | components | Sleep after calling prestop command. | 30 |
| `envoy.terminationGracePeriodSeconds` | components | Grace period to wait for prestop to finish for Envoy pods. | 120 |

### Access Log

| Key | Chart | Description | Default
| --- | --- | --- | --- |
| `envoy.enableAccesslog` | components | Whether to enable logging of requests. | true |
| `envoy.accesslogPath` | components | Path on disk to store logfile. This is only used when `enableAccesslog` is set. | /tmp/envoy-accesslog.txt |
| `envoy.includeSuccessfulRequests` | components | Whether to including successful requests. If set to false, then only failed requests are logged. This is only used when `enableAccesslog` is set.  | false |

## Autoscaling

### Native autoscaling (experimental)

| Key | Chart | Description | Default
| --- | --- | --- | --- |
| `agent.scalingStatsPeriodSeconds` | components | Sampling rate for metrics used for autoscaling. | 20 |
| `agent.modelInferenceLagThreshold` | components | Queue lag threshold to trigger scaling up of a model replica. | 30 |
| `agent.modelInactiveSecondsThreshold` | components | Period with no activity after which to trigger scaling down of a model replica. | 600 |
| `autoscaling.serverPackingEnabled` | components | Whether packing of models onto fewer servers is enabled. | false |
| `autoscaling.serverPackingPercentage` | components | Percentage of events where packing is allowed. Higher values represent more aggressive packing. This is only used when `serverPackingEnabled` is set. Range is from 0.0 to 1.0 | 0.0 |


## Server

### Prestop

| Key | Chart | Description | Default
| --- | --- | --- | --- |
| `serverConfig.terminationGracePeriodSeconds` | components | Grace period to wait for prestop process to finish for this particular Server pod. | 120 |


### Model Control Plane

| Key | Chart | Description | Default
| --- | --- | --- | --- |
| `agent.overcommitPercentage` | components | Overcommit percentage (of memory) allowed. Range is from 0 to 100 | 10 |
| `agent.maxLoadElapsedTimeMinutes` | components | Max time allowed for one model load command for a model on a particular server replica to take. Lower values allow errors to be exposed faster.  | 120 |
| `agent.maxLoadRetryCount` | components | Max number of retries for unsuccessful load command for a model on a particular server replica. Lower values allow control plane commands to fail faster.  | 5 |
| `agent.maxUnloadElapsedTimeMinutes` | components | Max time allowed for one model unload command for a model on a particular server replica to take. Lower values allow errors to be exposed faster.  | 15 |
| `agent.maxUnloadRetryCount` | components | Max number of retries for unsuccessful unload command for a model on a particular server replica. Lower values allow control plane commands to fail faster.  | 5 |
| `agent.unloadGracePeriodSeconds` | components | A period guarding against race conditions between Envoy actually applying the cluster change to remove a route and before proceeding with the model replica unloading command.  | 2 |