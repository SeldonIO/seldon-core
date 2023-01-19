# Multimodel serving with over-commit example

Note: this notebook requires access to internal services, so either `make start-all-host` or expose the relevant ports.

## `iris` model on `MLServer`

```python
%env INFER_ENDPOINT=0.0.0.0:9000
%env SCHEDULER_ENDPOINT=0.0.0.0:9004
%env MLSERVER_DEBUG=0.0.0.0:7777
%env TRITON_DEBUG=0.0.0.0:7778
```

By default if running locally there is 1 replica of `mlserver` with memory slots up to 10MB and 20% overcommit budget. We will load 11 `iris` models each requiring 1MB worth of memory slots as an example. These numbers allow for 10 models to be active at the same time and 1 model to be evicted to disk.

```bash
%%bash
for i in {1..11};
do

echo "loading model iris$i"

data='{
        "model":{
            "meta": {"name":"iris'"$i"'"},
            "modelSpec" : {
                "uri":"gs://seldon-models/mlserver/iris",
                "requirements":["sklearn"],
                "memoryBytes":1000000},
            "deploymentSpec": {"replicas":1}
            }
      }'

grpcurl -d "$data" \
-plaintext \
-import-path ../../apis \
-proto ../../apis/mlops/scheduler/scheduler.proto "$SCHEDULER_ENDPOINT" seldon.mlops.scheduler.Scheduler/LoadModel

sleep 0.01
done

```

Get the list of models on this mlserver replica and whether they are loaded in main memory

```bash
%%bash
grpcurl -d '{}' \
         -plaintext \
         -import-path ../../apis/ \
         -proto ../../apis/mlops/agent_debug/agent_debug.proto  ${MLSERVER_DEBUG} seldon.mlops.agent_debug.AgentDebugService/ReplicaStatus

```

```bash
%%bash
for i in {1..11};
do

url=http://${INFER_ENDPOINT}/v2/models/iris${i}/infer
ret=`curl -s -o /dev/null -w "%{http_code}" "${url}" -H "Content-Type: application/json" \
        -d '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'`
if [ $ret -ne 200 ]; then
    echo "Failed with code ${ret}"
    exit
fi

done
echo "All succeeded"

```

Doing inference for all models succeeds as swapping in and out models is handled automatically

```bash
%%bash
for i in {1..10};
do

for j in {1..11};
do
curl http://"$INFER_ENDPOINT"/v2/models/iris$j/infer -H "Content-Type: application/json"  \
        -d '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' &
done

done

```

Unload models

```bash
%%bash

for i in {1..11};
do

grpcurl -d '{"model": {"name" : "iris'"$i"'"}}' \
         -plaintext \
         -import-path ../../apis/ \
         -proto ../../apis/mlops/scheduler/scheduler.proto "$SCHEDULER_ENDPOINT" seldon.mlops.scheduler.Scheduler/UnloadModel

done

```

## `tfsimple` model on `triton`

With `tfsimple` on `triton` we will reduce the memory slot required to 100KB, which will allow us to load at least 100 models on the server in memory. The remaining models (10) will have to be evicted.

```bash
%%bash
for i in {1..110};
do

echo "loading model tfsimple$i"

data='{
        "model":{
            "meta": {"name":"tfsimple'"$i"'"},
            "modelSpec" : {
                "uri":"gs://seldon-models/triton/simple",
                "requirements":["tensorflow"],
                "memoryBytes":100000},
            "deploymentSpec": {"replicas":1}
            }
      }'

grpcurl -d "$data" \
-plaintext \
-import-path ../../apis \
-proto ../../apis/mlops/scheduler/scheduler.proto "$SCHEDULER_ENDPOINT" seldon.mlops.scheduler.Scheduler/LoadModel

sleep 0.01
done

```

```bash
%%bash
for i in {1..110};
do

url=http://${INFER_ENDPOINT}/v2/models/tfsimple${i}/infer
ret=`curl -s -o /dev/null -w "%{http_code}" curl -s -o /dev/null -w "%{http_code}" "${url}" -H "Content-Type: application/json" \
        -d '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}'`
if [ $ret -ne 200 ]; then
    echo "Failed with code ${ret}"
    exit
fi
done
echo "All succeeded"

```

```bash
%%bash
grpcurl -d '{}' \
         -plaintext \
         -import-path ../../apis/ \
         -proto ../../apis/mlops/agent_debug/agent_debug.proto  ${TRITON_DEBUG} seldon.mlops.agent_debug.AgentDebugService/ReplicaStatus

```

```bash
%%bash

for i in {1..110};
do

grpcurl -d '{"model": {"name" : "tfsimple'"$i"'"}}' \
         -plaintext \
         -import-path ../../apis/ \
         -proto ../../apis/mlops/scheduler/scheduler.proto "$SCHEDULER_ENDPOINT" seldon.mlops.scheduler.Scheduler/UnloadModel

done

```

```python

```
