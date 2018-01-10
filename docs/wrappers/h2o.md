# Packaging a H2o model for seldon core

In this readme we outline the steps needed to wrap any h2o model using seldon python wrappers into a docker image deployable with seldon core. 
The file H2oModel.py is a template to use in order to wrap a h2o model. The only modification  required on the user side consists of setting the MODEL_PATH variable at the top of the file.
We also provide a example of usage to train and save a prebuilded h2o model predicting bad loans. 

The session "General use" explain how to use the wrapper with any saved h2o model.

The session "Example of usage" provides a step-by-step guide for training, deploying and wrap the prebuilded h2o model for bad loans predictions as an example.

## General use

It is assumed you have already trained a h2o model and saved it in a file \<your_file_name> using the```h2o.save_model()``` python function.

### Preliminary steps

In order to wrap an H2o model with the python wrappers, you need  a python+java docker image avaliable to use as base-image. One way to build a suitable base-image locally is by using the [Dockerfile provided by h2o](https://h2o-release.s3.amazonaws.com/h2o/rel-turing/1/docs-website/h2o-docs/docker.html):

* Make sure you have docker deamon running.
* Download the [Dockerfile provided by h2o](https://github.com/h2oai/h2o-3/blob/master/Dockerfile) in any folder.
* Create the base docker image:

    ```bash
    docker build --force-rm=true -t <your_base_image> .
    ```

Building the image may take several minutes.

### Wrap:

You can now wrap the model using seldon python wrappers. 

1. Clone the ```seldon-core-examples``` git repository and copy the files H2oModel.py, requirements.txt and \<your_file_name> in a folder named \<your_model_folder> .
* Open the file H2oModel.py with your favorite editor and set the variable MODEL_PATH to:

    ```python
    MODEL_PATH=/microservice/<your_file_name>
    ```
       
2. Cd into ```seldon-core-examples``` directory and use the python wrapping scripts:

    ```bash
    ./wrap-model-in-minikube <path_to_your_model_folder> H2oModel <your_model_version> <your_docker_repo> --base-image <your_base_image> --force
    ```
    to build your docker image in minikube or

    ```bash
    ./wrap-model-in-host <path_to_your_model_folder> H2oModel <your_model_version> <your_docker_repo> --base-image <your_base_image> --force
    ```
    to build your docker image in your machine.
    
    This will create  a docker image named ```<your_docker_repo>/h2omodel:<your_model_version>``` which is ready for deployment in seldon-core.


## Example of usage

Here we give an example of usage step by step in which we will train and save a [h2o model for bad loan predictions](https://github.com/h2oai/h2o-tutorials/blob/master/h2o-open-tour-2016/chicago/intro-to-h2o.ipynb), we will create a base image supporting h2o named "seldonio/h2obase:0.1" and  we will use seldon wrappers to build  dockererized version of the model ready to be deployed with seldon-core. 

### Preliminary step: build  your base image locally

1. Have [h2o](http://docs.h2o.ai/h2o/latest-stable/h2o-docs/downloading.html) installed on your machine (h2o is only required to train the example. Seldon-core and seldon wrappers do not require h2o installed on your machine)
1. Make sure you have a  docker deamon running
* Download the [Dockerfile provided by h2o](https://github.com/h2oai/h2o-3/blob/master/Dockerfile) in any directory.
* Run ``` docker build --force-rm=true -t none/h2obase:0.0 .``` in the same directory. This will create the base image "none/h2obase:0.0" locally (may take several minutes).

### Train and wrap the model

1. Clone seldon-core-examples git repository

    ```bash
    git clone https://github.com/SeldonIO/seldon-core-examples
    ```
    
2. Train and save the H2o  model for bad loans prediction:

    ```bash
    cd seldon-core-examples/models/h2o_example/
    ```
    ```bash
    python train_model.py
    ````

    This will train the model and save it in a file named  "glm_fit1"" in the same directory.

* Wrap the model: 
    ```bash
    cd ../../
    ```
       
    ```bash 
    ./wrap-model-in-minikube models/h2o_example H2oModel 0.1 seldonio --base-image none/h2obase:0.0 --force
    ``` 

    This will create a docker image "seldonio/h2omodel:0.1", which is ready to be deployed in seldon-core.
