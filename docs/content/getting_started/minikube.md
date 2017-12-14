---
title: "Getting started on minikube"
date: 2017-12-09T17:49:41Z
---

Seldon core uses [helm](https://github.com/kubernetes/helm) charts to start and runs on [kubernetes](https://kubernetes.io/) clusters. 

## Getting started on minikube
### Prerequisites:

1. It is assumed you have a minikube local cluster running on your machine and that your kubetcl and docker cli are poiting at the cluster. If not, you have to
    * [Install minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/)
    * To start a minibuke local cluster, type on command line:
    
        ```minikube start --memory=8000```

    * Starting minikube should automatically point your kubectl cli to the minikube cluster, but not your docker cli. To  make sure your docker cli is pointing at the minikube cluster, type on command line:
	
        ```eval $(minikube docker-env)```
* It is assume that you have installed and initialized helm. If not, you have to
    * [Install helm](https://github.com/kubernetes/helm/blob/master/docs/install.md)
    * To initialize Helm, type on command line: 
    
        ```helm init```

### Starting seldon-core

1. The first step is to get seldon-core up and running on your minikube cluster. To  install seldon-core using helm, type on command line:
      	
    ```<get_helm_charts>```

    ```helm install <path_to_your_helm_charts_directory>/seldon-core --name seldon-core --set cluster_manager.image.tag=0.3-SNAPSHOT --set apife.image.tag=0.1-SNAPSHOT --set engine.image.tag=0.2-SNAPSHOT```
	
    Seldon-core should now be running on your cluster. You can verify if all the pods are up and running typing on command line:

    ```helm status seldon-core```

    or

    ```kubectl get pods```

2. You can now wrap one of seldon-example models using [seldon wrappers](link_to_wrappers_docs):

    * ...

    You should now have a local docker image named \<seldonio/\<image_name>:\<image_version>. You can verify typing on command line:

    ```docker images```

3. Deploy your wrapped model with seldon-core:

    * Open seldon json [deployment template](link_to_json_template) with your favorite editor and modify the "image" and "name" fields as follow:

        ```json
	{
	    ...
	    
	    "containers": [
	        {
	            "image": "<seldonio/<image_name>:<image_version>",
	            "imagePullPolicy": "IfNotPresent",
	            "name": "your-model-name",
	            ...
	        }
	    ]

        ...
	}
        ```

    * Save the json file as \<your_file_name>.json. To deploy it on seldon core, type on command line:

        ```kubectl apply -f <path_to_your_deployments_folder>/<your_file_name>.json```

4. Query and test your model:

    * ....

    * ....
    


