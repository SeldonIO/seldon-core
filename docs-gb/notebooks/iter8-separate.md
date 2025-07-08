# Progressive Rollouts using Two Seldon Deployments

In this example we will AB Test two Iris models: an SKLearn model and an XGBOOST model.
We will run a progressive rollout allowing Iter8 to control the traffic to the two Seldon Deployments and gradually move traffic to the best model.

## Install Depenendcies

  * Istio
  * Seldon Core
  * Seldon Core Analytics
  * Iter8
  
 You can create a Kind cluster with all dependencies installed with [Ansible](https://www.ansible.com/) with:
  
  ```
  pip install ansible openshift
  ansible-galaxy collection install git+https://github.com/SeldonIO/ansible-k8s-collection.git,v0.1.0
  ```
  
  Then from `example/iter8` folder run:
  
  ```
  ansible-playbook playbooks/iter8.yml
  ```

## Create ABTest with Two Seldon Deployments


```python
!cat baseline.yaml
```

    apiVersion: v1
    kind: Namespace
    metadata:
        name: ns-baseline
    ---
    apiVersion: machinelearning.seldon.io/v1
    kind: SeldonDeployment
    metadata:
      name: iris
      namespace: ns-baseline
    spec:
      predictors:
      - name: default
        graph:
          name: classifier
          modelUri: gs://seldon-models/v1.19.0-dev/sklearn/iris
          implementation: SKLEARN_SERVER



```python
!kubectl apply -f baseline.yaml
```

    namespace/ns-baseline created
    seldondeployment.machinelearning.seldon.io/iris created



```python
!cat candidate.yaml
```

    apiVersion: v1
    kind: Namespace
    metadata:
        name: ns-candidate
    ---
    apiVersion: machinelearning.seldon.io/v1
    kind: SeldonDeployment
    metadata:
      name: iris
      namespace: ns-candidate
    spec:
      predictors:
      - name: default
        graph:
          name: classifier
          modelUri: gs://seldon-models/xgboost/iris
          implementation: XGBOOST_SERVER
    



```python
!kubectl apply -f candidate.yaml
```

    namespace/ns-candidate created
    seldondeployment.machinelearning.seldon.io/iris created



```python
!kubectl wait --for condition=ready --timeout=600s pods --all -n ns-baseline
```

    pod/iris-default-0-classifier-5dc67f64bf-brmss condition met



```python
!kubectl wait --for condition=ready --timeout=600s pods --all -n ns-candidate
```

    pod/iris-default-0-classifier-7fff869d67-g5qnh condition met


## Create Virtual Service to Split Traffic


```python
!cat routing-rule.yaml
```

    apiVersion: networking.istio.io/v1alpha3
    kind: VirtualService
    metadata:
      name: routing-rule
      namespace: default
    spec:
      gateways:
      - istio-system/seldon-gateway
      hosts:
      - iris.example.com
      http:
      - route:
        - destination:
            host: iris-default.ns-baseline.svc.cluster.local
            port:
              number: 8000
          headers:
            response:
              set:
                version: iris-v1
          weight: 100
        - destination:
            host: iris-default.ns-candidate.svc.cluster.local
            port:
              number: 8000
          headers:
            response:
              set:
                version: iris-v2
          weight: 0



```python
!kubectl apply -f routing-rule.yaml
```

    virtualservice.networking.istio.io/routing-rule created


## Create some load on models.

We will send reqeusts which will be split by the Seldon AB Test as well as random feedback to both models with feedback favouring the candidate


```python
!cat fortio.yaml
```

    ---
    apiVersion: batch/v1
    kind: Job
    metadata:
      name: fortio-requests
      namespace: default
    spec:
      template:
        spec:
          volumes:
          - name: shared
            emptyDir: {}    
          containers:
          - name: fortio
            image: fortio/fortio
            command: [ 'fortio', 'load', '-t', '6000s', '-qps', "5", '-json', '/shared/fortiooutput.json', '-H', 'Host: iris.example.com', '-H', 'Content-Type: application/json', '-payload', '{"data": {"ndarray":[[6.8,2.8,4.8,1.4]]}}',  "$(URL)" ]
            env:
            - name: URL
              value: URL_VALUE/api/v1.0/predictions
            volumeMounts:
            - name: shared
              mountPath: /shared         
          - name: busybox
            image: busybox:1.28
            command: ['sh', '-c', 'echo busybox is running! && sleep 6000']          
            volumeMounts:
            - name: shared
              mountPath: /shared       
          restartPolicy: Never
    ---
    apiVersion: batch/v1
    kind: Job
    metadata:
      name: fortio-irisv1-rewards
      namespace: default
    spec:
      template:
        spec:
          volumes:
          - name: shared
            emptyDir: {}    
          containers:
          - name: fortio
            image: fortio/fortio
            command: [ 'fortio', 'load', '-t', '6000s', '-qps', "0.7", '-json', '/shared/fortiooutput.json', '-H', 'Content-Type: application/json', '-payload', '{"reward": 1}',  "$(URL)" ]
            env:
            - name: URL
              value: URL_VALUE/seldon/ns-baseline/iris/api/v1.0/feedback
            volumeMounts:
            - name: shared
              mountPath: /shared         
          - name: busybox
            image: busybox:1.28
            command: ['sh', '-c', 'echo busybox is running! && sleep 6000']          
            volumeMounts:
            - name: shared
              mountPath: /shared       
          restartPolicy: Never
    ---
    apiVersion: batch/v1
    kind: Job
    metadata:
      name: fortio-irisv2-rewards
      namespace: default
    spec:
      template:
        spec:
          volumes:
          - name: shared
            emptyDir: {}    
          containers:
          - name: fortio
            image: fortio/fortio
            command: [ 'fortio', 'load', '-t', '6000s', '-qps', "1", '-json', '/shared/fortiooutput.json', '-H', 'Content-Type: application/json', '-payload', '{"reward": 1}',  "$(URL)" ]
            env:
            - name: URL
              value: URL_VALUE/seldon/ns-candidate/iris/api/v1.0/feedback
            volumeMounts:
            - name: shared
              mountPath: /shared         
          - name: busybox
            image: busybox:1.28
            command: ['sh', '-c', 'echo busybox is running! && sleep 6000']          
            volumeMounts:
            - name: shared
              mountPath: /shared       
          restartPolicy: Never



```python
!URL_VALUE="http://$(kubectl -n istio-system get svc istio-ingressgateway -o jsonpath='{.spec.clusterIP}')" && \
  sed "s+URL_VALUE+${URL_VALUE}+g" fortio.yaml | kubectl apply -f -
```

    job.batch/fortio-requests created
    job.batch/fortio-irisv1-rewards created
    job.batch/fortio-irisv2-rewards created



```python
!kubectl wait --for condition=ready --timeout=600s pods --all -n default
```

    pod/fortio-irisv1-rewards-t5drl condition met
    pod/fortio-irisv2-rewards-rb9k8 condition met
    pod/fortio-requests-fkp95 condition met


## Create Metrics to evaluate 

These are a standard set of metrics we use in all examples.


```python
!cat ../../metrics.yaml
```

    apiVersion: v1
    kind: Namespace
    metadata:
      name: iter8-seldon
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      name: 95th-percentile-tail-latency
      namespace: iter8-seldon
    spec:
      description: 95th percentile tail latency
      jqExpression: .data.result[0].value[1] | tonumber
      params:
      - name: query
        value: |
          histogram_quantile(0.95, sum(rate(seldon_api_executor_client_requests_seconds_bucket{seldon_deployment_id='$sid',predictor_name='$predictor',kubernetes_namespace='$ns'}[${elapsedTime}s])) by (le))
      provider: prometheus
      sampleSize: iter8-seldon/request-count
      type: Gauge
      units: milliseconds
      urlTemplate: http://seldon-core-analytics-prometheus-seldon.seldon-system/api/v1/query
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      name: error-count
      namespace: iter8-seldon
    spec:
      description: Number of error responses
      jqExpression: .data.result[0].value[1] | tonumber
      params:
      - name: query
        value: |
          sum(increase(seldon_api_executor_server_requests_seconds_count{code!='200',seldon_deployment_id='$sid',predictor_name='$predictor',kubernetes_namespace='$ns'}[${elapsedTime}s])) or on() vector(0)
      provider: prometheus
      type: Counter
      urlTemplate: http://seldon-core-analytics-prometheus-seldon.seldon-system/api/v1/query  
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      name: error-rate
      namespace: iter8-seldon
    spec:
      description: Fraction of requests with error responses
      jqExpression: .data.result[0].value[1] | tonumber
      params:
      - name: query
        value: |
          (sum(increase(seldon_api_executor_server_requests_seconds_count{code!='200',seldon_deployment_id='$sid',predictor_name='$predictor',kubernetes_namespace='$ns'}[${elapsedTime}s])) or on() vector(0)) / (sum(increase(seldon_api_executor_server_requests_seconds_count{seldon_deployment_id='$sid',predictor_name='$predictor',kubernetes_namespace='$ns'}[${elapsedTime}s])) or on() vector(0))
      provider: prometheus
      sampleSize: iter8-seldon/request-count
      type: Gauge
      urlTemplate: http://seldon-core-analytics-prometheus-seldon.seldon-system/api/v1/query    
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      name: mean-latency
      namespace: iter8-seldon
    spec:
      description: Mean latency
      jqExpression: .data.result[0].value[1] | tonumber
      params:
      - name: query
        value: |
          (sum(increase(seldon_api_executor_client_requests_seconds_sum{seldon_deployment_id='$sid',predictor_name='$predictor',kubernetes_namespace='$ns'}[${elapsedTime}s])) or on() vector(0)) / (sum(increase(seldon_api_executor_client_requests_seconds_count{seldon_deployment_id='$sid',predictor_name='$predictor',kubernetes_namespace='$ns'}[${elapsedTime}s])) or on() vector(0))
      provider: prometheus
      sampleSize: iter8-seldon/request-count
      type: Gauge
      units: milliseconds
      urlTemplate: http://seldon-core-analytics-prometheus-seldon.seldon-system/api/v1/query      
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      name: request-count
      namespace: iter8-seldon
    spec:
      description: Number of requests
      jqExpression: .data.result[0].value[1] | tonumber
      params:
      - name: query
        value: |
          sum(increase(seldon_api_executor_client_requests_seconds_sum{seldon_deployment_id='$sid',predictor_name='$predictor',kubernetes_namespace='$ns'}[${elapsedTime}s])) or on() vector(0)
      provider: prometheus
      type: Counter
      urlTemplate: http://seldon-core-analytics-prometheus-seldon.seldon-system/api/v1/query
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      name: user-engagement
      namespace: iter8-seldon
    spec:
      description: Number of feedback requests
      jqExpression: .data.result[0].value[1] | tonumber
      params:
      - name: query
        value: |
          sum(increase(seldon_api_executor_server_requests_seconds_count{service='feedback',seldon_deployment_id='$sid',predictor_name='$predictor',kubernetes_namespace='$ns'}[${elapsedTime}s])) or on() vector(0)
      provider: prometheus
      type: Gauge
      urlTemplate: http://seldon-core-analytics-prometheus-seldon.seldon-system/api/v1/query



```python
!kubectl create -f ../../metrics.yaml
```

    namespace/iter8-seldon created
    metric.iter8.tools/95th-percentile-tail-latency created
    metric.iter8.tools/error-count created
    metric.iter8.tools/error-rate created
    metric.iter8.tools/mean-latency created
    metric.iter8.tools/request-count created
    metric.iter8.tools/user-engagement created



```python
!kubectl get metrics -n iter8-seldon
```

    NAME                           TYPE      DESCRIPTION
    95th-percentile-tail-latency   Gauge     95th percentile tail latency
    error-count                    Counter   Number of error responses
    error-rate                     Gauge     Fraction of requests with error responses
    mean-latency                   Gauge     Mean latency
    request-count                  Counter   Number of requests
    user-engagement                Gauge     Number of feedback requests


## Create Progressive Rollout Experiment

  * Run 15 iterations with 5 second gaps between default and candidate models
  * Both models must pass objectives
  * winnder will be chosen based on user engagement metric


```python
!cat experiment.yaml
```

    apiVersion: iter8.tools/v2alpha2
    kind: Experiment
    metadata:
      name: quickstart-exp
    spec:
      target: iris
      strategy:
        testingPattern: A/B
        deploymentPattern: Progressive
        actions:
          # when the experiment completes, promote the winning version using kubectl apply
          finish:
          - task: common/exec
            with:
              cmd: /bin/bash
              args: [ "-c", "kubectl apply -f {{ .promote }}" ]
      criteria:
        requestCount: iter8-seldon/request-count
        rewards: # Business rewards
        - metric: iter8-seldon/user-engagement
          preferredDirection: High # maximize user engagement
        objectives:
        - metric: iter8-seldon/mean-latency
          upperLimit: 2000
        - metric: iter8-seldon/95th-percentile-tail-latency
          upperLimit: 5000
        - metric: iter8-seldon/error-rate
          upperLimit: "0.01"
      duration:
        intervalSeconds: 10
        iterationsPerLoop: 10
      versionInfo:
        # information about model versions used in this experiment
        baseline:
          name: iris-v1
          weightObjRef:
            apiVersion: networking.istio.io/v1alpha3
            kind: VirtualService
            name: routing-rule
            namespace: default
            fieldPath: .spec.http[0].route[0].weight      
          variables:
          - name: ns
            value: ns-baseline
          - name: sid
            value: iris
          - name: predictor
            value: default
          - name: promote
            value: https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/seldon/quickstart/promote-v1.yaml
        candidates:
        - name: iris-v2
          weightObjRef:
            apiVersion: networking.istio.io/v1alpha3
            kind: VirtualService
            name: routing-rule
            namespace: default
            fieldPath: .spec.http[0].route[1].weight      
          variables:
          - name: ns
            value: ns-candidate
          - name: sid
            value: iris
          - name: predictor
            value: default
          - name: promote
            value: https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/seldon/quickstart/promote-v2.yaml



```python
!kubectl create -f experiment.yaml
```

    experiment.iter8.tools/quickstart-exp created


## Monitor Experiment

Download iter8ctl. 

```
GO111MODULE=on GOBIN=/usr/local/bin go get github.com/iter8-tools/iter8ctl@v0.1.3
```

Then:

```
while clear; do kubectl get experiment quickstart-exp -o yaml | iter8ctl describe -f -; sleep 8; done
```

By the end you should see the xgboost candidate model is promoted.


```python
!kubectl wait experiment quickstart-exp --for=condition=Completed --timeout=300s
```

    experiment.iter8.tools/quickstart-exp condition met



```python
!kubectl get experiment quickstart-exp
```

    NAME             TYPE   TARGET   STAGE       COMPLETED ITERATIONS   MESSAGE
    quickstart-exp   A/B    iris     Completed   10                     ExperimentCompleted: Experiment Completed


## Cleanup


```python
!kubectl delete -f fortio.yaml
!kubectl delete -f experiment.yaml
!kubectl delete -f ../../metrics.yaml
!kubectl delete -f routing-rule.yaml
!kubectl delete -f baseline.yaml
!kubectl delete -f candidate.yaml
```

    job.batch "fortio-requests" deleted
    job.batch "fortio-irisv1-rewards" deleted
    job.batch "fortio-irisv2-rewards" deleted
    experiment.iter8.tools "quickstart-exp" deleted
    namespace "iter8-seldon" deleted
    metric.iter8.tools "95th-percentile-tail-latency" deleted
    metric.iter8.tools "error-count" deleted
    metric.iter8.tools "error-rate" deleted
    metric.iter8.tools "mean-latency" deleted
    metric.iter8.tools "request-count" deleted
    metric.iter8.tools "user-engagement" deleted
    virtualservice.networking.istio.io "routing-rule" deleted
    namespace "ns-baseline" deleted
    seldondeployment.machinelearning.seldon.io "iris" deleted
    namespace "ns-candidate" deleted
    seldondeployment.machinelearning.seldon.io "iris" deleted



```python

```
