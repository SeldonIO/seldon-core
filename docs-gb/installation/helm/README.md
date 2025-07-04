---
description: Learn how to install and configure Seldon Core using Helm charts, including component setup and customization options.
---

# Helm Configuration Options

Seldon Core 2 provides a highly configurable deployment framework that allows you to fine-tune various components using Helm configuration options. These options offer control over deployment behavior, resource management, logging, autoscaling, and model lifecycle policies to optimize the performance and scalability of machine learning deployments.

This section details the key Helm configuration parameters for Envoy, Autoscaling, Server Prestop, and Model Control Plane, ensuring that you can customize deployment workflows and enhance operational efficiency.

* **Envoy**: Manage pre-stop behaviors and configure access logging to track request-level interactions.
* **Autoscaling** (Experimental): Fine-tune dynamic scaling policies for efficient resource allocation based on real-time inference workloads.
* **Servers**: Define grace periods for controlled shutdowns and optimize model control plane parameters for efficient model loading, unloading, and error handling.
* **Logging**: Define log levels for the different components of the system.


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
| `autoscaling.autoscalingModelEnabled` | components | Enable _native_ autoscaling for Models. This is orthogonal to external autoscaling services e.g. HPA. | false |
| `autoscaling.autoscalingServerEnabled` | components | Enable _native_ autoscaling for Models. This is orthogonal to external autoscaling services e.g. HPA. | true |
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


## Logging

### Component Log Level

| Key | Chart | Description | Default
| --- | --- | --- | --- |
| `logging.logLevel` | components | Components wide settings for logging level, if individual component levels are not set. Options are: debug, info, error. | info |
| `controller.logLevel` | components | check zap log level [here](https://pkg.go.dev/go.uber.org/zap#pkg-constants) | |
| `dataflow.logLevel` | components | | check klogging level [here](https://dokka.klogging.io/-klogging/io.klogging/-level/index.html) |
| `dataflow.logLevelKafka` | components | | check klogging level [here](https://dokka.klogging.io/-klogging/io.klogging/-level/index.html) |
| `scheduler.logLevel` | components | check logrus log level [here](https://pkg.go.dev/github.com/sirupsen/logrus#Level) | |
| `modelgateway.logLevel` | components | check logrus log level [here](https://pkg.go.dev/github.com/sirupsen/logrus#Level) | |
| `pipelinegateway.logLevel` | components | check logrus log level [here](https://pkg.go.dev/github.com/sirupsen/logrus#Level) | |
| `hodometer.logLevel` | components | check logrus log level [here](https://pkg.go.dev/github.com/sirupsen/logrus#Level) | |
| `serverConfig.rclone.logLevel` | components | check rclone `log-level` [here](https://rclone.org/docs/) | |
| `serverConfig.agent.logLevel` | components | check logrus log level [here](https://pkg.go.dev/github.com/sirupsen/logrus#Level)  | |

**Notes**:
- We set kafka client library log level from the log level that is passed to the component, which could be different to the level expected by `librdkafka` (syslog level). In this case we attempt to map the log level value to the best match.