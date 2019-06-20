
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
!s2i build . joelh1996/gpu-base:0.5 deep-mnist-gpu:0.1
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

## Install Seldon Core

**Before installing Seldon Core, we need to install HELM**

To do so, we need to creat a ClusterRoleBinding for us, a ServiceAccount and then a RoleBinding


```python
!kubectl create clusterrolebinding kube-system-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:default
```


```python
!kubectl create serviceaccount tiller --namespace kube-system
```


```python
!kubectl apply -f tiller-role-binding.yaml
```

**Once that is set-up we can install Tiller**


```python
!helm repo update
```


```python
!helm init --service-account tiller
```


```python
# Wait until Tiller finishes
!kubectl rollout status deploy/tiller-deploy -n kube-system
```

**Now we can install SELDON.**

We first start with the custom resource definitions (CRDs)


```python
!helm install seldon-core-operator --name seldon-core-operator --repo https://storage.googleapis.com/seldon-charts
```

And confirm they are running by getting the pods:


```python
!kubectl rollout status statefulset.apps/seldon-operator-controller-manager -n seldon-system
```

## Setup Ingress

This will allow you to reach the Seldon models from outside the kubernetes cluster.

In EKS it automatically creates an Elastic Load Balancer, which you can configure from the EC2 Console.


```python
!helm install stable/ambassador --name ambassador --set crds.keep=false
```

And let's wait until it's fully deployed


```python
!kubectl rollout status deployment.apps/ambassador
```

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
                                    "image": "gcr.io/<YOUR_PROJECT_ID>/deep-mnist-gpu:0.1",
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


**Change the image name in this file (line 24) to match the path to the image in your container registry.**

```
$vim deep_mnist_gpu.json
```

Next, we are ready to **build the seldon graph**.


```python
!kubectl create -f deep_mnist_gpu.json
```


```python
!kubectl rollout status deploy/deep-mnist-gpu-single-model-8969cc0
```

Check the deployment is running


```python
!kubectl get pods
```

## Test the deployment with test data

**Change the IP address to the External IP of your Ambassador deployment.**


```python
!kubectl get svc
```


```python
!seldon-core-api-tester contract.json <EXTERNAL_IP_ADDRESS> `kubectl get svc ambassador -o jsonpath='{.spec.ports[0].port}'` \
    deep-mnist-gpu --namespace default -p
```

## Clean up

Make sure you delete the cluster once you have finished with it to avoid any ongoing charges.


```python
!gcloud container clusters delete <YOUR_CLUSTER_NAME>
```
