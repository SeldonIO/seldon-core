---
title: "Model preparation"
date: 2017-12-09T17:49:41Z
---

### Packaging a model for seldon wrappers

Seldon python wrappers are designed to load a saved model and package it into a docker image. In order to use the wrappers, the saved model file need to be placed in a dedicated folder \<your_model_folder>. Moreover, you will need to use a base docker image  \<your_base_image>, a docker repository, a name and a version for the model image, respectivly  \<your_docker_repo>,  \<your_model_name>,  \<your_model_version> . 

*   \<your_model_folder>:  The model folder must include the following 3 files:
    1. \<your_model_name>.py: Needs to include a python class having the same name as the file, i,e. \<your_model_name>, and implementing the  methods \__init__()  and predict(). The following template shows the structure of the file:
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
 	    	
	3. \<your_saved_model>: The file with your saved model. Must be loadable with <your_loading_function>. For example, in the keras example n this file is "MnistClassifier.h5".
	
* \<your_model_name>: The name of the model.  The .py file in your model folder and the class implemented in it have to be called both \<your_model_name>, e.g MnistClassifier.

* \<your_model_version>: The version of your model, e.g.  0.0.

* \<your_docker_repo>: The repository for the image, e.g. seldonio.

* \<your_base_image>: The base image for the model, default is Python:2.
