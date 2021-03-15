
# AWS Elastic Kubernetes Service (EKS) Deep MNIST
In this example we will deploy a tensorflow MNIST model in Amazon Web Services' Elastic Kubernetes Service (EKS).

This tutorial will break down in the following sections:

1) Train a tensorflow model to predict mnist locally

2) Containerise the tensorflow model with our docker utility

3) Send some data to the docker model to test it

4) Install and configure AWS tools to interact with AWS

5) Use the AWS tools to create and setup EKS cluster with Seldon

6) Push and run docker image through the AWS Container Registry

7) Test our Elastic Kubernetes deployment by sending some data

#### Let's get started! 🚀🔥

## Dependencies:

* Helm v2.13.1+
* A Kubernetes cluster running v1.13 or above (minkube / docker-for-windows work well if enough RAM)
* kubectl v1.14+
* EKS CLI v0.1.32
* AWS Cli v1.16.163
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

    Extracting MNIST_data/train-images-idx3-ubyte.gz
    Extracting MNIST_data/train-labels-idx1-ubyte.gz
    Extracting MNIST_data/t10k-images-idx3-ubyte.gz
    Extracting MNIST_data/t10k-labels-idx1-ubyte.gz
    0.9194


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
!s2i build . seldonio/seldon-core-s2i-python36:0.5.1 deep-mnist:0.1
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

    5157ab4f516bd0dea11b159780f31121e9fb41df6394e0d6d631e6e0d572463b


Send some random features that conform to the contract


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


![png](aws_eks_deep_mnist_files/aws_eks_deep_mnist_11_0.png)


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

    LABEL 0:	 0.0068 %
    LABEL 1:	 0.0000 %
    LABEL 2:	 0.0085 %
    LABEL 3:	 0.3409 %
    LABEL 4:	 0.0002 %
    LABEL 5:	 0.0020 %
    LABEL 6:	 0.0000 %
    LABEL 7:	 99.5371 %
    LABEL 8:	 0.0026 %
    LABEL 9:	 0.1019 %



```python
!docker rm mnist_predictor --force
```

    mnist_predictor


## 4) Install and configure AWS tools to interact with AWS

First we install the awscli


```python
!pip install awscli --upgrade --user
```

    Collecting awscli
      Using cached https://files.pythonhosted.org/packages/f6/45/259a98719e7c7defc9be4cc00fbfb7ccf699fbd1f74455d8347d0ab0a1df/awscli-1.16.163-py2.py3-none-any.whl
    Collecting colorama<=0.3.9,>=0.2.5 (from awscli)
      Using cached https://files.pythonhosted.org/packages/db/c8/7dcf9dbcb22429512708fe3a547f8b6101c0d02137acbd892505aee57adf/colorama-0.3.9-py2.py3-none-any.whl
    Collecting PyYAML<=3.13,>=3.10 (from awscli)
    Collecting botocore==1.12.153 (from awscli)
      Using cached https://files.pythonhosted.org/packages/ec/3b/029218966ce62ae9824a18730de862ac8fc5a0e8083d07d1379815e7cca1/botocore-1.12.153-py2.py3-none-any.whl
    Requirement already satisfied, skipping upgrade: docutils>=0.10 in /home/alejandro/miniconda3/envs/reddit-classification/lib/python3.7/site-packages (from awscli) (0.14)
    Collecting rsa<=3.5.0,>=3.1.2 (from awscli)
      Using cached https://files.pythonhosted.org/packages/e1/ae/baedc9cb175552e95f3395c43055a6a5e125ae4d48a1d7a924baca83e92e/rsa-3.4.2-py2.py3-none-any.whl
    Requirement already satisfied, skipping upgrade: s3transfer<0.3.0,>=0.2.0 in /home/alejandro/miniconda3/envs/reddit-classification/lib/python3.7/site-packages (from awscli) (0.2.0)
    Requirement already satisfied, skipping upgrade: urllib3<1.25,>=1.20; python_version >= "3.4" in /home/alejandro/miniconda3/envs/reddit-classification/lib/python3.7/site-packages (from botocore==1.12.153->awscli) (1.24.2)
    Requirement already satisfied, skipping upgrade: python-dateutil<3.0.0,>=2.1; python_version >= "2.7" in /home/alejandro/miniconda3/envs/reddit-classification/lib/python3.7/site-packages (from botocore==1.12.153->awscli) (2.8.0)
    Requirement already satisfied, skipping upgrade: jmespath<1.0.0,>=0.7.1 in /home/alejandro/miniconda3/envs/reddit-classification/lib/python3.7/site-packages (from botocore==1.12.153->awscli) (0.9.4)
    Collecting pyasn1>=0.1.3 (from rsa<=3.5.0,>=3.1.2->awscli)
      Using cached https://files.pythonhosted.org/packages/7b/7c/c9386b82a25115cccf1903441bba3cbadcfae7b678a20167347fa8ded34c/pyasn1-0.4.5-py2.py3-none-any.whl
    Requirement already satisfied, skipping upgrade: six>=1.5 in /home/alejandro/miniconda3/envs/reddit-classification/lib/python3.7/site-packages (from python-dateutil<3.0.0,>=2.1; python_version >= "2.7"->botocore==1.12.153->awscli) (1.12.0)
    Installing collected packages: colorama, PyYAML, botocore, pyasn1, rsa, awscli
    Successfully installed PyYAML-3.13 awscli-1.16.163 botocore-1.12.153 colorama-0.3.9 pyasn1-0.4.5 rsa-3.4.2


#### Configure aws so it can talk to your server 
(if you are getting issues, make sure you have the permmissions to create clusters)


```bash
%%bash 
# You must make sure that the access key and secret are changed
aws configure << END_OF_INPUTS
YOUR_ACCESS_KEY
YOUR_ACCESS_SECRET
us-west-2
json
END_OF_INPUTS
```

    AWS Access Key ID [****************SF4A]: AWS Secret Access Key [****************WLHu]: Default region name [eu-west-1]: Default output format [json]: 

#### Install EKCTL
*IMPORTANT*: These instructions are for linux
Please follow the official installation of ekctl at: https://docs.aws.amazon.com/eks/latest/userguide/getting-started-eksctl.html


```python
!curl --silent --location "https://github.com/weaveworks/eksctl/releases/download/latest_release/eksctl_$(uname -s)_amd64.tar.gz" | tar xz 
```


```python
!chmod 755 ./eksctl
```


```python
!./eksctl version
```

    [36m[ℹ]  version.Info{BuiltAt:"", GitCommit:"", GitTag:"0.1.32"}
    [0m

## 5) Use the AWS tools to create and setup EKS cluster with Seldon
In this example we will create a cluster with 2 nodes, with a minimum of 1 and a max of 3. You can tweak this accordingly.

If you want to check the status of the deployment you can go to AWS CloudFormation or to the EKS dashboard.

It will take 10-15 minutes (so feel free to go grab a ☕). 

### IMPORTANT: If you get errors in this step...
It is most probably IAM role access requirements, which requires you to discuss with your administrator.


```bash
%%bash
./eksctl create cluster \
--name demo-eks-cluster \
--region us-west-2 \
--nodes 2 
```

    Process is interrupted.


### Configure local kubectl 
We want to now configure our local Kubectl so we can actually reach the cluster we've just created


```python
!aws eks --region us-west-2 update-kubeconfig --name demo-eks-cluster
```

    Updated context arn:aws:eks:eu-west-1:271049282727:cluster/deepmnist in /home/alejandro/.kube/config


And we can check if the context has been added to kubectl config (contexts are basically the different k8s cluster connections)
You should be able to see the context as "...aws:eks:eu-west-1:27...". 
If it's not activated you can activate that context with kubectlt config set-context <CONTEXT_NAME>


```python
!kubectl config get-contexts
```

    CURRENT   NAME                                                   CLUSTER                                                AUTHINFO                                               NAMESPACE
    *         arn:aws:eks:eu-west-1:271049282727:cluster/deepmnist   arn:aws:eks:eu-west-1:271049282727:cluster/deepmnist   arn:aws:eks:eu-west-1:271049282727:cluster/deepmnist   
              docker-desktop                                         docker-desktop                                         docker-desktop                                         
              docker-for-desktop                                     docker-desktop                                         docker-desktop                                         
              gke_ml-engineer_us-central1-a_security-cluster-1       gke_ml-engineer_us-central1-a_security-cluster-1       gke_ml-engineer_us-central1-a_security-cluster-1       


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
!helm install seldon-core-operator seldon-core-operator --repo https://storage.googleapis.com/seldon-charts --set usageMetrics.enabled=true --namespace seldon-system
```

    NAME:   seldon-core-operator
    LAST DEPLOYED: Wed May 22 16:24:10 2019
    NAMESPACE: seldon-system
    STATUS: DEPLOYED
    
    RESOURCES:
    ==> v1/ClusterRole
    NAME                          AGE
    seldon-operator-manager-role  2s
    
    ==> v1/ClusterRoleBinding
    NAME                                 AGE
    seldon-operator-manager-rolebinding  2s
    
    ==> v1/ConfigMap
    NAME                     DATA  AGE
    seldon-spartakus-config  3     2s
    
    ==> v1/Pod(related)
    NAME                                         READY  STATUS             RESTARTS  AGE
    seldon-operator-controller-manager-0         0/1    ContainerCreating  0         2s
    seldon-spartakus-volunteer-6954cffb89-qz4pq  0/1    ContainerCreating  0         1s
    
    ==> v1/Secret
    NAME                                   TYPE    DATA  AGE
    seldon-operator-webhook-server-secret  Opaque  0     2s
    
    ==> v1/Service
    NAME                                        TYPE       CLUSTER-IP      EXTERNAL-IP  PORT(S)  AGE
    seldon-operator-controller-manager-service  ClusterIP  10.100.198.157  <none>       443/TCP  2s
    
    ==> v1/ServiceAccount
    NAME                        SECRETS  AGE
    seldon-spartakus-volunteer  1        2s
    
    ==> v1/StatefulSet
    NAME                                READY  AGE
    seldon-operator-controller-manager  0/1    2s
    
    ==> v1beta1/ClusterRole
    NAME                        AGE
    seldon-spartakus-volunteer  2s
    
    ==> v1beta1/ClusterRoleBinding
    NAME                        AGE
    seldon-spartakus-volunteer  2s
    
    ==> v1beta1/CustomResourceDefinition
    NAME                                         AGE
    seldondeployments.machinelearning.seldon.io  2s
    
    ==> v1beta1/Deployment
    NAME                        READY  UP-TO-DATE  AVAILABLE  AGE
    seldon-spartakus-volunteer  0/1    1           0          2s
    
    
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

    NAME:   ambassador
    LAST DEPLOYED: Wed May 22 16:25:38 2019
    NAMESPACE: default
    STATUS: DEPLOYED
    
    RESOURCES:
    ==> v1/Deployment
    NAME        READY  UP-TO-DATE  AVAILABLE  AGE
    ambassador  0/3    3           0          0s
    
    ==> v1/Pod(related)
    NAME                         READY  STATUS             RESTARTS  AGE
    ambassador-6dbf99c886-frlfm  0/1    ContainerCreating  0         0s
    ambassador-6dbf99c886-kj56r  0/1    ContainerCreating  0         0s
    ambassador-6dbf99c886-v5mtv  0/1    ContainerCreating  0         0s
    
    ==> v1/Service
    NAME               TYPE          CLUSTER-IP      EXTERNAL-IP  PORT(S)                     AGE
    ambassador         LoadBalancer  10.100.59.146   <pending>    80:30911/TCP,443:31715/TCP  0s
    ambassador-admins  ClusterIP     10.100.152.178  <none>       8877/TCP                    0s
    
    ==> v1/ServiceAccount
    NAME        SECRETS  AGE
    ambassador  1        0s
    
    ==> v1beta1/ClusterRole
    NAME        AGE
    ambassador  0s
    
    ==> v1beta1/ClusterRoleBinding
    NAME        AGE
    ambassador  0s
    
    
    NOTES:
    Congratuations! You've successfully installed Ambassador.
    
    For help, visit our Slack at https://d6e.co/slack or view the documentation online at https://www.getambassador.io.
    
    To get the IP address of Ambassador, run the following commands:
    NOTE: It may take a few minutes for the LoadBalancer IP to be available.
         You can watch the status of by running 'kubectl get svc -w  --namespace default ambassador'
    
      On GKE/Azure:
      export SERVICE_IP=$(kubectl get svc --namespace default ambassador -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    
      On AWS:
      export SERVICE_IP=$(kubectl get svc --namespace default ambassador -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
    
      echo http://$SERVICE_IP:
    


And let's wait until it's fully deployed


```python
!kubectl rollout status deployment.apps/ambassador
```

## Push docker image
In order for the EKS seldon deployment to access the image we just built, we need to push it to the Elastic Container Registry (ECR).

If you have any issues please follow the official AWS documentation: https://docs.aws.amazon.com/AmazonECR/latest/userguide/docker-basics.html

### First we create a registry
You can run the following command, and then see the result at https://us-west-2.console.aws.amazon.com/ecr/repositories?#


```python
!aws ecr create-repository --repository-name seldon-repository --region us-west-2
```

    {
        "repository": {
            "repositoryArn": "arn:aws:ecr:us-west-2:271049282727:repository/seldon-repository",
            "registryId": "271049282727",
            "repositoryName": "seldon-repository",
            "repositoryUri": "271049282727.dkr.ecr.us-west-2.amazonaws.com/seldon-repository",
            "createdAt": 1558535798.0
        }
    }


### Now prepare docker image
We need to first tag the docker image before we can push it


```bash
%%bash
export AWS_ACCOUNT_ID=""
export AWS_REGION="us-west-2"
if [ -z "$AWS_ACCOUNT_ID" ]; then
    echo "ERROR: Please provide a value for the AWS variables"
    exit 1
fi

docker tag deep-mnist:0.1 "$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/seldon-repository"
```

### We now login to aws through docker so we can access the repository


```python
!`aws ecr get-login --no-include-email --region us-west-2`
```

    WARNING! Using --password via the CLI is insecure. Use --password-stdin.
    WARNING! Your password will be stored unencrypted in /home/alejandro/.docker/config.json.
    Configure a credential helper to remove this warning. See
    https://docs.docker.com/engine/reference/commandline/login/#credentials-store
    
    Login Succeeded


### And push the image
Make sure you add your AWS Account ID


```bash
%%bash
export AWS_ACCOUNT_ID=""
export AWS_REGION="us-west-2"
if [ -z "$AWS_ACCOUNT_ID" ]; then
    echo "ERROR: Please provide a value for the AWS variables"
    exit 1
fi

docker push "$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/seldon-repository"
```

    The push refers to repository [271049282727.dkr.ecr.us-west-2.amazonaws.com/seldon-repository]
    f7d0d000c138: Preparing
    987f3f1afb00: Preparing
    00d16a381c47: Preparing
    bb01f50d544a: Preparing
    fcb82c6941b5: Preparing
    67290e35c458: Preparing
    b813745f5bb3: Preparing
    ffecb18e9f0b: Preparing
    f50f856f49fa: Preparing
    80b43ad4adf9: Preparing
    14c77983a1cf: Preparing
    a22a5ac18042: Preparing
    6257fa9f9597: Preparing
    578414b395b9: Preparing
    abc3250a6c7f: Preparing
    13d5529fd232: Preparing
    67290e35c458: Waiting
    b813745f5bb3: Waiting
    ffecb18e9f0b: Waiting
    f50f856f49fa: Waiting
    80b43ad4adf9: Waiting
    6257fa9f9597: Waiting
    14c77983a1cf: Waiting
    a22a5ac18042: Waiting
    578414b395b9: Waiting
    abc3250a6c7f: Waiting
    13d5529fd232: Waiting
    987f3f1afb00: Pushed
    fcb82c6941b5: Pushed
    bb01f50d544a: Pushed
    f7d0d000c138: Pushed
    ffecb18e9f0b: Pushed
    b813745f5bb3: Pushed
    f50f856f49fa: Pushed
    67290e35c458: Pushed
    14c77983a1cf: Pushed
    578414b395b9: Pushed
    80b43ad4adf9: Pushed
    13d5529fd232: Pushed
    6257fa9f9597: Pushed
    abc3250a6c7f: Pushed
    00d16a381c47: Pushed
    a22a5ac18042: Pushed
    latest: digest: sha256:19aefaa9d87c1287eb46ec08f5d4f9a689744d9d0d0b75668b7d15e447819d74 size: 3691


## Running the Model
We will now run the model.

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
                                    "image": "271049282727.dkr.ecr.us-west-2.amazonaws.com/seldon-repository:latest",
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

We basically have a yaml file, where we want to replace the value "REPLACE_FOR_IMAGE_AND_TAG" for the image you pushed


```bash
%%bash
export AWS_ACCOUNT_ID=""
export AWS_REGION="us-west-2"
if [ -z "$AWS_ACCOUNT_ID" ]; then
    echo "ERROR: Please provide a value for the AWS variables"
    exit 1
fi

sed 's|REPLACE_FOR_IMAGE_AND_TAG|'"$AWS_ACCOUNT_ID"'.dkr.ecr.'"$AWS_REGION"'.amazonaws.com/seldon-repository|g' deep_mnist.json | kubectl apply -f -
```

    error: unable to recognize "STDIN": Get https://461835FD3FF52848655C8F09FBF5EEAA.yl4.us-west-2.eks.amazonaws.com/api?timeout=32s: dial tcp: lookup 461835FD3FF52848655C8F09FBF5EEAA.yl4.us-west-2.eks.amazonaws.com on 1.1.1.1:53: no such host



    ---------------------------------------------------------------------------

    CalledProcessError                        Traceback (most recent call last)

    <ipython-input-165-1129742af2c4> in <module>
    ----> 1 get_ipython().run_cell_magic('bash', '', 'export AWS_ACCOUNT_ID="2710"\nexport AWS_REGION="us-west-2"\nif [ -z "$AWS_ACCOUNT_ID" ]; then\n    echo "ERROR: Please provide a value for the AWS variables"\n    exit 1\nfi\n\nsed \'s|REPLACE_FOR_IMAGE_AND_TAG|\'"$AWS_ACCOUNT_ID"\'.dkr.ecr.\'"$AWS_REGION"\'.amazonaws.com/seldon-repository|g\' deep_mnist.json | kubectl apply -f -\n')
    

    ~/miniconda3/envs/reddit-classification/lib/python3.7/site-packages/IPython/core/interactiveshell.py in run_cell_magic(self, magic_name, line, cell)
       2350             with self.builtin_trap:
       2351                 args = (magic_arg_s, cell)
    -> 2352                 result = fn(*args, **kwargs)
       2353             return result
       2354 


    ~/miniconda3/envs/reddit-classification/lib/python3.7/site-packages/IPython/core/magics/script.py in named_script_magic(line, cell)
        140             else:
        141                 line = script
    --> 142             return self.shebang(line, cell)
        143 
        144         # write a basic docstring:


    </home/alejandro/miniconda3/envs/reddit-classification/lib/python3.7/site-packages/decorator.py:decorator-gen-110> in shebang(self, line, cell)


    ~/miniconda3/envs/reddit-classification/lib/python3.7/site-packages/IPython/core/magic.py in <lambda>(f, *a, **k)
        185     # but it's overkill for just that one bit of state.
        186     def magic_deco(arg):
    --> 187         call = lambda f, *a, **k: f(*a, **k)
        188 
        189         if callable(arg):


    ~/miniconda3/envs/reddit-classification/lib/python3.7/site-packages/IPython/core/magics/script.py in shebang(self, line, cell)
        243             sys.stderr.flush()
        244         if args.raise_error and p.returncode!=0:
    --> 245             raise CalledProcessError(p.returncode, cell, output=out, stderr=err)
        246 
        247     def _run_script(self, p, cell, to_close):


    CalledProcessError: Command 'b'export AWS_ACCOUNT_ID="2710"\nexport AWS_REGION="us-west-2"\nif [ -z "$AWS_ACCOUNT_ID" ]; then\n    echo "ERROR: Please provide a value for the AWS variables"\n    exit 1\nfi\n\nsed \'s|REPLACE_FOR_IMAGE_AND_TAG|\'"$AWS_ACCOUNT_ID"\'.dkr.ecr.\'"$AWS_REGION"\'.amazonaws.com/seldon-repository|g\' deep_mnist.json | kubectl apply -f -\n'' returned non-zero exit status 1.


And let's check that it's been created.

You should see an image called "deep-mnist-single-model...".

We'll wait until STATUS changes from "ContainerCreating" to "Running"


```python
!kubectl get pods
```

    NAME                                              READY   STATUS    RESTARTS   AGE
    ambassador-5475779f98-7bhcw                       1/1     Running   0          21m
    ambassador-5475779f98-986g5                       1/1     Running   0          21m
    ambassador-5475779f98-zcd28                       1/1     Running   0          21m
    deep-mnist-single-model-42ed9d9-fdb557d6b-6xv2h   2/2     Running   0          18m


## Test the model
Now we can test the model, let's first find out what is the URL that we'll have to use:


```python
!kubectl get svc ambassador -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 
```

    a68bbac487ca611e988060247f81f4c1-707754258.us-west-2.elb.amazonaws.com

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


![png](aws_eks_deep_mnist_files/aws_eks_deep_mnist_63_0.png)


    Expected label:  7.0 . One hot encoding:  [[0. 0. 0. 0. 0. 0. 0. 1. 0. 0.]]


We can now add the URL above to send our request:


```python
from seldon_core.seldon_client import SeldonClient
import math
import numpy as np

host = "a68bbac487ca611e988060247f81f4c1-707754258.us-west-2.elb.amazonaws.com"
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
      puid: "l6bv1r38mmb32l0hbinln2jjcl"
      requestPath {
        key: "classifier"
        value: "271049282727.dkr.ecr.us-west-2.amazonaws.com/seldon-repository:latest"
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
              number_value: 6.839015986770391e-05
            }
            values {
              number_value: 9.376968534979824e-09
            }
            values {
              number_value: 8.48581112222746e-05
            }
            values {
              number_value: 0.0034086888190358877
            }
            values {
              number_value: 2.3978568606253248e-06
            }
            values {
              number_value: 2.0100669644307345e-05
            }
            values {
              number_value: 3.0251623428512175e-08
            }
            values {
              number_value: 0.9953710436820984
            }
            values {
              number_value: 2.6070511012221687e-05
            }
            values {
              number_value: 0.0010185304563492537
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

    LABEL 0:	 0.0068 %
    LABEL 1:	 0.0000 %
    LABEL 2:	 0.0085 %
    LABEL 3:	 0.3409 %
    LABEL 4:	 0.0002 %
    LABEL 5:	 0.0020 %
    LABEL 6:	 0.0000 %
    LABEL 7:	 99.5371 %
    LABEL 8:	 0.0026 %
    LABEL 9:	 0.1019 %



```python

```
