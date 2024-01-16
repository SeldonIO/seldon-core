# About 

Seldon V2 APIs provide a state of the art solution for machine learning inference which can be run locally on a laptop as well as on Kubernetes for production.

## Features

 * A single platform for inference of wide range of standard and custom artifacts.
 * Deploy locally in Docker during development and testing of models.
 * Deploy at scale on Kubernetes for production.
 * Deploy single models to multi-step pipelines.
 * Save infrastructure costs by deploying multiple models transparently in inference servers.
 * Overcommit on resources to deploy more models than available memory.
 * Dynamically extended models with pipelines with a data-centric perspective backed by Kafka.
 * Explain individual models and pipelines with state of the art explanation techniques.
 * Deploy drift and outlier detectors alongside models.
 * Kubernetes Service mesh agnostic - use the service mesh of your choice.

## Publication

These features are influenced by our position paper on the next generation of ML model serving frameworks:

*Title*: [Desiderata for next generation of ML model serving](http://arxiv.org/abs/2210.14665)

*Workshop*: Challenges in deploying and monitoring ML systems workshop - NeurIPS 2022


## Getting started

### Local quick-start via `docker-compose`

Deploy via Docker Compose

```
make deploy-local
```

Run [local-examples.ipynb](samples/local-examples.ipynb)


Undeploy

```
make undeploy-local
```

### Kubernetes quick-start via `KinD`

Install Seldon ansible collection

```
pip install ansible openshift docker passlib
ansible-galaxy collection install git+https://github.com/SeldonIO/ansible-k8s-collection.git
```

Create a KinD cluster and install dependencies:

```
cd ansible
ansible-playbook playbooks/kind-cluster.yaml
ansible-playbook playbooks/setup-ecosystem.yaml
```

Deploy Seldon Core v2

```
cd ..
make deploy-k8s
```

Run [k8s-examples.ipynb](samples/k8s-examples.ipynb)

Undeploy Seldon Core v2

```
make undeploy-k8s
```


## Documentation

[Seldon Core v2 docs](https://docs.seldon.io/projects/seldon-core/en/v2/)

## License

[License](LICENSE)