# Graph Level Metadata testing grounds

## Test Models

Start by spinning up test models with
```bash
docker-compse up --build
```

This will start few docker containers with models defined with different metadata at ports:
* 9000 <- default model
* 9001 <- chain node 1 - Model 1
* 9002 <- chain node 2 - Model 2
* 9010 <- Model Combiner: combines outputs of Model A1 and Model A2
* 9011 <- Model A1
* 9012 <- Model A2

You can also use `run_docker_compose` Makefile's target.


## Start Executor (default single node scenario)

Now start executor with
```bash
make run_executor
```

and request metadata with
```bash
make metadata
```

You can also request predictions with
```bash
make request
```


## Start Executor (special scenarios)

As nature of graph-level metadata is to prepare a global metadata that combines information of all models present in the graph I prepared few different testing scenarios.

Each scenario uses models provided by the same `docker-compose.yml` file.
To start each scenario use a specific Makefile's target or choose a corresponding `yaml` definition and provide it as `GRAPH` environmental variable to `run_executor` targer.

Request `metadata` as before:
```bash
make metadata
```


### Chain of models
```bash
make run_executor_chain
```
or
```bash
GRAPH=chain.yaml make run_executor
```

Check component's metadata with
```bash
curl -s http://localhost:9001/metadata | jq
curl -s http://localhost:9002/metadata | jq
```


### Combiner of models
```bash
make run_executor_combiner
```
or
```bash
GRAPH=combiner.yaml make run_executor
```

Check component's metadata with
```bash
curl -s http://localhost:9010/metadata | jq
curl -s http://localhost:9011/metadata | jq
curl -s http://localhost:9012/metadata | jq
```
