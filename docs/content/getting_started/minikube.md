---
title: "Getting started on minikube"
date: 2017-12-09T17:49:41Z
---

Seldon core uses [helm](https://github.com/kubernetes/helm) charts to start and runs on [kubernetes](https://kubernetes.io/) clusters. 

## Getting started on minikube

1. The first step is to get seldon-core up and running on your minikube cluster:

    * [Install minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/)
    * Start a minibuke local cluster:
    
        ```minikube start --memory=8000```

    * Starting minikube should automatically point your kubectl cli to the minikube cluster, but not your docker cli. Make sure your docker is pointing at the minikube cluster:
	
        ```eval $(minikube docker-env)```
    
    * [Install helm](https://github.com/kubernetes/helm/blob/master/docs/install.md)
    * Initialize Helm and download seldon core helm charts:
    
        ```helm init```

        ```<get_helm_charts>```

    * Install seldon-core using helm:

        ```helm install <path_to_your_helm_charts_directory>/seldon-core --name seldon-core```
	
    Seldon-core should now be running on your cluster. You can verify if all the pods are up and running typing on terminal:

    ```helm status seldon-core```

    or

    ```kubectl get pods```

2. You can now wrap one of seldon-example models using [seldon wrappers](link_to_wrappers_docs):

    * ....

    * ....

    You should now have a local docker image named \<seldonio/\<image_name>:\<image_version>. You can verify typing on terminal:

    ```docker images```

3. Deploy your wrapped model with seldon-core:

    * ....

    * ....

4. Query and test your model:

    * ....

    * ....
    


