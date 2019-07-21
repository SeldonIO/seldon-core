# Helm Chart Configuration

## Seldon Core Operator Chart Configuration

|Parameter | Description | Default |
|----------|-------------|---------|
| ambassador.enabled | Whether to add Ambassador configuration to created services | true |
| ambassador.singleNamespace | Allow creation of Ambassador paths that don't include namespace | false |
| engine.grpc.port | gRPC port | 5001 |
| engine.image.name | Image to use for service orchestrator | ```<latest release image>``` |
| engine.image.tag | Tag for service orchestrator | ```<latest release image>``` |
| engine.image.pullPolicy | Pull policy for service orchestrator | IfNotPresent |
| engine.port | HTTP port | 8000 |
| engine.prometheus.path | Path to expose Prometheus metrics | /prometheus |
| engine.serviceAccount.name | Name of service account to use | default |
| engine.user | User ID to run the app | 8888 |
| image.repository | Operator image repo | seldonio/seldon-core-operator |
| image.tag | Image repo | Operator ```<latest release version>``` |
| image.pullPolicy | Operator pull policy | IfNotPresent |
| rbac.enabled | Whether to enabled RBAC | true |
| usageMetrics.enabled | Whether to send anonymous usage metrics | ```false``` |
| usageMetrics.datebase | URL for Spartakus DB | http://seldon-core-stats.seldon.io |


