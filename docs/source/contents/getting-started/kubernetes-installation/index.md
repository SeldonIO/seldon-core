# Kubernetes Installation

## Preparation

 1. Git clone seldon-core-v2
        git clone https://github.com/SeldonIO/seldon-core-v2
 2. Build [Seldon CLI](../cli.md)
 3. Create and authenticate to a Kubernetes cluster. For local testing you may want to try [Kind](https://kind.sigs.k8s.io/) or [Minikube](https://minikube.sigs.k8s.io/docs/).
 4. Install `make`.


## Extra Optional Requirements

 * To gain access to metrics you will need to install Prometheus.
 * To run Pipelines you will need to install Kafka.
 * To expose inference outside the cluster you will need to integration a service mesh of your choice. Some examples for Istio, Traefik and Ambassador are provided. We welcome help to extend these examples to other service meshes.

## Deploy

From the project root run:

```
make deploy-k8s
```

## Undeploy

From the project root run:

```
make undeploy-k8s
```

