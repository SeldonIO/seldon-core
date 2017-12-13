---
title: "H2o bad loan predictor example"
date: 2017-12-09T17:49:41Z
---
## General use

In this readme we outline the steps needed to wrap any h2o model using seldon python wrappers into a docker image deployable with seldon core. 

The file H2oModel.py is a template to use in order to wrap a h2o model. The only modification  required on the user side consists of setting the MODEL_PATH variable at the top of the file as explained below in session "Wrap" at point 2. 

We also provide a script to train and save a prebuilded h2o model predicting bad loans. 

The session "General steps" explain how to use the wrapper with any saved h2o model.

The session "Example of usage" provides a step-by-step guide for training, deploying and wrap the prebuilded h2o model for bad loans predictions as an example.


### Preliminary steps

1. It is assumed you have already trained a h2o model and saved it in a file \<your_file_name> using the ```h2o.save_model()``` python function.
* You have cloned the latest version of seldon-core git repository.
* You have a python+java docker image avaliable to use as base-image. One way to build a suitable base-image locally is by using the [Dockerfile provided by h2o](https://h2o-release.s3.amazonaws.com/h2o/rel-turing/1/docs-website/h2o-docs/docker.html):
	* Make sure you have docker deamon runnin	
	* Download the [Dockerfile provided by h2o](https://h2o-release.s3.amazonaws.com/h2o/rel-turing/1/docs-website/h2o-docs/docker.html) in any folder.
	* Create the base docker image:

		``` docker build --force-rm=true -t <your_base_image> .```

		it may take several minutes.

### Wrap:

You can now wrap the model using seldon python wrappers. 

1. Copy the files H2oModel.py, requirements.txt and \<your_file_name> in a folder named \<your_model_folder> .
* Open the file H2oModel.py with your favorite editor and set the variable MODEL_PATH to:

       ```MODEL_PATH=/microservice/<your_file_name>``` 
* Cd into the the python wrapper directory in seldon-core, ```/seldon-core/wrapper/python```
* Use the wrap_model.py script to wrap your model:

	```python wrap_model.py <path_to_your_model_folder> H2oModel <your_model_version> <your_docker_repo> --base-image <your_base_image> --force```

	This will create a "build" directory in \<your_model_folder>.	
* Enter the build directory created by wrap_model.py:

  	```cd <path_to_your_model_folder>/build```
* Build the docker image of your model:

  	```make build_docker_image```

	This will build a docker image named ```<your_docker_repo>/h2omodel:<your_model_version>```.

### Deploy in seldon-core:

You can now deploy the model as a docker image ```<your_docker_repo>/h2omodel:<your_model_version>``` using [seldon-core](link).

## Example of usage

Here we give an example of usage step by step in which we will train and save a [h2o model for bad loan predictions](https://github.com/h2oai/h2o-tutorials/blob/master/h2o-open-tour-2016/chicago/intro-to-h2o.ipynb). We will create a base image supporting h2o named "none/h2obase:0.0" and  we will use seldon wrappers to build model  docker image ready to be deployed with seldon-core. 

### Preliminary step:
The first step is to build locally your base image:

1. Make sure you have docker deamon running
* Download the [Dockerfile provided by h2o](https://h2o-release.s3.amazonaws.com/h2o/rel-turing/1/docs-website/h2o-docs/docker.html) in any directory.
* Run ``` docker build --force-rm=true -t none/h2obase:0.0 .``` in the same directory. This will create the base image "none/h2obase:0.0" locally (may take several minutes).

### Train and wrap the model

1. Clone seldon-core and seldon-core-plugins git repositories in the same directory.
* Train and saved the model:

  	```cd seldon-core-plugins/h2o_example```
	
	```python train_model.py```
	
	This will train the model and save it in a file named  "glm_fit1"" in the same directory.
* Wrap the model: 

       ```cd ../../seldon-core/wrappers/python``` 
       
       ```python wrap_model.py ../../../seldon-core-plugins/h2o_example H2oModel 0.0 none --base-image none/h2obase:0.0 --force``` 

       This will create a  directory "build" in "seldon-core-plugins/h2o_example".
* Create docker image: 

  	```cd ../../../seldon-core-plugins/h2o_example/build``` 
	
	```make build_docker_image``` 

	This will create a docker image "none/h2omodel:0.0", which is ready to be deployed in seldon-core.

Note that the  steps 3 and 4 are equivalent to steps 3-6 in the general use session above.