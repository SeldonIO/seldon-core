# Helm Chart Configuration

## Seldon Core Operator Chart Configuration

|Parameter | Description | Default |
|----------|-------------|---------|
| ambassador.enabled | Whether to add Ambassador configuration to created services | true |
| ambassador.singleNamespace | Allow creation of Ambassador paths that don't include namespace | false |
| certManager.enabled | Whether to assume cert manager for certificates | false |
| controllerId | The ID for the manager. Only for when you want manager to limit itself to resources labelled with same id. | '' |
| crd.create | Whether to install the Custom Resource Definition | true |
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
| rbac.create | Create RBAC roles and bindings | true |
| rbac.configmap.create |  Create cluster wide rbacs rules to read configmaps and secrets | true |
| usageMetrics.enabled | Whether to send anonymous usage metrics | ```false``` |
| usageMetrics.datebase | URL for Spartakus DB | http://seldon-core-stats.seldon.io |
| webhook.port | Port for webhook | 443 |
