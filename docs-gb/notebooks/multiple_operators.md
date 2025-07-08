# Multiple Seldon Core Operators

This notebook illustrate how multiple Seldon Core Operators can share the same cluster. In particular:

  * A Namespaced Operator that only manages Seldon Deployments inside its namespace. Only needs Role RBAC and Namespace labeled with `seldon.io/controller-id`
  * A Clusterwide Operator that manges SeldonDeployment with a matching `seldon.io/controller-id` label.
  * A Clusterwide Operator that manages Seldon Deployments not handled by the above.

## Setup Seldon Core

Use the setup notebook to [Setup Cluster](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Setup-Cluster) with [Ambassador Ingress](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Ambassador) and [Install Seldon Core](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Install-Seldon-Core). Instructions [also online](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html).

## Namespaced Seldon Core Operator


```python
!kubectl create namespace seldon-ns1
```


```python
!kubectl label namespace seldon-ns1 seldon.io/controller-id=seldon-ns1
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


```python
!kubectl rollout status deployment/seldon-controller-manager -n seldon-ns1
```


```python
!kubectl create -f resources/model.yaml -n seldon-ns1
```


```python
!kubectl rollout status deployment/seldon-model-example-0-classifier -n seldon-ns1
```


```python
!kubectl get sdep -n seldon-ns1
```


```python
NAME = !kubectl get sdep -n seldon-ns1 -o jsonpath='{.items[0].metadata.name}'
assert NAME[0] == "seldon-model"
```


```python
!kubectl config set-context $(kubectl config current-context) --namespace=seldon-ns1
```


```python
!kubectl delete -f resources/model.yaml -n seldon-ns1
```


```python
!helm delete seldon-namespaced
```

## Label Focused Seldon Core Operator

 * We set `crd.create=false` as the CRD already exists in the cluster.
 * We set `controllerId=seldon-id1`. SeldonDeployments with this label will be managed.


```python
!kubectl create namespace seldon-id1
```


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


```python
!kubectl rollout status deployment/seldon-controller-manager -n seldon-id1
```


```python
!pygmentize resources/model_controller_id.yaml
```


```python
!kubectl create -f resources/model_controller_id.yaml -n default
```


```python
!kubectl rollout status deployment/test-c1-example-0-classifier -n default
```


```python
!kubectl get sdep -n default
```


```python
NAME = !kubectl get sdep -n default -o jsonpath='{.items[0].metadata.name}'
assert NAME[0] == "test-c1"
```


```python
!kubectl config set-context $(kubectl config current-context) --namespace=seldon-id1
```


```python
!kubectl delete -f resources/model_controller_id.yaml -n default
```


```python
!helm delete seldon-controllerid
```
