## Pipeline Readiness Check and Metdata Calls


Local example settings.


```python
%env INFER_REST_ENDPOINT=http://0.0.0.0:9000
%env INFER_GRPC_ENDPOINT=0.0.0.0:9000
%env SELDON_SCHEDULE_HOST=0.0.0.0:9004
```

    env: INFER_REST_ENDPOINT=http://0.0.0.0:9000
    env: INFER_GRPC_ENDPOINT=0.0.0.0:9000
    env: SELDON_SCHEDULE_HOST=0.0.0.0:9004


Remote k8s cluster example settings - change as neeed for your needs.


```python
#%env INFER_REST_ENDPOINT=http://172.19.255.1:80
#%env INFER_GRPC_ENDPOINT=172.19.255.1:80
#%env SELDON_SCHEDULE_HOST=172.19.255.2:9004
```

### Model Chain - Ready Check

We will check the readiness of the Pipeline after every change to model and pipeline.


```python
!cat ./pipelines/tfsimples.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: tfsimples
    spec:
      steps:
        - name: tfsimple1
        - name: tfsimple2
          inputs:
          - tfsimple1
          tensorMap:
            tfsimple1.outputs.OUTPUT0: INPUT0
            tfsimple1.outputs.OUTPUT1: INPUT1
      output:
        steps:
        - tfsimple2



```python
!curl -Ik ${INFER_REST_ENDPOINT}/v2/pipelines/tfsimples/ready
```

    
    
    
    
    



```python
!grpcurl -d '{"name":"tfsimples"}' \
    -plaintext \
    -import-path ../apis \
    -proto ../apis/mlops/v2_dataplane/v2_dataplane.proto \
    -rpc-header seldon-model:tfsimples.pipeline \
    ${INFER_GRPC_ENDPOINT} inference.GRPCInferenceService/ModelReady
```

    ERROR:
      Code: Unimplemented
      Message: 



```python
!seldon pipeline load -f ./pipelines/tfsimples.yaml
!seldon pipeline status tfsimples -w PipelineReady
```

    {}
    {"pipelineName":"tfsimples","versions":[{"pipeline":{"name":"tfsimples","uid":"cdsa5mv20gbc73a16810","version":1,"steps":[{"name":"tfsimple1"},{"name":"tfsimple2","inputs":["tfsimple1.outputs"],"tensorMap":{"tfsimple1.outputs.OUTPUT0":"INPUT0","tfsimple1.outputs.OUTPUT1":"INPUT1"}}],"output":{"steps":["tfsimple2.outputs"]},"kubernetesMeta":{}},"state":{"pipelineVersion":1,"status":"PipelineReady","reason":"created pipeline","lastChangeTimestamp":"2022-11-19T09:33:16.449568356Z"}}]}



```python
!seldon pipeline status tfsimples | jq .versions[0].state.modelsReady
```

    [1;30mnull[0m



```python
!curl -Ik ${INFER_REST_ENDPOINT}/v2/pipelines/tfsimples/ready
```

    
    
    
    
    
    
    



```python
!grpcurl -d '{"name":"tfsimples"}' \
    -plaintext \
    -import-path ../apis \
    -proto ../apis/mlops/v2_dataplane/v2_dataplane.proto \
    -rpc-header seldon-model:tfsimples.pipeline \
    ${INFER_GRPC_ENDPOINT} inference.GRPCInferenceService/ModelReady
```

    {
      
    }



```python
!seldon model load -f ./models/tfsimple1.yaml 
!seldon model status tfsimple1 -w ModelAvailable 
```

    {}
    {}



```python
!curl -Ik ${INFER_REST_ENDPOINT}/v2/pipelines/tfsimples/ready
```

    
    
    
    
    
    
    



```python
!grpcurl -d '{"name":"tfsimples"}' \
    -plaintext \
    -import-path ../apis \
    -proto ../apis/mlops/v2_dataplane/v2_dataplane.proto \
    -rpc-header seldon-model:tfsimples.pipeline \
    ${INFER_GRPC_ENDPOINT} inference.GRPCInferenceService/ModelReady
```

    {
      
    }



```python
!seldon model load -f ./models/tfsimple2.yaml 
!seldon model status tfsimple2 -w ModelAvailable | jq -M .
```

    {}
    {}



```python
!curl -Ik ${INFER_REST_ENDPOINT}/v2/pipelines/tfsimples/ready
```

    
    
    
    
    
    
    



```python
!grpcurl -d '{"name":"tfsimples"}' \
    -plaintext \
    -import-path ../apis \
    -proto ../apis/mlops/v2_dataplane/v2_dataplane.proto \
    -rpc-header seldon-model:tfsimples.pipeline \
    ${INFER_GRPC_ENDPOINT} inference.GRPCInferenceService/ModelReady
```

    {
      "ready": true
    }



```python
!seldon pipeline status tfsimples | jq .versions[0].state.modelsReady
```

    [0;39mtrue[0m



```python
!seldon pipeline unload tfsimples
```

    {}



```python
!curl -Ik ${INFER_REST_ENDPOINT}/v2/pipelines/tfsimples/ready
```

    
    
    
    
    



```python
!grpcurl -d '{"name":"tfsimples"}' \
    -plaintext \
    -import-path ../apis \
    -proto ../apis/mlops/v2_dataplane/v2_dataplane.proto \
    -rpc-header seldon-model:tfsimples.pipeline \
    ${INFER_GRPC_ENDPOINT} inference.GRPCInferenceService/ModelReady
```

    ERROR:
      Code: Unimplemented
      Message: 


Models will still be ready even though Pipeline terminated


```python
!seldon pipeline status tfsimples | jq .versions[0].state.modelsReady
```

    [0;39mtrue[0m



```python
!seldon pipeline load -f ./pipelines/tfsimples.yaml
!seldon pipeline status tfsimples -w PipelineReady
```

    {}
    {"pipelineName":"tfsimples","versions":[{"pipeline":{"name":"tfsimples","uid":"cdsa5uv20gbc73a1681g","version":1,"steps":[{"name":"tfsimple1"},{"name":"tfsimple2","inputs":["tfsimple1.outputs"],"tensorMap":{"tfsimple1.outputs.OUTPUT0":"INPUT0","tfsimple1.outputs.OUTPUT1":"INPUT1"}}],"output":{"steps":["tfsimple2.outputs"]},"kubernetesMeta":{}},"state":{"pipelineVersion":1,"status":"PipelineReady","reason":"created pipeline","lastChangeTimestamp":"2022-11-19T09:33:47.581203463Z","modelsReady":true}}]}



```python
!curl -Ik ${INFER_REST_ENDPOINT}/v2/pipelines/tfsimples/ready
```

    
    
    
    
    
    
    



```python
!grpcurl -d '{"name":"tfsimples"}' \
    -plaintext \
    -import-path ../apis \
    -proto ../apis/mlops/v2_dataplane/v2_dataplane.proto \
    -rpc-header seldon-model:tfsimples.pipeline \
    ${INFER_GRPC_ENDPOINT} inference.GRPCInferenceService/ModelReady
```

    {
      "ready": true
    }



```python
!seldon pipeline status tfsimples | jq .versions[0].state.modelsReady
```

    [0;39mtrue[0m



```python
!seldon model unload tfsimple1
!seldon model unload tfsimple2
```

    {}
    {}



```python
!seldon pipeline status tfsimples | jq .versions[0].state.modelsReady
```

    [1;30mnull[0m



```python
!seldon pipeline unload tfsimples
```

    {}


### Kubernetes Resource Example


```python
import os
os.environ["NAMESPACE"] = "seldon-mesh"
```


```python
MESH_IP=!kubectl get svc seldon-mesh -n ${NAMESPACE} -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
MESH_IP=MESH_IP[0]
import os
os.environ['MESH_IP'] = MESH_IP
MESH_IP
```




    '172.19.255.1'




```python
!kubectl create -f ./pipelines/tfsimples.yaml -n ${NAMESPACE}
```

    pipeline.mlops.seldon.io/tfsimples created



```python
!kubectl wait --for condition=ready --timeout=1s pipeline --all -n ${NAMESPACE}
```

    error: timed out waiting for the condition on pipelines/tfsimples



```python
!kubectl get pipeline tfsimples -o jsonpath='{.status.conditions[0]}' -n ${NAMESPACE}
```

    {"lastTransitionTime":"2022-11-14T10:25:31Z","status":"False","type":"ModelsReady"}


```python
!kubectl create -f ./models/tfsimple1.yaml -n ${NAMESPACE}
!kubectl create -f ./models/tfsimple2.yaml -n ${NAMESPACE}
```

    model.mlops.seldon.io/tfsimple1 created
    model.mlops.seldon.io/tfsimple2 created



```python
!kubectl wait --for condition=ready --timeout=300s pipeline --all -n ${NAMESPACE}
```

    pipeline.mlops.seldon.io/tfsimples condition met



```python
!kubectl get pipeline tfsimples -o jsonpath='{.status.conditions[0]}' -n ${NAMESPACE}
```

    {"lastTransitionTime":"2022-11-14T10:25:49Z","status":"True","type":"ModelsReady"}


```python
!kubectl delete -f ./models/tfsimple1.yaml -n ${NAMESPACE}
!kubectl delete -f ./models/tfsimple2.yaml -n ${NAMESPACE}
!kubectl delete -f ./pipelines/tfsimples.yaml -n ${NAMESPACE}
```

    model.mlops.seldon.io "tfsimple1" deleted
    model.mlops.seldon.io "tfsimple2" deleted
    pipeline.mlops.seldon.io "tfsimples" deleted



```python

```
