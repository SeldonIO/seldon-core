---
title: "Getting started on minikube"
date: 2017-12-09T17:49:41Z
weight: 1
---

Seldon core uses [helm](https://github.com/kubernetes/helm) charts to start and runs on [kubernetes](https://kubernetes.io/) clusters. It can then run on a local minikube cluster. 

### Prerequisites

The following packages need to be installed on your machine in order to train the keras mnist example 

* python2.7 (we recommend [anconda distribution](link))
* grpcio-tools==1.1.3
* [sklearn==0.19.0](link)
* [minikube installed](https://kubernetes.io/docs/tasks/tools/install-minikube/)
* [helm installed](https://github.com/kubernetes/helm/blob/master/docs/install.md)

### Before starting: run a minikube cluster locally

Before starting, you need to have [minikube installed](https://kubernetes.io/docs/tasks/tools/install-minikube/) on your machine.

1. To start a minibuke local cluster, type on command line:

        minikube start --memory=8000

3. Starting minikube should automatically point your kubectl cli to the minikube cluster, but not your docker cli. To  make sure your docker cli is pointing at the minikube cluster, type on command line:
	
        eval $(minikube docker-env)

### Start seldon-core

You can now start seldon core in your minikube cluster.


1. Seldon core uses helm charts to start. To use seldon core, you need [helm installed](https://github.com/kubernetes/helm/blob/master/docs/install.md) in your machine. To Initialize helm, type on command line: 

         helm init

1. Clone seldon-core git repository and build all the required docker images locally using the provided bash script "build-all-in-minikube":

        git clone git@gitlab.com:seldon-dev/seldon-core.git

        cd seldon-core && ./build-all-in-minikube

1. Seldon-core repository include helm charts to start seldon-core. To start seldon-core using helm

        helm install helm-charts/seldon-core --name seldon-core --set grafana_prom_admin_password=password --set persistence.enabled=false --set cluster_manager.image.tag=0.3-SNAPSHOT --set apife.image.tag=0.1-SNAPSHOT --set engine.image.tag=0.2-SNAPSHOT


Seldon-core should now be running on your cluster. You can verify if all the pods are up and running typing on command line ```helm status seldon-core``` or ```kubectl get pods```

### Wrap your model

In this session, we show how to wrap the keras mnist classifier in the [seldon-core-example](link) repository using seldon-core python wrappers. 

1. Clone seldon-core-examples repositories in the same directory as seldon-core: 

        cd ../ && git clone git@gitlab.com:seldon-dev/seldon-core-examples.git

2. Train and save the keras mnist classifier example model using the provided scipt "train_mnist.py":

        cd seldon-core-examples/models/sklearn_iris/

        python train_iris.py

    This will train a keras convolutional neural network on mnist dataset for 2 epochs and save the model in the same folder.


3. Build protobuffers (this step requires grpc tools installed and has to be done only once. You can skip this step if done it before):

         cd ../../../seldon-core/wrappers

         make build_protos
    
4. Wrap the model using the wrap_model.py script:

        cd python

        python wrap_model.py ../../../seldon-core-examples/models/sklearn_iris IrisClassifier 0.0 seldonio --force
	
    This will create the folder build in keras_mnist. The --base-image argument is not specified and the wrapper will use the default base image Python:2.

5. Build a docker image of your model ready to deploy with seldon-core

	    cd ../../../seldon-core-examples/models/sklearn_iris/build/
	
	    make build_docker_image
    This will create the docker image ```seldonio/irisclassifier:0.0``` which is ready for [deployment with seldon-core](../../api/seldon-deployment).


### Deploy your model

The docker image version of your model is deployed through a json configuration file. A general template for the configuration can be found  [here](https://gitlab.com/seldon-dev/seldon-core-examples/blob/master/models/sklearn_iris/sklearn_iris_deployment.json). For the sklearn iris example, we have already created a deployment file "sklearn_iris_deployment.json":


    {
        "apiVersion": "machinelearning.seldon.io/v1alpha1",
        "kind": "SeldonDeployment",
        "metadata": {
            "labels": {
                "app": "seldon"
            },
            "name": "seldon-deployment-example"
        },
        "spec": {
            "annotations": {
                "project_name": "Iris classification",
                "deployment_version": "0.0"
            },
            "name": "sklearn-iris-deployment",
            "oauth_key": "oauth-key",
            "oauth_secret": "oauth-secret",
            "predictors": [
                {
                    "componentSpec": {
                        "spec": {
                            "containers": [
                                {
                                    "image": "seldonio/irisclassifier:0.0",
                                    "imagePullPolicy": "IfNotPresent",
                                    "name": "sklearn-iris-classifier",
                                    "resources": {
                                        "requests": {
                                            "memory": "1Mi"
                                        }
                                    }
                                }
                            ],
                            "terminationGracePeriodSeconds": 20
                        }
                    },
                    "graph": {
                        "children": [],
                        "name": "sklearn-iris-classifier",
                        "endpoint": {
                            "type" : "REST"
                        },
                        "subtype": "MICROSERVICE",
                        "type": "MODEL"
                    },
                    "name": "sklearn-iris-predictor",
                    "replicas": 1,
    	    	    "annotations": {
    	    	        "predictor_version" : "0.0"
                    }
                }
            ]
        }
    }


1. To deploy the model  in seldon core, type on command line:

        kubectl apply -f ../sklearn_iris_deployment.json
	
### Serve your  model:

1. Set the server host and port

        SERVER=192.168.99.100:30032

2. Get the authorization token:

        TOKEN=`curl -s -H "Accept: application/json" oauth-key:oauth-secret@${SERVER}/oauth/token -d grant_type=client_credentials | jq -r '.access_token'`

3. Send request prediction:

        curl -s -H "Content-Type:application/json" -H "Accept: application/json" -H "Authorization: Bearer $TOKEN" ${SERVER}/api/v0.1/predictions -d '{"data":{"names":["sepal length (cm)","sepal width (cm)", "petal length (cm)","petal width (cm)"],"ndarray":[[5.1,3.5,1.4,0.2]]}}'

You should see a response like

    {
         "meta": {
             "puid": "lhq41l3q736q7tnrij8o3lod8u",
             "tags": {
             },
             "routing": {
             }
         },
         "data": {
             "names": ["t:0", "t:1", "t:2"],
             "ndarray": [[0.8796816489561845, 0.12030753790659003, 1.0813137225507727E-5]]
        }
    }