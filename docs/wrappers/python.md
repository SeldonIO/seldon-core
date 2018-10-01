# Packaging a python model for Seldon Core using s2i


In this guide, we illustrate the steps needed to wrap your own python model in a docker image ready for deployment with Seldon Core using [source-to-image app s2i](https://github.com/openshift/source-to-image).

If you are not familiar with s2i you can read [general instructions on using s2i](./s2i.md) and then follow the steps below.


# Step 1 - Install s2i

 [Download and install s2i](https://github.com/openshift/source-to-image#installation)

 * Prerequisites for using s2i are:
   * Docker
   * Git (if building from a remote git repo)

To check everything is working you can run

```bash
s2i usage seldonio/seldon-core-s2i-python3:0.2
```


# Step 2 - Create your source code

To use our s2i builder image to package your python model you will need:

 * A python file with a class that runs your model
 * requirements.txt  or setup.py
 * .s2i/environment - model definitions used by the s2i builder to correctly wrap your model

We will go into detail for each of these steps:

## Python file
Your source code should contain a python file which defines a class of the same name as the file. For example, looking at our skeleton python model file at ```wrappers/s2i/python/test/model-template-app/MyModel.py```:

```python
class MyModel(object):
    """
    Model template. You can load your model parameters in __init__ from a location accessible at runtime
    """

    def __init__(self):
        """
        Add any initialization parameters. These will be passed at runtime from the graph definition parameters defined in your seldondeployment kubernetes resource manifest.
        """
        print("Initializing")

    def predict(self,X,features_names):
        """
        Return a prediction.

        Parameters
        ----------
        X : array-like
        feature_names : array of feature names (optional)
        """
        print("Predict called - will run identity function")
        return X
```

 * The file is called MyModel.py and it defines a class MyModel
 * The class contains a predict method that takes an array (numpy) X and feature_names and returns an array of predictions.
 * You can add any required initialization inside the class init method.
 * Your return array should be at least 2-dimensional.

## requirements.txt
Populate a requirements.txt with any software dependencies your code requires. These will be installed via pip when creating the image. You can instead provide a setup.py if you prefer.

## .s2i/environment

Define the core parameters needed by our python builder image to wrap your model. An example is:

```bash
MODEL_NAME=MyModel
API_TYPE=REST
SERVICE_TYPE=MODEL
PERSISTENCE=0
```

These values can also be provided or overridden on the command line when building the image.

# Step 3 - Build your image
Use ```s2i build``` to create your Docker image from source code. You will need Docker installed on the machine and optionally git if your source code is in a public git repo. You can choose from three python builder images

 * Python 2 : seldonio/seldon-core-s2i-python2:0.2
 * Python 3 : seldonio/seldon-core-s2i-python3:0.2
 * Python 3 plus ONNX support via [Intel nGraph](https://github.com/NervanaSystems/ngraph) : seldonio/seldon-core-s2i-python3-ngraph-onnx:0.1

Using s2i you can build directly from a git repo or from a local source folder. See the [s2i docs](https://github.com/openshift/source-to-image/blob/master/docs/cli.md#s2i-build) for further details. The general format is:

```bash
s2i build <git-repo> seldonio/seldon-core-s2i-python2:0.2 <my-image-name>
s2i build <src-folder> seldonio/seldon-core-s2i-python2:0.2 <my-image-name>
```

Change to seldonio/seldon-core-s2i-python3 if using python 3.

An example invocation using the test template model inside seldon-core:

```bash
s2i build https://github.com/seldonio/seldon-core.git --context-dir=wrappers/s2i/python/test/model-template-app seldonio/seldon-core-s2i-python2:0.2 seldon-core-template-model
```

The above s2i build invocation:

 * uses the GitHub repo: https://github.com/seldonio/seldon-core.git and the directory ```wrappers/s2i/python/test/model-template-app``` inside that repo.
 * uses the builder image ```seldonio/seldon-core-s2i-python2```
 * creates a docker image ```seldon-core-template-model```


For building from a local source folder, an example where we clone the seldon-core repo:

```bash
git clone https://github.com/seldonio/seldon-core.git
cd seldon-core
s2i build wrappers/s2i/python/test/model-template-app seldonio/seldon-core-s2i-python2:0.2 seldon-core-template-model
```

For more help see:

```
s2i usage seldonio/seldon-core-s2i-python2:0.2
s2i usage seldonio/seldon-core-s2i-python3:0.2
s2i build --help
```

# Reference

## Environment Variables
The required environment variables understood by the builder image are explained below. You can provide them in the ```.s2i/environment``` file or on the ```s2i build``` command line.


### MODEL_NAME
The name of the class containing the model. Also the name of the python file which will be imported.

### API_TYPE

API type to create. Can be REST or GRPC

### SERVICE_TYPE

The service type being created. Available options are:

 * MODEL
 * ROUTER
 * TRANSFORMER
 * COMBINER
 * OUTLIER_DETECTOR

### PERSISTENCE

Set either to 0 or 1. Default is 0. If set to 1 then your model will be saved periodically to redis and loaded from redis (if exists) or created fresh if not.


## Creating different service types

### MODEL

 * [A minimal skeleton for model source code](https://github.com/cliveseldon/seldon-core/tree/s2i/wrappers/s2i/python/test/model-template-app)
 * [Example models](https://github.com/SeldonIO/seldon-core/tree/master/examples/models)

### ROUTER

 * [A minimal skeleton for router source code](https://github.com/cliveseldon/seldon-core/tree/s2i/wrappers/s2i/python/test/router-template-app)
 * [Example routers](https://github.com/SeldonIO/seldon-core/tree/master/examples/routers)

### TRANSFORMER

 * [A minimal skeleton for transformer source code](https://github.com/cliveseldon/seldon-core/tree/s2i/wrappers/s2i/python/test/transformer-template-app)
 * [Example transformers](https://github.com/SeldonIO/seldon-core/tree/master/examples/routers)





