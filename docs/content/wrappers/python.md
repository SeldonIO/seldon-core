---
title: "Python wrappers "
date: 2017-12-09T17:49:41Z
weight: 1
---


In order to deploy a model using seldon-core the model must be packaged into a docker image. In this guide, we explain how to build a docker-image of your model which is ready to be [deployed with seldon-core](link_to_deployment_docs) using seldon python wrappers. The python wrappers are suitable to wrap any saved model that can be loaded and queried  using python functions.

### Preliminary steps 

Clone seldon-core, install grpc tools and buld the protobuffers (if not done before) 

* Clone the latest version of seldon-core git repository: 

    ```git clone seldon-core```
	
* Enter the seldon wrappers directory: 

    ```cd seldon-core/wrappers```
	
* You can skip this step if already done it before. Install grpc tools and build the protobuffers. You only have to do this only once:  

    ```python -m pip install grpcio-tools==1.1.3```
	
    ```make build_protos```

### Wrap a model

* Enter the python directory and run the wrap_model script 

    ```cd python```
	
    ```python wrap_model.py <path_to_your_model_folder> <your_model_name> <your_model_version> <your_docker_repo> --base-image <your_base_image>```
		 
    This will create a "build" directory in \<your_model_folder>.

* Enter  \<your_model_folder> and build your model docker image 

    ```cd <path_to_your_model_folder>/build``` 
	
    ```make build_docker_image``` 
	
    This will  build a docker image of your model locally ready to be [deployed with seldon-core](link_to_deployment_docs).

    
### Notes:


*   \<path_to_your_model_folder>: Your local path to the \<your_model_folder>. The model folder must include the following 3 files:
    1. \<your_model_name>.py: Needs to include a python class having the same name as the file, i,e. \<your_model_name>, and implementing the  methods \__init__  and predict. The following template shows the structure of the file:
        * General template:
            ```python
	    from <your_python_loading_library> import <your_loading_function>

            class <your_model_name>(object):

                def __init__(self):
                    self.model = <your_loading_function>(<your_saved_model>)
					  
                def predict(self,X,features_names):
                    return self.model.predict(X)
          ```

        * Keras mnist example:
	    ```python
	    from keras.models import load_model
	    
            class MnistClassifier(object):
	    
                def __init__(self):
                    self.model = load_model('MnistClassifier.h5')
		    
                def predict(self,X,features_names):
                    if X.shape[0]==784:
                        X = X.reshape(1,28,28,1)
                    else:
                        X = X.reshape(X.shape[0],28,28,1)
                    return self.model.predict(X)
	    ```
	
	2. requirements.txt: List of the packages required by your model. Such packages must be installable through ```pip install```. For example,   the requirements.txt file for the keras example presented in the next session is:
	
		    keras==2.0.6 
		    h5py
 	    	
	3. \<your_saved_model>: The file with your saved model. Must be loadable with <your_loading_function>. For example, in the keras example presented in the next session this file is "saved_model.h5".
	
* \<your_model_name>: The name of the .py file with your model. Has to be the same as the name of the python class implemented in the file, e.g MnistClassifier.

* \<your_model_version>: The version of your model, e.g.  0.0.

* \<your_docker_repo>: The repository for the image, e.g. seldonio.

* \<your_base_image>: THe base image for the model, default is Python:2.


# Example of usage.

Here we include a step-by-step guide to train, save and wrap a mnist classifier from scratch. The model is in "seldon-core-plugins/keras_mnist" and it is builded using keras. In reference to the session above and for the sake of clarity, in this example we have:

* \<your_model_folder> = keras_mnist
* \<path_to_your_model_folder> = ../../../seldon-core-plugins/keras_mnist
* \<your_model_name> = MnistClassifier
* \<your_model_version> = 0.0
* \<your_docker_repo> = seldonio
* \<your_base_image> = Python:2

### Preliminary steps

* Have seldon-core and seldon-core-plugins folders in the same directory and install grpc tools (if not done before).
* Run ```git clone seldon-core```
* Run ```git clone seldon-core-plugins```

### Train and save keras mnist classifier

* Run 

	```cd seldon-core-plugins/keras_mnist```
	
	```python train_mnist.py```

	This will train a keras convolutional neural network on mnist dataset for 2 epochs and save the model in the same folder.

### Wrap saved model

* If not done before, run 

	```python -m pip install grpcio-tools==1.1.3```

	```cd ../../seldon-core/wrappers``` 
	
	```make build_protos```

* Run 

	```cd python``` 
	
	```python wrap_model.py ../../../seldon-core-plugins/keras_mnist MnistClassifier 0.0 seldonio```
	
	This will create the folder build in keras_mnist. The --base-image argument is not specified and the wrapper will use the default base image Python:2.

* Run 

	```cd ../../../seldon-core-plugins/keras_mnist/build/``` 
	
	```make build_docker_image``` 
	
	This will create the docker image ```seldonio/mnistclassifier:0.0``` which is ready for [deployment with seldon-core](link_to_deployment_docs).
