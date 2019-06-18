
# Tensorflow GPU MNIST Model

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


3. Test using minikube


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

## 3) Test using Minikube

**Due to a [minikube/s2i issue](https://github.com/SeldonIO/seldon-core/issues/253) you will need [s2i >= 1.1.13](https://github.com/openshift/source-to-image/releases/tag/v1.1.13)**


```python
!minikube start --memory 4096
```


```python
!kubectl create clusterrolebinding kube-system-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:default
```


```python
!helm init
```


```python
!kubectl rollout status deploy/tiller-deploy -n kube-system
```


```python
!helm install ../../../helm-charts/seldon-core-operator --name seldon-core --set usageMetrics.enabled=true --namespace seldon-system
```


```python
!kubectl rollout status statefulset.apps/seldon-operator-controller-manager -n seldon-system
```

## Setup Ingress
There are gRPC issues with the latest Ambassador, so we rewcommend 0.40.2 until these are fixed.


```python
!helm install stable/ambassador --name ambassador --set crds.keep=false
```


```python
!kubectl rollout status deployment.apps/ambassador
```

## Wrap Model and Test


```python
!eval $(minikube docker-env) && s2i build . seldonio/seldon-core-s2i-python2:0.5.1 deep-mnist:0.1
```


```python
!kubectl create -f deep_mnist.json
```


```python
!kubectl rollout status deploy/deep-mnist-single-model-8969cc0
```


```python
!seldon-core-api-tester contract.json `minikube ip` `kubectl get svc ambassador -o jsonpath='{.spec.ports[0].nodePort}'` \
    deep-mnist --namespace default -p
```


```python
!minikube delete
```


```python

```
