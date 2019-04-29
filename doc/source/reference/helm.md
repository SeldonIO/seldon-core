# Helm Chart Configuration

The core choice in using the helm chart is to decide if you want to use Ambassador or the internal API OAuth gateway for ingress.

## Seldon Core Chart Configuration

### Seldon Core API OAuth Gateway (apife)

|Parameter | Description | Default |
|----------|-------------|---------|
|apife.enabled| Whether to enable the default Oauth API gateway | true |
|apife.image.name | The image to use for the API gateway | ```<latest release image>``` |
|apife.image.pull_policy | The pull policy for apife image | IfNotPresent |
|apife.service_type | The expose service type, e.g. NodePort, LoadBalancer | NodePort |
|apife.annotations | Configuration annotations | empty |

### Seldon Core Operator (ClusterManager)

|Parameter | Description | Default |
|----------|-------------|---------|
| cluster_manager.image.name | Image to use for Operator | ```<latest release image>```|
| cluster_manager.image.pull_policy | Pull policy for image | IfNotPresent |
| cluster_manager.java_opts | Extra Java Opts to pass to app | empty |
| cluster_manager.spring_opts | Spring specific opts to pass to app | empty |

### Service Orchestrator (engine)

|Parameter | Description | Default |
|----------|-------------|---------|
| engine.image.name | Image to use for service orchestrator | ```<latest release image>``` |

### Ambassador Reverse Proxy

|Parameter | Description | Default |
|----------|-------------|---------|
| ambassador.enabled | Whether to enable the ambbassador reverse proxy | false |
| ambassador.annotations | Configuration for Ambassador | default |
| ambassador.image.repository | Image name to use for ambassador | ```<tested release with seldon>``` |
| ambassador.image.tag | Image tag to use for ambassador | ```<tested release with seldon>``` |
| ambassador.service.type | How to expose the ambassador service, e.g. NodePort, LoadBalancer | NodePort |

For more see https://github.com/helm/charts/tree/master/stable/ambassador

### General Role Based Access Control Settings

These settings should generally be left untouched from their defaults. Use of non-rbac clusters will not be supported in the future. Most of these settings are used by the Google marketplace one click installation for Seldon Core as in this setting Google will create the service account and role bindings.

|Parameter | Description | Default |
|----------|-------------|---------|
| rbac.enabed | Whether to enabled RBAC | true |
| rbac.rolebinding.create | Whether to include role binding k8s settings | true |
| rbac.service_account.create | Whether to create the service account to use | true |
| rbac.service_account.name | The name of the service account to use | seldon |


### Redis

Redis is used by Seldon Core for :
  * Holding OAuth tokens for the API gateway
  * Saving state of some components

|Parameter | Description | Default |
|----------|-------------|---------|
| redis.image.name | Image to use for Redis | ```<latest tested with seldon>```|

