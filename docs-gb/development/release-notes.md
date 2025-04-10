# Release Notes — v2.9.0 (7 April 2025)

The release of v2.9.0, packed with new features, improvements, and important fixes. This version focuses on enhanced configurability, improved autoscaling support, more robust logging, and expanded Helm chart capabilities.

## New Features

* gRPC Model Streaming: Introduced support for gRPC-based model streaming to improve performance and flexibility in model serving.
* Native Autoscaling Support:
  Added server-native autoscaling configuration.
  Helm support to enable/disable native autoscaling.
* Access Log Enhancements: Enabled access log configuration for Envoy via Helm.
* Configurable Logging: Logging levels can now be customized via Helm; Kafka clients (librdkafka) respect log levels.
* Partial Scheduling Support: Scheduler now considers minReplicas during partial scheduling decisions.
* LLM CRD Extension: Included support for LLM specs in the operator’s Custom Resource Definitions (CRDs).
* CLI as Kubernetes Deployment: CLI can now be deployed as a Kubernetes resource for debugging purposes.
* PVC Retention Policy via Helm: Added support for persistent volume claim (PVC) retention configuration.

## Fixes and Improvements

* Scheduler & Operator Stability:
  Prevented unloading of inactive model versions.
  Fixed model equality check logic to ignore runtime info.
  Scaling logic improvements for cases with missing maxReplicas.
  Improved scheduling logic to handle multiple model instances per server.
  Runtime podSpec overrides are now respected.
* Documentation Enhancements:
  Expanded docs on autoscaling, HPA, Core 2 dependencies, and model scheduling logic.
  Added documentation for new Helm configs and native scaling options.
* Refinements & Cleanups:
  Fixed various spelling, formatting, and spacing issues across documentation.
  Updated summaries, API sections, and dashboards.

## Dependency & Image Updates

* Multiple upgrades across components for enhanced security and performance:
  envoyproxy/envoy, grafana/grafana, google.golang.org/grpc, rclone/rclone, and others.
* Upgraded to Go 1.22 and pinned relevant tooling versions (e.g., xk6).
* Updated base container images (e.g., ubi9/openjdk-17-runtime, ubi9/ubi-micro).
  
## Infrastructure & DevOps

* Improved GitHub Actions workflows:
  Updated upload-artifact version.
  Fixed k6 image builds and switched default service endpoints.
  License information regenerated across components for compliance.
  Ansible dependency upgrades included for installation scripts.
  
## Documentation

* New and updated pages:
  Operational Monitoring IA
  Managed Kafka integration
  Core 2 introduction
  Securing endpoints and API references
For a full list of changes, please refer to the complete changelog.
