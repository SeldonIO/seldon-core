# Basic Examples of Metrics with Prometheus Operator

## Prerequisites

 * A kubernetes cluster with kubectl configured
 * curl
 

## Setup Seldon Core

Install Seldon Core as described in [docs](../install/installation.md).

Then port-forward to that ingress on localhost:8003 in a separate terminal either with:
```bash
kubectl port-forward -n istio-system svc/istio-ingressgateway 8003:80
```


```bash
%%bash
kubectl create namespace seldon || echo "Seldon namespace already exists"
kubectl config set-context $(kubectl config current-context) --namespace=seldon
```

    Error from server (AlreadyExists): namespaces "seldon" already exists


    Seldon namespace already exists
    Context "kind-ansible" modified.


## Install Prometheus Operator


```bash
%%bash
kubectl create namespace seldon-monitoring

# Note: we set prometheus.scrapeInterval=1s for CI tests reliability here
helm upgrade --install seldon-monitoring kube-prometheus \
    --version 6.9.5 \
    --set fullnameOverride=seldon-monitoring \
    --set prometheus.scrapeInterval=1s \
    --namespace seldon-monitoring \
    --repo https://charts.bitnami.com/bitnami
```

    Error from server (AlreadyExists): namespaces "seldon-monitoring" already exists


    Release "seldon-monitoring" does not exist. Installing it now.


    W0509 15:22:58.165020  425776 warnings.go:70] policy/v1beta1 PodSecurityPolicy is deprecated in v1.21+, unavailable in v1.25+
    W0509 15:22:58.166767  425776 warnings.go:70] policy/v1beta1 PodSecurityPolicy is deprecated in v1.21+, unavailable in v1.25+
    W0509 15:22:58.168184  425776 warnings.go:70] policy/v1beta1 PodSecurityPolicy is deprecated in v1.21+, unavailable in v1.25+
    W0509 15:22:58.169785  425776 warnings.go:70] policy/v1beta1 PodSecurityPolicy is deprecated in v1.21+, unavailable in v1.25+
    W0509 15:22:58.171310  425776 warnings.go:70] policy/v1beta1 PodSecurityPolicy is deprecated in v1.21+, unavailable in v1.25+
    W0509 15:22:58.487981  425776 warnings.go:70] policy/v1beta1 PodSecurityPolicy is deprecated in v1.21+, unavailable in v1.25+
    W0509 15:22:58.489698  425776 warnings.go:70] policy/v1beta1 PodSecurityPolicy is deprecated in v1.21+, unavailable in v1.25+
    W0509 15:22:58.489729  425776 warnings.go:70] policy/v1beta1 PodSecurityPolicy is deprecated in v1.21+, unavailable in v1.25+
    W0509 15:22:58.489886  425776 warnings.go:70] policy/v1beta1 PodSecurityPolicy is deprecated in v1.21+, unavailable in v1.25+
    W0509 15:22:58.489921  425776 warnings.go:70] policy/v1beta1 PodSecurityPolicy is deprecated in v1.21+, unavailable in v1.25+


    NAME: seldon-monitoring
    LAST DEPLOYED: Mon May  9 15:22:57 2022
    NAMESPACE: seldon-monitoring
    STATUS: deployed
    REVISION: 1
    TEST SUITE: None
    NOTES:
    CHART NAME: kube-prometheus
    CHART VERSION: 6.9.5
    APP VERSION: 0.55.1
    
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



```bash
%%bash
# Extra sleep as statefulset is not always present right away
sleep 5 
kubectl rollout status -n seldon-monitoring deployment/seldon-monitoring-operator
kubectl rollout status -n seldon-monitoring deployment/prometheus-kube-state-metrics
kubectl rollout status -n seldon-monitoring statefulsets/prometheus-seldon-monitoring-prometheus
```

    Waiting for deployment "seldon-monitoring-operator" rollout to finish: 0 of 1 updated replicas are available...
    deployment "seldon-monitoring-operator" successfully rolled out


    Error from server (NotFound): deployments.apps "prometheus-kube-state-metrics" not found


    statefulset rolling update complete 1 pods at revision prometheus-seldon-monitoring-prometheus-5f84fbf5d4...



```bash
%%bash
cat <<EOF | kubectl apply -f -
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
EOF
```

    Warning: resource podmonitors/seldon-podmonitor is missing the kubectl.kubernetes.io/last-applied-configuration annotation which is required by kubectl apply. kubectl apply should only be used on resources created declaratively by either kubectl create --save-config or kubectl apply. The missing annotation will be patched automatically.


    podmonitor.monitoring.coreos.com/seldon-podmonitor configured


## Deploy Example Model


```python
%%writefile echo-sdep.yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: echo
  namespace: seldon
spec:
  predictors:
  - name: default
    replicas: 1
    graph:
      name: classifier
      type: MODEL
    componentSpecs:
    - spec:
        containers:
        - image: seldonio/echo-model:1.19.0-dev
          name: classifier
```

    Overwriting echo-sdep.yaml



```python
!kubectl apply -f echo-sdep.yaml
```

    seldondeployment.machinelearning.seldon.io/echo created



```bash
%%bash
deployment=$(kubectl get deploy -l seldon-deployment-id=echo -o jsonpath='{.items[0].metadata.name}')
kubectl rollout status deploy/${deployment}
```

    Waiting for deployment "echo-default-0-classifier" rollout to finish: 0 of 1 updated replicas are available...
    deployment "echo-default-0-classifier" successfully rolled out


## Sent series of REST requests


```bash
%%bash

# Wait for the model to become fully ready
echo "Waiting 5s for model to fully ready"
sleep 5

# Send 20 requests to REST endpoint
for i in `seq 1 10`; do sleep 0.1 && \
   curl -s -H "Content-Type: application/json" \
   -d '{"data": {"ndarray":[[1.0, 2.0, 5.0]]}}' \
   http://localhost:8003/seldon/seldon/echo/api/v1.0/predictions > /dev/null ; \
done

# Give time for metrics to get collected by Prometheus
echo "Waiting 10s for Prometheus to scrape metrics"
sleep 10
```

    Waiting 5s for model to fully ready
    Waiting 10s for Prometheus to scrape metrics


## Check Metrics (REST)


```python
import json
```


```python
%%writefile get-metrics.sh
QUERY='query=seldon_api_executor_client_requests_seconds_count{deployment_name=~"echo",namespace=~"seldon",method=~"post"}'
QUERY_URL=http://seldon-monitoring-prometheus.seldon-monitoring.svc.cluster.local:9090/api/v1/query

kubectl run --quiet=true -it --rm curlmetrics-$(date +%s) --image=radial/busyboxplus:curl --restart=Never -- \
    curl --data-urlencode ${QUERY} ${QUERY_URL}
```

    Overwriting get-metrics.sh



```python
metrics = ! bash get-metrics.sh
metrics = json.loads(metrics[0])
```


```python
metrics
```




    {'status': 'success',
     'data': {'resultType': 'vector',
      'result': [{'metric': {'__name__': 'seldon_api_executor_client_requests_seconds_count',
         'code': '200',
         'container': 'seldon-container-engine',
         'deployment_name': 'echo',
         'endpoint': 'metrics',
         'instance': '10.244.0.43:8000',
         'job': 'seldon-monitoring/seldon-podmonitor',
         'method': 'post',
         'model_image': 'seldonio/echo-model',
         'model_name': 'classifier',
         'model_version': '1.15.0-dev',
         'namespace': 'seldon',
         'pod': 'echo-default-0-classifier-6fcd878bc5-pzzsc',
         'predictor_name': 'default',
         'service': '/predict'},
        'value': [1652106256.269, '10']},
       {'metric': {'__name__': 'seldon_api_executor_client_requests_seconds_count',
         'code': '200',
         'container': 'seldon-container-engine',
         'deployment_name': 'echo',
         'endpoint': 'metrics',
         'instance': '10.244.0.43:8000',
         'job': 'seldon-system/seldon-podmonitor',
         'method': 'post',
         'model_image': 'seldonio/echo-model',
         'model_name': 'classifier',
         'model_version': '1.15.0-dev',
         'namespace': 'seldon',
         'pod': 'echo-default-0-classifier-6fcd878bc5-pzzsc',
         'predictor_name': 'default',
         'service': '/predict'},
        'value': [1652106256.269, '10']}]}}




```python
counter = int(metrics["data"]["result"][0]["value"][1])
assert counter == 10, f"expected 10 requests, got {counter}"
```

## Send series GRPC requests


```bash
%%bash
cd ../../../executor/proto && for i in `seq 1 10`; do sleep 0.1 && \
    grpcurl -d '{"data": {"ndarray":[[1.0, 2.0, 5.0]]}}' \
    -rpc-header seldon:echo -rpc-header namespace:seldon \
    -plaintext -proto ./prediction.proto \
     0.0.0.0:8003 seldon.protos.Seldon/Predict > /dev/null ; \
done

# Give time for metrics to get collected by Prometheus
echo "Waiting 10s for Prometheus to scrape metrics"
sleep 10
```

    Waiting 10s for Prometheus to scrape metrics


## Check metrics (GRPC)


```python
%%writefile get-metrics.sh
QUERY='query=seldon_api_executor_client_requests_seconds_count{deployment_name=~"echo",namespace=~"seldon",method=~"unary"}'
QUERY_URL=http://seldon-monitoring-prometheus.seldon-monitoring.svc.cluster.local:9090/api/v1/query

kubectl run --quiet=true -it --rm curlmetrics-$(date +%s) --image=radial/busyboxplus:curl --restart=Never -- \
    curl --data-urlencode ${QUERY} ${QUERY_URL}
```

    Overwriting get-metrics.sh



```python
metrics = ! bash get-metrics.sh
metrics = json.loads(metrics[0])
```


```python
counter = int(metrics["data"]["result"][0]["value"][1])
assert counter == 10, f"expected 10 requests, got {counter}"
```

## Check Custom Metrics

This model defines a few custom metrics in its `.py` class definition:
```Python
    def metrics(self):
        print("metrics called")
        return [
            # a counter which will increase by the given value
            {"type": "COUNTER", "key": "mycounter", "value": 1},

            # a gauge which will be set to given value
            {"type": "GAUGE", "key": "mygauge", "value": 100},

            # a timer (in msecs) which  will be aggregated into HISTOGRAM
            {"type": "TIMER", "key": "mytimer", "value": 20.2},
        ]
```      

We will be checking value of `mygaguge` metrics.


```python
%%writefile get-metrics.sh
QUERY='query=mygauge{deployment_name=~"echo",namespace=~"seldon"}'
QUERY_URL=http://seldon-monitoring-prometheus.seldon-monitoring.svc.cluster.local:9090/api/v1/query

kubectl run --quiet=true -it --rm curlmetrics-$(date +%s) --image=radial/busyboxplus:curl --restart=Never -- \
    curl --data-urlencode ${QUERY} ${QUERY_URL}
```

    Overwriting get-metrics.sh



```python
metrics = ! bash get-metrics.sh
metrics = json.loads(metrics[0])
```


```python
metrics
```




    {'status': 'success',
     'data': {'resultType': 'vector',
      'result': [{'metric': {'__name__': 'mygauge',
         'container': 'classifier',
         'deployment_name': 'echo',
         'endpoint': 'metrics',
         'image_name': 'seldonio/echo-model',
         'image_version': '1.15.0-dev',
         'instance': '10.244.0.43:6000',
         'job': 'seldon-monitoring/seldon-podmonitor',
         'method': 'predict',
         'model_image': 'seldonio/echo-model',
         'model_name': 'classifier',
         'model_version': '1.15.0-dev',
         'namespace': 'seldon',
         'pod': 'echo-default-0-classifier-6fcd878bc5-pzzsc',
         'predictor_name': 'default',
         'predictor_version': 'default',
         'seldon_deployment_name': 'echo',
         'worker_id': '50'},
        'value': [1652106270.106, '100']},
       {'metric': {'__name__': 'mygauge',
         'container': 'classifier',
         'deployment_name': 'echo',
         'endpoint': 'metrics',
         'image_name': 'seldonio/echo-model',
         'image_version': '1.15.0-dev',
         'instance': '10.244.0.43:6000',
         'job': 'seldon-monitoring/seldon-podmonitor',
         'method': 'predict',
         'model_image': 'seldonio/echo-model',
         'model_name': 'classifier',
         'model_version': '1.15.0-dev',
         'namespace': 'seldon',
         'pod': 'echo-default-0-classifier-6fcd878bc5-pzzsc',
         'predictor_name': 'default',
         'predictor_version': 'default',
         'seldon_deployment_name': 'echo',
         'worker_id': '58'},
        'value': [1652106270.106, '100']},
       {'metric': {'__name__': 'mygauge',
         'container': 'classifier',
         'deployment_name': 'echo',
         'endpoint': 'metrics',
         'image_name': 'seldonio/echo-model',
         'image_version': '1.15.0-dev',
         'instance': '10.244.0.43:6000',
         'job': 'seldon-system/seldon-podmonitor',
         'method': 'predict',
         'model_image': 'seldonio/echo-model',
         'model_name': 'classifier',
         'model_version': '1.15.0-dev',
         'namespace': 'seldon',
         'pod': 'echo-default-0-classifier-6fcd878bc5-pzzsc',
         'predictor_name': 'default',
         'predictor_version': 'default',
         'seldon_deployment_name': 'echo',
         'worker_id': '50'},
        'value': [1652106270.106, '100']},
       {'metric': {'__name__': 'mygauge',
         'container': 'classifier',
         'deployment_name': 'echo',
         'endpoint': 'metrics',
         'image_name': 'seldonio/echo-model',
         'image_version': '1.15.0-dev',
         'instance': '10.244.0.43:6000',
         'job': 'seldon-system/seldon-podmonitor',
         'method': 'predict',
         'model_image': 'seldonio/echo-model',
         'model_name': 'classifier',
         'model_version': '1.15.0-dev',
         'namespace': 'seldon',
         'pod': 'echo-default-0-classifier-6fcd878bc5-pzzsc',
         'predictor_name': 'default',
         'predictor_version': 'default',
         'seldon_deployment_name': 'echo',
         'worker_id': '58'},
        'value': [1652106270.106, '100']}]}}




```python
gauge = int(metrics["data"]["result"][0]["value"][1])
assert gauge == 100, f"expected 100 on guage, got {gauge}"
```

## Cleanup


```python
!kubectl delete sdep -n seldon echo
!helm uninstall -n seldon-monitoring seldon-monitoring
```

    seldondeployment.machinelearning.seldon.io "echo" deleted
    W0509 15:24:31.442158  428564 warnings.go:70] policy/v1beta1 PodSecurityPolicy is deprecated in v1.21+, unavailable in v1.25+
    W0509 15:24:31.443380  428564 warnings.go:70] policy/v1beta1 PodSecurityPolicy is deprecated in v1.21+, unavailable in v1.25+
    W0509 15:24:31.443570  428564 warnings.go:70] policy/v1beta1 PodSecurityPolicy is deprecated in v1.21+, unavailable in v1.25+
    W0509 15:24:31.443720  428564 warnings.go:70] policy/v1beta1 PodSecurityPolicy is deprecated in v1.21+, unavailable in v1.25+
    W0509 15:24:31.443934  428564 warnings.go:70] policy/v1beta1 PodSecurityPolicy is deprecated in v1.21+, unavailable in v1.25+
    release "seldon-monitoring" uninstalled

