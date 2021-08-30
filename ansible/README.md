# Seldon Core Ansible Playbooks

To use these playbooks follow the installation of the Ansible collection at https://github.com/SeldonIO/ansible-k8s-collection

Once installed you can use the following Playbooks.

## Create Kind Cluster

```
ansible-playbook playbooks/kind.yaml
```


## Install Seldon Core with Istio

```
ansible-playbook playbooks/seldon_core.yaml
```


## Install Kafka

```
ansible-playbook playbooks/kafka.yaml
```

