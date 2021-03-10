
# Tensorflow GPU MNIST Model with GKE

**Please note: This tutorial uses Tensorflow-gpu=1.13.1, CUDA 10.0 and cuDNN 7.6**

**Requirements: Ubuntu 18.+ and Python 3.6**

In this tutorial we will run a deep MNIST Tensorflow example with GPU.

The tutorial will be broken down into the following sections:

1. Install all dependencies to run Tensorflow-GPU
    
    1.1 Installing CUDA 10.0
    
    1.2 Installing cuDNN 7.6
    
    1.3 Configure CUDA and cuDNN
    
    1.4 Install Tensorflow GPU
    
    
2. Train the MNIST model locally


3. Push the Image to your proejcts Container Registry


4. Deploy the model on GKE using Seldon Core


## Local Testing Environment

For the development of this example a GCE Virtual Machine was used to allow access to a GPU. The configuration for this VM is as follows:

* VM Image: TensorFlow from NVIDIA
* 8 vCPUs
* 32 GB memory
* 1x NVIDIA Tesla V100 GPU


## 1) Installing all dependencies to run Tensorflow-GPU

* Dependencies installed in this section:
    * Nvidia compute 3.0 onwards
    * CUDA 10.0
    * cuDNN 7.6
    * tensorflow-gpu 1.13.1

**Check Nvidia drivers >= 3.0**


```python
!nvidia-smi
```

## 1.1) Install CUDA 10.0

* **Download the CUDA 10.0 runfile**


```python
!wget https://developer.nvidia.com/compute/cuda/10.0/Prod/local_installers/cuda_10.0.130_410.48_linux
```

* **Unpack the separate files:**


```python
! chmod +x cuda_10.0.130_410.48_linux
! ./cuda_10.0.130_410.48_linux --extract=$HOME
```

* **Install the Cuda 10.0 Toolkit file**:

From the terminal, run the following command
```
$ sudo ./cuda-linux.10.0.130-24817639.run
```
Hold 'd' to scroll to the bottom of the license agreement.

Accept the licencing agreement and all of the default settings.

* **Verify the install, by installing the sample test:**
```
$ sudo ./cuda-samples.10.0.130-24817639-linux.run
```
Again, accept the agreement and all default settings

* **Configure the runtime library:**

```
$ sudo bash -c "echo /usr/local/cuda/lib64/ > /etc/ld.so.conf.d/cuda.conf"
```

```
$ sudo ldconfig
```

* **Add the cuda bin to the file system:**

```
$ sudo vim /etc/environment
```

Add ‘:/usr/local/cuda/bin’ to the end of the PATH (inside quotes)

* **Reboot the system**


```python
!sudo shutdown -r now
```

* **Run the tests that we set up** - this takes some time to complete, so let it run for a little while...

```
$ cd /usr/local/cuda-10.0/samples

$ sudo make
```

If run into an error involving the GCC version:

```
$ sudo update-alternatives --install /usr/bin/g++ g++ /usr/bin/g++-6 10
```

```
$ sudo update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-6 10
```

And run again, otherwise, skip this step.

* After complete, **run a devicequery and bandwidth test**:


```bash
%%bash

cd /usr/local/cuda/samples/bin/x86_64/linux/release
./deviceQuery
```

**Remember to clean up by removing all of the downloaded runtime packages**

## 1.2) Install cuDNN 7.6

* **Download all 3 .deb files for CUDA10.0 and Ubuntu 18.04**

You will have to create a Nvidia account for this and go to the archive section of the cuDNN downloads

Ensure you download all 3 files:
- Runtime
- Developer
- Code Samples


**Unpackage the three files in this order**


```bash
%%bash 
sudo dpkg -i ~/libcudnn7_7.6.0.64-1+cuda10.0_amd64.deb
sudo dpkg -i ~/libcudnn7-dev_7.6.0.64-1+cuda10.0_amd64.deb
sudo dpkg -i ~/libcudnn7-doc_7.6.0.64-1+cuda10.0_amd64.deb
```

* **Verify the install is successful with the MNIST example**

From the download folder. Copy the files to somewhere with write access: 


```python
! cp -r /usr/src/cudnn_samples_v7/ ~
```

**Go to the MNIST example code, compile and run it**


```bash
%%bash
cd ~/cudnn_samples_v7/mnistCUDNN
sudo make
sudo ./mnistCUDNN
```

**Remember to clean up by removing all of the downloaded runtime packages**

## 1.3) Configure CUDA and cuDNN

**Add LD_LIBRARY_PATH in your .bashrc file:**

Add the following line in the end or your .bashrc file export export:

```
LD_LIBRARY_PATH=/usr/local/cuda/lib64:${LD_LIBRARY_PATH:+:${LD_LIBRARY_PATH}}
```

And source it with:

```
$ source ~/.bashrc
```

## 1.4) Install tensorflow with GPU

**Require v=1.13.1 as with CUDA 10.0**


```python
! pip3 install --upgrade tensorflow-gpu==1.13.1
```


```python
import tensorflow as tf
sess = tf.Session(config=tf.ConfigProto(log_device_placement=True))
```

## 2) Train the MNIST model locally

* Wrap a Tensorflow MNIST python model for use as a prediction microservice in seldon-core
 
   * Run locally on Docker to test
   * Deploy on seldon-core running on minikube
 
## Dependencies

 * [Helm](https://github.com/kubernetes/helm)
 * [Minikube](https://github.com/kubernetes/minikube)
 * [S2I](https://github.com/openshift/source-to-image)

```bash
pip3 install seldon-core
```

## Train locally
 


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

Wrap model using s2i


```python
!s2i build . seldonio/seldon-core-s2i-python3-tf-gpu:0.1 deep-mnist-gpu:0.1
```


```python
!docker run --name "mnist_predictor" -d --rm -p 5000:5000 deep-mnist-gpu:0.1
```

Send some random features that conform to the contract


```python
!seldon-core-tester contract.json 0.0.0.0 5000 -p
```


```python
!docker rm mnist_predictor --force
```

## 3) Push the image to Google Container Registry

**Configure access to container registry** (follow the configuration to link to your own project).

```
$ gcloud auth configure-docker
```

**Tag Image with your project's registry path** (Edit the command below)


```python
!docker tag deep-mnist-gpu:0.1 gcr.io/<YOUR_PROJECT_ID>/deep-mnist-gpu:0.1
```

**Push the Image to the Container Registry** (Again edit command below)


```python
!docker push gcr.io/<YOUR_PROJECT_ID>/deep-mnist-gpu:0.1
```

## 4) Deploy in GKE

## Spin up a GKE Cluster

For this example only one node is needed within the cluster. The cluster should have the following **config**:

* 8 CPUs
* 30 GB Total Memory
* 1 Node with 1X NVIDIA Tesla V100 GPU
* Ubuntu Node image

Leave the rest of the config as default. 

**Connect to your cluster and check the context.**


```python
!gcloud config set project <YOUR_PROJECT_ID>
!gcloud container clusters get-credentials <YOUR_CLUSTER_NAME>
!kubectl config current-context
```

**Installing NVIDIA GPU device drivers**

(The below command is for the Ubuntu Node Image - if using a COS image, please see the Google Cloud Documentation for the correct command).


```python
!kubectl apply -f https://raw.githubusercontent.com/GoogleCloudPlatform/container-engine-accelerators/master/nvidia-driver-installer/ubuntu/daemonset-preloaded.yaml
```

## Install Seldon Core

**Before installing Seldon Core, we need to install HELM**

To do so, we need to creat a ClusterRoleBinding for us, a ServiceAccount and then a RoleBinding


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


**Once that is set-up we can install Tiller**


```python
!helm repo update
```

    Hang tight while we grab the latest from your chart repositories...
    ...Skip local chart repository
    ...Successfully got an update from the "stable" chart repository
    Update Complete.



```python
!helm init --service-account tiller
```

    $HELM_HOME has been configured at /Users/Seldon/.helm.
    
    Tiller (the Helm server-side component) has been installed into your Kubernetes Cluster.
    
    Please note: by default, Tiller is deployed with an insecure 'allow unauthenticated users' policy.
    To prevent this, run `helm init` with the --tiller-tls-verify flag.
    For more information on securing your installation see: https://docs.helm.sh/using_helm/#securing-your-helm-installation



```python
# Wait until Tiller finishes
!kubectl rollout status deploy/tiller-deploy -n kube-system
```

    Waiting for deployment "tiller-deploy" rollout to finish: 0 of 1 updated replicas are available...
    deployment "tiller-deploy" successfully rolled out


**Now we can install SELDON.**

We first start with the custom resource definitions (CRDs)


```python
!helm install seldon-core-operator seldon-core-operator --repo https://storage.googleapis.com/seldon-charts
```

    NAME:   seldon-core-operator
    E0624 14:57:46.960571   83748 portforward.go:372] error copying from remote stream to local connection: readfrom tcp4 127.0.0.1:64632->127.0.0.1:64637: write tcp4 127.0.0.1:64632->127.0.0.1:64637: write: broken pipe
    LAST DEPLOYED: Mon Jun 24 14:57:44 2019
    NAMESPACE: default
    STATUS: DEPLOYED
    
    RESOURCES:
    ==> v1/ClusterRole
    NAME                          AGE
    seldon-operator-manager-role  2s
    
    ==> v1/ClusterRoleBinding
    NAME                                 AGE
    seldon-operator-manager-rolebinding  2s
    
    ==> v1/Pod(related)
    NAME                                  READY  STATUS             RESTARTS  AGE
    seldon-operator-controller-manager-0  0/1    ContainerCreating  0         2s
    
    ==> v1/Secret
    NAME                                   TYPE    DATA  AGE
    seldon-operator-webhook-server-secret  Opaque  0     3s
    
    ==> v1/Service
    NAME                                        TYPE       CLUSTER-IP   EXTERNAL-IP  PORT(S)  AGE
    seldon-operator-controller-manager-service  ClusterIP  10.76.8.100  <none>       443/TCP  2s
    
    ==> v1/StatefulSet
    NAME                                READY  AGE
    seldon-operator-controller-manager  0/1    2s
    
    ==> v1beta1/CustomResourceDefinition
    NAME                                         AGE
    seldondeployments.machinelearning.seldon.io  2s
    
    
    NOTES:
    NOTES: TODO
    
    


And confirm they are running by getting the pods:


```python
!kubectl rollout status deployment/seldon-operator-controller-manager -n seldon-system
```

    Error from server (NotFound): namespaces "seldon-system" not found


## Setup Ingress

This will allow you to reach the Seldon models from outside the kubernetes cluster.

In EKS it automatically creates an Elastic Load Balancer, which you can configure from the EC2 Console.


```python
%%bash
helm repo add datawire https://www.getambassador.io
helm repo update
helm install ambassador datawire/ambassador \
    --set image.repository=quay.io/datawire/ambassador \
    --set enableAES=false \
    --set crds.keep=false
```

    NAME:   ambassador
    LAST DEPLOYED: Mon Jun 24 14:58:01 2019
    NAMESPACE: default
    STATUS: DEPLOYED
    
    RESOURCES:
    ==> v1/Deployment
    NAME        READY  UP-TO-DATE  AVAILABLE  AGE
    ambassador  0/3    3           0          1s
    
    ==> v1/Pod(related)
    NAME                         READY  STATUS             RESTARTS  AGE
    ambassador-865c877494-2td9s  0/1    ContainerCreating  0         0s
    ambassador-865c877494-2vsk2  0/1    ContainerCreating  0         0s
    ambassador-865c877494-qzh4c  0/1    ContainerCreating  0         0s
    
    ==> v1/Service
    NAME               TYPE          CLUSTER-IP    EXTERNAL-IP  PORT(S)                     AGE
    ambassador         LoadBalancer  10.76.8.138   <pending>    80:30783/TCP,443:32277/TCP  1s
    ambassador-admins  ClusterIP     10.76.12.144  <none>       8877/TCP                    1s
    
    ==> v1/ServiceAccount
    NAME        SECRETS  AGE
    ambassador  1        1s
    
    ==> v1beta1/ClusterRole
    NAME        AGE
    ambassador  1s
    
    ==> v1beta1/ClusterRoleBinding
    NAME        AGE
    ambassador  1s
    
    ==> v1beta1/CustomResourceDefinition
    NAME                                          AGE
    authservices.getambassador.io                 1s
    consulresolvers.getambassador.io              1s
    kubernetesendpointresolvers.getambassador.io  1s
    kubernetesserviceresolvers.getambassador.io   1s
    mappings.getambassador.io                     1s
    modules.getambassador.io                      1s
    ratelimitservices.getambassador.io            1s
    tcpmappings.getambassador.io                  1s
    tlscontexts.getambassador.io                  1s
    tracingservices.getambassador.io              1s
    
    
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

    Waiting for deployment "ambassador" rollout to finish: 0 of 3 updated replicas are available...
    Waiting for deployment "ambassador" rollout to finish: 1 of 3 updated replicas are available...
    Waiting for deployment "ambassador" rollout to finish: 2 of 3 updated replicas are available...
    deployment "ambassador" successfully rolled out


## Build the Seldon Graph

First lets look at the Seldon Graph Yaml file:


```python
!cat deep_mnist_gpu.json
```

    {
        "apiVersion": "machinelearning.seldon.io/v1alpha2",
        "kind": "SeldonDeployment",
        "metadata": {
            "labels": {
                "app": "seldon"
            },
            "name": "deep-mnist-gpu"
        },
        "spec": {
            "annotations": {
                "project_name": "Tensorflow MNIST",
                "deployment_version": "v1"
            },
            "name": "deep-mnist-gpu",
            "oauth_key": "oauth-key",
            "oauth_secret": "oauth-secret",
            "predictors": [
                {
                    "componentSpecs": [{
                        "spec": {
                            "containers": [
                                {
                                    "image": "gcr.io/dev-joel/deep-mnist-gpu:0.1",
                                    "imagePullPolicy": "IfNotPresent",
                                    "name": "classifier",
                                    "resources": {
                                        "requests": {
                                            "memory": "1Mi"
                                        },
    				    "limits": {
    					"nvidia.com/gpu": 2
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


**Change the image name in this file (line 24) to match the path to the image in your container registry.**

```
$vim deep_mnist_gpu.json
```

Next, we are ready to **build the seldon graph**.


```python
!kubectl create -f deep_mnist_gpu.json
```

    seldondeployment.machinelearning.seldon.io/deep-mnist-gpu created



```python
!kubectl rollout status deploy/deep-mnist-gpu-single-model-8969cc0
```

    Error from server (NotFound): deployments.extensions "deep-mnist-gpu-single-model-8969cc0" not found


Check the deployment is running


```python
!kubectl get pods
```

    NAME                                                   READY   STATUS    RESTARTS   AGE
    ambassador-865c877494-2td9s                            1/1     Running   0          101m
    ambassador-865c877494-2vsk2                            1/1     Running   0          101m
    ambassador-865c877494-qzh4c                            1/1     Running   0          101m
    deep-mnist-gpu-single-model-0588ac2-865d745b7d-kqcp9   2/2     Running   0          71m
    seldon-operator-controller-manager-0                   1/1     Running   1          101m


## Test the deployment with test data

**Change the IP address to the External IP of your Ambassador deployment.**


```python
!kubectl get svc
```

    NAME                                         TYPE           CLUSTER-IP     EXTERNAL-IP     PORT(S)                      AGE
    ambassador                                   LoadBalancer   10.76.8.138    104.197.71.69   80:30783/TCP,443:32277/TCP   101m
    ambassador-admins                            ClusterIP      10.76.12.144   <none>          8877/TCP                     101m
    deep-mnist-gpu-deep-mnist-gpu                ClusterIP      10.76.5.205    <none>          8000/TCP,5001/TCP            71m
    kubernetes                                   ClusterIP      10.76.0.1      <none>          443/TCP                      107m
    seldon-87fe3957f4554e9b5af993717a0b9327      ClusterIP      10.76.14.160   <none>          9000/TCP                     71m
    seldon-operator-controller-manager-service   ClusterIP      10.76.8.100    <none>          443/TCP                      101m
    webhook-server-service                       ClusterIP      10.76.7.151    <none>          443/TCP                      101m



```python
!seldon-core-api-tester contract.json <EXTERNAL_IP_ADDRESS> `kubectl get svc ambassador -o jsonpath='{.spec.ports[0].port}'` \
    deep-mnist-gpu --namespace default -p
```

    ----------------------------------------
    SENDING NEW REQUEST:
    
    [[0.798 0.827 0.034 0.384 0.938 0.036 0.135 0.555 0.86  0.263 0.411 0.894
      0.327 0.865 0.906 0.914 0.133 0.565 0.803 0.417 0.825 0.678 0.805 0.206
      0.017 0.698 0.41  0.503 0.984 0.214 0.468 0.366 0.132 0.973 0.472 0.346
      0.001 0.662 0.412 0.537 0.522 0.242 0.289 0.676 0.379 0.542 0.452 0.467
      0.392 1.    0.771 0.442 0.352 0.505 0.259 0.505 0.664 0.942 0.457 0.417
      0.895 0.42  0.322 0.885 0.578 0.528 0.222 0.283 0.137 0.605 0.915 0.182
      0.42  0.94  0.262 0.599 0.552 0.437 0.179 0.928 0.831 0.193 0.391 0.416
      0.315 0.012 0.815 0.925 0.52  0.773 0.93  0.673 0.757 0.979 0.151 0.459
      0.621 0.553 0.605 0.176 0.702 0.814 0.784 0.952 0.513 0.125 0.68  0.043
      0.377 0.67  0.466 0.824 0.245 0.221 0.324 0.749 0.182 0.992 0.243 0.855
      0.477 0.176 0.262 0.537 0.69  0.717 0.059 0.711 0.26  0.149 0.34  0.71
      0.041 0.623 0.447 0.319 0.089 0.954 0.435 0.267 0.416 0.275 0.923 0.254
      0.542 0.995 0.782 0.337 0.991 0.187 0.183 0.479 0.73  0.288 0.6   0.583
      0.392 0.389 0.572 0.281 0.016 0.097 0.745 0.161 0.053 0.994 0.998 0.21
      0.348 0.531 0.423 0.894 0.153 0.759 0.277 0.002 0.113 0.236 0.171 0.979
      0.315 0.171 0.217 0.328 0.995 0.231 0.134 0.69  0.468 0.437 0.536 0.198
      0.412 0.15  0.465 0.402 0.975 0.698 0.057 0.885 0.433 0.463 0.73  0.285
      0.429 0.068 0.942 0.367 0.96  0.042 0.383 0.498 0.563 0.606 0.139 0.148
      0.151 0.4   0.946 0.805 0.954 0.739 0.925 0.305 0.909 0.222 0.475 0.729
      0.679 0.43  0.7   0.085 0.103 0.3   0.073 0.263 0.472 0.998 0.615 0.218
      0.677 0.555 0.155 0.093 0.36  0.149 0.343 0.801 0.896 0.106 0.253 0.875
      0.245 0.853 0.909 0.958 0.362 0.663 0.674 0.298 0.139 0.118 0.242 0.282
      0.095 0.755 0.635 0.168 0.259 0.515 0.77  0.196 0.185 0.659 0.379 0.64
      0.351 0.184 0.723 0.639 0.893 0.132 0.833 0.377 0.486 0.262 0.091 0.694
      0.043 0.957 0.927 0.469 0.47  0.407 0.166 0.673 0.065 0.582 0.403 0.795
      0.39  0.991 0.723 0.863 0.347 0.612 0.63  0.628 0.298 0.398 0.788 0.491
      0.497 0.669 0.016 0.609 0.778 0.379 0.454 0.113 0.4   0.649 0.155 0.687
      0.317 0.248 0.044 0.933 0.615 0.335 0.022 0.661 0.582 0.418 0.053 0.924
      0.69  0.723 0.007 0.149 0.703 0.1   0.799 0.991 0.877 0.626 0.191 0.829
      0.07  0.814 0.989 0.664 0.192 0.849 0.611 0.78  0.397 0.281 0.688 0.876
      0.423 0.185 0.036 0.476 0.417 0.804 0.336 0.498 0.653 0.585 0.339 0.155
      0.438 0.781 0.321 0.462 0.595 0.324 0.463 0.065 0.655 0.534 0.01  0.906
      0.836 0.389 0.457 0.629 0.831 0.145 0.082 0.889 0.231 0.075 0.404 0.408
      0.035 0.226 0.371 0.961 0.907 0.366 0.937 0.818 0.373 0.813 0.645 0.009
      0.16  0.797 0.81  0.48  0.76  0.464 0.127 0.842 0.531 0.362 0.546 0.95
      0.788 0.069 0.276 0.79  0.287 0.64  0.797 0.262 0.132 0.317 0.766 0.759
      0.714 0.642 0.601 0.482 0.529 0.43  0.934 0.07  0.137 0.794 0.5   0.065
      0.157 0.672 0.858 0.336 0.991 0.054 0.352 0.163 0.981 0.481 0.29  0.3
      0.38  0.136 0.911 0.231 0.556 0.798 0.496 0.407 0.237 0.474 0.676 0.356
      0.757 0.954 0.217 0.165 0.948 0.746 0.986 0.501 0.216 0.638 0.398 0.863
      0.462 0.924 0.889 0.448 0.325 0.922 0.895 0.331 0.491 0.626 0.207 0.133
      0.68  0.304 0.126 0.835 0.233 0.485 0.217 0.405 0.44  0.124 0.71  0.332
      0.546 0.58  0.151 0.447 0.104 0.206 0.257 0.053 0.716 0.804 0.67  0.789
      0.804 0.473 0.008 0.318 0.033 0.381 0.634 0.407 0.659 0.62  0.497 0.689
      0.83  0.384 0.67  0.911 0.101 0.668 0.355 0.579 0.111 0.446 0.596 0.814
      0.318 0.355 0.07  0.542 0.017 0.21  0.327 0.599 0.059 0.252 0.951 0.56
      0.367 0.813 0.074 0.964 0.079 0.68  0.446 0.019 0.7   0.903 0.918 0.74
      0.22  0.241 0.656 0.283 0.625 0.209 0.154 0.862 0.254 0.151 0.323 0.789
      0.393 0.023 0.668 0.55  0.408 0.54  0.207 0.064 0.844 0.323 0.216 0.688
      0.273 0.71  0.542 0.32  0.277 0.535 0.621 0.014 0.272 0.235 0.959 0.067
      0.027 0.585 0.001 0.853 0.189 0.687 0.059 0.284 0.419 0.995 0.151 0.391
      0.184 0.741 0.752 0.956 0.646 0.84  0.619 0.993 0.37  0.499 0.491 0.318
      0.782 0.724 0.748 0.552 0.485 0.667 0.206 0.813 0.511 0.128 0.936 0.33
      0.937 0.484 0.157 0.878 0.834 0.133 0.809 0.977 0.567 0.366 0.964 0.535
      0.678 0.64  0.076 0.866 0.211 0.853 0.619 0.103 0.433 0.667 0.73  0.136
      0.519 0.612 0.184 0.044 0.448 0.233 0.885 0.38  0.172 0.804 0.106 0.724
      0.107 0.619 0.554 0.548 0.812 0.587 0.577 0.417 0.962 0.774 0.364 0.485
      0.881 0.533 0.714 0.52  0.963 0.718 0.651 0.375 0.889 0.239 0.148 0.715
      0.551 0.768 0.073 0.599 0.671 0.947 0.059 0.453 0.356 0.271 0.156 0.096
      0.975 0.454 0.594 0.605 0.689 0.151 0.823 0.286 0.107 0.031 0.59  0.801
      0.847 0.291 0.516 0.977 0.883 0.169 0.848 0.954 0.371 0.632 0.313 0.397
      0.944 0.937 0.051 0.193 0.221 0.446 0.327 0.456 0.619 0.924 0.326 0.848
      0.496 0.515 0.668 0.703 0.942 0.712 0.533 0.656 0.691 0.669 0.407 0.42
      0.659 0.933 1.    0.244 0.566 0.613 0.747 0.896 0.236 0.355 0.338 0.243
      0.069 0.416 0.684 0.923 0.392 0.654 0.523 0.38  0.319 0.327 0.522 0.985
      0.01  0.316 0.938 0.907]]
    RECEIVED RESPONSE:
    meta {
      puid: "14k74obmqhus06jl6pai9hcg7r"
      requestPath {
        key: "classifier"
        value: "gcr.io/dev-joel/deep-mnist-gpu:0.1"
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
              number_value: 0.0025008211378008127
            }
            values {
              number_value: 7.924897005295861e-08
            }
            values {
              number_value: 0.057240355759859085
            }
            values {
              number_value: 0.21792393922805786
            }
            values {
              number_value: 6.878228759887861e-06
            }
            values {
              number_value: 0.5588285326957703
            }
            values {
              number_value: 0.0005614690016955137
            }
            values {
              number_value: 0.0004520844086073339
            }
            values {
              number_value: 0.161981999874115
            }
            values {
              number_value: 0.0005038614035584033
            }
          }
        }
      }
    }
    
    


## Clean up

Make sure you delete the cluster once you have finished with it to avoid any ongoing charges.


```python
!gcloud container clusters delete <YOUR_CLUSTER_NAME>
```
