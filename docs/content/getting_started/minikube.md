---
title: "Getting started on minikube"
date: 2017-12-09T17:49:41Z
weight: 1
---

Seldon core uses [helm](https://github.com/kubernetes/helm) charts to start and runs on [kubernetes](https://kubernetes.io/) clusters. It can then run on a local minikube cluster. 

### Before starting

To start a minikube cluster lacally in your machine, you have to

* [Install minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/)
* To start a minibuke local cluster, type on command line:
    
        minikube start --memory=8000

* Starting minikube should automatically point your kubectl cli to the minikube cluster, but not your docker cli. To  make sure your docker cli is pointing at the minikube cluster, type on command line:
	
        eval $(minikube docker-env)

### Start seldon-core

You can now start seldon core in your minikube cluster.

1. To use seldon core, you need helm installed and initialized:

    * [Install helm](https://github.com/kubernetes/helm/blob/master/docs/install.md)
    * Initialize helm. Type on command line: 

            helm init
* To  install seldon-core using helm, type on command line:

        helm install <seldon_core_helm_charts>
	

Seldon-core should now be running on your cluster. You can verify if all the pods are up and running typing on command line ```helm status seldon-core``` or ```kubectl get pods```

### Wrap your model

In this session, we show how to wrap the keras mnist classifier in the [seldon-core-example](link) repository using seldon-core python wrappers. 

1. Clone seldon-core and seldon-core-examples repositories in the same directory: 

        git clone seldon-core 

        git clone seldon-core-examples

2. Train and save the keras mnist classifier example model using the provided scipt "train_mnist.py":

        cd seldon-core-examples/keras_mnist

        python train_mnist.py

    This will train a keras convolutional neural network on mnist dataset for 2 epochs and save the model in the same folder.


3. Build protobuffers (this step requires grpc tools installed and has to be done only once. You can skip this step if done it before):

         cd ../../seldon-core/wrappers

         make build_protos
    
4. Wrap the model using the wrap_model.py script:

        cd python

        python wrap_model.py ../../../seldon-core-examples/keras_mnist MnistClassifier 0.0 seldonio
	
    This will create the folder build in keras_mnist. The --base-image argument is not specified and the wrapper will use the default base image Python:2.

5. Build a docker image of your model ready to deploy with seldon-core

	    cd ../../../seldon-core-plugins/keras_mnist/build/
	
	    make build_docker_image
    This will create the docker image ```seldonio/mnistclassifier:0.0``` which is ready for [deployment with seldon-core](../../api/seldon-deployment).


### Deploy and serve your model

1. Open seldon json [deployment template](../../api/seldon-deployment) and enter your deployment specifications. For the keras mnist example the deployment json file should look like this

    ```json
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
                "project_name": "Keras MNIST classifier",
                "deployment_version": "0.0"
            },
            "name": "test-deployment",
            "oauth_key": "oauth-key",
            "oauth_secret": "oauth-secret",
            "predictors": [
                {
                    "componentSpec": {
                        "spec": {
                            "containers": [
                                {
                                    "image": "seldonio/mnistclassifier:0.0",
                                    "imagePullPolicy": "IfNotPresent",
                                    "name": "keras-mnist-classifier",
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
                        "name": "keras-mnist-classifier",
                        "endpoint": {
                            "type" : "REST"
                        },
                        "subtype": "MICROSERVICE",
                        "type": "MODEL"
                    },
                    "name": "keras-mnist-classifier",
                    "replicas": 1,
    	    	    "annotations": {
    	    	        "predictor_version" : "0.0"
                    }
                }
            ]
        }
    }
    ```

2. Save the json file as deployment_keras_mnist.json. To deploy it on seldon core, type on command line:

        kubectl apply -f <path_to_your_deployments_folder>/deployment_keras_mnist.json
