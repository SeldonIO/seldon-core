# Ansible Setup for Seldon Core v1

> :warning: **NOTE:** Ansible way of installing Seldon Core and associated ecosystem is meant for dev/testing purposes. For production use cases follow [Helm installation](https://docs.seldon.io/projects/seldon-core/en/stable/workflow/install.html)

## Installing Ansible

Provided Ansible playbooks and roles depends on [kubernetes.core](https://github.com/ansible-collections/kubernetes.core) Ansible Collection for performing `kubectl` and `helm` operations.
Check Ansible [documentation] for further information.


To install Ansible and required collections
```bash
pip install ansible openshift kubernetes docker
ansible-galaxy collection install kubernetes.core
```

We have tested provided instructions on Python 3.8 - 3.11 with following version of Python libraries

| Python | Ansible | Docker | Kubernetes |
|--------|---------|--------|------------|
| 3.8    | 6.7.0   | 6.0.1  | 26.1.0     |
| 3.9    | 7.2.0   | 6.0.1  | 26.1.0     |
| 3.10   | 7.2.0   | 6.0.1  | 26.1.0     |
| 3.11   | 7.2.0   | 6.0.1  | 26.1.0     |

and `kubernetes.core` collection in version `2.4.0`.


Once installed you can use the following Playbooks that you will find in [Ansible](https://github.com/SeldonIO/seldon-core/tree/v2/ansible) folder of Seldon Core V2 repository.

You also need to have installed [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/).

## Create Kind Cluster

```bash
ansible-playbook playbooks/kind-cluster.yaml
```

To deploy Kind cluster with 4 workers add `-e kind_use_many_workers=yes` flag.


## Install Seldon Core with Istio

```
ansible-playbook playbooks/main.yaml
```


## Install Kafka

```bash
ansible-playbook playbooks/kafka.yaml
```


## Side notes

__N.B:__ If you are using MacOS and have an error saying `in progress in another thread when fork() was called` when installing something with ansible. You might want to set `export OBJC_DISABLE_INITIALIZE_FORK_SAFETY=YES`

See issue [here](https://github.com/ansible/ansible/issues/32499#issuecomment-341578864)
