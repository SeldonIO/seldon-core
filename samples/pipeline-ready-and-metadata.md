## Pipeline Readiness Check and Metdata Calls

Local example settings.

```python
%env INFER_REST_ENDPOINT=http://0.0.0.0:9000
%env INFER_GRPC_ENDPOINT=0.0.0.0:9000
%env SELDON_SCHEDULE_HOST=0.0.0.0:9004
```

```yaml
env: INFER_REST_ENDPOINT=http://0.0.0.0:9000
env: INFER_GRPC_ENDPOINT=0.0.0.0:9000
env: SELDON_SCHEDULE_HOST=0.0.0.0:9004

```

Remote k8s cluster example settings - change as neeed for your needs.

```python
#%env INFER_REST_ENDPOINT=http://172.19.255.1:80
#%env INFER_GRPC_ENDPOINT=172.19.255.1:80
#%env SELDON_SCHEDULE_HOST=172.19.255.2:9004
```

### Model Chain - Ready Check

We will check the readiness of the Pipeline after every change to model and pipeline.

```bash
cat ./pipelines/tfsimples.yaml
```

```yaml
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

```

```bash
curl -Ik ${INFER_REST_ENDPOINT}/v2/pipelines/tfsimples/ready
```

```

```

```bash
grpcurl -d '{"name":"tfsimples"}' \
    -plaintext \
    -import-path ../apis \
    -proto ../apis/mlops/v2_dataplane/v2_dataplane.proto \
    -rpc-header seldon-model:tfsimples.pipeline \
    ${INFER_GRPC_ENDPOINT} inference.GRPCInferenceService/ModelReady
```

```yaml
ERROR:
  Code: Unimplemented
  Message:

```

```bash
seldon pipeline load -f ./pipelines/tfsimples.yaml
seldon pipeline status tfsimples -w PipelineReady
```

```json
{}
{"pipelineName":"tfsimples","versions":[{"pipeline":{"name":"tfsimples","uid":"cdsa5mv20gbc73a16810","version":1,"steps":[{"name":"tfsimple1"},{"name":"tfsimple2","inputs":["tfsimple1.outputs"],"tensorMap":{"tfsimple1.outputs.OUTPUT0":"INPUT0","tfsimple1.outputs.OUTPUT1":"INPUT1"}}],"output":{"steps":["tfsimple2.outputs"]},"kubernetesMeta":{}},"state":{"pipelineVersion":1,"status":"PipelineReady","reason":"created pipeline","lastChangeTimestamp":"2022-11-19T09:33:16.449568356Z"}}]}

```

```bash
seldon pipeline status tfsimples | jq .versions[0].state.modelsReady
```

```
[1;30mnull[0m

```

```bash
curl -Ik ${INFER_REST_ENDPOINT}/v2/pipelines/tfsimples/ready
```

```

```

```bash
grpcurl -d '{"name":"tfsimples"}' \
    -plaintext \
    -import-path ../apis \
    -proto ../apis/mlops/v2_dataplane/v2_dataplane.proto \
    -rpc-header seldon-model:tfsimples.pipeline \
    ${INFER_GRPC_ENDPOINT} inference.GRPCInferenceService/ModelReady
```

```json
{

}

```

```bash
seldon model load -f ./models/tfsimple1.yaml
seldon model status tfsimple1 -w ModelAvailable
```

```json
{}
{}

```

```bash
curl -Ik ${INFER_REST_ENDPOINT}/v2/pipelines/tfsimples/ready
```

```

```

```bash
grpcurl -d '{"name":"tfsimples"}' \
    -plaintext \
    -import-path ../apis \
    -proto ../apis/mlops/v2_dataplane/v2_dataplane.proto \
    -rpc-header seldon-model:tfsimples.pipeline \
    ${INFER_GRPC_ENDPOINT} inference.GRPCInferenceService/ModelReady
```

```json
{

}

```

```bash
seldon model load -f ./models/tfsimple2.yaml
seldon model status tfsimple2 -w ModelAvailable | jq -M .
```

```json
{}
{}

```

```bash
curl -Ik ${INFER_REST_ENDPOINT}/v2/pipelines/tfsimples/ready
```

```

```

```bash
grpcurl -d '{"name":"tfsimples"}' \
    -plaintext \
    -import-path ../apis \
    -proto ../apis/mlops/v2_dataplane/v2_dataplane.proto \
    -rpc-header seldon-model:tfsimples.pipeline \
    ${INFER_GRPC_ENDPOINT} inference.GRPCInferenceService/ModelReady
```

```json
{
  "ready": true
}

```

```bash
seldon pipeline status tfsimples | jq .versions[0].state.modelsReady
```

```
[0;39mtrue[0m

```

```bash
seldon pipeline unload tfsimples
```

```json
{}

```

```bash
curl -Ik ${INFER_REST_ENDPOINT}/v2/pipelines/tfsimples/ready
```

```

```

```bash
grpcurl -d '{"name":"tfsimples"}' \
    -plaintext \
    -import-path ../apis \
    -proto ../apis/mlops/v2_dataplane/v2_dataplane.proto \
    -rpc-header seldon-model:tfsimples.pipeline \
    ${INFER_GRPC_ENDPOINT} inference.GRPCInferenceService/ModelReady
```

```yaml
ERROR:
  Code: Unimplemented
  Message:

```

Models will still be ready even though Pipeline terminated

```bash
seldon pipeline status tfsimples | jq .versions[0].state.modelsReady
```

```
[0;39mtrue[0m

```

```bash
seldon pipeline load -f ./pipelines/tfsimples.yaml
seldon pipeline status tfsimples -w PipelineReady
```

```json
{}
{"pipelineName":"tfsimples","versions":[{"pipeline":{"name":"tfsimples","uid":"cdsa5uv20gbc73a1681g","version":1,"steps":[{"name":"tfsimple1"},{"name":"tfsimple2","inputs":["tfsimple1.outputs"],"tensorMap":{"tfsimple1.outputs.OUTPUT0":"INPUT0","tfsimple1.outputs.OUTPUT1":"INPUT1"}}],"output":{"steps":["tfsimple2.outputs"]},"kubernetesMeta":{}},"state":{"pipelineVersion":1,"status":"PipelineReady","reason":"created pipeline","lastChangeTimestamp":"2022-11-19T09:33:47.581203463Z","modelsReady":true}}]}

```

```bash
curl -Ik ${INFER_REST_ENDPOINT}/v2/pipelines/tfsimples/ready
```

```

```

```bash
grpcurl -d '{"name":"tfsimples"}' \
    -plaintext \
    -import-path ../apis \
    -proto ../apis/mlops/v2_dataplane/v2_dataplane.proto \
    -rpc-header seldon-model:tfsimples.pipeline \
    ${INFER_GRPC_ENDPOINT} inference.GRPCInferenceService/ModelReady
```

```json
{
  "ready": true
}

```

```bash
seldon pipeline status tfsimples | jq .versions[0].state.modelsReady
```

```
[0;39mtrue[0m

```

```bash
seldon model unload tfsimple1
seldon model unload tfsimple2
```

```json
{}
{}

```

```bash
seldon pipeline status tfsimples | jq .versions[0].state.modelsReady
```

```
[1;30mnull[0m

```

```bash
seldon pipeline unload tfsimples
```

```json
{}

```

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

```
'172.19.255.1'

```

```bash
kubectl create -f ./pipelines/tfsimples.yaml -n ${NAMESPACE}
```

```
pipeline.mlops.seldon.io/tfsimples created

```

```bash
kubectl wait --for condition=ready --timeout=1s pipeline --all -n ${NAMESPACE}
```

```yaml
error: timed out waiting for the condition on pipelines/tfsimples

```

```bash
kubectl get pipeline tfsimples -o jsonpath='{.status.conditions[0]}' -n ${NAMESPACE}
```

```json
{"lastTransitionTime":"2022-11-14T10:25:31Z","status":"False","type":"ModelsReady"}

```

```bash
kubectl create -f ./models/tfsimple1.yaml -n ${NAMESPACE}
kubectl create -f ./models/tfsimple2.yaml -n ${NAMESPACE}
```

```
model.mlops.seldon.io/tfsimple1 created
model.mlops.seldon.io/tfsimple2 created

```

```bash
kubectl wait --for condition=ready --timeout=300s pipeline --all -n ${NAMESPACE}
```

```
pipeline.mlops.seldon.io/tfsimples condition met

```

```bash
kubectl get pipeline tfsimples -o jsonpath='{.status.conditions[0]}' -n ${NAMESPACE}
```

```json
{"lastTransitionTime":"2022-11-14T10:25:49Z","status":"True","type":"ModelsReady"}

```

```bash
kubectl delete -f ./models/tfsimple1.yaml -n ${NAMESPACE}
kubectl delete -f ./models/tfsimple2.yaml -n ${NAMESPACE}
kubectl delete -f ./pipelines/tfsimples.yaml -n ${NAMESPACE}
```

```
model.mlops.seldon.io "tfsimple1" deleted
model.mlops.seldon.io "tfsimple2" deleted
pipeline.mlops.seldon.io "tfsimples" deleted

```

```python

```
