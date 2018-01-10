# Packaging a python  model for seldon core
In this guide, we illustrate the steps needed to wrap your own python  model in a docker image ready for deployment with seldon-core. 
The steps are general and can be used to package any model that can be loaded in python (either via a pure python function or python API function) for seldon wrappers.

## Create a model folder

Seldon python wrappers are designed to load a saved model and package it into a docker image. In order to use the wrappers, the loadable containing your model needs to be placed in a dedicated folder \<your_model_folder>.

Here we illustrate the content of the ```keras_mnist``` model folder which can be found in [seldon-core-example/models/](https://github.com/SeldonIO/seldon-core-examples). In this example we have \<your_model_folder> = seldon-core-examples/models/keras_mnist.

Any model folder must include the following 3 files (if you build your own model, rename the files where appropriate):
1. MnistClassifier.py: Needs to include a python class having the same name as the file, in this case MnistClassifier, and implementing the  methods \__init__()  and predict(). The following template shows the structure of the file:
    * General template:
        ```python
        from <your_python_loading_library> import <your_loading_function>
            
        class <your_model_name>(object): #Must be the same as the name of the module

            def __init__(self):
                self.model = <your_loading_function>(<your_saved_model>)
				  
            def predict(self,X,features_names):
                return self.model.predict(X)
        ```
    * Keras mnist example:
        ```python
        from keras.models import load_model
	    
        class MnistClassifier(object): #Must be the same as the name of the module
	    
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
 	    	
3. MnistClassifier.h5: The file with your saved model, loadable with load_model() function. 

## Wrap the model

After you have copied the required files in your model folder, you can use seldon wrappers to create a docker image of your model. The seldon wrapper script requires  a model name, a model version, a docker repository and a base docker image as parameters. In our example: 
	
* \<your_model_name> = MnistClassifier: 

    The name of the model.  The .py file in your model folder and the class implemented in it have to be called both \<your_model_name>, e.g MnistClassifier.

* \<your_model_version> = 0.1: 

    The version of your model, e.g.  0.1.

* \<your_docker_repo> = seldonio: 

    The repository for the image, e.g. seldonio.

* \<your_base_image> = Python:2: 
    
    The base image for the model, default is Python:2.

If you are using Minikube, the wrapping can be done as in the [getting started on Minikube session](../getting_started/minikube.md)

```bash
 git clone https://github.com/SeldonIO/seldon-core-examples && cd seldon-core-example 
```
```bash 
./wrap-model-in-minikube models/keras_mnist MnistClassifier 0.1 seldonio --force
```

This will create a docker image ```seldonio/mnistclassifier:0.1``` which is ready for deployment on seldon-core.
