# Docker Installation

## Preparation

 1. `git clone https://github.com/SeldonIO/seldon-core --branch=v2`
 2. Build [Seldon CLI](../cli.md)
 3. Install [Docker Compose](https://docs.docker.com/compose/install/) (or directly from GitHub [release](https://github.com/docker/compose#linux) if not using Docker Desktop).
 4. Install `make`. This will depend on your version of Linux, for example on Ubuntu run `sudo apt-get install build-essential`.


## Deploy

From the project root run:

```
make deploy-local
```

This will run with `latest` images for the components.

Note: Triton and MLServer are large images at present (11G and 9G respectively) so will take time to download on first usage.

### Run a particular version

To run a particular release set the environment variable `CUSTOM_IMAGE_TAG` to the desired version before running the command, e.g.:

```
export CUSTOM_IMAGE_TAG=0.2.0
make deploy-local
```

### GPU support

To enable GPU on servers:

1. Make sure that `nvidia-container-runtime` is installed, follow [link](https://docs.docker.com/config/containers/resource_constraints/#gpu)
2. Enable GPU: `export GPU_ENABLED=1`


### Local Models

To deploy with a local folder available for loading models set the environment variable `LOCAL_MODEL_FOLDER` to the folder, e.g.:

```bash
export LOCAL_MODEL_FOLDER=/home/seldon/models
make deploy-local
```

This folder will be mounted at `/mnt/models`. You can then specify models as shown below:

```{literalinclude} ../../../../../samples/models/sklearn-iris-local.yaml
:language: yaml
```

If you have set the local model folder as above then this would be looking at `/home/seldon/models/iris`.

## Tracing

The default local install will provide Jaeger tracing at `http://0.0.0.0:16686/search`.

## Metrics

The default local install will expose Grafana at `http://localhost:3000`.

## Undeploy

From the project root run:

```
make undeploy-local
```


```{toctree}
:maxdepth: 1
:hidden:

```
