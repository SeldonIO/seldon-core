# Defining Disruption Budgets for Seldon Deployments

## Prerequisites
 
* A kubernetes cluster with kubectl configured
* pygmentize

## Setup Seldon Core

Use the setup notebook to [Setup Cluster](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Setup-Cluster) with [Ambassador Ingress](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Ambassador) and [Install Seldon Core](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Install-Seldon-Core). Instructions [also online](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html).


```python
!kubectl create namespace seldon
```


```python
!kubectl config set-context $(kubectl config current-context) --namespace=seldon
```

## Create model with Pod Disruption Budget

To create a model with a Pod Disruption Budget, it is first important to understand how you would like your application to respond to [voluntary disruptions](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#voluntary-and-involuntary-disruptions).  Depending on the type of disruption budgeting your application needs, you will either define either of the following:

* `minAvailable` which is a description of the number of pods from that set that must still be available after the eviction, even in the absence of the evicted pod. `minAvailable` can be either an absolute number or a percentage.
* `maxUnavailable` which is a description of the number of pods from that set that can be unavailable after the eviction. It can be either an absolute number or a percentage.

The full SeldonDeployment spec is shown below.


```python
!pygmentize model_with_pdb.yaml
```


```python
!kubectl apply -f model_with_pdb.yaml
```


```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=seldon-model -o jsonpath='{.items[0].metadata.name}')
```

## Validate Disruption Budget Configuration


```python
import json


def getPdbConfig():
    dp = !kubectl get pdb seldon-model-example-0-classifier -o json
    dp = json.loads("".join(dp))
    return dp["spec"]["maxUnavailable"]


assert getPdbConfig() == 2
```


```python
!kubectl get pods,deployments,pdb
```

## Update Disruption Budget and Validate Change

Next, we'll update the maximum number of unavailable pods and check that the PDB is properly updated to match.


```python
!pygmentize model_with_patched_pdb.yaml
```


```python
!kubectl apply -f model_with_patched_pdb.yaml
```


```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=seldon-model -o jsonpath='{.items[0].metadata.name}')
```


```python
assert getPdbConfig() == 1
```

## Clean Up


```python
!kubectl get pods,deployments,pdb
```


```python
!kubectl delete -f model_with_patched_pdb.yaml
```


```python

```
