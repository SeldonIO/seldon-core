# Packaging a model for seldon wrappers

Seldon python wrappers are designed to load a saved model and package it into a docker image. In order to use the wrappers, the saved model file need to be placed in a dedicated folder \<your_model_folder>. Moreover, you will need a base docker image  \<your_base_image>, a docker repository, a name and a version for the model image, respectivly  \<your_docker_repo>,  \<your_model_name>,  \<your_model_version> .

Here we illustrate the content of the ```keras_mnist``` model folder which can be found in ["seldon-core-example/models/keras_mnist](link to github). In this example we have:

* \<your_model_folder> = seldon-core-examples/models/keras_mnist
* \<your_model_name> = MnistClassifier
* \<your_model_version> = 0.0
* \<your_docker_repo> = seldonio
* \<your_base_image> = Python:2

The steps are general and can be used to package your own model for seldon wrappers

# Keras mnist classifier

1.  seldon-core-examples/models/keras_mnist:  The model folder must include the following 3 files:
    * MnistClassifier.py: Needs to include a python class having the same name as the file, i,e. MnistClassifier, and implementing the  methods \__init__()  and predict(). The following template shows the structure of the file:
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
	
	* requirements.txt: List of the packages required by your model. Such packages must be installable through ```pip install```. For example,   the requirements.txt file for the keras example presented in the next session is:
	
		    keras==2.0.6 
		    h5py
 	    	
	* MnistClassifier.h5: The file with your saved model, loadable with load_model() function. 
	
2. MnistClassifier: The name of the model.  The .py file in your model folder and the class implemented in it have to be called both \<your_model_name>, e.g MnistClassifier.

3. 0.0: The version of your model, e.g.  0.0.

4. seldonio: The repository for the image, e.g. seldonio.

5. Python:2: The base image for the model, default is Python:2.
