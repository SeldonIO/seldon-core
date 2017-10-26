Use wrappers to create docker image for a model

1 Clone seldon-core and install grpc tools (if not done before) 

    git clone seldon-core
    python -m pip install grpcio-tools
       
2 Build the protos. You only have to do this once (?)

    cd seldon-core/wrappers
    make build_protos

3 Wrap model:

    cd python
    python wrap_model.py <path_to_model_folder> <model_name> <model_verion> <docker_repo> 

4 Build docker image of your model locally

    cd path_to_model_folder
    cd build
    make build_docker_image
    
Notes:
------

*   model_folder is the folder with your model. In the folder you need at least

        model_name.py:
	    Needs to include a class called model_name implementing 2 methods, __init__ and predict.
	    The following example show the content of  model_name.py that load a keras model previusly saved in h5 format.
	
	    from keras.models import load_model
	
	    class model_name(object):

	        def __init__(self):
                    self.model = load_model('saved_model_folder/saved_model.h5) #loads a keras  model saved in saved_model_folder

	        def predict(self,X,features_names):
	            return self.model.predict(X) #return predictions for batch X

        requirements.txt:
	    Lists the requirements needed fot you model:
	    for example:
	
	    keras==2.0.6
 	    h5py

*   model_version is the version of your model, e.g.  0.0

*   docker_repo is the repository for the image. e.g. seldonio

*   Example of usage from scratch with seldon example keras_mnist:

        1. Have seldon-core and seldon-core-plugins folders in the same directory

        2. cd seldon-core-plugins/keras_mnist

        3. python train_mnist.py
	(this will train a keras convolutional neural network on mnist dataset for 2 epochs and save the model in the same folder)

        4. cd ../../seldon-core/wrappers

        5. python -m pip install grpcio-tools
	(if not done before)

        6. make build_protos
	(if not done before)

        7. cd python

        8. python wrap_model.py ../../../seldon-core-plugins/keras_mnist MnistClassifier 0.0 seldonio
	(this will create the folder build in the keras_mnist)

        9. cd ../../../seldon-core-plugins/keras_mnist/build/

        10. make build_docker_image
