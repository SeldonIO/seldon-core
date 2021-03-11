
# Azure Kubernetes Service (AKS) Deep MNIST
In this example we will deploy a tensorflow MNIST model in the Azure Kubernetes Service (AKS).

This tutorial will break down in the following sections:

1) Train a tensorflow model to predict mnist locally

2) Containerise the tensorflow model with our docker utility

3) Send some data to the docker model to test it

4) Install and configure Azure tools to interact with your cluster

5) Use the Azure tools to create and setup AKS cluster with Seldon

6) Push and run docker image through the Azure Container Registry

7) Test our Elastic Kubernetes deployment by sending some data

#### Let's get started! 🚀🔥

## Dependencies:

* Helm v2.13.1+
* A Kubernetes cluster running v1.13 or above (minkube / docker-for-windows work well if enough RAM)
* kubectl v1.14+
* az CLI v2.0.66+
* Python 3.6+
* Python DEV requirements


## 1) Train a tensorflow model to predict mnist locally
We will load the mnist images, together with their labels, and then train a tensorflow model to predict the right labels


```python
from tensorflow.examples.tutorials.mnist import input_data
mnist = input_data.read_data_sets("MNIST_data/", one_hot = True)
import tensorflow as tf

if __name__ == '__main__':
    
    x = tf.placeholder(tf.float32, [None,784], name="x")

    W = tf.Variable(tf.zeros([784,10]))
    b = tf.Variable(tf.zeros([10]))

    y = tf.nn.softmax(tf.matmul(x,W) + b, name="y")

    y_ = tf.placeholder(tf.float32, [None, 10])

    cross_entropy = tf.reduce_mean(-tf.reduce_sum(y_ * tf.log(y), reduction_indices=[1]))

    train_step = tf.train.GradientDescentOptimizer(0.5).minimize(cross_entropy)

    init = tf.initialize_all_variables()

    sess = tf.Session()
    sess.run(init)

    for i in range(1000):
        batch_xs, batch_ys = mnist.train.next_batch(100)
        sess.run(train_step, feed_dict={x: batch_xs, y_: batch_ys})

    correct_prediction = tf.equal(tf.argmax(y,1), tf.argmax(y_,1))
    accuracy = tf.reduce_mean(tf.cast(correct_prediction, tf.float32))
    print(sess.run(accuracy, feed_dict = {x: mnist.test.images, y_:mnist.test.labels}))

    saver = tf.train.Saver()

    saver.save(sess, "model/deep_mnist_model")
```

    WARNING:tensorflow:From <ipython-input-1-559b63ab8b48>:2: read_data_sets (from tensorflow.contrib.learn.python.learn.datasets.mnist) is deprecated and will be removed in a future version.
    Instructions for updating:
    Please use alternatives such as official/mnist/dataset.py from tensorflow/models.
    WARNING:tensorflow:From /home/alejandro/miniconda3/envs/reddit-classification/lib/python3.7/site-packages/tensorflow/contrib/learn/python/learn/datasets/mnist.py:260: maybe_download (from tensorflow.contrib.learn.python.learn.datasets.base) is deprecated and will be removed in a future version.
    Instructions for updating:
    Please write your own downloading logic.
    WARNING:tensorflow:From /home/alejandro/miniconda3/envs/reddit-classification/lib/python3.7/site-packages/tensorflow/contrib/learn/python/learn/datasets/base.py:252: _internal_retry.<locals>.wrap.<locals>.wrapped_fn (from tensorflow.contrib.learn.python.learn.datasets.base) is deprecated and will be removed in a future version.
    Instructions for updating:
    Please use urllib or similar directly.
    Successfully downloaded train-images-idx3-ubyte.gz 9912422 bytes.
    WARNING:tensorflow:From /home/alejandro/miniconda3/envs/reddit-classification/lib/python3.7/site-packages/tensorflow/contrib/learn/python/learn/datasets/mnist.py:262: extract_images (from tensorflow.contrib.learn.python.learn.datasets.mnist) is deprecated and will be removed in a future version.
    Instructions for updating:
    Please use tf.data to implement this functionality.
    Extracting MNIST_data/train-images-idx3-ubyte.gz
    Successfully downloaded train-labels-idx1-ubyte.gz 28881 bytes.
    WARNING:tensorflow:From /home/alejandro/miniconda3/envs/reddit-classification/lib/python3.7/site-packages/tensorflow/contrib/learn/python/learn/datasets/mnist.py:267: extract_labels (from tensorflow.contrib.learn.python.learn.datasets.mnist) is deprecated and will be removed in a future version.
    Instructions for updating:
    Please use tf.data to implement this functionality.
    Extracting MNIST_data/train-labels-idx1-ubyte.gz
    WARNING:tensorflow:From /home/alejandro/miniconda3/envs/reddit-classification/lib/python3.7/site-packages/tensorflow/contrib/learn/python/learn/datasets/mnist.py:110: dense_to_one_hot (from tensorflow.contrib.learn.python.learn.datasets.mnist) is deprecated and will be removed in a future version.
    Instructions for updating:
    Please use tf.one_hot on tensors.
    Successfully downloaded t10k-images-idx3-ubyte.gz 1648877 bytes.
    Extracting MNIST_data/t10k-images-idx3-ubyte.gz
    Successfully downloaded t10k-labels-idx1-ubyte.gz 4542 bytes.
    Extracting MNIST_data/t10k-labels-idx1-ubyte.gz
    WARNING:tensorflow:From /home/alejandro/miniconda3/envs/reddit-classification/lib/python3.7/site-packages/tensorflow/contrib/learn/python/learn/datasets/mnist.py:290: DataSet.__init__ (from tensorflow.contrib.learn.python.learn.datasets.mnist) is deprecated and will be removed in a future version.
    Instructions for updating:
    Please use alternatives such as official/mnist/dataset.py from tensorflow/models.
    WARNING:tensorflow:From /home/alejandro/miniconda3/envs/reddit-classification/lib/python3.7/site-packages/tensorflow/python/framework/op_def_library.py:263: colocate_with (from tensorflow.python.framework.ops) is deprecated and will be removed in a future version.
    Instructions for updating:
    Colocations handled automatically by placer.
    WARNING:tensorflow:From /home/alejandro/miniconda3/envs/reddit-classification/lib/python3.7/site-packages/tensorflow/python/ops/math_ops.py:3066: to_int32 (from tensorflow.python.ops.math_ops) is deprecated and will be removed in a future version.
    Instructions for updating:
    Use tf.cast instead.
    WARNING:tensorflow:From /home/alejandro/miniconda3/envs/reddit-classification/lib/python3.7/site-packages/tensorflow/python/util/tf_should_use.py:193: initialize_all_variables (from tensorflow.python.ops.variables) is deprecated and will be removed after 2017-03-02.
    Instructions for updating:
    Use `tf.global_variables_initializer` instead.
    0.915


## 2) Containerise the tensorflow model with our docker utility

First you need to make sure that you have added the .s2i/environment configuration file in this folder with the following content:


```python
!cat .s2i/environment
```

    MODEL_NAME=DeepMnist
    API_TYPE=REST
    SERVICE_TYPE=MODEL
    PERSISTENCE=0


Now we can build a docker image named "deep-mnist" with the tag 0.1


```python
!s2i build . seldonio/seldon-core-s2i-python36:0.7 deep-mnist:0.1
```

    ---> Installing application source...
    ---> Installing dependencies ...
    Looking in links: /whl
    Requirement already satisfied: tensorflow>=1.12.0 in /usr/local/lib/python3.6/site-packages (from -r requirements.txt (line 1)) (1.13.1)
    Requirement already satisfied: keras-preprocessing>=1.0.5 in /usr/local/lib/python3.6/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (1.0.9)
    Requirement already satisfied: gast>=0.2.0 in /usr/local/lib/python3.6/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (0.2.2)
    Requirement already satisfied: absl-py>=0.1.6 in /usr/local/lib/python3.6/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (0.7.1)
    Requirement already satisfied: astor>=0.6.0 in /usr/local/lib/python3.6/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (0.7.1)
    Requirement already satisfied: keras-applications>=1.0.6 in /usr/local/lib/python3.6/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (1.0.7)
    Requirement already satisfied: six>=1.10.0 in /usr/local/lib/python3.6/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (1.12.0)
    Requirement already satisfied: termcolor>=1.1.0 in /usr/local/lib/python3.6/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (1.1.0)
    Requirement already satisfied: grpcio>=1.8.6 in /usr/local/lib/python3.6/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (1.19.0)
    Requirement already satisfied: wheel>=0.26 in /usr/local/lib/python3.6/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (0.33.1)
    Requirement already satisfied: tensorboard<1.14.0,>=1.13.0 in /usr/local/lib/python3.6/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (1.13.1)
    Requirement already satisfied: numpy>=1.13.3 in /usr/local/lib/python3.6/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (1.16.2)
    Requirement already satisfied: protobuf>=3.6.1 in /usr/local/lib/python3.6/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (3.7.0)
    Requirement already satisfied: tensorflow-estimator<1.14.0rc0,>=1.13.0 in /usr/local/lib/python3.6/site-packages (from tensorflow>=1.12.0->-r requirements.txt (line 1)) (1.13.0)
    Requirement already satisfied: h5py in /usr/local/lib/python3.6/site-packages (from keras-applications>=1.0.6->tensorflow>=1.12.0->-r requirements.txt (line 1)) (2.9.0)
    Requirement already satisfied: markdown>=2.6.8 in /usr/local/lib/python3.6/site-packages (from tensorboard<1.14.0,>=1.13.0->tensorflow>=1.12.0->-r requirements.txt (line 1)) (3.0.1)
    Requirement already satisfied: werkzeug>=0.11.15 in /usr/local/lib/python3.6/site-packages (from tensorboard<1.14.0,>=1.13.0->tensorflow>=1.12.0->-r requirements.txt (line 1)) (0.15.0)
    Requirement already satisfied: setuptools in /usr/local/lib/python3.6/site-packages (from protobuf>=3.6.1->tensorflow>=1.12.0->-r requirements.txt (line 1)) (40.8.0)
    Requirement already satisfied: mock>=2.0.0 in /usr/local/lib/python3.6/site-packages (from tensorflow-estimator<1.14.0rc0,>=1.13.0->tensorflow>=1.12.0->-r requirements.txt (line 1)) (2.0.0)
    Requirement already satisfied: pbr>=0.11 in /usr/local/lib/python3.6/site-packages (from mock>=2.0.0->tensorflow-estimator<1.14.0rc0,>=1.13.0->tensorflow>=1.12.0->-r requirements.txt (line 1)) (5.1.3)
    Url '/whl' is ignored. It is either a non-existing path or lacks a specific scheme.
    You are using pip version 19.0.3, however version 19.1.1 is available.
    You should consider upgrading via the 'pip install --upgrade pip' command.
    Build completed successfully


## 3) Send some data to the docker model to test it
We first run the docker image we just created as a container called "mnist_predictor"


```python
!docker run --name "mnist_predictor" -d --rm -p 5000:5000 deep-mnist:0.1
```

    9087047e368ac8f285e1f742704b4c0c7bceac7d29ee90b3b0a6ef2d61ebd15c


Send some random features that conform to the contract


```python
import matplotlib.pyplot as plt
import numpy as np
# This is the variable that was initialised at the beginning of the file
i = [0]
x = mnist.test.images[i]
y = mnist.test.labels[i]
plt.imshow(x.reshape((28, 28)), cmap='gray')
plt.show()
print("Expected label: ", np.sum(range(0,10) * y), ". One hot encoding: ", y)
```


![png](azure_aks_deep_mnist_files/azure_aks_deep_mnist_11_0.png)


    Expected label:  7.0 . One hot encoding:  [[0. 0. 0. 0. 0. 0. 0. 1. 0. 0.]]



```python
from seldon_core.seldon_client import SeldonClient
import math
import numpy as np

# We now test the REST endpoint expecting the same result
endpoint = "0.0.0.0:5000"
batch = x
payload_type = "ndarray"

sc = SeldonClient(microservice_endpoint=endpoint)

# We use the microservice, instead of the "predict" function
client_prediction = sc.microservice(
    data=batch,
    method="predict",
    payload_type=payload_type,
    names=["tfidf"])

for proba, label in zip(client_prediction.response.data.ndarray.values[0].list_value.ListFields()[0][1], range(0,10)):
    print(f"LABEL {label}:\t {proba.number_value*100:6.4f} %")
```

    LABEL 0:	 0.0064 %
    LABEL 1:	 0.0000 %
    LABEL 2:	 0.0155 %
    LABEL 3:	 0.2862 %
    LABEL 4:	 0.0003 %
    LABEL 5:	 0.0027 %
    LABEL 6:	 0.0000 %
    LABEL 7:	 99.6643 %
    LABEL 8:	 0.0020 %
    LABEL 9:	 0.0227 %



```python
!docker rm mnist_predictor --force
```

    mnist_predictor


## 4) Install and configure Azure tools 

First we install the azure cli - follow specific instructions at https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest


```python
!curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash
```

#### Configure the azure CLI so it can talk to your server 
(if you are getting issues, make sure you have the permmissions to create clusters)

You must run this through a terminal and follow the instructions:
```
az login
```

Once you are logged in, we can create our cluster. Run the following command, it may take a while so feel free to get a ☕.


```bash
%%bash 
# We'll create a resource group
az group create --name SeldonResourceGroup --location westus
# Now we create the cluster
az aks create \
    --resource-group SeldonResourceGroup \
    --name SeldonCluster \
    --node-count 1 \
    --enable-addons monitoring \
    --generate-ssh-keys
    --kubernetes-version 1.13.5
```

Once it's created we can authenticate our local `kubectl` to make sure we can talk to the azure cluster:


```python
!az aks get-credentials --resource-group SeldonResourceGroup --name SeldonCluster
```

And now we can check that this has been successful by making sure that our `kubectl` context is pointing to the cluster:


```python
!kubectl config get-contexts
```

## Install Seldon Core

### Before we install seldon core, we need to install HELM
For that, we need to create a ClusterRoleBinding for us, a ServiceAccount, and then a RoleBinding


```python
!kubectl create clusterrolebinding kube-system-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:default
```

    clusterrolebinding.rbac.authorization.k8s.io/kube-system-cluster-admin created



```python
!kubectl create serviceaccount tiller --namespace kube-system
```

    serviceaccount/tiller created



```python
!kubectl apply -f tiller-role-binding.yaml
```

    clusterrolebinding.rbac.authorization.k8s.io/tiller-role-binding created


### Once that is set-up we can install Tiller


```python
!helm repo update
```


```python
!helm init --service-account tiller
```

    $HELM_HOME has been configured at /home/alejandro/.helm.
    
    Tiller (the Helm server-side component) has been installed into your Kubernetes Cluster.
    
    Please note: by default, Tiller is deployed with an insecure 'allow unauthenticated users' policy.
    To prevent this, run `helm init` with the --tiller-tls-verify flag.
    For more information on securing your installation see: https://docs.helm.sh/using_helm/#securing-your-helm-installation
    Happy Helming!



```python
# Wait until Tiller finishes
!kubectl rollout status deploy/tiller-deploy -n kube-system
```

    deployment "tiller-deploy" successfully rolled out


### Now we can install SELDON. 
We first start with the custom resource definitions (CRDs)


```python
!helm install seldon-core-operator seldon-core-operator --repo https://storage.googleapis.com/seldon-charts
```

    NAME:   seldon-core-operator
    LAST DEPLOYED: Thu Jun  6 12:03:45 2019
    NAMESPACE: default
    STATUS: DEPLOYED
    
    RESOURCES:
    ==> v1/ClusterRole
    NAME                          AGE
    seldon-operator-manager-role  3s
    
    ==> v1/ClusterRoleBinding
    NAME                                 AGE
    seldon-operator-manager-rolebinding  3s
    
    ==> v1/Pod(related)
    NAME                                  READY  STATUS   RESTARTS  AGE
    seldon-operator-controller-manager-0  1/1    Running  0         3s
    
    ==> v1/Secret
    NAME                                   TYPE    DATA  AGE
    seldon-operator-webhook-server-secret  Opaque  4     3s
    
    ==> v1/Service
    NAME                                        TYPE       CLUSTER-IP    EXTERNAL-IP  PORT(S)  AGE
    seldon-operator-controller-manager-service  ClusterIP  10.0.224.128  <none>       443/TCP  3s
    
    ==> v1/StatefulSet
    NAME                                READY  AGE
    seldon-operator-controller-manager  1/1    3s
    
    ==> v1beta1/CustomResourceDefinition
    NAME                                         AGE
    seldondeployments.machinelearning.seldon.io  3s
    
    
    NOTES:
    NOTES: TODO
    
    


And confirm they are running by getting the pods:


```python
!kubectl rollout status deployment/seldon-operator-controller-manager -n seldon-system
```

    partitioned roll out complete: 1 new pods have been updated...


### Now we set-up the ingress
This will allow you to reach the Seldon models from outside the kubernetes cluster. 

In EKS it automatically creates an Elastic Load Balancer, which you can configure from the EC2 Console


```python
!helm install ambassador stable/ambassador --set crds.keep=false
```

    Error: release ambassador failed: serviceaccounts "ambassador" already exists


And let's wait until it's fully deployed


```python
!kubectl rollout status deployment.apps/ambassador
```

    deployment "ambassador" successfully rolled out


## Push docker image
In order for the EKS seldon deployment to access the image we just built, we need to push it to the Azure Container Registry (ACR) - you can check if it's been successfully created in the dashboard https://portal.azure.com/#blade/HubsExtension/BrowseResourceBlade/resourceType/Microsoft.ContainerRegistry%2Fregistries

If you have any issues please follow the official Azure documentation: https://docs.microsoft.com/en-us/azure/container-registry/container-registry-get-started-azure-cli

### First we create a registry
Make sure you keep the `loginServer` value in the output dictionary as we'll use it below.


```python
!az acr create --resource-group SeldonResourceGroup --name SeldonContainerRegistry --sku Basic
```

    [K{- Finished ..
      "adminUserEnabled": false,
      "creationDate": "2019-06-06T10:51:55.288108+00:00",
      "id": "/subscriptions/df7969cc-8033-4c83-b027-14c0424f039d/resourceGroups/KlawClusterResourceGroup/providers/Microsoft.ContainerRegistry/registries/SeldonContainerRegistry",
      "location": "westus",
      "loginServer": "seldoncontainerregistry.azurecr.io",
      "name": "SeldonContainerRegistry",
      "networkRuleSet": null,
      "provisioningState": "Succeeded",
      "resourceGroup": "KlawClusterResourceGroup",
      "sku": {
        "name": "Basic",
        "tier": "Basic"
      },
      "status": null,
      "storageAccount": null,
      "tags": {},
      "type": "Microsoft.ContainerRegistry/registries"
    }
    [0m

### Make sure your local docker instance has access to the registry


```python
!az acr login --name SeldonContainerRegistry
```

    Login Succeeded
    WARNING! Your password will be stored unencrypted in /home/alejandro/.docker/config.json.
    Configure a credential helper to remove this warning. See
    https://docs.docker.com/engine/reference/commandline/login/#credentials-store
    


### Now prepare docker image
We need to first tag the docker image before we can push it.

NOTE: if you named your registry different make sure you change the value of `seldoncontainerregistry.azurecr.io`


```python
!docker tag deep-mnist:0.1 seldoncontainerregistry.azurecr.io/deep-mnist:0.1
```

### And push the image

NOTE: if you named your registry different make sure you change the value of `seldoncontainerregistry.azurecr.io`


```python
!docker push seldoncontainerregistry.azurecr.io/deep-mnist:0.1
```

    The push refers to repository [seldoncontainerregistry.azurecr.io/deep-mnist]
    
    [1Bb4fe3076: Preparing 
    [1Be4a983d1: Preparing 
    [1B74b2c556: Preparing 
    [1B9472b523: Preparing 
    [1Ba2a7ea60: Preparing 
    [1Beddb328a: Preparing 
    [1B1393f8e7: Preparing 
    [1B67d6e30e: Preparing 
    [1Bf19da2c9: Preparing 
    [1B9ec591c4: Preparing 
    [1B32b1ff99: Preparing 
    [1B64f96dbc: Preparing 
    [1Be6d76fd9: Preparing 
    [1B11a84ad4: Preparing 
    [13B4b2c556: Pushing  187.1MB/648.8MB[14A[1K[K[12A[1K[K[15A[1K[K[15A[1K[K[11A[1K[K[15A[1K[K[11A[1K[K[15A[1K[K[15A[1K[K[15A[1K[K[13A[1K[K[15A[1K[K[11A[1K[K[15A[1K[K[11A[1K[K[1K[K[15A[1K[K[11A[1K[K[15A[1K[K[15A[1K[K[11A[1K[K[13A[1K[K[15A[1K[K[11A[1K[K[14A[1K[K[12A[1K[K[11A[1K[K[15A[1K[K[13A[1K[K[15A[1K[K[15A[1K[K[11A[1K[K[11A[1K[K[15A[1K[K[11A[1K[K[15A[1K[K[15A[1K[K[15A[1K[K[15A[1K[K[13A[1K[K[15A[1K[K[15A[1K[K[15A[1K[K[11A[1K[K[15A[1K[K[13A[1K[K[15A[1K[K[15A[1K[K[11A[1K[K[15A[1K[K[11A[1K[K[15A[1K[K[15A[1K[K[11A[1K[K[15A[1K[K[11A[1K[K[15A[1K[K[15A[1K[K[11A[1K[K[15A[1K[K[11A[1K[K[11A[1K[K[11A[1K[K[13A[1K[K[11A[1K[K[11A[1K[K[11A[1K[K[13A[1K[K[11A[1K[K[11A[1K[K[11A[1K[K[11A[1K[K[9A[1K[K[11A[1K[K[11A[1K[K[11A[1K[K[11A[1K[K[13A[1K[K[11A[1K[K[9A[1K[K[10A[1K[K[11A[1K[K[10A[1K[K[9A[1K[K[11A[1K[K[9A[1K[K[13A[1K[K[10A[1K[K[13A[1K[K[9A[1K[K[11A[1K[K[10A[1K[K[9A[1K[K[15A[1K[K[13A[1K[K[11A[1K[K[9A[1K[K[11A[1K[K[10A[1K[K[9A[1K[K[11A[1K[K[9A[1K[K[10A[1K[K[9A[1K[K[9A[1K[K[10A[1K[K[11A[1K[K[9A[1K[K[9A[1K[K[10A[1K[K[11A[1K[K[13A[1K[K[10A[1K[K[10A[1K[K[13A[1K[K[11A[1K[K[10A[1K[K[10A[1K[K[11A[1K[K[10A[1K[K[8A[1K[K[11A[1K[K[11A[1K[K[11A[1K[K[10A[1K[K[11A[1K[K[11A[1K[K[10A[1K[K[13A[1K[K[11A[1K[K[10A[1K[K[11A[1K[K[13A[1K[K[10A[1K[K[11A[1K[K[10A[1K[K[13A[1K[K[10A[1K[K[13A[1K[K[10A[1K[K[11A[1K[K[10A[1K[K[11A[1K[K[10A[1K[K[13A[1K[K[10A[1K[K[11A[1K[K[10A[1K[K[11A[1K[K[13A[1K[K[11A[1K[K[10A[1K[K[11A[1K[K[13A[1K[K[10A[1K[K[11A[1K[K[13A[1K[K[10A[1K[K[10A[1K[K[10A[1K[K[11A[1K[K[10A[1K[K[11A[1K[K[11A[1K[K[10A[1K[K[11A[1K[K[13A[1K[K[11A[1K[K[11A[1K[K[11A[1K[K[11A[1K[K[13A[1K[K[7A[1K[K[11A[1K[K[7A[1K[K[11A[1K[K[11A[1K[K[13A[1K[K[7A[1K[K[6A[1K[K[6A[1K[K[13A[1K[K[7A[1K[K[6A[1K[K[7A[1K[K[6A[1K[K[13A[1K[K[10A[1K[K[7A[1K[K[7A[1K[K[6A[1K[K[7A[1K[K[6A[1K[K[6A[1K[K[11A[1K[K[13A[1K[K[7A[1K[K[6A[1K[K[7A[1K[K[6A[1K[K[7A[1K[K[6A[1K[K[13A[1K[KPushing   28.9MB/648.8MB[6A[1K[K[13A[1K[K[6A[1K[K[7A[1K[K[6A[1K[K[7A[1K[K[7A[1K[K[6A[1K[K[13A[1K[K[6A[1K[K[7A[1K[K[6A[1K[K[5A[1K[K[13A[1K[K[6A[1K[K[7A[1K[K[5A[1K[K[13A[1K[K[6A[1K[K[4A[1K[K[5A[1K[K[13A[1K[K[7A[1K[K[5A[1K[K[4A[1K[K[13A[1K[K[7A[1K[K[5A[1K[K[7A[1K[K[4A[1K[K[7A[1K[K[5A[1K[K[13A[1K[K[4A[1K[K[5A[1K[K[4A[1K[K[5A[1K[K[7A[1K[K[4A[1K[K[13A[1K[K[5A[1K[K[7A[1K[K[4A[1K[K[5A[1K[K[4A[1K[K[7A[1K[K[4A[1K[K[7A[1K[K[5A[1K[K[5A[1K[K[6A[1K[K[13A[1K[K[5A[1K[K[7A[1K[K[5A[1K[K[7A[1K[K[7A[1K[K[4A[1K[K[7A[1K[K[4A[1K[K[5A[1K[K[4A[1K[K[5A[1K[K[7A[1K[K[5A[1K[K[7A[1K[K[13A[1K[K[7A[1K[K[5A[1K[K[13A[1K[K[4A[1K[K[4A[1K[K[5A[1K[K[7A[1K[K[7A[1K[K[7A[1K[K[5A[1K[K[4A[1K[K[5A[1K[K[4A[1K[K[7A[1K[K[4A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[4A[1K[K[5A[1K[K[4A[1K[K[7A[1K[K[5A[1K[K[4A[1K[K[7A[1K[K[4A[1K[K[7A[1K[K[5A[1K[K[4A[1K[K[5A[1K[K[3A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[4A[1K[K[7A[1K[K[4A[1K[K[4A[1K[K[3A[1K[K[4A[1K[K[3A[1K[K[13A[1K[K[4A[1K[K[7A[1K[K[13A[1K[K[5A[1K[K[7A[1K[K[3A[1K[K[5A[1K[K[5A[1K[K[13A[1K[K[7A[1K[K[7A[1K[K[5A[1K[K[4A[1K[K[5A[1K[K[3A[1K[K[4A[1K[K[13A[1K[K[3A[1K[K[4A[1K[K[13A[1K[K[13A[1K[K[4A[1K[K[13A[1K[K[4A[1K[K[5A[1K[K[4A[1K[K[13A[1K[K[3A[1K[K[13A[1K[K[4A[1K[K[5A[1K[K[13A[1K[K[4A[1K[K[5A[1K[K[5A[1K[K[4A[1K[K[13A[1K[K[5A[1K[K[4A[1K[K[4A[1K[K[3A[1K[K[5A[1K[K[4A[1K[K[13A[1K[K[5A[1K[K[5A[1K[K[4A[1K[K[13A[1K[K[4A[1K[K[3A[1K[K[4A[1K[K[5A[1K[K[4A[1K[K[5A[1K[K[5A[1K[K[3A[1K[K[13A[1K[K[5A[1K[K[4A[1K[K[5A[1K[K[4A[1K[K[4A[1K[K[13A[1K[K[5A[1K[K[4A[1K[K[4A[1K[K[5A[1K[K[4A[1K[K[5A[1K[K[4A[1K[K[5A[1K[K[13A[1K[K[4A[1K[K[5A[1K[K[4A[1K[K[13A[1K[K[4A[1K[K[5A[1K[K[4A[1K[K[5A[1K[K[4A[1K[K[5A[1K[K[13A[1K[K[4A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[4A[1K[K[13A[1K[K[4A[1K[K[4A[1K[KPushing  60.12MB/141.8MB[4A[1K[KPushing  60.64MB/141.8MB[7A[1K[K[4A[1K[K[13A[1K[K[4A[1K[K[5A[1K[K[4A[1K[K[5A[1K[K[4A[1K[K[5A[1K[K[3A[1K[K[13A[1K[K[5A[1K[K[4A[1K[K[5A[1K[K[5A[1K[K[4A[1K[K[13A[1K[K[4A[1K[K[3A[1K[K[5A[1K[K[5A[1K[K[4A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[4A[1K[K[13A[1K[K[4A[1K[K[5A[1K[K[4A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[4A[1K[K[13A[1K[K[4A[1K[K[4A[1K[K[4A[1K[K[5A[1K[K[13A[1K[K[4A[1K[K[5A[1K[K[5A[1K[K[4A[1K[K[13A[1K[K[4A[1K[K[13A[1K[K[4A[1K[K[13A[1K[K[5A[1K[K[4A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[4A[1K[K[13A[1K[K[4A[1K[K[13A[1K[K[4A[1K[K[5A[1K[K[5A[1K[K[4A[1K[K[4A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[5A[1K[K[4A[1K[K[4A[1K[K[4A[1K[K[5A[1K[K[5A[1K[K[4A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[4A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[4A[1K[K[5A[1K[K[4A[1K[K[4A[1K[K[13A[1K[K[4A[1K[K[5A[1K[K[4A[1K[K[13A[1K[K[4A[1K[K[5A[1K[K[4A[1K[K[5A[1K[K[5A[1K[K[4A[1K[K[5A[1K[K[4A[1K[K[4A[1K[K[5A[1K[K[4A[1K[K[13A[1K[K[5A[1K[K[5A[1K[K[2A[1K[K[5A[1K[K[K[4A[1K[K[5A[1K[K[4A[1K[K[13A[1K[K[4A[1K[K[4A[1K[K[13A[1K[K[4A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[4A[1K[K[13A[1K[K[4A[1K[K[5A[1K[K[13A[1K[K[4A[1K[K[5A[1K[K[4A[1K[K[5A[1K[K[4A[1K[K[13A[1K[K[4A[1K[K[4A[1K[K[5A[1K[K[4A[1K[K[13A[1K[K[4A[1K[K[13A[1K[K[5A[1K[K[4A[1K[K[4A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[4A[1K[K[13A[1K[K[4A[1K[K[5A[1K[K[13A[1K[K[4A[1K[KPushing  141.9MB[2A[1K[K[4A[1K[K[2A[1K[K[4A[1K[K[4A[1K[K[5A[1K[K[4A[1K[K[13A[1K[K[4A[1K[K[4A[1K[K[5A[1K[K[4A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[2A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[5A[1K[K[2A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[2A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[4A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[2A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[2A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[2A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[5A[1K[K[2A[1K[K[13A[1K[K[5A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[2A[1K[K[13A[1K[K[5A[1K[K[13A[1K[KPushing  185.5MB/648.8MB[5A[1K[K[5B32b1ff99: Pushed   570.6MB/556.5MB[13A[1K[K[5A[1K[K[13A[1K[K[2A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[2A[1K[K[5A[1K[K[5A[1K[KPushing  254.7MB/556.5MB[13A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[2A[1K[K[5A[1K[K[2A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[2A[1K[K[13A[1K[K[2A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[1A[1K[K[13A[1K[K[2A[1K[K[5A[1K[K[13A[1K[K[2A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[1A[1K[K[13A[1K[K[5A[1K[K[5A[1K[K[2A[1K[K[2A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[2A[1K[K[13A[1K[K[2A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[2A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[1A[1K[K[2A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[2A[1K[K[13A[1K[K[5A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[1A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[2A[1K[K[13A[1K[K[5A[1K[K[K[5A[1K[K[13A[1K[K[13A[1K[K[2A[1K[K[13A[1K[K[5A[1K[K[5A[1K[K[5A[1K[KPushing  268.7MB/648.8MB[2A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[1A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[2A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[2A[1K[K[5A[1K[K[2A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[1A[1K[K[13A[1K[K[5A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[KPushing  6.456MB/100.6MB[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[5A[1K[K[1A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[13A[1K[K[1A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[5A[1K[K[13A[1K[K[1A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[1A[1K[K[5A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[1A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[KPushing  345.2MB/648.8MB[13A[1K[K[1A[1K[K[5A[1K[K[5A[1K[K[1A[1K[K[5A[1K[K[1A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[1A[1K[K[13A[1K[K[1A[1K[K[5A[1K[K[1A[1K[K[5A[1K[K[5A[1K[K[1A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[1A[1K[K[5A[1K[K[1A[1K[K[13A[1K[K[1A[1K[K[13A[1K[K[1A[1K[K[5A[1K[K[5A[1K[K[1A[1K[K[5A[1K[K[13A[1K[K[1A[1K[K[13A[1K[K[1A[1K[K[1A[1K[K[13A[1K[K[1A[1K[K[13A[1K[K[1A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[1A[1K[K[5A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[1A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[1A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[1A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[1A[1K[K[5A[1K[K[13A[1K[K[1A[1K[K[13A[1K[K[13A[1K[K[13A[1K[KPushing  398.1MB/648.8MB[5A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[1A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[1A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[1A[1K[K[13A[1K[K[1A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[1A[1K[K[5A[1K[K[13A[1K[K[1A[1K[K[5A[1K[K[13A[1K[K[1A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[1A[1K[K[1A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[1A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[1A[1K[K[13A[1K[K[1A[1K[KPushing  55.64MB/100.6MB[13A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[1A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[1A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[1A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[1A[1K[K[5A[1K[K[1A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[13A[1K[KPushing  481.1MB/648.8MB[13A[1K[K[13A[1K[K[13A[1K[K[1A[1K[K[13A[1K[K[13A[1K[K[1A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[1A[1K[K[13A[1K[K[1A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[1A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[1A[1K[K[13A[1K[K[1A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[5A[1K[KPushing  473.8MB/556.5MB[13A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[1A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[1A[1K[K[13A[1K[K[1A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[1A[1K[K[1A[1K[K[13A[1K[K[5A[1K[K[1A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[1A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[1A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[1A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[1A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[1A[1K[K[5A[1K[K[1A[1K[K[5A[1K[K[13A[1K[K[1A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[1A[1K[K[1A[1K[K[13A[1K[K[1A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[1A[1K[K[13A[1K[K[1A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[1A[1K[K[1A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[1A[1K[K[13A[1K[K[1A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[1A[1K[K[1A[1K[K[13A[1K[K[13A[1K[K[1A[1K[K[13A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[1A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[1A[1K[K[5A[1K[K[13A[1K[K[13A[1K[K[13A[1K[K[1A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[1A[1K[K[5A[1K[K[5A[1K[K[1A[1K[K[5A[1K[K[5A[1K[K[1A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[1A[1K[K[1A[1K[K[5A[1K[K[5A[1K[K[1A[1K[K[1A[1K[K[5A[1K[K[1A[1K[K[K[1A[1K[K[5A[1K[K[5A[1K[K[13A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[1A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K[5A[1K[K0.1: digest: sha256:b6110da62719e103bfd8c4b187f868b4341c35be16d288018d529da1cfa2585c size: 3482


## Running the Model
We will now run the model. As you can see we have a placeholder `"REPLACE_FOR_IMAGE_AND_TAG"`, which we'll replace to point to our registry.

Let's first have a look at the file we'll be using to trigger the model:


```python
!cat deep_mnist.json
```

    {
        "apiVersion": "machinelearning.seldon.io/v1alpha2",
        "kind": "SeldonDeployment",
        "metadata": {
            "labels": {
                "app": "seldon"
            },
            "name": "deep-mnist"
        },
        "spec": {
            "annotations": {
                "project_name": "Tensorflow MNIST",
                "deployment_version": "v1"
            },
            "name": "deep-mnist",
            "oauth_key": "oauth-key",
            "oauth_secret": "oauth-secret",
            "predictors": [
                {
                    "componentSpecs": [{
                        "spec": {
                            "containers": [
                                {
                                    "image": "REPLACE_FOR_IMAGE_AND_TAG",
                                    "imagePullPolicy": "IfNotPresent",
                                    "name": "classifier",
                                    "resources": {
                                        "requests": {
                                            "memory": "1Mi"
                                        }
                                    }
                                }
                            ],
                            "terminationGracePeriodSeconds": 20
                        }
                    }],
                    "graph": {
                        "children": [],
                        "name": "classifier",
                        "endpoint": {
    			"type" : "REST"
    		    },
                        "type": "MODEL"
                    },
                    "name": "single-model",
                    "replicas": 1,
    		"annotations": {
    		    "predictor_version" : "v1"
    		}
                }
            ]
        }
    }


Now let's trigger seldon to run the model.

### Run the deployment in your cluster

NOTE: In order for this to work you need to make sure that your cluster has the permissions to pull the images. You can do this by:

1) Go into the Azure Container Registry

2) Select the SeldonContainerRegistry you created

3) Click on "Add a role assignment"

4) Select the AcrPull role

5) Select service principle

6) Find the SeldonCluster

7) Wait until the role has been added

We basically have a yaml file, where we want to replace the value "REPLACE_FOR_IMAGE_AND_TAG" for the image you pushed


```bash
%%bash
# Change accordingly if your registry is called differently
sed 's|REPLACE_FOR_IMAGE_AND_TAG|seldoncontainerregistry.azurecr.io/deep-mnist:0.1|g' deep_mnist.json | kubectl apply -f -
```

    seldondeployment.machinelearning.seldon.io/deep-mnist created


And let's check that it's been created.

You should see an image called "deep-mnist-single-model...".

We'll wait until STATUS changes from "ContainerCreating" to "Running"


```python
!kubectl get pods
```

## Test the model
Now we can test the model, let's first find out what is the URL that we'll have to use:


```python
!kubectl get svc ambassador -o jsonpath='{.status.loadBalancer.ingress[0].ip}'  
```

    52.160.64.65

We'll use a random example from our dataset


```python
import matplotlib.pyplot as plt
# This is the variable that was initialised at the beginning of the file
i = [0]
x = mnist.test.images[i]
y = mnist.test.labels[i]
plt.imshow(x.reshape((28, 28)), cmap='gray')
plt.show()
print("Expected label: ", np.sum(range(0,10) * y), ". One hot encoding: ", y)
```


![png](azure_aks_deep_mnist_files/azure_aks_deep_mnist_58_0.png)


    Expected label:  7.0 . One hot encoding:  [[0. 0. 0. 0. 0. 0. 0. 1. 0. 0.]]


We can now add the URL above to send our request:


```python
from seldon_core.seldon_client import SeldonClient
import math
import numpy as np

host = "52.160.64.65"
port = "80" # Make sure you use the port above
batch = x
payload_type = "ndarray"

sc = SeldonClient(
    gateway="ambassador", 
    ambassador_endpoint=host + ":" + port,
    namespace="default",
    oauth_key="oauth-key", 
    oauth_secret="oauth-secret")

client_prediction = sc.predict(
    data=batch, 
    deployment_name="deep-mnist",
    names=["text"],
    payload_type=payload_type)

print(client_prediction)
```

    Success:True message:
    Request:
    data {
      names: "text"
      ndarray {
        values {
          list_value {
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.3294117748737335
            }
            values {
              number_value: 0.7254902124404907
            }
            values {
              number_value: 0.6235294342041016
            }
            values {
              number_value: 0.5921568870544434
            }
            values {
              number_value: 0.2352941334247589
            }
            values {
              number_value: 0.1411764770746231
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.8705883026123047
            }
            values {
              number_value: 0.9960784912109375
            }
            values {
              number_value: 0.9960784912109375
            }
            values {
              number_value: 0.9960784912109375
            }
            values {
              number_value: 0.9960784912109375
            }
            values {
              number_value: 0.9450981020927429
            }
            values {
              number_value: 0.7764706611633301
            }
            values {
              number_value: 0.7764706611633301
            }
            values {
              number_value: 0.7764706611633301
            }
            values {
              number_value: 0.7764706611633301
            }
            values {
              number_value: 0.7764706611633301
            }
            values {
              number_value: 0.7764706611633301
            }
            values {
              number_value: 0.7764706611633301
            }
            values {
              number_value: 0.7764706611633301
            }
            values {
              number_value: 0.6666666865348816
            }
            values {
              number_value: 0.2039215862751007
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.26274511218070984
            }
            values {
              number_value: 0.44705885648727417
            }
            values {
              number_value: 0.2823529541492462
            }
            values {
              number_value: 0.44705885648727417
            }
            values {
              number_value: 0.6392157077789307
            }
            values {
              number_value: 0.8901961445808411
            }
            values {
              number_value: 0.9960784912109375
            }
            values {
              number_value: 0.8823530077934265
            }
            values {
              number_value: 0.9960784912109375
            }
            values {
              number_value: 0.9960784912109375
            }
            values {
              number_value: 0.9960784912109375
            }
            values {
              number_value: 0.9803922176361084
            }
            values {
              number_value: 0.8980392813682556
            }
            values {
              number_value: 0.9960784912109375
            }
            values {
              number_value: 0.9960784912109375
            }
            values {
              number_value: 0.5490196347236633
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.06666667014360428
            }
            values {
              number_value: 0.25882354378700256
            }
            values {
              number_value: 0.05490196496248245
            }
            values {
              number_value: 0.26274511218070984
            }
            values {
              number_value: 0.26274511218070984
            }
            values {
              number_value: 0.26274511218070984
            }
            values {
              number_value: 0.23137256503105164
            }
            values {
              number_value: 0.08235294371843338
            }
            values {
              number_value: 0.9254902601242065
            }
            values {
              number_value: 0.9960784912109375
            }
            values {
              number_value: 0.41568630933761597
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.32549020648002625
            }
            values {
              number_value: 0.9921569228172302
            }
            values {
              number_value: 0.8196079134941101
            }
            values {
              number_value: 0.07058823853731155
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.08627451211214066
            }
            values {
              number_value: 0.9137255549430847
            }
            values {
              number_value: 1.0
            }
            values {
              number_value: 0.32549020648002625
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.5058823823928833
            }
            values {
              number_value: 0.9960784912109375
            }
            values {
              number_value: 0.9333333969116211
            }
            values {
              number_value: 0.1725490242242813
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.23137256503105164
            }
            values {
              number_value: 0.9764706492424011
            }
            values {
              number_value: 0.9960784912109375
            }
            values {
              number_value: 0.24313727021217346
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.5215686559677124
            }
            values {
              number_value: 0.9960784912109375
            }
            values {
              number_value: 0.7333333492279053
            }
            values {
              number_value: 0.019607843831181526
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.03529411926865578
            }
            values {
              number_value: 0.803921639919281
            }
            values {
              number_value: 0.9725490808486938
            }
            values {
              number_value: 0.22745099663734436
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.4941176772117615
            }
            values {
              number_value: 0.9960784912109375
            }
            values {
              number_value: 0.7137255072593689
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.29411765933036804
            }
            values {
              number_value: 0.9843137860298157
            }
            values {
              number_value: 0.9411765336990356
            }
            values {
              number_value: 0.22352942824363708
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.07450980693101883
            }
            values {
              number_value: 0.8666667342185974
            }
            values {
              number_value: 0.9960784912109375
            }
            values {
              number_value: 0.6509804129600525
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.011764707043766975
            }
            values {
              number_value: 0.7960785031318665
            }
            values {
              number_value: 0.9960784912109375
            }
            values {
              number_value: 0.8588235974311829
            }
            values {
              number_value: 0.13725490868091583
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.14901961386203766
            }
            values {
              number_value: 0.9960784912109375
            }
            values {
              number_value: 0.9960784912109375
            }
            values {
              number_value: 0.3019607961177826
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.12156863510608673
            }
            values {
              number_value: 0.8784314393997192
            }
            values {
              number_value: 0.9960784912109375
            }
            values {
              number_value: 0.45098042488098145
            }
            values {
              number_value: 0.003921568859368563
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.5215686559677124
            }
            values {
              number_value: 0.9960784912109375
            }
            values {
              number_value: 0.9960784912109375
            }
            values {
              number_value: 0.2039215862751007
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.2392157018184662
            }
            values {
              number_value: 0.9490196704864502
            }
            values {
              number_value: 0.9960784912109375
            }
            values {
              number_value: 0.9960784912109375
            }
            values {
              number_value: 0.2039215862751007
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.4745098352432251
            }
            values {
              number_value: 0.9960784912109375
            }
            values {
              number_value: 0.9960784912109375
            }
            values {
              number_value: 0.8588235974311829
            }
            values {
              number_value: 0.1568627506494522
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.4745098352432251
            }
            values {
              number_value: 0.9960784912109375
            }
            values {
              number_value: 0.8117647767066956
            }
            values {
              number_value: 0.07058823853731155
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
            values {
              number_value: 0.0
            }
          }
        }
      }
    }
    
    Response:
    meta {
      puid: "hndttmvl5qua3beavsre2n8hrj"
      requestPath {
        key: "classifier"
        value: "seldoncontainerregistry.azurecr.io/deep-mnist:0.1"
      }
    }
    data {
      names: "class:0"
      names: "class:1"
      names: "class:2"
      names: "class:3"
      names: "class:4"
      names: "class:5"
      names: "class:6"
      names: "class:7"
      names: "class:8"
      names: "class:9"
      ndarray {
        values {
          list_value {
            values {
              number_value: 6.386729364749044e-05
            }
            values {
              number_value: 7.228476039955467e-09
            }
            values {
              number_value: 0.00015463074669241905
            }
            values {
              number_value: 0.00286240060813725
            }
            values {
              number_value: 2.6505524601816433e-06
            }
            values {
              number_value: 2.7261585273663513e-05
            }
            values {
              number_value: 2.7168187699544433e-08
            }
            values {
              number_value: 0.9966427087783813
            }
            values {
              number_value: 1.9810671801678836e-05
            }
            values {
              number_value: 0.00022661285765934736
            }
          }
        }
      }
    }
    


### Let's visualise the probability for each label
It seems that it correctly predicted the number 7


```python
for proba, label in zip(client_prediction.response.data.ndarray.values[0].list_value.ListFields()[0][1], range(0,10)):
    print(f"LABEL {label}:\t {proba.number_value*100:6.4f} %")
```

    LABEL 0:	 0.0064 %
    LABEL 1:	 0.0000 %
    LABEL 2:	 0.0155 %
    LABEL 3:	 0.2862 %
    LABEL 4:	 0.0003 %
    LABEL 5:	 0.0027 %
    LABEL 6:	 0.0000 %
    LABEL 7:	 99.6643 %
    LABEL 8:	 0.0020 %
    LABEL 9:	 0.0227 %



```python

```
