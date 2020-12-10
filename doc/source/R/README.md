# Packaging an R model for Seldon Core using s2i (incubating)

In this guide, we illustrate the steps needed to wrap your own R model in a docker image ready for deployment with Seldon Core using [source-to-image app s2i](https://github.com/openshift/source-to-image). If you prefer to use plain Docker, see the [Docker instructions](r_wrapping_docker.md).

If you are not familiar with s2i you can read [general instructions on using s2i](../wrappers/s2i.md) and then follow the steps below.

## Step 1 - Install s2i

[Download and install s2i](https://github.com/openshift/source-to-image#installation)

- Prerequisites for using s2i are:
  - Docker
  - Git (if building from a remote git repo)

To check everything is working you can run

```bash
s2i usage seldonio/seldon-core-s2i-r:0.1
```

## Step 2 - Create your source code

To use our s2i builder image to package your R model you will need:

- An R file which provides an S3 class for your model via an `initialise_seldon` function and that has appropriate generics for your component, e.g. predict for a model.
- An optional install.R to be run to install any libraries needed
- .s2i/environment - model definitions used by the s2i builder to correctly wrap your model

We will go into detail for each of these steps:

### R Runtime Model file

Your source code should contain an R file which defines an S3 class for your model. For example, looking at our skeleton R model file at `incubating/wrappers/s2i/R/test/model-template-app/MyModel.R`:

```R
library(methods)

predict.mymodel <- function(mymodel,newdata=list()) {
  write("MyModel predict called", stdout())
  newdata
}


new_mymodel <- function() {
  structure(list(), class = "mymodel")
}


initialise_seldon <- function(params) {
  new_mymodel()
}
```

- A `seldon_initialise` function creates an S3 class for my model via a constructor `new_mymodel`. This will be called on startup and you can use this to load any parameters your model needs.
- A generic `predict` function is created for my model class. This will be called with a `newdata` field with the `data.frame` to be predicted.

There are similar templates for ROUTERS and TRANSFORMERS.

### install.R

Populate an `install.R` with any software dependencies your code requires. For example:

```R
install.packages('rpart')
```

### .s2i/environment

Define the core parameters needed by our R builder image to wrap your model. An example is:

```bash
MODEL_NAME=MyModel.R
API_TYPE=REST
SERVICE_TYPE=MODEL
PERSISTENCE=0
```

These values can also be provided or overridden on the command line when building the image.

## Step 3 - Build your image

Use `s2i build` to create your Docker image from source code. You will need Docker installed on the machine and optionally git if your source code is in a public git repo.

Using s2i you can build directly from a git repo or from a local source folder. See the [s2i docs](https://github.com/openshift/source-to-image/blob/master/docs/cli.md#s2i-build) for further details. The general format is:

```bash
s2i build <git-repo> seldonio/seldon-core-s2i-r:0.1 <my-image-name>
s2i build <src-folder> seldonio/seldon-core-s2i-r:0.1 <my-image-name>
```

An example invocation using the test template model inside seldon-core:

```bash
s2i build https://github.com/seldonio/seldon-core --context-dir=incubating/wrappers/s2i/R/test/model-template-app seldonio/seldon-core-s2i-r:0.1 seldon-core-template-model
```

The above s2i build invocation:

- uses the GitHub repo: https://github.com/seldonio/seldon-core and the directory `incubating/wrappers/s2i/R/test/model-template-app` inside that repo.
- uses the builder image `seldonio/seldon-core-s2i-r`
- creates a docker image `seldon-core-template-model`

For building from a local source folder, an example where we clone the seldon-core repo:

```bash
git clone https://github.com/seldonio/seldon-core
cd seldon-core
s2i build incubating/wrappers/s2i/R/test/model-template-app seldonio/seldon-core-s2i-r:0.1 seldon-core-template-model
```

For more help see:

```bash
s2i usage seldonio/seldon-core-s2i-r:0.1
s2i build --help
```

## Reference

### Environment Variables

The required environment variables understood by the builder image are explained below. You can provide them in the `.s2i/environment` file or on the `s2i build` command line.

#### MODEL_NAME

The name of the R file containing the model.

#### API_TYPE

API type to create. Can be REST only at present.

#### SERVICE_TYPE

The service type being created. Available options are:

- MODEL
- ROUTER
- TRANSFORMER

#### PERSISTENCE

Can only by 0 at present. In future, will allow the state of the component to be saved periodically.

### Creating different service types

#### MODEL

- [A minimal skeleton for model source code](https://github.com/SeldonIO/seldon-core/tree/master/incubating/wrappers/s2i/R/test/model-template-app)
- [Example models](../examples/notebooks.html)

#### ROUTER
- [A minimal skeleton for router source code](https://github.com/SeldonIO/seldon-core/tree/master/incubating/wrappers/s2i/R/test/router-template-app)

#### TRANSFORMER

- [A minimal skeleton for transformer source code](https://github.com/SeldonIO/seldon-core/tree/master/incubating/wrappers/s2i/R/test/transformer-template-app)
