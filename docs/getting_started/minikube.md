
# Getting started on Minikube


In this guide, we will show how to create, deploy and serve a iris classification model using seldon-core running on a Minikube cluster. Seldon-core uses [helm](https://github.com/kubernetes/helm) charts to start and runs on [kubernetes](https://kubernetes.io/) clusters. Minikube is a tool that makes it easy to run Kubernetes locally,  and runs a single-node Kubernetes cluster inside a VM on your laptop. 


### Prerequisites

The following packages need to be installed on your machine.

* [minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/) >= 0.24.0
* [helm](https://github.com/kubernetes/helm/blob/master/docs/install.md) >= 2.7.0

* [sklearn](http://scikit-learn.org/stable/) 
  - Sklearn is needed only to train the iris classifier example below. Seldon-core doesn't not require sklearn installed on your machine  to run.


### Before starting: run a Minikube cluster locally

Before starting, you need to have [minikube installed](https://kubernetes.io/docs/tasks/tools/install-minikube/) on your machine.

1. Start a Kubernetes local cluster in your machine using Minikube:

    ```bash
    minikube start --memory=8000 --feature-gates=CustomResourceValidation=true
    ```
    
Once the cluster is created, Minikube should automatically point your kubectl cli to the minikube cluster.

### Start seldon-core

You can now start seldon-core in your minikube cluster.


1. Seldon-core uses helm charts to start. To use seldon-core, you need [helm installed](https://github.com/kubernetes/helm/blob/master/docs/install.md) in your machine. To Initialize helm, type on command line:

    ```bash
    helm init
    ```

1. Seldon-core uses helm charts to start which are stored in google storage. To start seldon-core using helm:

    ```bash
     helm install seldon-core --name seldon-core \
     --repo https://storage.googleapis.com/seldon-charts
    ```

Seldon-core should now be running on your cluster. You can verify if all the pods are up and running typing on command line ```helm status seldon-core``` or ```kubectl get pods```

### Wrap your model

In this session, we show how to wrap the sklearn iris classifier in the [seldon-core-example](link) repository using seldon-core python wrappers. The example consist of a logistic regression model trained on the  [iris dataset](link_iris).

1. Clone seldon-core-examples repository:

    ```bash
    git clone https://github.com/SeldonIO/seldon-core-examples
    ```

2. Train and save the sklearn iris classifier example model using the provided script ```train_iris.py```:

    ```bash
    cd seldon-core-examples/models/sklearn_iris/
    ```
    ```bash
    python train_iris.py
    ````
    
    This will train a simple logistic regression model on the iris dataset and save the model in the same folder.


3. Wrap your saved model using the bash script wrap-model-in-minikube in ```seldon-core-examples```::
    ```bash
    cd ../../
    ```
    ```bash
    ./wrap-model-in-minikube models/sklearn_iris IrisClassifier 0.1 seldonio --force
    ```
    
    This will create the docker image ```seldonio/irisclassifier:0.1``` inside the minikube cluster  which is ready for deployment with seldon-core.


### Deploy your model

The docker image version of your model is deployed through a json configuration file. A general template for the configuration can be found  [here](https://github.com/SeldonIO/seldon-core-examples/blob/master/models/sklearn_iris/sklearn_iris_deployment.json). For the sklearn iris example, we have already created a deployment file ```sklearn_iris_deployment.json``` in the ```sklearn_iris``` folder:

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
            "project_name": "Iris classification",
            "deployment_version": "0.1"
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
                                "image": "seldonio/irisclassifier:0.1",
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
                    "predictor_version" : "0.1"
                }
            }
        ]
    }
}
```

1. To deploy the model  in seldon core, type on command line:

    ```bash
    kubectl apply -f models/sklearn_iris/sklearn_iris_deployment.json
    ```

2. The deployment will take a few seconds to be ready. To check if your deployment is ready:

    ```bash
    kubectl describe seldondeployments seldon-deployment-example
    ```

        
	
### Serve your  model:

In order to send a prediction request to your model, you need to query the seldon-core api server and obtain an authentication token from your model key and secret (in this example key and secret are set to "oauth-key" and "oauth-secret" for simplicity). The api server is running on minikube with default port 30032. To query your model you need to

1. Set the api server IP and port:

    ```bash
     SERVER=$(minikube ip):30032
    ```

2. Get the authorization token using your key ("oauth-key" here) and secret ("oauth-secret"):

    ```bash
    TOKEN=`curl -s -H "Accept: application/json" \
    oauth-key:oauth-secret@${SERVER}/oauth/token -d grant_type=client_credentials | jq -r '.access_token'`
    ````

3. Query the api server prediction endpoint. The json object at the end is your message containing the values for your features:
    ```bash
    curl -s -H "Content-Type:application/json" -H "Accept: application/json" \
    -H "Authorization: Bearer $TOKEN" ${SERVER}/api/v0.1/predictions -d \
    '{"meta":{},"data":{"names":["sepal length (cm)","sepal width (cm)", "petal length (cm)","petal width (cm)"],"ndarray":[[5.1,3.5,1.4,0.2]]}}'

The response from the server should be a json object of this type:

```json
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
```

The response contains:

* a "meta" dictionary: This dictionary contains various metadata:
    * "puid": A unique identifier for the prediction
    * "tags": Optional tags. Empty in this case
    * "routing": This field is relevant when the deployment contain a more complex graph (see [A/B test example](link)). In this case is empty since we are deploying a single model
* "data" dictionary: This dictionary contains the predictions for your model classes
    * "names": The names of your classes.
    * "ndarray": The predicted  probabilities for each class