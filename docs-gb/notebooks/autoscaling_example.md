# Autoscaling Seldon Deployments


## Prerequisites
 
- The cluster should have `metric-server` running in the `kube-system` namespace
- For Kind install `../../testing/scripts/metrics.yaml` See https://github.com/kubernetes-sigs/kind/issues/398
- For Minikube run:
    
    ```
    minikube addons enable metrics-server
    ```
    

## Setup Seldon Core

Use the setup notebook to [Setup Cluster](https://docs.seldon.ai/seldon-core-1/tutorials/notebooks/seldon-core-setup#setup-cluster) with [Ambassador Ingress](https://docs.seldon.ai/seldon-core-1/tutorials/notebooks/seldon-core-setup#install-ingress).


```python
!kubectl create namespace seldon
```

    Error from server (AlreadyExists): namespaces "seldon" already exists


## Create model with v2beta1 autoscaler

To create a model with an HorizontalPodAutoscaler there are three steps:


  1. Ensure you have a resource request for the metric you want to scale on if it is a standard metric such as cpu or memory, e.g.:
  
```
          resources:
            requests:
              cpu: '0.5'
     
```
     
  1. Add an v2beta1 HPA Spec referring to this Deployment, e.g.:
  
```
    - hpaSpec:
        maxReplicas: 3
        minReplicas: 1
        metrics:
        - resource:
            name: cpu
            targetAverageUtilization: 10
          type: Resource

```

The full SeldonDeployment spec is shown below.


```python
!pygmentize model_with_hpa_v2beta1.yaml
```

    [94mapiVersion[39;49;00m:[37m [39;49;00mmachinelearning.seldon.io/v1[37m[39;49;00m
    [94mkind[39;49;00m:[37m [39;49;00mSeldonDeployment[37m[39;49;00m
    [94mmetadata[39;49;00m:[37m[39;49;00m
    [37m  [39;49;00m[94mname[39;49;00m:[37m [39;49;00mseldon-model[37m[39;49;00m
    [94mspec[39;49;00m:[37m[39;49;00m
    [37m  [39;49;00m[94mname[39;49;00m:[37m [39;49;00mtest-deployment[37m[39;49;00m
    [37m  [39;49;00m[94mpredictors[39;49;00m:[37m[39;49;00m
    [37m  [39;49;00m-[37m [39;49;00m[94mcomponentSpecs[39;49;00m:[37m[39;49;00m
    [37m    [39;49;00m-[37m [39;49;00m[94mhpaSpec[39;49;00m:[37m[39;49;00m
    [37m        [39;49;00m[94mmaxReplicas[39;49;00m:[37m [39;49;00m3[37m[39;49;00m
    [37m        [39;49;00m[94mmetrics[39;49;00m:[37m[39;49;00m
    [37m        [39;49;00m-[37m [39;49;00m[94mresource[39;49;00m:[37m[39;49;00m
    [37m            [39;49;00m[94mname[39;49;00m:[37m [39;49;00mcpu[37m[39;49;00m
    [37m            [39;49;00m[94mtargetAverageUtilization[39;49;00m:[37m [39;49;00m10[37m[39;49;00m
    [37m          [39;49;00m[94mtype[39;49;00m:[37m [39;49;00mResource[37m[39;49;00m
    [37m        [39;49;00m[94mminReplicas[39;49;00m:[37m [39;49;00m1[37m[39;49;00m
    [37m      [39;49;00m[94mspec[39;49;00m:[37m[39;49;00m
    [37m        [39;49;00m[94mcontainers[39;49;00m:[37m[39;49;00m
    [37m        [39;49;00m-[37m [39;49;00m[94mimage[39;49;00m:[37m [39;49;00mseldonio/mock_classifier:1.19.0-dev[37m[39;49;00m
    [37m          [39;49;00m[94mimagePullPolicy[39;49;00m:[37m [39;49;00mIfNotPresent[37m[39;49;00m
    [37m          [39;49;00m[94mname[39;49;00m:[37m [39;49;00mclassifier[37m[39;49;00m
    [37m          [39;49;00m[94mresources[39;49;00m:[37m[39;49;00m
    [37m            [39;49;00m[94mrequests[39;49;00m:[37m[39;49;00m
    [37m              [39;49;00m[94mcpu[39;49;00m:[37m [39;49;00m[33m'[39;49;00m[33m0.5[39;49;00m[33m'[39;49;00m[37m[39;49;00m
    [37m        [39;49;00m[94mterminationGracePeriodSeconds[39;49;00m:[37m [39;49;00m1[37m[39;49;00m
    [37m    [39;49;00m[94mgraph[39;49;00m:[37m[39;49;00m
    [37m      [39;49;00m[94mchildren[39;49;00m:[37m [39;49;00m[][37m[39;49;00m
    [37m      [39;49;00m[94mname[39;49;00m:[37m [39;49;00mclassifier[37m[39;49;00m
    [37m      [39;49;00m[94mtype[39;49;00m:[37m [39;49;00mMODEL[37m[39;49;00m
    [37m    [39;49;00m[94mname[39;49;00m:[37m [39;49;00mexample[37m[39;49;00m



```python
!kubectl create -f model_with_hpa_v2beta1.yaml -n seldon
```

    seldondeployment.machinelearning.seldon.io/seldon-model created



```python
!kubectl wait sdep/seldon-model \
  --for=condition=ready \
  --timeout=120s \
  -n seldon
```

    seldondeployment.machinelearning.seldon.io/seldon-model condition met


### Create Load

We label some nodes for the loadtester. We attempt the first two as for Kind the first node shown will be the master.


```python
!kubectl label nodes $(kubectl get nodes -o jsonpath='{.items[0].metadata.name}') role=locust
```

    node/kind-control-plane not labeled



```python
!helm install loadtester ../../../helm-charts/seldon-core-loadtesting -n seldon  \
    --set locust.host=http://seldon-model-example:8000 \
    --set oauth.enabled=false \
    --set locust.hatchRate=1 \
    --set locust.clients=1 \
    --set loadtest.sendFeedback=0 \
    --set locust.minWait=0 \
    --set locust.maxWait=0 \
    --set replicaCount=1
```

    NAME: loadtester
    LAST DEPLOYED: Thu Dec  4 11:32:40 2025
    NAMESPACE: seldon
    STATUS: deployed
    REVISION: 1
    TEST SUITE: None


After a few mins you should see the deployment `my-dep` scaled to 3 deployments


```python
import json
import time


def getNumberPods():
    dp = !kubectl get deployment -n seldon seldon-model-example-0-classifier -o json
    dp = json.loads("".join(dp))
    return dp["status"]["replicas"]


scaled = False
for i in range(60):
    pods = getNumberPods()
    print(pods)
    if pods > 1:
        scaled = True
        break
    time.sleep(5)
assert scaled
```

    1
    1
    1
    1
    1
    1
    1
    1
    1
    3



```python
!kubectl get pods,deployments,hpa -n seldon
```

    NAME                                                     READY   STATUS    RESTARTS   AGE
    pod/locust-master-1-92n86                                1/1     Running   0          75s
    pod/locust-slave-1-v4cjz                                 1/1     Running   0          75s
    pod/seldon-model-example-0-classifier-7b5659c8f9-pcd22   2/2     Running   0          25s
    pod/seldon-model-example-0-classifier-7b5659c8f9-vpr2p   2/2     Running   0          115s
    pod/seldon-model-example-0-classifier-7b5659c8f9-wpbfm   2/2     Running   0          25s
    
    NAME                                                READY   UP-TO-DATE   AVAILABLE   AGE
    deployment.apps/seldon-model-example-0-classifier   3/3     3            3           116s
    
    NAME                                                                    REFERENCE                                      TARGETS        MINPODS   MAXPODS   REPLICAS   AGE
    horizontalpodautoscaler.autoscaling/seldon-model-example-0-classifier   Deployment/seldon-model-example-0-classifier   cpu: 51%/10%   1         3         3          116s



```python
!helm delete loadtester -n seldon
```

    release "loadtester" uninstalled



```python
!kubectl delete -f model_with_hpa_v2beta1.yaml -n seldon
```

    seldondeployment.machinelearning.seldon.io "seldon-model" deleted


## Create model with v2 autoscaler

To create a model with an HorizontalPodAutoscaler there are three steps:


  1. Ensure you have a resource request for the metric you want to scale on if it is a standard metric such as cpu or memory, e.g.:
  
```
          resources:
            requests:
              cpu: '0.5'
     
```
     
  1. Add an v2beta1 HPA Spec referring to this Deployment, e.g.:
  
```
    - hpaSpec:
        maxReplicas: 3
        minReplicas: 1
        metricsv2:
        - resource:
            name: cpu
            target:
              type: Utilization
              averageUtilization: 10
          type: Resource
```

The full SeldonDeployment spec is shown below.


```python
!pygmentize model_with_hpa_v2.yaml
```

    [94mapiVersion[39;49;00m:[37m [39;49;00mmachinelearning.seldon.io/v1[37m[39;49;00m
    [94mkind[39;49;00m:[37m [39;49;00mSeldonDeployment[37m[39;49;00m
    [94mmetadata[39;49;00m:[37m[39;49;00m
    [37m  [39;49;00m[94mname[39;49;00m:[37m [39;49;00mseldon-model[37m[39;49;00m
    [94mspec[39;49;00m:[37m[39;49;00m
    [37m  [39;49;00m[94mname[39;49;00m:[37m [39;49;00mtest-deployment[37m[39;49;00m
    [37m  [39;49;00m[94mpredictors[39;49;00m:[37m[39;49;00m
    [37m  [39;49;00m-[37m [39;49;00m[94mcomponentSpecs[39;49;00m:[37m[39;49;00m
    [37m    [39;49;00m-[37m [39;49;00m[94mhpaSpec[39;49;00m:[37m[39;49;00m
    [37m        [39;49;00m[94mmaxReplicas[39;49;00m:[37m [39;49;00m3[37m[39;49;00m
    [37m        [39;49;00m[94mmetricsv2[39;49;00m:[37m[39;49;00m
    [37m        [39;49;00m-[37m [39;49;00m[94mresource[39;49;00m:[37m[39;49;00m
    [37m            [39;49;00m[94mname[39;49;00m:[37m [39;49;00mcpu[37m[39;49;00m
    [37m            [39;49;00m[94mtarget[39;49;00m:[37m[39;49;00m
    [37m              [39;49;00m[94mtype[39;49;00m:[37m [39;49;00mUtilization[37m[39;49;00m
    [37m              [39;49;00m[94maverageUtilization[39;49;00m:[37m [39;49;00m10[37m[39;49;00m
    [37m          [39;49;00m[94mtype[39;49;00m:[37m [39;49;00mResource[37m[39;49;00m
    [37m        [39;49;00m[94mminReplicas[39;49;00m:[37m [39;49;00m1[37m[39;49;00m
    [37m      [39;49;00m[94mspec[39;49;00m:[37m[39;49;00m
    [37m        [39;49;00m[94mcontainers[39;49;00m:[37m[39;49;00m
    [37m        [39;49;00m-[37m [39;49;00m[94mimage[39;49;00m:[37m [39;49;00mseldonio/mock_classifier:1.19.0-dev[37m[39;49;00m
    [37m          [39;49;00m[94mimagePullPolicy[39;49;00m:[37m [39;49;00mIfNotPresent[37m[39;49;00m
    [37m          [39;49;00m[94mname[39;49;00m:[37m [39;49;00mclassifier[37m[39;49;00m
    [37m          [39;49;00m[94mresources[39;49;00m:[37m[39;49;00m
    [37m            [39;49;00m[94mrequests[39;49;00m:[37m[39;49;00m
    [37m              [39;49;00m[94mcpu[39;49;00m:[37m [39;49;00m[33m'[39;49;00m[33m0.5[39;49;00m[33m'[39;49;00m[37m[39;49;00m
    [37m        [39;49;00m[94mterminationGracePeriodSeconds[39;49;00m:[37m [39;49;00m1[37m[39;49;00m
    [37m    [39;49;00m[94mgraph[39;49;00m:[37m[39;49;00m
    [37m      [39;49;00m[94mchildren[39;49;00m:[37m [39;49;00m[][37m[39;49;00m
    [37m      [39;49;00m[94mname[39;49;00m:[37m [39;49;00mclassifier[37m[39;49;00m
    [37m      [39;49;00m[94mtype[39;49;00m:[37m [39;49;00mMODEL[37m[39;49;00m
    [37m    [39;49;00m[94mname[39;49;00m:[37m [39;49;00mexample[37m[39;49;00m



```python
!kubectl create -f model_with_hpa_v2.yaml -n seldon
```

    seldondeployment.machinelearning.seldon.io/seldon-model created



```python
!kubectl wait sdep/seldon-model \
  --for=condition=ready \
  --timeout=120s \
  -n seldon
```

    seldondeployment.machinelearning.seldon.io/seldon-model condition met


### Create Load

We label some nodes for the loadtester. We attempt the first two as for Kind the first node shown will be the master.


```python
!kubectl label nodes $(kubectl get nodes -o jsonpath='{.items[0].metadata.name}') role=locust
```

    node/kind-control-plane not labeled



```python
!helm install loadtester ../../../helm-charts/seldon-core-loadtesting -n seldon  \
    --set locust.host=http://seldon-model-example:8000 \
    --set oauth.enabled=false \
    --set locust.hatchRate=1 \
    --set locust.clients=1 \
    --set loadtest.sendFeedback=0 \
    --set locust.minWait=0 \
    --set locust.maxWait=0 \
    --set replicaCount=1
```

    NAME: loadtester
    LAST DEPLOYED: Thu Dec  4 11:35:11 2025
    NAMESPACE: seldon
    STATUS: deployed
    REVISION: 1
    TEST SUITE: None


After a few mins you should see the deployment `my-dep` scaled to 3 deployments


```python
import json
import time


def getNumberPods():
    dp = !kubectl get deployment -n seldon seldon-model-example-0-classifier -o json
    dp = json.loads("".join(dp))
    return dp["status"]["replicas"]


scaled = False
for i in range(60):
    pods = getNumberPods()
    print(pods)
    if pods > 1:
        scaled = True
        break
    time.sleep(5)
assert scaled
```

    1
    1
    1
    1
    1
    1
    1
    3



```python
!kubectl get pods,deployments,hpa -n seldon
```

    NAME                                                     READY   STATUS    RESTARTS   AGE
    pod/locust-master-1-vztf6                                1/1     Running   0          59s
    pod/locust-slave-1-98phc                                 1/1     Running   0          59s
    pod/seldon-model-example-0-classifier-789cd4b649-47b5j   0/2     Running   0          18s
    pod/seldon-model-example-0-classifier-789cd4b649-rkmz2   2/2     Running   0          93s
    pod/seldon-model-example-0-classifier-789cd4b649-wrhm9   0/2     Running   0          17s
    
    NAME                                                READY   UP-TO-DATE   AVAILABLE   AGE
    deployment.apps/seldon-model-example-0-classifier   1/3     3            1           93s
    
    NAME                                                                    REFERENCE                                      TARGETS        MINPODS   MAXPODS   REPLICAS   AGE
    horizontalpodautoscaler.autoscaling/seldon-model-example-0-classifier   Deployment/seldon-model-example-0-classifier   cpu: 51%/10%   1         3         3          93s



```python
!helm delete loadtester -n seldon
```

    release "loadtester" uninstalled



```python
!kubectl delete -f model_with_hpa_v2.yaml -n seldon
```

    seldondeployment.machinelearning.seldon.io "seldon-model" deleted

