# Scale Seldon Deployments based on Prometheus Metrics.
This notebook shows how you can scale Seldon Deployments based on Prometheus metrics via KEDA. 

[KEDA](https://keda.sh/) is a Kubernetes-based Event Driven Autoscaler. With KEDA, you can drive the scaling of any container in Kubernetes based on the number of events needing to be processed. 

With the support of KEDA in Seldon, you can scale your seldon deployments with any scalers listed [here](https://keda.sh/docs/2.0/scalers/).
In this example we will scale the seldon deployment with Prometheus metrics as an example.

## Install Seldon Core

Install Seldon Core as described in [docs](https://docs.seldon.ai/seldon-core-1/tutorials/notebooks/seldon-core-setup)

Make sure add `--set keda.enabled=true`

## Install Prometheus



```python
!kubectl create namespace seldon-monitoring
!helm upgrade --install seldon-monitoring kube-prometheus-stack \
    --version 44.4.1 \
    --set fullnameOverride=seldon-monitoring \
    --namespace seldon-monitoring \
    --repo https://prometheus-community.github.io/helm-charts/ \
    --wait
```

    Error from server (AlreadyExists): namespaces "seldon-monitoring" already exists
    Release "seldon-monitoring" has been upgraded. Happy Helming!
    NAME: seldon-monitoring
    LAST DEPLOYED: Thu Dec  4 11:20:53 2025
    NAMESPACE: seldon-monitoring
    STATUS: deployed
    REVISION: 2
    NOTES:
    kube-prometheus-stack has been installed. Check its status by running:
      kubectl --namespace seldon-monitoring get pods -l "release=seldon-monitoring"
    
    Visit https://github.com/prometheus-operator/kube-prometheus for instructions on how to create & configure Alertmanager and Prometheus instances using the Operator.



```python
!kubectl rollout status -n seldon-monitoring statefulsets/prometheus-seldon-monitoring-prometheus
```

    statefulset rolling update complete 1 pods at revision prometheus-seldon-monitoring-prometheus-58fb79649...



```python
!cat pod-monitor.yaml
```

    apiVersion: monitoring.coreos.com/v1
    kind: PodMonitor
    metadata:
      name: seldon-podmonitor
      namespace: seldon-monitoring
      labels:
        release: seldon-monitoring
    spec:
      namespaceSelector:
        matchNames:
          - seldon
      selector:
        matchLabels:
          seldon.io/model: "true"
      podMetricsEndpoints:
        - port: metrics
          path: /prometheus



```python
!kubectl apply -f pod-monitor.yaml
```

    podmonitor.monitoring.coreos.com/seldon-podmonitor unchanged


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




    '1.19.0-dev'




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
        - image: seldonio/mock_classifier:1.19.0-dev
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
!kubectl apply -f model_with_keda_prom.yaml -n seldon
```

    seldondeployment.machinelearning.seldon.io/seldon-model created



```python
!kubectl wait sdep/seldon-model \
  --for=condition=ready \
  --timeout=120s \
  -n seldon
```

    seldondeployment.machinelearning.seldon.io/seldon-model condition met


## Create Load

We label some nodes for the loadtester. We attempt the first two as for Kind the first node shown will be the master.


```python
!kubectl label nodes $(kubectl get nodes -o jsonpath='{.items[0].metadata.name}') role=locust
!kubectl label nodes $(kubectl get nodes -o jsonpath='{.items[1].metadata.name}') role=locust
```

    node/kind-control-plane labeled
    node/kind-worker labeled


Before add loads to the model, there is only one replica


```python
!kubectl get deployment seldon-model-example-0-classifier -n seldon
```

    NAME                                READY   UP-TO-DATE   AVAILABLE   AGE
    seldon-model-example-0-classifier   1/1     1            1           41s



```python
!helm install seldon-core-loadtesting seldon-core-loadtesting -n seldon --repo https://storage.googleapis.com/seldon-charts \
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
    LAST DEPLOYED: Thu Dec  4 11:23:04 2025
    NAMESPACE: seldon
    STATUS: deployed
    REVISION: 1
    TEST SUITE: None


After a few mins you should see the deployment scaled to 5 replicas


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
!kubectl get deployment/seldon-model-example-0-classifier -n seldon scaledobject/seldon-model-example-0-classifier
```

    NAME                                                READY   UP-TO-DATE   AVAILABLE   AGE
    deployment.apps/seldon-model-example-0-classifier   5/5     5            5           3m9s
    
    NAME                                                     SCALETARGETKIND      SCALETARGETNAME                     MIN   MAX   TRIGGERS     AUTHENTICATION   READY   ACTIVE   FALLBACK   PAUSED    AGE
    scaledobject.keda.sh/seldon-model-example-0-classifier   apps/v1.Deployment   seldon-model-example-0-classifier   1     5     prometheus                    True    True     False      Unknown   3m9s


## Remove Load


```python
!helm delete seldon-core-loadtesting -n seldon
```

    release "seldon-core-loadtesting" uninstalled


After 5-10 mins you should see the deployment replica number decrease to 1

## Cleanup


```python
!kubectl delete -f model_with_keda_prom.yaml -n seldon
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

