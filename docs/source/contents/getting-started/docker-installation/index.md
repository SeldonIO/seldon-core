# Docker Installation

## Preparation

 1. Git clone seldon-core-v2

  ```bash
  git clone https://github.com/SeldonIO/seldon-core-v2
  ```
 3. Build [Seldon CLI](../cli.md)
 4. Install [Docker Compose](https://docs.docker.com/compose/install/).
 5. Install `make`.


## Deploy

From the project root run:

```
make deploy-local
```

## Example Use

Try out some [simple examples](examples.md).

## Undeploy

From the project root run:

```
make undeploy-local
```

