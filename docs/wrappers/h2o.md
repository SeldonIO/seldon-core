# Packaging a H2O model for Seldon Core

This document outlines the steps needed to wrap any H2O model using Seldon's python wrappers into a docker image ready for deployment with Seldon Core. The process is nearly identical to [wrapping a python model](python.md), so be sure to read the documentation on this process first. 
The main differences are:
* The data sent to the model needs to be transformed from numpy arrays into H2O Frames and back;
* The base docker image has to be changed, because H2O needs a Java Virtual Machine installed in the container.

You will find below explanations for:
* How to build the H2O base docker image;
* How to wrap your H2O model;
* A detailed example where we train and wrap a bad loans prediction model.

## Building the H2O base docker image

In order to wrap a H2O model with the python wrappers, you need a python+java docker image available to use as base image. One way to build a suitable base image locally is by using the [Dockerfile provided by H2O](https://h2o-release.s3.amazonaws.com/h2o/rel-turing/1/docs-website/h2o-docs/docker.html):

* Make sure you have docker deamon running.
* Download the [Dockerfile provided by H2O](https://github.com/h2oai/h2o-3/blob/master/Dockerfile) in any folder.
* Create the base docker image (we will call it H2OBase:1.0 in this example):

    ```bash
    docker build --force-rm=true -t H2OBase:1.0 .
    ```

Building the image may take several minutes.

## Wrapping the model


It is assumed you have already trained a H2O model and saved it in a file (called in what follows SavedModel.h2o). If you use the H2O python API, you can save your model using the```h2o.save_model()``` method.

You can now wrap the model using Seldon's python wrappers. This is similar to the general python model wrapping process except that you need to specify the H2O base image as an argument when calling the wrapping script.

We provide a file [H2OModel.py](https://github.com/SeldonIO/seldon-core/blob/master/examples/models/h2o_example/H2OModel.py) as a template for the model entrypoint, which handles loading the H2OModel and transforming the data between numpy and H2O Frames.  In what follows we assume you are using this template. The H2O model is loaded in the class constructor and the numpy arrays are turned into H2O Frames when received in the predict method. 

Detailed steps:
1. Put the files H2OModel.py, requirements.txt and SavedModel.h2o in a directory created for this purpose.
2. Open the file H2OModel.py with your favorite text editor and set the variable MODEL_PATH to:

    ```python
    MODEL_PATH=./SavedModel.h2o
    ```
       
3. Run the python wrapping scripts, with the additional ````--base-image``` argument:

	```bash
	docker run -v /path/to/your/model/folder:/model seldonio/core-python-wrapper:0.6 /model H2OModel 0.1 myrepo --base-image=H2OBase:1.0
	```
	
	"0.1" is the version of the docker image that will be created. "myrepo" is the name of your dockerhub repository.
	
4. CD into the newly generated "build" directory and run:

   ```bash
   ./build_image.sh
   ./push_image.sh
   ```

    This will build and push to dockerhub a docker image named ```myrepo/h2omodel:0.1``` which is ready for deployment in seldon-core.

## Example

Here we give a step by step example in which we will train and save a [H2O model for bad loans predictions](https://github.com/h2oai/h2o-tutorials/blob/master/h2o-open-tour-2016/chicago/intro-to-h2o.ipynb), before turning it into a dockerized microservice.

### Preliminary Requirements

1. Have [H2O](http://docs.h2o.ai/h2o/latest-stable/h2o-docs/downloading.html) installed on your machine (H2O is only required to train the example. Seldon Core and Seldon wrappers do not require H2O installed on your machine)
2. You need to have built the base H2O docker image (see the [dedicated section](#building-the-h2o-base-docker-image) above)

### Train and wrap the model

1. Clone the Seldon Core git repository

    ```bash
    git clone https://github.com/SeldonIO/seldon-core
    ```
    
2. Train and save the H2O model for bad loans prediction:

    ```bash
    cd seldon-core/examples/models/h2o_example/
    python train_model.py
    ```

    This will train the model and save it in a file named  "glm_fit1"" in the same directory.

3. Wrap the model:

    ```bash
    cd ../../
	docker run -v models/h2o_example:my_model seldonio/core-python-wrapper:0.6 my_model H2OModel 0.1 myrepo --base-image=H2OBase:1.0
    ``` 

    This will create a docker image "seldonio/h2omodel:0.1", which is ready to be deployed in seldon-core.
