# Multimodel serving with over-commit example

Note: this notebook requires access to internal services, so either `make start-all-host` (under `scheduler` sub directory) or expose the relevant ports.

## `iris` model on `MLServer`

```python
%env INFER_ENDPOINT=0.0.0.0:9000
%env SCHEDULER_ENDPOINT=0.0.0.0:9004
%env MLSERVER_DEBUG=0.0.0.0:7777
%env TRITON_DEBUG=0.0.0.0:7778
```

```
env: INFER_ENDPOINT=0.0.0.0:9000
env: SCHEDULER_ENDPOINT=0.0.0.0:9004
env: MLSERVER_DEBUG=0.0.0.0:7777
env: TRITON_DEBUG=0.0.0.0:7778

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
-import-path ../apis \
-proto ../apis/mlops/scheduler/scheduler.proto "$SCHEDULER_ENDPOINT" seldon.mlops.scheduler.Scheduler/LoadModel

sleep 0.05
done

```

```
loading model iris1
{

}
loading model iris2
{

}
loading model iris3
{

}
loading model iris4
{

}
loading model iris5
{

}
loading model iris6
{

}
loading model iris7
{

}
loading model iris8
{

}
loading model iris9
{

}
loading model iris10
{

}
loading model iris11
{

}

```

Get the list of models on this mlserver replica and whether they are loaded in main memory

```bash
%%bash
grpcurl -d '{}' \
         -plaintext \
         -import-path ../apis/ \
         -proto ../apis/mlops/agent_debug/agent_debug.proto  ${MLSERVER_DEBUG} seldon.mlops.agent_debug.AgentDebugService/ReplicaStatus

```

```json
{
  "models": [
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:45:40Z",
      "name": "iris10_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:44:48Z",
      "name": "iris3_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:44:48Z",
      "name": "iris6_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:44:48Z",
      "name": "iris1_1"
    },
    {
      "lastAccessed": "0001-01-01T00:00:00Z",
      "name": "iris11_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:44:48Z",
      "name": "iris9_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:44:48Z",
      "name": "iris7_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:44:48Z",
      "name": "iris8_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:44:48Z",
      "name": "iris5_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:44:48Z",
      "name": "iris4_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:44:48Z",
      "name": "iris2_1"
    }
  ]
}

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

```
All succeeded

```

Doing inference for all models succeeds as swapping in and out models is handled automatically

```bash
%%bash
for i in {1..10};
do

for j in {1..11};
do
curl -s -o /dev/null http://"$INFER_ENDPOINT"/v2/models/iris$j/infer -H "Content-Type: application/json"  \
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
         -import-path ../apis/ \
         -proto ../apis/mlops/scheduler/scheduler.proto "$SCHEDULER_ENDPOINT" seldon.mlops.scheduler.Scheduler/UnloadModel

done

```

```json
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}

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
-import-path ../apis \
-proto ../apis/mlops/scheduler/scheduler.proto "$SCHEDULER_ENDPOINT" seldon.mlops.scheduler.Scheduler/LoadModel

sleep 0.05
done

```

```
loading model tfsimple1
{

}
loading model tfsimple2
{

}
loading model tfsimple3
{

}
loading model tfsimple4
{

}
loading model tfsimple5
{

}
loading model tfsimple6
{

}
loading model tfsimple7
{

}
loading model tfsimple8
{

}
loading model tfsimple9
{

}
loading model tfsimple10
{

}
loading model tfsimple11
{

}
loading model tfsimple12
{

}
loading model tfsimple13
{

}
loading model tfsimple14
{

}
loading model tfsimple15
{

}
loading model tfsimple16
{

}
loading model tfsimple17
{

}
loading model tfsimple18
{

}
loading model tfsimple19
{

}
loading model tfsimple20
{

}
loading model tfsimple21
{

}
loading model tfsimple22
{

}
loading model tfsimple23
{

}
loading model tfsimple24
{

}
loading model tfsimple25
{

}
loading model tfsimple26
{

}
loading model tfsimple27
{

}
loading model tfsimple28
{

}
loading model tfsimple29
{

}
loading model tfsimple30
{

}
loading model tfsimple31
{

}
loading model tfsimple32
{

}
loading model tfsimple33
{

}
loading model tfsimple34
{

}
loading model tfsimple35
{

}
loading model tfsimple36
{

}
loading model tfsimple37
{

}
loading model tfsimple38
{

}
loading model tfsimple39
{

}
loading model tfsimple40
{

}
loading model tfsimple41
{

}
loading model tfsimple42
{

}
loading model tfsimple43
{

}
loading model tfsimple44
{

}
loading model tfsimple45
{

}
loading model tfsimple46
{

}
loading model tfsimple47
{

}
loading model tfsimple48
{

}
loading model tfsimple49
{

}
loading model tfsimple50
{

}
loading model tfsimple51
{

}
loading model tfsimple52
{

}
loading model tfsimple53
{

}
loading model tfsimple54
{

}
loading model tfsimple55
{

}
loading model tfsimple56
{

}
loading model tfsimple57
{

}
loading model tfsimple58
{

}
loading model tfsimple59
{

}
loading model tfsimple60
{

}
loading model tfsimple61
{

}
loading model tfsimple62
{

}
loading model tfsimple63
{

}
loading model tfsimple64
{

}
loading model tfsimple65
{

}
loading model tfsimple66
{

}
loading model tfsimple67
{

}
loading model tfsimple68
{

}
loading model tfsimple69
{

}
loading model tfsimple70
{

}
loading model tfsimple71
{

}
loading model tfsimple72
{

}
loading model tfsimple73
{

}
loading model tfsimple74
{

}
loading model tfsimple75
{

}
loading model tfsimple76
{

}
loading model tfsimple77
{

}
loading model tfsimple78
{

}
loading model tfsimple79
{

}
loading model tfsimple80
{

}
loading model tfsimple81
{

}
loading model tfsimple82
{

}
loading model tfsimple83
{

}
loading model tfsimple84
{

}
loading model tfsimple85
{

}
loading model tfsimple86
{

}
loading model tfsimple87
{

}
loading model tfsimple88
{

}
loading model tfsimple89
{

}
loading model tfsimple90
{

}
loading model tfsimple91
{

}
loading model tfsimple92
{

}
loading model tfsimple93
{

}
loading model tfsimple94
{

}
loading model tfsimple95
{

}
loading model tfsimple96
{

}
loading model tfsimple97
{

}
loading model tfsimple98
{

}
loading model tfsimple99
{

}
loading model tfsimple100
{

}
loading model tfsimple101
{

}
loading model tfsimple102
{

}
loading model tfsimple103
{

}
loading model tfsimple104
{

}
loading model tfsimple105
{

}
loading model tfsimple106
{

}
loading model tfsimple107
{

}
loading model tfsimple108
{

}
loading model tfsimple109
{

}
loading model tfsimple110
{

}

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

```
All succeeded

```

```bash
%%bash
grpcurl -d '{}' \
         -plaintext \
         -import-path ../apis/ \
         -proto ../apis/mlops/agent_debug/agent_debug.proto  ${TRITON_DEBUG} seldon.mlops.agent_debug.AgentDebugService/ReplicaStatus

```

```json
{
  "models": [
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:28Z",
      "name": "tfsimple89_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:26Z",
      "name": "tfsimple34_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:27Z",
      "name": "tfsimple59_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:28Z",
      "name": "tfsimple76_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:27Z",
      "name": "tfsimple45_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:27Z",
      "name": "tfsimple60_1"
    },
    {
      "lastAccessed": "0001-01-01T00:00:00Z",
      "name": "tfsimple8_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:26Z",
      "name": "tfsimple19_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:26Z",
      "name": "tfsimple31_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:28Z",
      "name": "tfsimple85_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:29Z",
      "name": "tfsimple107_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:26Z",
      "name": "tfsimple27_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:27Z",
      "name": "tfsimple51_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:28Z",
      "name": "tfsimple91_1"
    },
    {
      "lastAccessed": "0001-01-01T00:00:00Z",
      "name": "tfsimple2_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:26Z",
      "name": "tfsimple24_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:27Z",
      "name": "tfsimple53_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:29Z",
      "name": "tfsimple98_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:26Z",
      "name": "tfsimple17_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:26Z",
      "name": "tfsimple35_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:26Z",
      "name": "tfsimple38_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:26Z",
      "name": "tfsimple12_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:28Z",
      "name": "tfsimple83_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:29Z",
      "name": "tfsimple99_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:26Z",
      "name": "tfsimple20_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:27Z",
      "name": "tfsimple58_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:29Z",
      "name": "tfsimple101_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:26Z",
      "name": "tfsimple37_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:27Z",
      "name": "tfsimple63_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:28Z",
      "name": "tfsimple87_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:27Z",
      "name": "tfsimple40_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:28Z",
      "name": "tfsimple72_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:29Z",
      "name": "tfsimple94_1"
    },
    {
      "lastAccessed": "0001-01-01T00:00:00Z",
      "name": "tfsimple4_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:28Z",
      "name": "tfsimple80_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:29Z",
      "name": "tfsimple95_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:26Z",
      "name": "tfsimple26_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:27Z",
      "name": "tfsimple48_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:27Z",
      "name": "tfsimple55_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:28Z",
      "name": "tfsimple75_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:29Z",
      "name": "tfsimple100_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:27Z",
      "name": "tfsimple41_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:27Z",
      "name": "tfsimple43_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:27Z",
      "name": "tfsimple56_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:28Z",
      "name": "tfsimple73_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:28Z",
      "name": "tfsimple74_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:26Z",
      "name": "tfsimple21_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:27Z",
      "name": "tfsimple42_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:28Z",
      "name": "tfsimple82_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:28Z",
      "name": "tfsimple86_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:29Z",
      "name": "tfsimple97_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:29Z",
      "name": "tfsimple106_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:29Z",
      "name": "tfsimple108_1"
    },
    {
      "lastAccessed": "0001-01-01T00:00:00Z",
      "name": "tfsimple10_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:26Z",
      "name": "tfsimple15_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:28Z",
      "name": "tfsimple67_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:29Z",
      "name": "tfsimple105_1"
    },
    {
      "lastAccessed": "0001-01-01T00:00:00Z",
      "name": "tfsimple7_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:26Z",
      "name": "tfsimple22_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:27Z",
      "name": "tfsimple61_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:26Z",
      "name": "tfsimple25_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:27Z",
      "name": "tfsimple54_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:27Z",
      "name": "tfsimple64_1"
    },
    {
      "lastAccessed": "0001-01-01T00:00:00Z",
      "name": "tfsimple6_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:28Z",
      "name": "tfsimple90_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:29Z",
      "name": "tfsimple96_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:26Z",
      "name": "tfsimple16_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:27Z",
      "name": "tfsimple62_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:28Z",
      "name": "tfsimple71_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:26Z",
      "name": "tfsimple28_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:29Z",
      "name": "tfsimple103_1"
    },
    {
      "lastAccessed": "0001-01-01T00:00:00Z",
      "name": "tfsimple5_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:28Z",
      "name": "tfsimple93_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:26Z",
      "name": "tfsimple14_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:28Z",
      "name": "tfsimple69_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:28Z",
      "name": "tfsimple84_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:26Z",
      "name": "tfsimple36_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:29Z",
      "name": "tfsimple104_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:28Z",
      "name": "tfsimple70_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:29Z",
      "name": "tfsimple110_1"
    },
    {
      "lastAccessed": "0001-01-01T00:00:00Z",
      "name": "tfsimple9_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:27Z",
      "name": "tfsimple57_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:27Z",
      "name": "tfsimple65_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:29Z",
      "name": "tfsimple109_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:26Z",
      "name": "tfsimple30_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:27Z",
      "name": "tfsimple66_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:28Z",
      "name": "tfsimple68_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:27Z",
      "name": "tfsimple39_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:27Z",
      "name": "tfsimple44_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:28Z",
      "name": "tfsimple88_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:26Z",
      "name": "tfsimple13_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:26Z",
      "name": "tfsimple23_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:26Z",
      "name": "tfsimple29_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:27Z",
      "name": "tfsimple50_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:28Z",
      "name": "tfsimple79_1"
    },
    {
      "lastAccessed": "0001-01-01T00:00:00Z",
      "name": "tfsimple3_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:26Z",
      "name": "tfsimple18_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:29Z",
      "name": "tfsimple102_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:28Z",
      "name": "tfsimple92_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:27Z",
      "name": "tfsimple49_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:28Z",
      "name": "tfsimple77_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:28Z",
      "name": "tfsimple78_1"
    },
    {
      "lastAccessed": "0001-01-01T00:00:00Z",
      "name": "tfsimple1_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:27Z",
      "name": "tfsimple47_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:28Z",
      "name": "tfsimple81_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:26Z",
      "name": "tfsimple11_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:26Z",
      "name": "tfsimple33_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:27Z",
      "name": "tfsimple46_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:27Z",
      "name": "tfsimple52_1"
    },
    {
      "state": "InMemory",
      "lastAccessed": "2023-01-20T18:48:26Z",
      "name": "tfsimple32_1"
    }
  ]
}

```

```bash
%%bash

for i in {1..110};
do

grpcurl -d '{"model": {"name" : "tfsimple'"$i"'"}}' \
         -plaintext \
         -import-path ../apis/ \
         -proto ../apis/mlops/scheduler/scheduler.proto "$SCHEDULER_ENDPOINT" seldon.mlops.scheduler.Scheduler/UnloadModel

done

```

```json
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}
{

}

```

```python

```
