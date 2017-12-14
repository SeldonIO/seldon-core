---
title: "Starting seldon core"
date: 2017-12-09T17:49:41Z
---

Seldon core used [helm](link to helm) charts to start and runs on [kubernetes clusters](link_to_kubernetes). 

### Getting started on minikube

1. The first step is to get seldon-core up and running on your cluster:

    * [Install minikube](link_to_minikube)
    * Start a minibuke local cluster:
    
        ```minikube start --memory=8000```

    * Starting minikube should automatically point your kubectl cli to the minikube cluster, but not your docker cli. Make sure your docker is pointing at the minikube cluster
	
        ```eval $(minikube docker-env)```
    
    * [Install helm](https://github.com/kubernetes/helm/blob/master/docs/install.md)
    * Initialize Helm:

        ```cd <your_helm_directory>```
    
        ```helm init```

    * Install seldon-core using helm:

        ```helm install seldon-core .....```

    Seldon-core should now be running on your cluster. You can verify if all the pods are up and running typing on terminal:

    ```kubectl get pods```

2. Wrap one of seldon-example models using [seldon wrappers](link_to_wrappers_docs):

    * ....

    * ....

    You should now have a local docker image named <seldonio/<image_name>:<image_version>>. You can verify typing on terminal:

    ```docker images```

3. Deploy  model with seldon-core:

    * ....

    * ....

4. Query and test model:

    * ....

    * ....
    


