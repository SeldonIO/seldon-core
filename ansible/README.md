## Ansible Setup for Seldon Core V2

To use these playbooks follow the installation of the Ansible collection at https://github.com/SeldonIO/ansible-k8s-collection
```bash
ansible-galaxy collection install git+https://github.com/SeldonIO/ansible-k8s-collection.git
```

Once installed you can use the following Playbooks.

### Create Kind Cluster

```bash
ansible-playbook playbooks/kind_cluster.yaml
```


### Setup Ecosystem

Run the following from the ansible folder:

```bash
ansible-playbook playbooks/setup-ecosystem.yaml
```

### Ecosystem configuration options

The ecosystem setup can be parametrized by providing extra Ansible variables, e.g. using `-e` flag to `ansible-playbook` command.

For example run the following from the ansible folder:
```bash
ansible-playbook playbooks/setup-ecosystem.yaml -e full_install=no -e install_kafka=yes
```
will only install Kafka when setting up the ecosystem.


|                         | type | default                     | comment                                                 |
|-------------------------|------|-----------------------------|---------------------------------------------------------|
| full_install            | bool | yes                         | enables full ecosystem installation                     |
| install_kafka           | bool | {{ full_install }}          | installs Kafka using seldonio.k8s.strimzi_kafka         |
| install_prometeus       | bool | {{ full_install }}          | installs Prometheus using seldonio.k8s.prometheus       |
| install_certmanager     | bool | {{ full_install }}          | installs certmanager using seldonio.k8s.certmanager     |
| install_jaeger          | bool | {{ full_install }}          | installs Jaeger using seldonio.k8s.jaeger               |
| install_opentelemetry   | bool | {{ full_install }}          | installs OpenTelemetry using seldonio.k8s.opentelemetry |
| configure_kafka         | bool | {{ install_kafka }}         | configures Kafka using V2 specific resources            |
| configure_prometheus    | bool | {{ install_prometheus }}    | configure Prometheus using V2 specific resources        |
| configure_jaeger        | bool | {{ install_jaeger }}        | configure Jaeger using V2 specific resoruces            |
| configure_opentelemetry | bool | {{ install_opentelemetry }} | configure OpenTelemetry using V2 specific resources     |


### Install Seldon Core V2

Run the following from the ansible folder:

```bash
ansible-playbook playbooks/setup-seldon-v2.yaml
```
