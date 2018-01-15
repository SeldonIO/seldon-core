# Packaging a H2O model for seldon core

In this readme we outline the steps needed to wrap any H2O model using seldon python wrappers into a docker image ready for deployment with seldon core. The process is nearly identical to [wrapping a python model](python.md), so be sure to read the documentation on this process first. The only difference is that you need to provide a different base image for your docker image, because H2O needs java installed on the image to work.
We provide a file H2oModel.py as a template that can be used with minimal work. The only modification required is setting the MODEL_PATH variable at the top of the file.
We also provide an example of use to train and save a prebuilt H2O model predicting bad loans. 

It is assumed you have already trained a H2O model and saved it in a file (called in what follows MyH2OModel.h5). If you use the H2O python API, you can save your model using the```h2o.save_model()``` method.

## Building the H2O base docker image

In order to wrap an H2O model with the python wrappers, you need a python+java docker image available to use as base image. One way to build a suitable base image locally is by using the [Dockerfile provided by H2O](https://h2o-release.s3.amazonaws.com/h2o/rel-turing/1/docs-website/h2o-docs/docker.html):

* Make sure you have docker deamon running.
* Download the [Dockerfile provided by H2O](https://github.com/h2oai/h2o-3/blob/master/Dockerfile) in any folder.
* Create the base docker image (we will call it H2OBase:1.0 in this example):

    ```bash
    docker build --force-rm=true -t H2OBase:1.0 .
    ```

Building the image may take several minutes.

## Wrapping the model

You can now wrap the model using seldon python wrappers. This is similar to the general python model wrapping process except that you need to specify the H2O base image as an argument when callilng the wrapping script.

You can write your own entrypoint python file or use the H2oModel.py file we provide [here](https://github.com/SeldonIO/seldon-core-examples/blob/master/models/h2o_example/H2oModel.py). In what follows we assume you are using our file.

1. Put the files H2oModel.py, requirements.txt and MyH2OModel.h5 in a directory created for this purpose (h2o_model in what follows).
2. Open the file H2oModel.py with your favorite editor and set the variable MODEL_PATH to:

    ```python
    MODEL_PATH=/microservice/MyH2OModel.h5
    ```
       
3. Run the python wrapping scripts:

	```bash
	docker run -v /path/to/your/model/folder:/my_model seldonio/core-python-wrapper:0.4 /my_model H2oModel 0.1 myrepo --base-image=H2OBase:1.0
	```
	
4. CD into the newly generated "build" directory and run:

   ```bash
   make build_image
   make publish_image
   ```

    This will build and publish a docker image named ```myrepo/h2omodel:0.1``` which is ready for deployment in seldon-core.

## Example

Here we give a step by step example in which we will train and save a [H2O model for bad loans predictions](https://github.com/h2oai/h2o-tutorials/blob/master/h2o-open-tour-2016/chicago/intro-to-h2o.ipynb), and we use seldon wrappers to turn it into a dockerized microservice.

### Preliminary Requirements

1. Have [H2O](http://docs.h2o.ai/h2o/latest-stable/h2o-docs/downloading.html) installed on your machine (H2O is only required to train the example. Seldon-core and seldon wrappers do not require h2o installed on your machine)
2. You need to have built the base H2O docker image (see the [dedicated section](#building-the-h2o-base-docker-image) above)
2. Make sure you have a  docker deamon running
3. Download the [Dockerfile provided by H2O](https://github.com/h2oai/h2o-3/blob/master/Dockerfile).
4. Run ``` docker build --force-rm=true -t none/h2obase:1.0 .``` in the same directory. This will create the base image "none/h2obase:1.0" locally (may take several minutes).

### Train and wrap the model

1. Clone the seldon-core-examples git repository

    ```bash
    git clone https://github.com/SeldonIO/seldon-core-examples
    ```
    
2. Train and save the H2O model for bad loans prediction:

    ```bash
    cd seldon-core-examples/models/h2o_example/
    python train_model.py
    ```

    This will train the model and save it in a file named  "glm_fit1"" in the same directory.

* Wrap the model: 
    ```bash
    cd ../../
	docker run -v models/h2o_example:my_model seldonio/core-python-wrapper:0.4 my_model H2oModel 0.1 myrepo --base-image=H2OBase:1.0
    ``` 

    This will create a docker image "seldonio/h2omodel:0.1", which is ready to be deployed in seldon-core.
