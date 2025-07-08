# Scale Seldon Deployments based on Prometheus Metrics.
This notebook shows how you can scale Seldon Deployments based on Prometheus metrics via KEDA. 

[KEDA](https://keda.sh/) is a Kubernetes-based Event Driven Autoscaler. With KEDA, you can drive the scaling of any container in Kubernetes based on the number of events needing to be processed. 

With the support of KEDA in Seldon, you can scale your seldon deployments with any scalers listed [here](https://keda.sh/docs/2.0/scalers/).
In this example we will scale the seldon deployment with Prometheus metrics as an example.

## Install Seldon Core

Install Seldon Core as described in [docs](https://docs.seldon.io/projects/seldon-core/en/latest/workflow/install.html)

Make sure add `--set keda.enabled=true`

## Install Prometheus



```python
!kubectl create namespace seldon-monitoring
!helm upgrade --install seldon-monitoring kube-prometheus \
    --version 8.3.2 \
    --set fullnameOverride=seldon-monitoring \
    --namespace seldon-monitoring \
    --repo https://charts.bitnami.com/bitnami \
    --wait
```

    namespace/seldon-monitoring created
    Release "seldon-monitoring" does not exist. Installing it now.
    NAME: seldon-monitoring
    LAST DEPLOYED: Sun Feb  5 08:41:12 2023
    NAMESPACE: seldon-monitoring
    STATUS: deployed
    REVISION: 1
    TEST SUITE: None
    NOTES:
    CHART NAME: kube-prometheus
    CHART VERSION: 8.3.2
    APP VERSION: 0.62.0
    
    ** Please be patient while the chart is being deployed **
    
    Watch the Prometheus Operator Deployment status using the command:
    
        kubectl get deploy -w --namespace seldon-monitoring -l app.kubernetes.io/name=kube-prometheus-operator,app.kubernetes.io/instance=seldon-monitoring
    
    Watch the Prometheus StatefulSet status using the command:
    
        kubectl get sts -w --namespace seldon-monitoring -l app.kubernetes.io/name=kube-prometheus-prometheus,app.kubernetes.io/instance=seldon-monitoring
    
    Prometheus can be accessed via port "9090" on the following DNS name from within your cluster:
    
        seldon-monitoring-prometheus.seldon-monitoring.svc.cluster.local
    
    To access Prometheus from outside the cluster execute the following commands:
    
        echo "Prometheus URL: http://127.0.0.1:9090/"
        kubectl port-forward --namespace seldon-monitoring svc/seldon-monitoring-prometheus 9090:9090
    
    Watch the Alertmanager StatefulSet status using the command:
    
        kubectl get sts -w --namespace seldon-monitoring -l app.kubernetes.io/name=kube-prometheus-alertmanager,app.kubernetes.io/instance=seldon-monitoring
    
    Alertmanager can be accessed via port "9093" on the following DNS name from within your cluster:
    
        seldon-monitoring-alertmanager.seldon-monitoring.svc.cluster.local
    
    To access Alertmanager from outside the cluster execute the following commands:
    
        echo "Alertmanager URL: http://127.0.0.1:9093/"
        kubectl port-forward --namespace seldon-monitoring svc/seldon-monitoring-alertmanager 9093:9093



```python
!kubectl rollout status -n seldon-monitoring statefulsets/prometheus-seldon-monitoring-prometheus
```

    statefulset rolling update complete 1 pods at revision prometheus-seldon-monitoring-prometheus-b99bd7cb6...



```python
!cat pod-monitor.yaml
```

    apiVersion: monitoring.coreos.com/v1
    kind: PodMonitor
    metadata:
      name: seldon-podmonitor
      namespace: seldon-monitoring
    spec:
      selector:
        matchLabels:
          app.kubernetes.io/managed-by: seldon-core
      podMetricsEndpoints:
        - port: metrics
          path: /prometheus
      namespaceSelector:
        any: true



```python
!kubectl apply -f pod-monitor.yaml
```

    podmonitor.monitoring.coreos.com/seldon-podmonitor created


## Install KEDA

Follow the [docs for KEDA](https://keda.sh/docs/) to install.

## Create model with KEDA

To create a model with KEDA autoscaling you just need to add a KEDA spec referring in the Deployment, e.g.:
```yaml
kedaSpec:
  pollingInterval: 15                                # Optional. Default: 30 seconds
  minReplicaCount: 1                                 # Optional. Default: 0
  maxReplicaCount: 5                                 # Optional. Default: 100
  triggers:
  - type: prometheus
          metadata:
            # Required
            serverAddress: http://seldon-monitoring-prometheus.seldon-monitoring.svc.cluster.local:9090
            metricName: access_frequency
            threshold: '10'
            query: rate(seldon_api_executor_client_requests_seconds_count{model_name="classifier"}[1m])
```
The full SeldonDeployment spec is shown below.


```python
VERSION = !cat ../../version.txt
VERSION = VERSION[0]
VERSION
```




    '1.16.0-dev'




```python
%%writefile model_with_keda_prom.yaml
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
        - image: seldonio/mock_classifier:1.16.0-dev
          imagePullPolicy: IfNotPresent
          name: classifier
          resources:
            requests:
              cpu: '0.5'
      kedaSpec:
        pollingInterval: 15                                # Optional. Default: 30 seconds
        minReplicaCount: 1                                 # Optional. Default: 0
        maxReplicaCount: 5                                 # Optional. Default: 100
        triggers:
        - type: prometheus
          metadata:
            # Required
            serverAddress: http://seldon-monitoring-prometheus.seldon-monitoring.svc.cluster.local:9090
            metricName: access_frequency
            threshold: '10'
            query: rate(seldon_api_executor_client_requests_seconds_count{model_name="classifier"}[1m])
    graph:
      children: []
      endpoint:
        type: REST
      name: classifier
      type: MODEL
    name: example

```

    Overwriting model_with_keda_prom.yaml



```python
!kubectl create -f model_with_keda_prom.yaml
```

    seldondeployment.machinelearning.seldon.io/seldon-model created



```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=seldon-model -o jsonpath='{.items[0].metadata.name}')
```

    Waiting for deployment "seldon-model-example-0-classifier" rollout to finish: 0 of 1 updated replicas are available...
    deployment "seldon-model-example-0-classifier" successfully rolled out


## Create Load

We label some nodes for the loadtester. We attempt the first two as for Kind the first node shown will be the master.


```python
!kubectl label nodes $(kubectl get nodes -o jsonpath='{.items[0].metadata.name}') role=locust
!kubectl label nodes $(kubectl get nodes -o jsonpath='{.items[1].metadata.name}') role=locust
```

    node/kind-control-plane not labeled
    node/kind-worker not labeled


Before add loads to the model, there is only one replica


```python
!kubectl get deployment seldon-model-example-0-classifier
```

    NAME                                READY   UP-TO-DATE   AVAILABLE   AGE
    seldon-model-example-0-classifier   1/1     1            1           34s



```python
!helm install seldon-core-loadtesting seldon-core-loadtesting --repo https://storage.googleapis.com/seldon-charts \
    --set locust.host=http://seldon-model-example:8000 \
    --set oauth.enabled=false \
    --set locust.hatchRate=1 \
    --set locust.clients=1 \
    --set loadtest.sendFeedback=0 \
    --set locust.minWait=0 \
    --set locust.maxWait=0 \
    --set replicaCount=1
```

    NAME: seldon-core-loadtesting
    LAST DEPLOYED: Sun Feb  5 08:48:08 2023
    NAMESPACE: seldon
    STATUS: deployed
    REVISION: 1
    TEST SUITE: None


After a few mins you should see the deployment scaled to 5 replicas


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
    1
    1
    1
    1
    4



```python
!kubectl get deployment/seldon-model-example-0-classifier scaledobject/seldon-model-example-0-classifier
```

    NAME                                                READY   UP-TO-DATE   AVAILABLE   AGE
    deployment.apps/seldon-model-example-0-classifier   5/5     5            5           3m51s
    
    NAME                                                     SCALETARGETKIND      SCALETARGETNAME                     TRIGGERS     AUTHENTICATION   READY   ACTIVE   AGE
    scaledobject.keda.sh/seldon-model-example-0-classifier   apps/v1.Deployment   seldon-model-example-0-classifier   prometheus                    True    True     3m51s


## Remove Load


```python
!helm delete seldon-core-loadtesting
```

    release "seldon-core-loadtesting" uninstalled


After 5-10 mins you should see the deployment replica number decrease to 1

## Cleanup


```python
!kubectl delete -f model_with_keda_prom.yaml
```

    seldondeployment.machinelearning.seldon.io "seldon-model" deleted



```python
!helm delete seldon-monitoring -n seldon-monitoring
```

    release "seldon-monitoring" uninstalled



```python
!kubectl delete namespace seldon-monitoring
```

    namespace "seldon-monitoring" deleted



```python

```
