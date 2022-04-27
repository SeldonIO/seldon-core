# Ansible Setup for Seldon Core V2

To use these playbooks follow the installation of the Ansible collection at https://github.com/SeldonIO/ansible-k8s-collection
```bash
ansible-galaxy collection install git+https://github.com/SeldonIO/ansible-k8s-collection.git
```

Once installed you can use the following Playbooks.

## Create Kind Cluster

```bash
ansible-playbook playbooks/kind_cluster.yaml
```


## Setup Ecosystem

```bash
ansible-playbook playbooks/setup-ecosystem.yaml
```

### Ecosystem configuration options

The ecosystem setup can be parametrized by providing extra Ansible variables, e.g. using `-e` flag to `ansible-playbook` command.

For example
```bash
ansible-playbook playbooks/setup-ecosystem.yaml -e full_install=no -e install_kafka=yes
```
will only install Kafka when setting up the ecosystem.


| flag                 | type | default                  | comment                                           |
|----------------------|------|--------------------------|---------------------------------------------------|
| full_install         | bool | yes                      | enables full ecosystem installation               |
| install_kafka        | bool | {{ full_install }}       | installs Kafka using seldonio.k8s.strimzi_kafka   |
| install_prometeus    | bool | {{ full_install }}       | installs Prometheus using seldonio.k8s.prometheus |
| configure_kafka      | bool | {{ install_kafka }}      | configures Kafka using V2 specific resources      |
| configure_prometheus | bool | {{ install_prometheus }} | configure Prometheus using V2 specific resources  |


## Install Seldon Core V2
```bash
ansible-playbook playbooks/setup-seldon-v2.yaml
```
