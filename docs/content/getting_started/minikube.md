---
title: "Getting started on minikube"
date: 2017-12-09T17:49:41Z
---

Seldon core uses [helm](https://github.com/kubernetes/helm) charts to start and runs on [kubernetes](https://kubernetes.io/) clusters. It can then run on a local minikube cluster. 

### Getting started on minikube

To start a minikube cluster lacally in your machine, you have to

* [Install minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/)
* To start a minibuke local cluster, type on command line:
    
    ```minikube start --memory=8000```

* Starting minikube should automatically point your kubectl cli to the minikube cluster, but not your docker cli. To  make sure your docker cli is pointing at the minikube cluster, type on command line:
	
    ```eval $(minikube docker-env)```

### Starting seldon-core

You can now start seldon core in your minikube cluster as explained in [getting started](../helm_start).

