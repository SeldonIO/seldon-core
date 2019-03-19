# Build using private local repository

## Prerequisites

* Local Docker
* Kubernetes cluster access
* Private local repository setup in the cluster with local access
    * use the project [k8s-local-docker-registry](https://github.com/SeldonIO/k8s-local-docker-registry)
    * "127.0.0.1:5000" will be used as the repo host url

## Prerequisite check

Ensure the prerequisites are in place and the correct ports available.

```bash
# Check that the private local registry works
(set -x && curl -X GET http://127.0.0.1:5000/v2/_catalog && \
        docker pull busybox && docker tag busybox 127.0.0.1:5000/busybox && \
        docker push 127.0.0.1:5000/busybox)
```

## Updating components and redeploying into cluster

Basic process of how to test code changes in cluster.

1. Stop seldon core if its running.
1. Build and push the component that was updated or all components if necessary.
1. Start seldon core.
1. Deploy models.

Below are details to achieve this.

### Building all components

Build all images and push to private local repository.

```bash
./build-all-private-repo
./push-all-private-repo
```

### start/stop Seldon Core

```bash
./start-seldon-core-private-repo
./stop-seldon-core-private-repo
```

### Building individual components

```bash
./cluster-manager/build-private-repo
./cluster-manager/push-private-repo

./api-frontend/build-private-repo
./api-frontend/push-private-repo

./engine/build-private-repo
./engine/push-private-repo
```

