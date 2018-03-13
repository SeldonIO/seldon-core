# Packaging a python model for Seldon Core using Seldon Wrapper
In this guide, we illustrate the steps needed to wrap your own python model in a docker image ready for deployment with Seldon Core, using the Seldon wrapper script. This script is designed to take your python model and turn it into a dockerised microservice that conforms to Seldon's internal API, thus avoiding the hassle to write your own dockerised microservice.

You can use these wrappers with any model that offers a python API. Some examples are:

 * [TensorFlow](https://www.tensorflow.org/)
 * [Keras](https://keras.io/)
 * [pyTorch](http://pytorch.org/)
 * [StatsModels](http://www.statsmodels.org/stable/index.html)
 * [Scikit-learn](http://scikit-learn.org/stable/)
 * [XGBoost](https://github.com/dmlc/xgboost)

The global process is as follows:
* Regroup your files under a single directory and create a standard python class that will be used as an entrypoint
* Run the wrapper script that will package your model for docker
* Build and publish a docker image from the generated files


## Create a model folder

The Seldon python wrappers are designed to take your model and turn it into a dockerised microservice that conforms to Seldon's internal API. 
To wrap a model, there are 2 requirements:
* All the files that will be used at runtime need to be put into a single directory;
* You need a file that contains a standardised python class that will serve as an entrypoint for runtime predictions.

Additionally, if you are making use of specific python libraries, you need to list them in a requirements.txt file that will be used by pip to install the packages in the docker image.

Here we illustrate the content of the ```keras_mnist``` model folder which can be found in [seldon-core/examples/models/](https://github.com/SeldonIO/seldon-core/tree/master/examples).

This folder contains the following 3 files: 

1. MnistClassifier.py: This is the entrypoint for the model. It needs to include a python class having the same name as the file, in this case MnistClassifier, that implements a method called predict that takes as arguments a multi-dimensional numpy array (X) and a list of strings (feature_names), and returns a numpy array of predictions. 

	
    ```python
    from keras.models import load_model
	    
    class MnistClassifier(object): # The file is called MnistClassifier.py
	    
        def __init__(self):
			""" You can load your pre-trained model in here. The instance will be created once when the docker container starts running on the cluster. """
            self.model = load_model('MnistClassifier.h5')
		    
        def predict(self,X,feature_names):
			""" X is a 2-dimensional numpy array, feature_names is a list of strings. 
			This methods needs to return a numpy array of predictions."""
            return self.model.predict(X)
    ```
	
2. requirements.txt: List of the packages required by your model, that will be installed via ```pip install```.

   ```
   keras==2.0.6 
   h5py==2.7.0
   ```
 	    	
3. MnistClassifier.h5: This hdf file contains the saved keras model. 

## Wrap the model

After you have copied the required files in your model folder, you run the Seldon wrapper script to turn your model into a dockerised microservice. The wrapper script requires as arguments the path to your model directory, the model name, a version for the docker image, and the name of a docker repository. It will generate a "build" directory that contains the microservice, Dockerfile, etc.

In order to make things even simpler (and because we love Docker!) we have dockerised the wrapper script so that you don't need to install anything on your machine to run it - except Docker.

```
docker run -v /path/to/model/dir:/my_model seldonio/core-python-wrapper:0.7 /my_model MnistClassifier 0.1 seldonio
```

Let's explain each piece of this command in more details.


``` docker run seldonio/core-python-wrapper:0.7 ``` : run the core-python-wrapper container.

``` -v /path/to/model/dir:/my_model ``` : Tells docker to mount your local folder to /my_model in the container. This is used to access your files and generate the wrapped model files. 

``` /my_model MnistClassifier 0.1 seldonio ``` : These are the command line arguments that are passed to the script. The bare minimum, as in this example, are the path where your model folder has been mounted in the container, the model name, the docker image version and the docker hub repository.

For reference, here is the complete list of arguments that can be passed to the script.

```
docker run -v /path:<model_path> seldonio/core-python-wrapper:0.7 
	<model_path>
	<model_name>
	<image_version>
	<docker_repo>
	--out-folder=<out_folder>
	--service-type=<service_type>
	--base-image=<base_image>
	--image-name=<image_name>
	--force
	--persistence
	--grpc
```

Required:
* model_path: The path to the model folder inside the container - the same as the mount you have chosen, in our above example my_model
* model_name: The name of your model class and file, as defined abobe. In our example, MnistClassifier
* image_version: The version of the docker image you will create. By default the name of the image will be the name of your model in lowercase (more on how to change this later). In our example 0.1
* docker_repo: The name of your dockerhub repository. In our example seldonio.

Optional:
* out-folder: The folder that will be created to contain the output files. Defaults to ./build
* service-type: The type of Seldon Service API the model will use. Defaults to MODEL. Other options are ROUTER, COMBINER, TRANSFORMER, OUTPUT_TRANSFORMER
* base-image: The docker image your docker container will inherit from. Defaults to python:2.
* image-name: The name of your docker image. Defaults to model_name in lowercase
* force: When this flag is present, the build folder will be overwritten if it already exists. The wrapping is aborted by default.
* persistence: When this flag is present, the model will be made persistent, its state being saved at a regular interval on redis.
* grpc: When this flag is present, the model will expose a GRPC API rather than the default REST API

Note that you can access the command line help of the script by using the -h or --help argument as follows:

```
docker run seldonio/core-python-wrapper:0.7 -h
```

Note also that you could use the python script directly if you feel so enclined, but you would have to check out seldon-core and install some python libraries on your local machine - by using the docker image you don't have to care about these dependencies.

## Build and push the Docker image

A folder named "build" should have appeared in your model directory. It contains all the files needed to build and publish your model's docker image.

To do so, run:

```
cd /path/to/model/dir/build
./build_image.sh
./push_image.sh
```

And voila, the docker image for your model is now available in your docker repository, and Seldon Core can deploy it into production.
