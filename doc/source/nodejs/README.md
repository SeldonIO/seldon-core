# Packaging a NodeJS model for Seldon Core using s2i

In this guide, we illustrate the steps needed to wrap your own JS model running on a node engine in a docker image ready for deployment with Seldon Core using [source-to-image app s2i](https://github.com/openshift/source-to-image).

If you are not familiar with s2i you can read [general instructions on using s2i](../wrappers/s2i.md) and then follow the steps below.

## Step 1 - Install s2i

[Download and install s2i](https://github.com/openshift/source-to-image#installation)

- Prerequisites for using s2i are:
  - Docker
  - Git (if building from a remote git repo)

To check everything is working you can run

```bash
s2i usage seldonio/seldon-core-s2i-nodejs:0.1
```

## Step 2 - Create your source code

To use our s2i builder image to package your NodeJS model you will need:

- An JS file which provides an ES5 Function object or an ES6 class for your model and that has appropriate generics for your component, i.e. an `init` and a `predict` for the model.
- A package.json that contains all the dependencies and meta data for the model
- .s2i/environment - model definitions used by the s2i builder to correctly wrap your model

We will go into detail for each of these steps:

### NodeJS Runtime Model file

Your source code should which provides an ES5 Function object or an ES6 class for your model. For example, looking at our skeleton JS structure:

```js
let MyModel = function() {};

MyModel.prototype.init = async function() {
  // A mandatory init method for the class to load run-time dependencies
  this.model = "My Awesome model";
};

MyModel.prototype.predict = function(newdata, feature_names) {
  //A mandatory predict function for the model predictions
  console.log("Predicting ...");
  return newdata;
};

module.exports = MyModel;
```

Also the model could be an ES6 class as follows

```js
class MyModel {
  async init() {
    // A mandatory init method for the class to load run-time dependencies
    this.model = "My Awesome ES6 model";
  }
  predict(newdata, feature_names) {
    //A mandatory predict function for the model predictions
    console.log("ES6 Predicting ...");
    return newdata;
  }
}
module.exports = MyModel;
```

- A `init` method for the model object. This will be called on startup and you can use this to load any parameters your model needs. This function may also be an async,for example in case if it has to load the model weights from a remote location.
- A generic `predict` method is created for my model class. This will be called with a `newdata` field with the data object to be predicted.

### package.json

Populate an `package.json` with any software dependencies your code requires using an `npm init` command and save your dependencies to the file.

### .s2i/environment

Define the core parameters needed by our node JS builder image to wrap your model. An example is:

```bash
MODEL_NAME=MyModel.js
API_TYPE=REST
SERVICE_TYPE=MODEL
PERSISTENCE=0
```

These values can also be provided or overridden on the command line when building the image.

## Step 3 - Build your image

Use `s2i build` to create your Docker image from source code. You will need Docker installed on the machine and optionally git if your source code is in a public git repo.

Using s2i you can build directly from a git repo or from a local source folder. See the [s2i docs](https://github.com/openshift/source-to-image/blob/master/docs/cli.md#s2i-build) for further details. The general format is:

```bash
s2i build <git-repo> seldonio/seldon-core-s2i-nodejs:0.1 <my-image-name>
s2i build <src-folder> seldonio/seldon-core-s2i-nodejs:0.1 <my-image-name>
```

An example invocation using the test template model inside seldon-core:

```bash
s2i build https://github.com/seldonio/seldon-core.git --context-dir=incubating/wrappers/s2i/nodejs/test/model-template-app seldonio/seldon-core-s2i-nodejs:0.1 seldon-core-template-model
```

The above s2i build invocation:

- uses the GitHub repo: https://github.com/seldonio/seldon-core.git and the directory `incubating/wrappers/s2i/nodejs/test/model-template-app` inside that repo.
- uses the builder image `seldonio/seldon-core-s2i-nodejs`
- creates a docker image `seldon-core-template-model`

For building from a local source folder, an example where we clone the seldon-core repo:

```bash
git clone https://github.com/seldonio/seldon-core.git
cd seldon-core
s2i build incubating/wrappers/s2i/nodejs/test/model-template-app seldonio/seldon-core-s2i-nodejs:0.1 seldon-core-template-model
```

For more help see:

```bash
s2i usage seldonio/seldon-core-s2i-nodejs:0.1
s2i build --help
```

## Reference

### Environment Variables

The required environment variables understood by the builder image are explained below. You can provide them in the `.s2i/environment` file or on the `s2i build` command line.

### MODEL_NAME

The name of the JS file containing the model.

### API_TYPE

API type to create. Can be REST or GRPC.

### SERVICE_TYPE

The service type being created. Available options are:

- MODEL
- TRANSFORMER

### PERSISTENCE

Can only by 0 at present.

## Creating different service types

### MODEL

- [Example model](../examples/notebooks.html)
