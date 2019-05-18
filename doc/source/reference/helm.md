# Helm Chart Configuration

## Seldon Core Controller Chart Configuration

### Usage Metrics

|Parameter | Description | Default |
|----------|-------------|---------|
| usageMetrics.enabled | Whether to send anonymous usage metrics | ```true``` |


### Controller settings

|Parameter | Description | Default |
|----------|-------------|---------|
| image.repository | Image repo | seldonio/seldon-core-operator |
| image.tag | Image repo | ```<latest release version>``` |
| image.pull_policy | Pull policy for image | IfNotPresent |

### Service Orchestrator (engine)

|Parameter | Description | Default |
|----------|-------------|---------|
| engine.image.name | Image to use for service orchestrator | ```<latest release image>``` |

### General Role Based Access Control Settings

These settings should generally be left untouched from their defaults. Use of non-rbac clusters will not be supported in the future. Most of these settings are used by the Google marketplace one click installation for Seldon Core as in this setting Google will create the service account and role bindings.

|Parameter | Description | Default |
|----------|-------------|---------|
| rbac.enabed | Whether to enabled RBAC | true |
| rbac.rolebinding.create | Whether to include role binding k8s settings | true |
| rbac.service_account.create | Whether to create the service account to use | true |
| rbac.service_account.name | The name of the service account to use | seldon |


## Seldon Core OAuth Gateway

### Seldon Core API OAuth Gateway (apife)

|Parameter | Description | Default |
|----------|-------------|---------|
| enabled| Whether to enable the default Oauth API gateway | true |
| image.repository | The image repository for the API gateway | seldonio/apife |
| image.tag | The tag for the image | ```<latest release image>``` |
| image.pull_policy | The pull policy for apife image | IfNotPresent |
| service_type | The expose service type, e.g. NodePort, LoadBalancer | NodePort |
| annotations | Configuration annotations | empty |

