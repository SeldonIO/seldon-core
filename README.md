## Kubernetes Local Setup

## Local Quick Start

Deploy via Docker Compose

```
make deploy-local
```

Undeploy

```
make undeploy-local
```

Run [local-examples.ipynb](sample/local-examples.ipynb)

## Kind Quickstart

Install Seldon ansible collection

```
pip install ansible openshift docker passlib
ansible-galaxy collection install git+https://github.com/SeldonIO/ansible-k8s-collection.git
```

Create a KinD cluster:

```
ansible-playbook seldonio.k8s.kind
```

Deploy Seldon Core v2

```
make deploy-k8s
```

Undeploy Seldon Core v2

```
make undeploy-k8s
```

