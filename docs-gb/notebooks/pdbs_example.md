# Defining Disruption Budgets for Seldon Deployments

## Prerequisites
 
* A kubernetes cluster with kubectl configured
* pygmentize

## Setup Seldon Core

Use the setup notebook to [Setup Cluster](https://docs.seldon.ai/seldon-core-1/tutorials/notebooks/seldon-core-setup) with [Ambassador Ingress](https://docs.seldon.ai/seldon-core-1/tutorials/notebooks/seldon-core-setup#ambassador).


```python
!kubectl create namespace seldon
```

    Error from server (AlreadyExists): namespaces "seldon" already exists


## Create model with Pod Disruption Budget

To create a model with a Pod Disruption Budget, it is first important to understand how you would like your application to respond to [voluntary disruptions](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#voluntary-and-involuntary-disruptions).  Depending on the type of disruption budgeting your application needs, you will either define either of the following:

* `minAvailable` which is a description of the number of pods from that set that must still be available after the eviction, even in the absence of the evicted pod. `minAvailable` can be either an absolute number or a percentage.
* `maxUnavailable` which is a description of the number of pods from that set that can be unavailable after the eviction. It can be either an absolute number or a percentage.

The full SeldonDeployment spec is shown below.


```python
!pygmentize model_with_pdb.yaml
```

    [94mapiVersion[39;49;00m:[37m [39;49;00mmachinelearning.seldon.io/v1[37m[39;49;00m
    [94mkind[39;49;00m:[37m [39;49;00mSeldonDeployment[37m[39;49;00m
    [94mmetadata[39;49;00m:[37m[39;49;00m
    [37m  [39;49;00m[94mname[39;49;00m:[37m [39;49;00mseldon-model[37m[39;49;00m
    [94mspec[39;49;00m:[37m[39;49;00m
    [37m  [39;49;00m[94mname[39;49;00m:[37m [39;49;00mtest-deployment[37m[39;49;00m
    [37m  [39;49;00m[94mreplicas[39;49;00m:[37m [39;49;00m2[37m[39;49;00m
    [37m  [39;49;00m[94mpredictors[39;49;00m:[37m[39;49;00m
    [37m  [39;49;00m-[37m [39;49;00m[94mcomponentSpecs[39;49;00m:[37m[39;49;00m
    [37m    [39;49;00m-[37m [39;49;00m[94mpdbSpec[39;49;00m:[37m[39;49;00m
    [37m        [39;49;00m[94mmaxUnavailable[39;49;00m:[37m [39;49;00m2[37m[39;49;00m
    [37m      [39;49;00m[94mspec[39;49;00m:[37m[39;49;00m
    [37m        [39;49;00m[94mcontainers[39;49;00m:[37m[39;49;00m
    [37m        [39;49;00m-[37m [39;49;00m[94mimage[39;49;00m:[37m [39;49;00mseldonio/mock_classifier_rest:1.3[37m[39;49;00m
    [37m          [39;49;00m[94mimagePullPolicy[39;49;00m:[37m [39;49;00mIfNotPresent[37m[39;49;00m
    [37m          [39;49;00m[94mname[39;49;00m:[37m [39;49;00mclassifier[37m[39;49;00m
    [37m          [39;49;00m[94mresources[39;49;00m:[37m[39;49;00m
    [37m            [39;49;00m[94mrequests[39;49;00m:[37m[39;49;00m
    [37m              [39;49;00m[94mcpu[39;49;00m:[37m [39;49;00m[33m'[39;49;00m[33m0.5[39;49;00m[33m'[39;49;00m[37m[39;49;00m
    [37m        [39;49;00m[94mterminationGracePeriodSeconds[39;49;00m:[37m [39;49;00m1[37m[39;49;00m
    [37m    [39;49;00m[94mgraph[39;49;00m:[37m[39;49;00m
    [37m      [39;49;00m[94mchildren[39;49;00m:[37m [39;49;00m[][37m[39;49;00m
    [37m      [39;49;00m[94mendpoint[39;49;00m:[37m[39;49;00m
    [37m        [39;49;00m[94mtype[39;49;00m:[37m [39;49;00mREST[37m[39;49;00m
    [37m      [39;49;00m[94mname[39;49;00m:[37m [39;49;00mclassifier[37m[39;49;00m
    [37m      [39;49;00m[94mtype[39;49;00m:[37m [39;49;00mMODEL[37m[39;49;00m
    [37m    [39;49;00m[94mname[39;49;00m:[37m [39;49;00mexample[37m[39;49;00m



```python
!kubectl apply -f model_with_pdb.yaml -n seldon
```

    seldondeployment.machinelearning.seldon.io/seldon-model created



```python
!kubectl wait sdep/seldon-model \
  --for=condition=ready \
  --timeout=120s \
  -n seldon
```

    seldondeployment.machinelearning.seldon.io/seldon-model condition met


## Validate Disruption Budget Configuration


```python
import json


def getPdbConfig():
    dp = !kubectl get pdb -n seldon seldon-model-example-0-classifier -o json
    dp = json.loads("".join(dp))
    return dp["spec"]["maxUnavailable"]


assert getPdbConfig() == 2
```


```python
!kubectl get pods,deployments,pdb -n seldon
```

    NAME                                                     READY   STATUS    RESTARTS   AGE
    pod/seldon-model-example-0-classifier-7964b5c9f8-89x9h   2/2     Running   0          29s
    pod/seldon-model-example-0-classifier-7964b5c9f8-p5fg9   2/2     Running   0          29s
    
    NAME                                                READY   UP-TO-DATE   AVAILABLE   AGE
    deployment.apps/seldon-model-example-0-classifier   2/2     2            2           29s
    
    NAME                                                           MIN AVAILABLE   MAX UNAVAILABLE   ALLOWED DISRUPTIONS   AGE
    poddisruptionbudget.policy/seldon-model-example-0-classifier   N/A             2                 2                     29s


## Update Disruption Budget and Validate Change

Next, we'll update the maximum number of unavailable pods and check that the PDB is properly updated to match.


```python
!pygmentize model_with_patched_pdb.yaml
```

    [94mapiVersion[39;49;00m:[37m [39;49;00mmachinelearning.seldon.io/v1[37m[39;49;00m
    [94mkind[39;49;00m:[37m [39;49;00mSeldonDeployment[37m[39;49;00m
    [94mmetadata[39;49;00m:[37m[39;49;00m
    [37m  [39;49;00m[94mname[39;49;00m:[37m [39;49;00mseldon-model[37m[39;49;00m
    [94mspec[39;49;00m:[37m[39;49;00m
    [37m  [39;49;00m[94mname[39;49;00m:[37m [39;49;00mtest-deployment[37m[39;49;00m
    [37m  [39;49;00m[94mreplicas[39;49;00m:[37m [39;49;00m2[37m[39;49;00m
    [37m  [39;49;00m[94mpredictors[39;49;00m:[37m[39;49;00m
    [37m  [39;49;00m-[37m [39;49;00m[94mcomponentSpecs[39;49;00m:[37m[39;49;00m
    [37m    [39;49;00m-[37m [39;49;00m[94mpdbSpec[39;49;00m:[37m[39;49;00m
    [37m        [39;49;00m[94mmaxUnavailable[39;49;00m:[37m [39;49;00m1[37m[39;49;00m
    [37m      [39;49;00m[94mspec[39;49;00m:[37m[39;49;00m
    [37m        [39;49;00m[94mcontainers[39;49;00m:[37m[39;49;00m
    [37m        [39;49;00m-[37m [39;49;00m[94mimage[39;49;00m:[37m [39;49;00mseldonio/mock_classifier_rest:1.3[37m[39;49;00m
    [37m          [39;49;00m[94mimagePullPolicy[39;49;00m:[37m [39;49;00mIfNotPresent[37m[39;49;00m
    [37m          [39;49;00m[94mname[39;49;00m:[37m [39;49;00mclassifier[37m[39;49;00m
    [37m          [39;49;00m[94mresources[39;49;00m:[37m[39;49;00m
    [37m            [39;49;00m[94mrequests[39;49;00m:[37m[39;49;00m
    [37m              [39;49;00m[94mcpu[39;49;00m:[37m [39;49;00m[33m'[39;49;00m[33m0.5[39;49;00m[33m'[39;49;00m[37m[39;49;00m
    [37m        [39;49;00m[94mterminationGracePeriodSeconds[39;49;00m:[37m [39;49;00m1[37m[39;49;00m
    [37m    [39;49;00m[94mgraph[39;49;00m:[37m[39;49;00m
    [37m      [39;49;00m[94mchildren[39;49;00m:[37m [39;49;00m[][37m[39;49;00m
    [37m      [39;49;00m[94mendpoint[39;49;00m:[37m[39;49;00m
    [37m        [39;49;00m[94mtype[39;49;00m:[37m [39;49;00mREST[37m[39;49;00m
    [37m      [39;49;00m[94mname[39;49;00m:[37m [39;49;00mclassifier[37m[39;49;00m
    [37m      [39;49;00m[94mtype[39;49;00m:[37m [39;49;00mMODEL[37m[39;49;00m
    [37m    [39;49;00m[94mname[39;49;00m:[37m [39;49;00mexample[37m[39;49;00m



```python
!kubectl apply -f model_with_patched_pdb.yaml -n seldon
```

    seldondeployment.machinelearning.seldon.io/seldon-model configured



```python
!kubectl wait sdep/seldon-model \
  --for=condition=ready \
  --timeout=120s \
  -n seldon
```

    seldondeployment.machinelearning.seldon.io/seldon-model condition met



```python
assert getPdbConfig() == 1
```

## Clean Up


```python
!kubectl get pods,deployments,pdb -n seldon
```

    NAME                                                     READY   STATUS    RESTARTS   AGE
    pod/seldon-model-example-0-classifier-778dc959fd-dhts9   2/2     Running   0          33s
    pod/seldon-model-example-0-classifier-778dc959fd-dpsnc   0/2     Running   0          9s
    pod/seldon-model-example-0-classifier-7964b5c9f8-89x9h   2/2     Running   0          73s
    
    NAME                                                READY   UP-TO-DATE   AVAILABLE   AGE
    deployment.apps/seldon-model-example-0-classifier   2/2     2            2           73s
    
    NAME                                                           MIN AVAILABLE   MAX UNAVAILABLE   ALLOWED DISRUPTIONS   AGE
    poddisruptionbudget.policy/seldon-model-example-0-classifier   N/A             1                 1                     73s



```python
!kubectl delete -f model_with_patched_pdb.yaml -n seldon
```

    seldondeployment.machinelearning.seldon.io "seldon-model" deleted

