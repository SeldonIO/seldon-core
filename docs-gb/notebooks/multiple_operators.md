# Multiple Seldon Core Operators

This notebook illustrate how multiple Seldon Core Operators can share the same cluster. In particular:

  * A Namespaced Operator that only manages Seldon Deployments inside its namespace. Only needs Role RBAC and Namespace labeled with `seldon.io/controller-id`
  * A Clusterwide Operator that manges SeldonDeployment with a matching `seldon.io/controller-id` label.
  * A Clusterwide Operator that manages Seldon Deployments not handled by the above.

## Setup Seldon Core

Use the setup notebook to [Setup Cluster](https://docs.seldon.ai/seldon-core-1/tutorials/notebooks/seldon-core-setup#setup-cluster) with [Ambassador Ingress](https://docs.seldon.ai/seldon-core-1/tutorials/notebooks/seldon-core-setup#install-ingress).

## Namespaced Seldon Core Operator


```python
!kubectl create namespace seldon-ns1
```

    namespace/seldon-ns1 created



```python
!kubectl label namespace seldon-ns1 seldon.io/controller-id=seldon-ns1
```

    namespace/seldon-ns1 labeled



```python
VERSION=!cat ../version.txt
VERSION=VERSION[0]
VERSION
```




    '1.19.0-dev'




```python
from IPython.core.magic import register_line_cell_magic


@register_line_cell_magic
def writetemplate(line, cell):
    with open(line, "w") as f:
        f.write(cell.format(**globals()))
```


```python
!helm install seldon-namespaced  ../helm-charts/seldon-core-operator  \
    --set singleNamespace=true \
    --set image.pullPolicy=IfNotPresent \
    --set usageMetrics.enabled=false \
    --set crd.create=false \
    --namespace seldon-ns1 \
    --wait
```

    NAME: seldon-namespaced
    LAST DEPLOYED: Thu Dec  4 10:08:54 2025
    NAMESPACE: seldon-ns1
    STATUS: deployed
    REVISION: 1
    TEST SUITE: None



```python
!kubectl rollout status deployment/seldon-controller-manager -n seldon-ns1
```

    deployment "seldon-controller-manager" successfully rolled out



```python
%%writetemplate resources/model.yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: seldon-model
spec:
  name: test-deployment
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier:{VERSION}
          name: classifier
    graph:
      name: classifier
      type: MODEL
      endpoint:
        type: REST
    name: example
    replicas: 1
```


```python
!kubectl create -f resources/model.yaml -n seldon-ns1
```

    seldondeployment.machinelearning.seldon.io/seldon-model created



```python
!kubectl wait sdep/seldon-model \
  --for=condition=ready \
  --timeout=120s \
  -n seldon-ns1
```

    seldondeployment.machinelearning.seldon.io/seldon-model condition met



```python
NAME = !kubectl get sdep -n seldon-ns1 -o jsonpath='{.items[0].metadata.name}'
assert NAME[0] == "seldon-model"
```


```python
!kubectl config set-context $(kubectl config current-context) --namespace=seldon-ns1
```

    Context "kind-kind" modified.



```python
!kubectl delete -f resources/model.yaml -n seldon-ns1
```

    seldondeployment.machinelearning.seldon.io "seldon-model" deleted



```python
!helm delete seldon-namespaced
```

    release "seldon-namespaced" uninstalled


## Label Focused Seldon Core Operator

 * We set `crd.create=false` as the CRD already exists in the cluster.
 * We set `controllerId=seldon-id1`. SeldonDeployments with this label will be managed.


```python
!kubectl create namespace seldon-id1
```

    namespace/seldon-id1 created



```python
!helm install seldon-controllerid  ../helm-charts/seldon-core-operator  \
    --set singleNamespace=false \
    --set image.pullPolicy=IfNotPresent \
    --set usageMetrics.enabled=false \
    --set crd.create=false \
    --set controllerId=seldon-id1 \
    --namespace seldon-id1 \
    --wait
```

    NAME: seldon-controllerid
    LAST DEPLOYED: Thu Dec  4 10:13:08 2025
    NAMESPACE: seldon-id1
    STATUS: deployed
    REVISION: 1
    TEST SUITE: None



```python
!kubectl rollout status deployment/seldon-controller-manager -n seldon-id1
```

    deployment "seldon-controller-manager" successfully rolled out



```python
%%writetemplate resources/model_controller_id.yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  labels:
    app: seldon
    seldon.io/controller-id: seldon-id1
  name: test-c1
spec:
  name: test-c1
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier:{VERSION}
          imagePullPolicy: IfNotPresent
          name: classifier
          resources:
            requests:
              memory: 1Mi
        terminationGracePeriodSeconds: 1
    graph:
      children: []
      endpoint:
        type: REST
      name: classifier
      type: MODEL
    labels:
      version: v1
    name: example
    replicas: 1
```


```python
!pygmentize resources/model_controller_id.yaml
```

    [94mapiVersion[39;49;00m:[37m [39;49;00mmachinelearning.seldon.io/v1alpha2[37m[39;49;00m
    [94mkind[39;49;00m:[37m [39;49;00mSeldonDeployment[37m[39;49;00m
    [94mmetadata[39;49;00m:[37m[39;49;00m
    [37m  [39;49;00m[94mlabels[39;49;00m:[37m[39;49;00m
    [37m    [39;49;00m[94mapp[39;49;00m:[37m [39;49;00mseldon[37m[39;49;00m
    [37m    [39;49;00m[94mseldon.io/controller-id[39;49;00m:[37m [39;49;00mseldon-id1[37m[39;49;00m
    [37m  [39;49;00m[94mname[39;49;00m:[37m [39;49;00mtest-c1[37m[39;49;00m
    [94mspec[39;49;00m:[37m[39;49;00m
    [37m  [39;49;00m[94mname[39;49;00m:[37m [39;49;00mtest-c1[37m[39;49;00m
    [37m  [39;49;00m[94mpredictors[39;49;00m:[37m[39;49;00m
    [37m  [39;49;00m-[37m [39;49;00m[94mcomponentSpecs[39;49;00m:[37m[39;49;00m
    [37m    [39;49;00m-[37m [39;49;00m[94mspec[39;49;00m:[37m[39;49;00m
    [37m        [39;49;00m[94mcontainers[39;49;00m:[37m[39;49;00m
    [37m        [39;49;00m-[37m [39;49;00m[94mimage[39;49;00m:[37m [39;49;00mseldonio/mock_classifier:1.19.0-dev[37m[39;49;00m
    [37m          [39;49;00m[94mimagePullPolicy[39;49;00m:[37m [39;49;00mIfNotPresent[37m[39;49;00m
    [37m          [39;49;00m[94mname[39;49;00m:[37m [39;49;00mclassifier[37m[39;49;00m
    [37m          [39;49;00m[94mresources[39;49;00m:[37m[39;49;00m
    [37m            [39;49;00m[94mrequests[39;49;00m:[37m[39;49;00m
    [37m              [39;49;00m[94mmemory[39;49;00m:[37m [39;49;00m1Mi[37m[39;49;00m
    [37m        [39;49;00m[94mterminationGracePeriodSeconds[39;49;00m:[37m [39;49;00m1[37m[39;49;00m
    [37m    [39;49;00m[94mgraph[39;49;00m:[37m[39;49;00m
    [37m      [39;49;00m[94mchildren[39;49;00m:[37m [39;49;00m[][37m[39;49;00m
    [37m      [39;49;00m[94mendpoint[39;49;00m:[37m[39;49;00m
    [37m        [39;49;00m[94mtype[39;49;00m:[37m [39;49;00mREST[37m[39;49;00m
    [37m      [39;49;00m[94mname[39;49;00m:[37m [39;49;00mclassifier[37m[39;49;00m
    [37m      [39;49;00m[94mtype[39;49;00m:[37m [39;49;00mMODEL[37m[39;49;00m
    [37m    [39;49;00m[94mlabels[39;49;00m:[37m[39;49;00m
    [37m      [39;49;00m[94mversion[39;49;00m:[37m [39;49;00mv1[37m[39;49;00m
    [37m    [39;49;00m[94mname[39;49;00m:[37m [39;49;00mexample[37m[39;49;00m
    [37m    [39;49;00m[94mreplicas[39;49;00m:[37m [39;49;00m1[37m[39;49;00m



```python
!kubectl create -f resources/model_controller_id.yaml -n default
```

    seldondeployment.machinelearning.seldon.io/test-c1 created



```python
!kubectl wait sdep/test-c1 \
  --for=condition=ready \
  --timeout=120s \
  -n default
```

    seldondeployment.machinelearning.seldon.io/test-c1 condition met



```python
!kubectl get sdep -n default
```

    NAME      AGE
    test-c1   26s



```python
NAME = !kubectl get sdep -n default -o jsonpath='{.items[0].metadata.name}'
assert NAME[0] == "test-c1"
```


```python
!kubectl config set-context $(kubectl config current-context) --namespace=seldon-id1
```

    Context "kind-kind" modified.



```python
!kubectl delete -f resources/model_controller_id.yaml -n default
```

    seldondeployment.machinelearning.seldon.io "test-c1" deleted



```python
!helm delete seldon-controllerid
```

    release "seldon-controllerid" uninstalled

