# Packaging an R model for Seldon Core using Docker

In this guide, we illustrate the steps needed to wrap your own R model in a docker image ready for deployment with Seldon Core using Docker.

## Step 1 - Create your source code

You will need an R file which provides an S3 class for your model via an `initialise_seldon` function and that has appropriate generics for your component, e.g. predict for a model. You will also need to declare any dependencies needed by your code.

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

### Dependencies

For running your code outside of docker you can populate an `install.R` with any software dependencies your code requires. For example:

```R
install.packages('rpart')
```

These same dependencies will need to be installed in the docker image, as explained in the next section.

## Step 2 - Build your image

How you install your dependencies in your docker image depends on the [base image that you choose](https://www.r-bloggers.com/2019/01/docker-images-for-r-r-base-versus-r-apt/) and whether binary versions of the dependencies are available. Using `rocker/r-apt:bionic` as a base image and install dependencies as binaries, if possible, results in a faster and smaller build.

An example docker file can be seen in the [seldon kubeflow example](https://github.com/kubeflow/example-seldon/blob/master/models/r_mnist/runtime/Dockerfile):

```dockerfile
FROM rocker/r-apt:bionic

RUN apt-get update && \
    apt-get install -y -qq \
    	r-cran-plumber \
    	r-cran-jsonlite \
    	r-cran-optparse \
    	r-cran-stringr \
    	r-cran-urltools \
    	r-cran-caret \
    	r-cran-pls \
    	curl

ENV MODEL_NAME mnist.R
ENV API_TYPE REST
ENV SERVICE_TYPE MODEL
ENV PERSISTENCE 0

RUN mkdir microservice
COPY . /microservice
WORKDIR /microservice

RUN curl -OL https://raw.githubusercontent.com/SeldonIO/seldon-core/v0.5.0/incubating/wrappers/s2i/R/microservice.R > /microservice/microservice.R

EXPOSE 5000

CMD Rscript microservice.R --model $MODEL_NAME --api $API_TYPE --service $SERVICE_TYPE --persistence $PERSISTENCE
```

Here binary versions of libraries are installed at the top. The dependencies 'plumber', 'jsonlite', 'optparse', 'stringr', 'urltools' and 'caret' are all required for the seldon wrapper - beyond these you can add your own dependencies ('pls' here is part of the example and not needed by the wrapper).

Then environment variables are set which will be passed as parameters into the R microservice in CMD at the end. The meaning of the environment variables is explained below.

A directory is created and the local source code is coped into the directory, which is then set as the working directory. The seldon microservice wrapper file is then copied into this directory. This wraps the model to run as a seldon microservice. The expose command sets 5000 as the port for the service.

The image can then be built with `docker build . -t $ORG/$MODEL_NAME:$TAG` to create your Docker image from source code. A simple name can be used but convention is to use the ORG/IMAGE:TAG format.

## Reference

### Environment Variables

The required environment variables understood by the builder image are explained below. You can provide them in the Dockerfile or as `-e` parameters to `docker run`.

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
- [A minimal skeleton for router source code](https://github.com/seldonio/seldon-core/tree/master/incubating/wrappers/s2i/R/test/router-template-app)

#### TRANSFORMER

- [A minimal skeleton for transformer source code](https://github.com/seldonio/seldon-core/tree/master/incubating/wrappers/s2i/R/test/transformer-template-app)
