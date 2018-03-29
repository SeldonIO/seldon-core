
# Getting started on Minikube


In this guide, we will show how to create, deploy and serve an Iris classification model using Seldon Core running on a Minikube cluster. Seldon Core uses [helm](https://github.com/kubernetes/helm) charts to start and runs on [kubernetes](https://kubernetes.io/) clusters. Minikube is a tool that makes it easy to run Kubernetes locally,  and runs a single-node Kubernetes cluster inside a virtual machine on your laptop. 


### Prerequisites

The following packages need to be installed on your machine.

* [minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/) >= 0.24.0
* [helm](https://github.com/kubernetes/helm/blob/master/docs/install.md) >= 2.7.0

* [sklearn](http://scikit-learn.org/stable/) 
  - Sklearn is needed only to train the iris classifier example below. Seldon Core doesn't not require sklearn installed on your machine  to run.


### Before starting: run a Minikube cluster locally

Before starting, you need to have [minikube installed](https://kubernetes.io/docs/tasks/tools/install-minikube/) on your machine.

1. Start a Kubernetes local cluster in your machine using Minikube with RBAC enabled:

    ```bash
    minikube start --memory=8000 --feature-gates=CustomResourceValidation=true
    ```
    
Once the cluster is created, Minikube should automatically point your kubectl cli to the minikube cluster.

### Start seldon-core

You can now start Seldon Core in your minikube cluster.


1. Seldon Core uses helm charts to start. To use Seldon Core, you need [helm installed](https://github.com/kubernetes/helm/blob/master/docs/install.md) in your machine. To Initialize helm, type on command line:

    ```bash
    helm init
    ```

2. Seldon Core uses helm charts to start, which are stored in google storage. 
Use the charts to install the CRD and then the core components. Enabling reporting of anonymous usage metrics is optional, see [Usage Reporting](/docs/developer/readme.md#usage-reporting).


    ```bash
     helm install seldon-core-crd --name seldon-core-crd \
            --repo https://storage.googleapis.com/seldon-charts \
	    --set usage_metrics.enabled=true \
	    --set rbac.enabled=false
     helm install seldon-core --name seldon-core \
            --repo https://storage.googleapis.com/seldon-charts \
	    --set rbac.enabled=false
    ```

Seldon Core should now be running on your cluster. You can verify if all the pods are up and running typing on command line ```helm status seldon-core``` or ```kubectl get pods```

### Wrap your model

In this session, we show how to wrap the sklearn iris classifier in the [seldon-core](https://github.com/SeldonIO/seldon-core) repository using Seldon Core python wrappers. The example consists of a logistic regression model trained on the  [Iris dataset](http://scikit-learn.org/stable/auto_examples/datasets/plot_iris_dataset.html).

1. Clone the seldon-core repository:

    ```bash
    git clone https://github.com/SeldonIO/seldon-core
    ```

2. Train and save the sklearn iris classifier example model using the provided script ```train_iris.py```:

    ```bash
    cd seldon-core/examples/models/sklearn_iris/
    ```
    ```bash
    python train_iris.py
    ````
    
    This will train a simple logistic regression model on the iris dataset and save the model in the same folder.


3. Wrap your saved model using the core-python-wrapper docker image:
    ```bash
    docker run -v $(pwd):/model seldonio/core-python-wrapper:0.7 /model IrisClassifier 0.1 seldonio --force
    ```
    
4. Build the docker image locally
    ```bash
    cd build
    ./build_image.sh
    ```
    This will create the docker image ```seldonio/irisclassifier:0.1``` inside the minikube cluster which is ready for deployment with Seldon Core.


### Deploy your model

The docker image version of your model is deployed through a json configuration file. A general template for the configuration can be found  [here](https://github.com/SeldonIO/seldon-core/blob/master/examples/models/sklearn_iris/sklearn_iris_deployment.json). For the sklearn iris example, we have already created a deployment file ```sklearn_iris_deployment.json``` in the ```sklearn_iris``` folder:

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

1. To deploy the model in seldon core, type on command line (from the that contains the sklearn_iris_deployment.json file):

    ```bash
    kubectl apply -f sklearn_iris_deployment.json
    ```

2. The deployment will take a few seconds to be ready. To check if your deployment is ready:

    ```bash
    kubectl describe seldondeployments seldon-deployment-example
    ```

        
	
### Serve your  model:

In order to send a prediction request to your model, you need to query the Seldon Core api server and obtain an authentication token from your model key and secret (in this example key and secret are set to "oauth-key" and "oauth-secret" for simplicity). The api server is running on minikube. To query your model you need to

1. Set the api server IP and port:

    ```bash
    SERVER=$(minikube ip):$(kubectl get svc -l app=seldon-apiserver-container-app -o jsonpath='{.items[0].spec.ports[0].nodePort}')
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
    * "puid": A unique identifier for the prediction.
    * "tags": Optional tags. Empty in this case.
    * "routing": This field is relevant when the deployment contain a more complex graph. In this case is empty since we are deploying a single model.
* "data" dictionary: This dictionary contains the predictions for your model classes
    * "names": The names of your classes.
    * "ndarray": The predicted  probabilities for each class.

## Next Steps

 * You can run several notebooks that show various examples on minikube and Google cloud platform
   *  [Jupyter Notebook showing deployment of prebuilt model using Minikube](https://github.com/SeldonIO/seldon-core/blob/master/notebooks/kubectl_demo_minikube.ipynb)
   * [Jupyter Notebook showing deployment of prebuilt model using GCP cluster](https://github.com/SeldonIO/seldon-core/blob/master/notebooks/kubectl_demo_gcp.ipynb)
   * [Epsilon-greedy multi-armed bandits for real time optimization of models](https://github.com/SeldonIO/seldon-core/blob/master/notebooks/epsilon_greedy_gcp.ipynb)
   * [Advanced graphs showing the various types of runtime prediction graphs that can be built](https://github.com/cliveseldon/seldon-core/blob/master/notebooks/advanced_graphs.ipynb) 
   * [Jupyter notebook to create seldon-core with ksonnet and expose APIs using Ambassador.](https://github.com/SeldonIO/seldon-core/blob/master/notebooks/ksonnet_ambassador_minikube.ipynb)

