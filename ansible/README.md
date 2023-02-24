## Ansible Setup for Seldon Core V2


### Installing Ansible

Provided Ansible playbooks and roles depends on [kubernetes.core](https://github.com/ansible-collections/kubernetes.core) Ansible Collection for performing `kubectl` and `helm` operations.
Check Ansible [documentation] for further information.


To install Ansible and required collections
```bash
pip install ansible openshift kubernetes docker
ansible-galaxy collection install kubernetes.core
```

We have tested provided instructions on Python 3.8 - 3.11 with following version Python libraries

| Python | Ansible | Docker | Kubernetes |
|--------|---------|--------|------------|
| 3.8    | 6.7.0   | 6.0.1  | 26.1.0     |
| 3.9    | 7.2.0   | 6.0.1  | 26.1.0     |
| 3.10   | 7.2.0   | 6.0.1  | 26.1.0     |
| 3.11   | 7.2.0   | 6.0.1  | 26.1.0     |

and `kubernetes.core` collection in version `2.4.0`.


Once installed you can use the following Playbooks that you will find in [Ansible](https://github.com/SeldonIO/seldon-core/tree/v2/ansible) folder of Seldon Core V2 repository.

You also need to have installed `kubectl`

### Create Kind Cluster

```bash
ansible-playbook playbooks/kind-cluster.yaml
```

### Setup Ecosystem

Seldon runs by default in `seldon-mesh` namespace and a Jaeger pod and  and OpenTelemtry collector are installed in the chosen namespace. Run the following:

```bash
ansible-playbook playbooks/setup-ecosystem.yaml
```

### Ecosystem configuration options

The ecosystem setup can be parametrized by providing extra Ansible variables, e.g. using `-e` flag to `ansible-playbook` command.

For example run the following from the `ansible/` folder:
```bash
ansible-playbook playbooks/setup-ecosystem.yaml -e full_install=no -e install_kafka=yes
```
will only install Kafka when setting up the ecosystem.

|                         | type   | default                     | comment                                                 |
|-------------------------|--------|-----------------------------|---------------------------------------------------------|
| seldon_kafka_namespace  | string | seldon-mesh                 | namespace to install Kafka                              |
| seldon_mesh_namespace   | string | seldon-mesh                 | namespace to install Seldon                             |
| full_install            | bool   | yes                         | enables full ecosystem installation                     |
| install_kafka           | bool   | {{ full_install }}          | installs Kafka using seldonio.k8s.strimzi_kafka         |
| install_prometheus      | bool   | {{ full_install }}          | installs Prometheus using seldonio.k8s.prometheus       |
| install_certmanager     | bool   | {{ full_install }}          | installs certmanager using seldonio.k8s.certmanager     |
| install_jaeger          | bool   | {{ full_install }}          | installs Jaeger using seldonio.k8s.jaeger               |
| install_opentelemetry   | bool   | {{ full_install }}          | installs OpenTelemetry using seldonio.k8s.opentelemetry |
| configure_kafka         | bool   | {{ install_kafka }}         | configures Kafka using V2 specific resources            |
| configure_prometheus    | bool   | {{ install_prometheus }}    | configure Prometheus using V2 specific resources        |
| configure_jaeger        | bool   | {{ install_jaeger }}        | configure Jaeger using V2 specific resoruces            |
| configure_opentelemetry | bool   | {{ install_opentelemetry }} | configure OpenTelemetry using V2 specific resources     |

The most common change will be to install in another namespace with:

```bash
ansible-playbook playbooks/setup-ecosystem.yaml -e seldon_mesh_namespace=<mynamespace>
```

### Install Seldon Core V2

Run the following from the `ansible/` folder:

```bash
ansible-playbook playbooks/setup-seldon.yaml
```

If you have changed the namespace you wish to use you will need to run with:

```bash
ansible-playbook playbooks/setup-seldon.yaml -e seldon_mesh_namespace=<mynamespace>
```

|                         | type   | default                     | comment                                                 |
|-------------------------|--------|-----------------------------|---------------------------------------------------------|
| seldon_kafka_namespace  | string | seldon-mesh                 | namespace to install Kafka                              |
| seldon_mesh_namespace   | string | seldon-mesh                 | namespace to install Seldon                             |
| seldon_crds_namespace   | string | default                     | namespace to install Seldon CRDs                        |
| full_install            | bool   | yes                         | enables full ecosystem installation                     |
| install_crds            | bool   | {{ full_install }}          | installs Seldon CRDs                                    |
| install_components      | bool   | {{ full_install }}          | install Seldon components                               |
| install_servers         | bool   | {{ full_install }}          | install Seldon servers                                  |
