# Seldon Core Ansible Playbooks

To use these playbooks follow the installation of the Ansible collection at https://github.com/SeldonIO/ansible-k8s-collection

Once installed you can use the following Playbooks.

## Create Kind Cluster

```
ansible-playbook playbooks/kind_cluster.yaml
```


## Install Seldon Core with Istio

```
ansible-playbook playbooks/seldon_core.yaml
```


## Install Kafka

```
ansible-playbook playbooks/kafka.yaml
```


__N.B:__ If you are using MacOS and have an error saying `in progress in another thread when fork() was called` when installing something with ansible. You might want to set `export OBJC_DISABLE_INITIALIZE_FORK_SAFETY=YES` 

See issue [here](https://github.com/ansible/ansible/issues/32499#issuecomment-341578864)
