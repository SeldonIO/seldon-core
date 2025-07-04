# Runtime Metrics / Tags Example


## Prerequisites

 * Kind cluster with Seldon Installed
 * curl
 * s2i
 * seldon-core-analytics

 
## Setup Seldon Core

Use the setup notebook to [Setup Cluster](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html) to setup Seldon Core with an ingress.


```python
!kubectl create namespace seldon
```

    Error from server (AlreadyExists): namespaces "seldon" already exists



```python
!kubectl config set-context $(kubectl config current-context) --namespace=seldon
```

    Context "kind-kind" modified.


## Install Seldon Core Analytics


```python
!helm install seldon-core-analytics ../../../helm-charts/seldon-core-analytics -n seldon-system --wait
```

    Error: cannot re-use a name that is still in use


## Define Model


```python
%%writefile Model.py
import logging

from seldon_core.user_model import SeldonResponse


def reshape(x):
    if len(x.shape) < 2:
        return x.reshape(1, -1)
    else:
        return x


class Model:
    def predict(self, features, names=[], meta={}):
        X = reshape(features)

        logging.info(f"model features: {features}")
        logging.info(f"model names: {names}")
        logging.info(f"model meta: {meta}")

        logging.info(f"model X: {X}")

        runtime_metrics = [{"type": "COUNTER", "key": "instance_counter", "value": len(X)}]
        runtime_tags = {"runtime": "tag", "shared": "right one"}
        return SeldonResponse(data=X, metrics=runtime_metrics, tags=runtime_tags)

    def metrics(self):
        return [{"type": "COUNTER", "key": "requests_counter", "value": 1}]

    def tags(self):
        return {"static": "tag", "shared": "not right one"}      
```

    Overwriting Model.py


## Build Image and load into kind cluster


```bash
%%bash
s2i build -E ENVIRONMENT_REST . seldonio/seldon-core-s2i-python37-ubi8:1.7.0-dev runtime-metrics-tags:0.1
kind load docker-image runtime-metrics-tags:0.1
```

    ---> Installing application source...
    Collecting pip-licenses
    Downloading https://files.pythonhosted.org/packages/c5/50/6c4b4e69a0c43bd9f03a30579695093062ba72da4e3e4026cd2144dbcc71/pip_licenses-2.3.0-py3-none-any.whl
    Collecting PTable (from pip-licenses)
    Downloading https://files.pythonhosted.org/packages/ab/b3/b54301811173ca94119eb474634f120a49cd370f257d1aae5a4abaf12729/PTable-0.9.2.tar.gz
    Building wheels for collected packages: PTable
    Building wheel for PTable (setup.py): started
    Building wheel for PTable (setup.py): finished with status 'done'
    Created wheel for PTable: filename=PTable-0.9.2-cp37-none-any.whl size=22906 sha256=fe30596e3606620d3cfba1d38ee16568d716eebc86368394bfaf62cbe9a905c3
    Stored in directory: /root/.cache/pip/wheels/22/cc/2e/55980bfe86393df3e9896146a01f6802978d09d7ebcba5ea56
    Successfully built PTable
    Installing collected packages: PTable, pip-licenses
    Successfully installed PTable-0.9.2 pip-licenses-2.3.0
    created path: ./licenses/license_info.csv
    created path: ./licenses/license.txt
    Build completed successfully
    Image: "runtime-metrics-tags:0.1" with ID "sha256:75b9a64cf21c3ae335eb62fadf76d9841b057b899fdf2778833cdba5e26295f8" not yet present on node "kind-control-plane", loading...


## Deploy Model


```python
%%writefile deployment.yaml

apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: seldon-model-runtime-data
spec:
  name: test-deployment
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: runtime-metrics-tags:0.1
          name: my-model
    graph:
      name: my-model
      type: MODEL
    name: example
    replicas: 1
```

    Overwriting deployment.yaml



```python
!kubectl apply -f deployment.yaml
```

    seldondeployment.machinelearning.seldon.io/seldon-model-runtime-data created



```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=seldon-model-runtime-data -o jsonpath='{.items[0].metadata.name}')
```

    Waiting for deployment "seldon-model-runtime-data-example-0-my-model" rollout to finish: 0 of 1 updated replicas are available...
    deployment "seldon-model-runtime-data-example-0-my-model" successfully rolled out


## Send few inference requests


```bash
%%bash
curl -s -H 'Content-Type: application/json' -d '{"data": {"ndarray": [[1, 2, 3]]}}' \
    http://localhost:8003/seldon/seldon/seldon-model-runtime-data/api/v1.0/predictions
```

    {"data":{"names":["t:0","t:1","t:2"],"ndarray":[[1,2,3]]},"meta":{"metrics":[{"key":"requests_counter","type":"COUNTER","value":1},{"key":"instance_counter","type":"COUNTER","value":1}],"tags":{"runtime":"tag","shared":"right one","static":"tag"}}}



```bash
%%bash
curl -s -H 'Content-Type: application/json' -d '{"data": {"ndarray": [[1, 2, 3], [4, 5, 6]]}}' \
    http://localhost:8003/seldon/seldon/seldon-model-runtime-data/api/v1.0/predictions
```

    {"data":{"names":["t:0","t:1","t:2"],"ndarray":[[1,2,3],[4,5,6]]},"meta":{"metrics":[{"key":"requests_counter","type":"COUNTER","value":1},{"key":"instance_counter","type":"COUNTER","value":2}],"tags":{"runtime":"tag","shared":"right one","static":"tag"}}}


## Check metrics


```python
import json
```


```python
metrics =! kubectl run --quiet=true -it --rm curlmetrics --image=radial/busyboxplus:curl --restart=Never -- \
    curl -s seldon-core-analytics-prometheus-seldon.seldon-system/api/v1/query?query=instance_counter_total

json.loads(metrics[0])["data"]["result"][0]["value"][1]
```




    '3'




```python
metrics =! kubectl run --quiet=true -it --rm curlmetrics --image=radial/busyboxplus:curl --restart=Never -- \
    curl -s seldon-core-analytics-prometheus-seldon.seldon-system/api/v1/query?query=requests_counter_total

json.loads(metrics[0])["data"]["result"][0]["value"][1]
```




    '2'



## Cleanup


```python
!kubectl delete -f deployment.yaml
```

    seldondeployment.machinelearning.seldon.io "seldon-model-runtime-data" deleted



```python
!helm delete seldon-core-analytics --namespace seldon-system
```

    release "seldon-core-analytics" uninstalled

