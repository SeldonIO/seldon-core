# Autoscaling Seldon Deployments


## Prerequisites
 
- The cluster should have `metric-server` running in the `kube-system` namespace
- For Kind install `../../testing/scripts/metrics.yaml` See https://github.com/kubernetes-sigs/kind/issues/398
- For Minikube run:
    
    ```
    minikube addons enable metrics-server
    ```
    

## Setup Seldon Core

Use the setup notebook to [Setup Cluster](../notebooks/seldon-core-setup.md#setup-cluster) with [Ambassador Ingress](../notebooks/seldon-core-setup.md#ambassador) and [Install Seldon Core](../notebooks/seldon-core-setup.md#Install-Seldon-Core). Instructions [also online](../notebooks/seldon-core-setup.md).


```python
!kubectl create namespace seldon
```

    Error from server (AlreadyExists): namespaces "seldon" already exists



```python
!kubectl config set-context $(kubectl config current-context) --namespace=seldon
```

    Context "kind-ansible" modified.


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

    [38;2;0;128;0;01mapiVersion[39;00m:[38;2;187;187;187m [39mmachinelearning.seldon.io/v1
    [38;2;0;128;0;01mkind[39;00m:[38;2;187;187;187m [39mSeldonDeployment
    [38;2;0;128;0;01mmetadata[39;00m:
    [38;2;187;187;187m  [39m[38;2;0;128;0;01mname[39;00m:[38;2;187;187;187m [39mseldon-model
    [38;2;0;128;0;01mspec[39;00m:
    [38;2;187;187;187m  [39m[38;2;0;128;0;01mname[39;00m:[38;2;187;187;187m [39mtest-deployment
    [38;2;187;187;187m  [39m[38;2;0;128;0;01mpredictors[39;00m:
    [38;2;187;187;187m  [39m-[38;2;187;187;187m [39m[38;2;0;128;0;01mcomponentSpecs[39;00m:
    [38;2;187;187;187m    [39m-[38;2;187;187;187m [39m[38;2;0;128;0;01mhpaSpec[39;00m:
    [38;2;187;187;187m        [39m[38;2;0;128;0;01mmaxReplicas[39;00m:[38;2;187;187;187m [39m3
    [38;2;187;187;187m        [39m[38;2;0;128;0;01mmetrics[39;00m:
    [38;2;187;187;187m        [39m-[38;2;187;187;187m [39m[38;2;0;128;0;01mresource[39;00m:
    [38;2;187;187;187m            [39m[38;2;0;128;0;01mname[39;00m:[38;2;187;187;187m [39mcpu
    [38;2;187;187;187m            [39m[38;2;0;128;0;01mtargetAverageUtilization[39;00m:[38;2;187;187;187m [39m10
    [38;2;187;187;187m          [39m[38;2;0;128;0;01mtype[39;00m:[38;2;187;187;187m [39mResource
    [38;2;187;187;187m        [39m[38;2;0;128;0;01mminReplicas[39;00m:[38;2;187;187;187m [39m1
    [38;2;187;187;187m      [39m[38;2;0;128;0;01mspec[39;00m:
    [38;2;187;187;187m        [39m[38;2;0;128;0;01mcontainers[39;00m:
    [38;2;187;187;187m        [39m-[38;2;187;187;187m [39m[38;2;0;128;0;01mimage[39;00m:[38;2;187;187;187m [39mseldonio/mock_classifier:1.5.0-dev
    [38;2;187;187;187m          [39m[38;2;0;128;0;01mimagePullPolicy[39;00m:[38;2;187;187;187m [39mIfNotPresent
    [38;2;187;187;187m          [39m[38;2;0;128;0;01mname[39;00m:[38;2;187;187;187m [39mclassifier
    [38;2;187;187;187m          [39m[38;2;0;128;0;01mresources[39;00m:
    [38;2;187;187;187m            [39m[38;2;0;128;0;01mrequests[39;00m:
    [38;2;187;187;187m              [39m[38;2;0;128;0;01mcpu[39;00m:[38;2;187;187;187m [39m[38;2;186;33;33m'[39m[38;2;186;33;33m0.5[39m[38;2;186;33;33m'[39m
    [38;2;187;187;187m        [39m[38;2;0;128;0;01mterminationGracePeriodSeconds[39;00m:[38;2;187;187;187m [39m1
    [38;2;187;187;187m    [39m[38;2;0;128;0;01mgraph[39;00m:
    [38;2;187;187;187m      [39m[38;2;0;128;0;01mchildren[39;00m:[38;2;187;187;187m [39m[]
    [38;2;187;187;187m      [39m[38;2;0;128;0;01mname[39;00m:[38;2;187;187;187m [39mclassifier
    [38;2;187;187;187m      [39m[38;2;0;128;0;01mtype[39;00m:[38;2;187;187;187m [39mMODEL
    [38;2;187;187;187m    [39m[38;2;0;128;0;01mname[39;00m:[38;2;187;187;187m [39mexample



```python
!kubectl create -f model_with_hpa_v2beta1.yaml
```

    seldondeployment.machinelearning.seldon.io/seldon-model created



```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=seldon-model -o jsonpath='{.items[0].metadata.name}')
```

    Waiting for deployment "seldon-model-example-0-classifier" rollout to finish: 0 of 1 updated replicas are available...
    deployment "seldon-model-example-0-classifier" successfully rolled out


### Create Load

We label some nodes for the loadtester. We attempt the first two as for Kind the first node shown will be the master.


```python
!kubectl label nodes $(kubectl get nodes -o jsonpath='{.items[0].metadata.name}') role=locust
```

    node/ansible-control-plane not labeled



```python
!helm install loadtester ../../../helm-charts/seldon-core-loadtesting  \
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
    LAST DEPLOYED: Sat Mar  4 09:13:46 2023
    NAMESPACE: seldon
    STATUS: deployed
    REVISION: 1
    TEST SUITE: None


After a few mins you should see the deployment `my-dep` scaled to 3 deployments


```python
import json
import time


def getNumberPods():
    dp = !kubectl get deployment seldon-model-example-0-classifier -o json
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

    3



```python
!kubectl get pods,deployments,hpa
```

    NAME                                                     READY   STATUS    RESTARTS   AGE
    pod/locust-master-1-xjplw                                1/1     Running   0          85s
    pod/locust-slave-1-gljjf                                 1/1     Running   0          85s
    pod/seldon-model-example-0-classifier-795b9cc8b6-7jfgp   0/2     Running   0          15s
    pod/seldon-model-example-0-classifier-795b9cc8b6-bqwg9   2/2     Running   0          80m
    pod/seldon-model-example-0-classifier-795b9cc8b6-fms5f   0/2     Running   0          15s
    
    NAME                                                READY   UP-TO-DATE   AVAILABLE   AGE
    deployment.apps/seldon-model-example-0-classifier   1/3     3            1           80m
    
    NAME                                                                    REFERENCE                                      TARGETS   MINPODS   MAXPODS   REPLICAS   AGE
    horizontalpodautoscaler.autoscaling/seldon-model-example-0-classifier   Deployment/seldon-model-example-0-classifier   60%/10%   1         3         1          80m



```python
!helm delete loadtester -n seldon
```

    release "loadtester" uninstalled



```python
!kubectl delete -f model_with_hpa_v2beta1.yaml
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

    [38;2;0;128;0;01mapiVersion[39;00m:[38;2;187;187;187m [39mmachinelearning.seldon.io/v1
    [38;2;0;128;0;01mkind[39;00m:[38;2;187;187;187m [39mSeldonDeployment
    [38;2;0;128;0;01mmetadata[39;00m:
    [38;2;187;187;187m  [39m[38;2;0;128;0;01mname[39;00m:[38;2;187;187;187m [39mseldon-model
    [38;2;0;128;0;01mspec[39;00m:
    [38;2;187;187;187m  [39m[38;2;0;128;0;01mname[39;00m:[38;2;187;187;187m [39mtest-deployment
    [38;2;187;187;187m  [39m[38;2;0;128;0;01mpredictors[39;00m:
    [38;2;187;187;187m  [39m-[38;2;187;187;187m [39m[38;2;0;128;0;01mcomponentSpecs[39;00m:
    [38;2;187;187;187m    [39m-[38;2;187;187;187m [39m[38;2;0;128;0;01mhpaSpec[39;00m:
    [38;2;187;187;187m        [39m[38;2;0;128;0;01mmaxReplicas[39;00m:[38;2;187;187;187m [39m3
    [38;2;187;187;187m        [39m[38;2;0;128;0;01mmetricsv2[39;00m:
    [38;2;187;187;187m        [39m-[38;2;187;187;187m [39m[38;2;0;128;0;01mresource[39;00m:
    [38;2;187;187;187m            [39m[38;2;0;128;0;01mname[39;00m:[38;2;187;187;187m [39mcpu
    [38;2;187;187;187m            [39m[38;2;0;128;0;01mtarget[39;00m:
    [38;2;187;187;187m              [39m[38;2;0;128;0;01mtype[39;00m:[38;2;187;187;187m [39mUtilization
    [38;2;187;187;187m              [39m[38;2;0;128;0;01maverageUtilization[39;00m:[38;2;187;187;187m [39m10
    [38;2;187;187;187m          [39m[38;2;0;128;0;01mtype[39;00m:[38;2;187;187;187m [39mResource
    [38;2;187;187;187m        [39m[38;2;0;128;0;01mminReplicas[39;00m:[38;2;187;187;187m [39m1
    [38;2;187;187;187m      [39m[38;2;0;128;0;01mspec[39;00m:
    [38;2;187;187;187m        [39m[38;2;0;128;0;01mcontainers[39;00m:
    [38;2;187;187;187m        [39m-[38;2;187;187;187m [39m[38;2;0;128;0;01mimage[39;00m:[38;2;187;187;187m [39mseldonio/mock_classifier:1.5.0-dev
    [38;2;187;187;187m          [39m[38;2;0;128;0;01mimagePullPolicy[39;00m:[38;2;187;187;187m [39mIfNotPresent
    [38;2;187;187;187m          [39m[38;2;0;128;0;01mname[39;00m:[38;2;187;187;187m [39mclassifier
    [38;2;187;187;187m          [39m[38;2;0;128;0;01mresources[39;00m:
    [38;2;187;187;187m            [39m[38;2;0;128;0;01mrequests[39;00m:
    [38;2;187;187;187m              [39m[38;2;0;128;0;01mcpu[39;00m:[38;2;187;187;187m [39m[38;2;186;33;33m'[39m[38;2;186;33;33m0.5[39m[38;2;186;33;33m'[39m
    [38;2;187;187;187m        [39m[38;2;0;128;0;01mterminationGracePeriodSeconds[39;00m:[38;2;187;187;187m [39m1
    [38;2;187;187;187m    [39m[38;2;0;128;0;01mgraph[39;00m:
    [38;2;187;187;187m      [39m[38;2;0;128;0;01mchildren[39;00m:[38;2;187;187;187m [39m[]
    [38;2;187;187;187m      [39m[38;2;0;128;0;01mname[39;00m:[38;2;187;187;187m [39mclassifier
    [38;2;187;187;187m      [39m[38;2;0;128;0;01mtype[39;00m:[38;2;187;187;187m [39mMODEL
    [38;2;187;187;187m    [39m[38;2;0;128;0;01mname[39;00m:[38;2;187;187;187m [39mexample



```python
!kubectl create -f model_with_hpa_v2.yaml
```

    seldondeployment.machinelearning.seldon.io/seldon-model created



```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=seldon-model -o jsonpath='{.items[0].metadata.name}')
```

    Waiting for deployment "seldon-model-example-0-classifier" rollout to finish: 0 of 1 updated replicas are available...
    deployment "seldon-model-example-0-classifier" successfully rolled out


### Create Load

We label some nodes for the loadtester. We attempt the first two as for Kind the first node shown will be the master.


```python
!kubectl label nodes $(kubectl get nodes -o jsonpath='{.items[0].metadata.name}') role=locust
```

    node/ansible-control-plane not labeled



```python
!helm install loadtester ../../../helm-charts/seldon-core-loadtesting  \
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
    LAST DEPLOYED: Sat Mar  4 09:20:04 2023
    NAMESPACE: seldon
    STATUS: deployed
    REVISION: 1
    TEST SUITE: None


After a few mins you should see the deployment `my-dep` scaled to 3 deployments


```python
import json
import time


def getNumberPods():
    dp = !kubectl get deployment seldon-model-example-0-classifier -o json
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
    3



```python
!kubectl get pods,deployments,hpa
```

    NAME                                                     READY   STATUS    RESTARTS   AGE
    pod/locust-master-1-qhvt6                                1/1     Running   0          11m
    pod/locust-slave-1-gnz8h                                 1/1     Running   0          11m
    pod/seldon-model-example-0-classifier-5f6445c99c-6t42q   2/2     Running   0          10m
    pod/seldon-model-example-0-classifier-5f6445c99c-fqfd9   2/2     Running   0          10m
    pod/seldon-model-example-0-classifier-5f6445c99c-s4wrv   2/2     Running   0          11m
    
    NAME                                                READY   UP-TO-DATE   AVAILABLE   AGE
    deployment.apps/seldon-model-example-0-classifier   3/3     3            3           11m
    
    NAME                                                                    REFERENCE                                      TARGETS   MINPODS   MAXPODS   REPLICAS   AGE
    horizontalpodautoscaler.autoscaling/seldon-model-example-0-classifier   Deployment/seldon-model-example-0-classifier   21%/10%   1         3         3          11m



```python
!helm delete loadtester -n seldon
```

    release "loadtester" uninstalled



```python
!kubectl delete -f model_with_hpa_v2.yaml
```

    seldondeployment.machinelearning.seldon.io "seldon-model" deleted

