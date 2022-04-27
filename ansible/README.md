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


## Install Seldon Core V2
```bash
ansible-playbook playbooks/setup-seldon-v2.yaml
```
