# Content

Seldon-core-examples repository provides out-of-the-box machine learning models examples to deploy using [seldon-core](https://github.com/SeldonIO/seldon-core). Since seldon-core deploys dockerized versions of your models, the repository also includes wrapping scripts that allow you to create docker images of such models which are deployable with seldon-core.

## Wrapping scripts

The repository contains two wrapping scripts at the moment
* wrap-model-in-host : If you are using docker on your machine, this script will build a docker image of your model locally.
* wrap-model-in-minikube: If you are using minikube, this script will build a docker image of your model directly on your minikube cluster (for usage see [seldon-core docs](https://github.com/SeldonIO/seldon-core/blob/master/docs/wrappers/readme.md)).

## Examples

The examples in the "models" folder are out-of-the-box machine learning models packaged as required by seldon wrappers. Each model folder usually includes a script to create and save the model, a model python file and a requirements file.
As an example, we describe the content of the folder  "models/sklearn_iris". Check out [seldon wrappers guidelines](https://github.com/SeldonIO/seldon-core/blob/master/docs/wrappers/readme.md)) for more details about packaging models.

* train_iris.py : Script to train and save a sklearn iris classifier
* IrisClassifier.py : The file used by seldon-wrappers to load and serve your saved model.
* requirements.txt : A list of packages required by your model
* sklearn_iris_deployment.json : A configuration json file used to deploy your model in  [seldon-core](https://github.com/SeldonIO/seldon-core#quick-start).