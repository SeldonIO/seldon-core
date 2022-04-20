# Docker Installation

## Preparation

 1. Git clone seldon-core-v2
        git clone https://github.com/SeldonIO/seldon-core-v2
 2. Build [Seldon CLI](../cli.md)
 3. Install [Docker Compose](https://docs.docker.com/compose/install/).
 4. Install `make`.


## Deploy

From the project root run:

```
make deploy-local
```

### Local Models

To deploy with a local folder available for loading models set the enviroment variable `LOCAL_MODEL_FOLDER` to the folder, e.g.:

```bash
export LOCAL_MODEL_FOLDER=/home/seldon/models
make deploy-local
```

This folder will be mounted at `/mnt/models`. You can then specify models as shown below:

```{literalinclude} ../../../../../samples/models/sklearn-iris-local.yaml 
:language: yaml
```

If you have set the local model folder as above then this would be looking at `/home/seldon/models/mlserver/iris`.

## Undeploy

From the project root run:

```
make undeploy-local
```


```{toctree}
:maxdepth: 1
:hidden:

```
