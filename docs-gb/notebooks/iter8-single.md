# Progressive Rollouts with Single Seldon Deployment

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

## Create ABTest with Two Predictors


```python
!cat abtest.yaml
```

    apiVersion: v1
    kind: Namespace
    metadata:
        name: ns-production
    ---
    apiVersion: machinelearning.seldon.io/v1
    kind: SeldonDeployment
    metadata:
      name: iris
      namespace: ns-production
    spec:
      predictors:
      - name: baseline
        traffic: 100    
        graph:
          name: classifier
          modelUri: gs://seldon-models/v1.19.0-dev/sklearn/iris
          implementation: SKLEARN_SERVER
      - name: candidate
        traffic: 0
        graph:
          name: classifier
          modelUri: gs://seldon-models/xgboost/iris
          implementation: XGBOOST_SERVER
    



```python
!kubectl apply -f abtest.yaml
```

    namespace/ns-production created
    seldondeployment.machinelearning.seldon.io/iris created



```python
!kubectl wait --for condition=ready --timeout=600s pods --all -n ns-production
```

    pod/iris-baseline-0-classifier-5759fd6c8-j4hc2 condition met
    pod/iris-candidate-0-classifier-6d6d54786c-nthhm condition met


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
            command: [ 'fortio', 'load', '-t', '6000s', '-qps', "5", '-json', '/shared/fortiooutput.json', '-H', 'Content-Type: application/json', '-payload', '{"data": {"ndarray":[[6.8,2.8,4.8,1.4]]}}',  "$(URL)" ]
            env:
            - name: URL
              value: URL_VALUE/seldon/ns-production/iris/api/v1.0/predictions
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
              value: iris-baseline.ns-production:8000/api/v1.0/feedback
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
              value: iris-candidate.ns-production:8000/api/v1.0/feedback
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

    pod/fortio-irisv1-rewards-5srd4 condition met
    pod/fortio-irisv2-rewards-jzs4s condition met
    pod/fortio-requests-lrzhp condition met


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
        iterationsPerLoop: 15
      versionInfo:
        # information about model versions used in this experiment
        baseline:
          name: iris-v1
          weightObjRef:
            apiVersion: machinelearning.seldon.io/v1
            kind: SeldonDeployment
            name: iris
            namespace: ns-production
            fieldPath: .spec.predictors[0].traffic
          variables:
          - name: ns
            value: ns-production
          - name: sid
            value: iris
          - name: predictor
            value: baseline
          - name: promote
            value: https://gist.githubusercontent.com/cliveseldon/acac9b7e6ba3c52cde556323be0fc776/raw/78781a1f5c86a6cc24c3c7e64e3df211bc083207/promote-v1.yaml
        candidates:
        - name: iris-v2
          weightObjRef:
            apiVersion: machinelearning.seldon.io/v1
            kind: SeldonDeployment
            name: iris
            namespace: ns-production
            fieldPath: .spec.predictors[1].traffic
          variables:
          - name: ns
            value: ns-production
          - name: sid
            value: iris
          - name: predictor
            value: candidate
          - name: promote
            value: https://gist.githubusercontent.com/cliveseldon/3766b9315a187aa2800422205832ad9b/raw/ba00718fcacb8014e826cc6410a8190aa19116d4/promote-v2.yaml



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
!kubectl get experiment quickstart-exp -o yaml
```

    apiVersion: iter8.tools/v2alpha2
    kind: Experiment
    metadata:
      creationTimestamp: "2021-06-15T15:20:18Z"
      generation: 1
      managedFields:
      - apiVersion: iter8.tools/v2alpha2
        fieldsType: FieldsV1
        fieldsV1:
          f:spec:
            .: {}
            f:criteria:
              .: {}
              f:requestCount: {}
              f:rewards: {}
            f:duration:
              .: {}
              f:intervalSeconds: {}
              f:iterationsPerLoop: {}
            f:strategy:
              .: {}
              f:actions:
                .: {}
                f:finish: {}
              f:deploymentPattern: {}
              f:testingPattern: {}
            f:target: {}
            f:versionInfo:
              .: {}
              f:baseline:
                .: {}
                f:name: {}
                f:variables: {}
                f:weightObjRef:
                  .: {}
                  f:apiVersion: {}
                  f:fieldPath: {}
                  f:kind: {}
                  f:name: {}
                  f:namespace: {}
              f:candidates: {}
        manager: kubectl-create
        operation: Update
        time: "2021-06-15T15:20:18Z"
      - apiVersion: iter8.tools/v2alpha2
        fieldsType: FieldsV1
        fieldsV1:
          f:spec:
            f:criteria:
              f:objectives: {}
              f:strength: {}
            f:duration:
              f:maxLoops: {}
            f:strategy:
              f:weights:
                .: {}
                f:maxCandidateWeight: {}
                f:maxCandidateWeightIncrement: {}
          f:status:
            .: {}
            f:analysis:
              .: {}
              f:aggregatedMetrics:
                .: {}
                f:data:
                  .: {}
                  f:iter8-seldon/95th-percentile-tail-latency:
                    .: {}
                    f:data:
                      .: {}
                      f:iris-v1:
                        .: {}
                        f:value: {}
                      f:iris-v2:
                        .: {}
                        f:value: {}
                  f:iter8-seldon/error-rate:
                    .: {}
                    f:data:
                      .: {}
                      f:iris-v1:
                        .: {}
                        f:value: {}
                      f:iris-v2:
                        .: {}
                        f:value: {}
                  f:iter8-seldon/mean-latency:
                    .: {}
                    f:data:
                      .: {}
                      f:iris-v1:
                        .: {}
                        f:value: {}
                      f:iris-v2:
                        .: {}
                        f:value: {}
                  f:iter8-seldon/request-count:
                    .: {}
                    f:data:
                      .: {}
                      f:iris-v1:
                        .: {}
                        f:value: {}
                      f:iris-v2:
                        .: {}
                        f:value: {}
                  f:iter8-seldon/user-engagement:
                    .: {}
                    f:data:
                      .: {}
                      f:iris-v1:
                        .: {}
                        f:value: {}
                      f:iris-v2:
                        .: {}
                        f:value: {}
                f:message: {}
                f:provenance: {}
                f:timestamp: {}
              f:versionAssessments:
                .: {}
                f:data:
                  .: {}
                  f:iris-v1: {}
                  f:iris-v2: {}
                f:message: {}
                f:provenance: {}
                f:timestamp: {}
              f:weights:
                .: {}
                f:data: {}
                f:message: {}
                f:provenance: {}
                f:timestamp: {}
              f:winnerAssessment:
                .: {}
                f:data:
                  .: {}
                  f:winner: {}
                  f:winnerFound: {}
                f:message: {}
                f:provenance: {}
                f:timestamp: {}
            f:completedIterations: {}
            f:conditions: {}
            f:currentWeightDistribution: {}
            f:initTime: {}
            f:lastUpdateTime: {}
            f:message: {}
            f:metrics: {}
            f:stage: {}
            f:startTime: {}
            f:versionRecommendedForPromotion: {}
        manager: manager
        operation: Update
        time: "2021-06-15T15:20:51Z"
      name: quickstart-exp
      namespace: seldon
      resourceVersion: "68311"
      selfLink: /apis/iter8.tools/v2alpha2/namespaces/seldon/experiments/quickstart-exp
      uid: 5cefaea5-48f3-448c-9f76-c11df0a54b10
    spec:
      criteria:
        objectives:
        - metric: iter8-seldon/mean-latency
          upperLimit: 2000
        - metric: iter8-seldon/95th-percentile-tail-latency
          upperLimit: 5000
        - metric: iter8-seldon/error-rate
          upperLimit: "0.01"
        requestCount: iter8-seldon/request-count
        rewards:
        - metric: iter8-seldon/user-engagement
          preferredDirection: High
      duration:
        intervalSeconds: 10
        iterationsPerLoop: 15
      strategy:
        actions:
          finish:
          - task: common/exec
            with:
              args:
              - -c
              - kubectl apply -f {{ .promote }}
              cmd: /bin/bash
        deploymentPattern: Progressive
        testingPattern: A/B
      target: iris
      versionInfo:
        baseline:
          name: iris-v1
          variables:
          - name: ns
            value: ns-production
          - name: sid
            value: iris
          - name: predictor
            value: baseline
          - name: promote
            value: https://gist.githubusercontent.com/cliveseldon/acac9b7e6ba3c52cde556323be0fc776/raw/78781a1f5c86a6cc24c3c7e64e3df211bc083207/promote-v1.yaml
          weightObjRef:
            apiVersion: machinelearning.seldon.io/v1
            fieldPath: .spec.predictors[0].traffic
            kind: SeldonDeployment
            name: iris
            namespace: ns-production
        candidates:
        - name: iris-v2
          variables:
          - name: ns
            value: ns-production
          - name: sid
            value: iris
          - name: predictor
            value: candidate
          - name: promote
            value: https://gist.githubusercontent.com/cliveseldon/3766b9315a187aa2800422205832ad9b/raw/ba00718fcacb8014e826cc6410a8190aa19116d4/promote-v2.yaml
          weightObjRef:
            apiVersion: machinelearning.seldon.io/v1
            fieldPath: .spec.predictors[1].traffic
            kind: SeldonDeployment
            name: iris
            namespace: ns-production
    status:
      analysis:
        aggregatedMetrics:
          data:
            iter8-seldon/95th-percentile-tail-latency:
              data:
                iris-v1:
                  value: 24804743n
                iris-v2:
                  value: 24029853n
            iter8-seldon/error-rate:
              data:
                iris-v1:
                  value: "0"
                iris-v2:
                  value: "0"
            iter8-seldon/mean-latency:
              data:
                iris-v1:
                  value: 13352538n
                iris-v2:
                  value: 10753143n
            iter8-seldon/request-count:
              data:
                iris-v1:
                  value: 7500676139n
                iris-v2:
                  value: 7384333605n
            iter8-seldon/user-engagement:
              data:
                iris-v1:
                  value: 133384533334n
                iris-v2:
                  value: 188753612054n
          message: 'Error: ; Warning: ; Info: '
          provenance: http://iter8-analytics.iter8-system:8080/v2/analytics_results
          timestamp: "2021-06-15T15:23:23Z"
        versionAssessments:
          data:
            iris-v1:
            - true
            - true
            - true
            iris-v2:
            - true
            - true
            - true
          message: 'Error: ; Warning: ; Info: '
          provenance: http://iter8-analytics.iter8-system:8080/v2/analytics_results
          timestamp: "2021-06-15T15:23:23Z"
        weights:
          data:
          - name: iris-v1
            value: 5
          - name: iris-v2
            value: 95
          message: 'Error: ; Warning: ; Info: all ok'
          provenance: http://iter8-analytics.iter8-system:8080/v2/analytics_results
          timestamp: "2021-06-15T15:23:23Z"
        winnerAssessment:
          data:
            winner: iris-v2
            winnerFound: true
          message: 'Error: ; Warning: ; Info: found unique winner'
          provenance: http://iter8-analytics.iter8-system:8080/v2/analytics_results
          timestamp: "2021-06-15T15:23:23Z"
      completedIterations: 15
      conditions:
      - lastTransitionTime: "2021-06-15T15:23:38Z"
        message: Experiment Completed
        reason: ExperimentCompleted
        status: "True"
        type: Completed
      - lastTransitionTime: "2021-06-15T15:20:18Z"
        status: "False"
        type: Failed
      - lastTransitionTime: "2021-06-15T15:20:18Z"
        message: ""
        reason: TargetAcquired
        status: "True"
        type: TargetAcquired
      currentWeightDistribution:
      - name: iris-v1
        value: 5
      - name: iris-v2
        value: 95
      initTime: "2021-06-15T15:20:18Z"
      lastUpdateTime: "2021-06-15T15:23:25Z"
      message: 'ExperimentCompleted: Experiment Completed'
      metrics:
      - metricObj:
          apiVersion: iter8.tools/v2alpha2
          kind: Metric
          metadata:
            creationTimestamp: "2021-06-15T12:21:06Z"
            generation: 1
            managedFields:
            - apiVersion: iter8.tools/v2alpha2
              fieldsType: FieldsV1
              fieldsV1:
                f:spec:
                  .: {}
                  f:description: {}
                  f:jqExpression: {}
                  f:method: {}
                  f:params: {}
                  f:provider: {}
                  f:type: {}
                  f:urlTemplate: {}
              manager: kubectl-create
              operation: Update
              time: "2021-06-15T12:21:06Z"
            name: request-count
            namespace: iter8-seldon
            resourceVersion: "19711"
            selfLink: /apis/iter8.tools/v2alpha2/namespaces/iter8-seldon/metrics/request-count
            uid: cce03836-477f-4527-9979-2992b4ff89f6
          spec:
            description: Number of requests
            jqExpression: .data.result[0].value[1] | tonumber
            method: GET
            params:
            - name: query
              value: |
                sum(increase(seldon_api_executor_client_requests_seconds_sum{seldon_deployment_id='$sid',predictor_name='$predictor',kubernetes_namespace='$ns'}[${elapsedTime}s])) or on() vector(0)
            provider: prometheus
            type: Counter
            urlTemplate: http://seldon-core-analytics-prometheus-seldon.seldon-system/api/v1/query
        name: iter8-seldon/request-count
      - metricObj:
          apiVersion: iter8.tools/v2alpha2
          kind: Metric
          metadata:
            creationTimestamp: "2021-06-15T12:21:06Z"
            generation: 1
            managedFields:
            - apiVersion: iter8.tools/v2alpha2
              fieldsType: FieldsV1
              fieldsV1:
                f:spec:
                  .: {}
                  f:description: {}
                  f:jqExpression: {}
                  f:method: {}
                  f:params: {}
                  f:provider: {}
                  f:type: {}
                  f:urlTemplate: {}
              manager: kubectl-create
              operation: Update
              time: "2021-06-15T12:21:06Z"
            name: user-engagement
            namespace: iter8-seldon
            resourceVersion: "19712"
            selfLink: /apis/iter8.tools/v2alpha2/namespaces/iter8-seldon/metrics/user-engagement
            uid: b33392f0-8bb4-4e30-b044-63b8750721af
          spec:
            description: Number of feedback requests
            jqExpression: .data.result[0].value[1] | tonumber
            method: GET
            params:
            - name: query
              value: |
                sum(increase(seldon_api_executor_server_requests_seconds_count{service='feedback',seldon_deployment_id='$sid',predictor_name='$predictor',kubernetes_namespace='$ns'}[${elapsedTime}s])) or on() vector(0)
            provider: prometheus
            type: Gauge
            urlTemplate: http://seldon-core-analytics-prometheus-seldon.seldon-system/api/v1/query
        name: iter8-seldon/user-engagement
      - metricObj:
          apiVersion: iter8.tools/v2alpha2
          kind: Metric
          metadata:
            creationTimestamp: "2021-06-15T12:21:06Z"
            generation: 1
            managedFields:
            - apiVersion: iter8.tools/v2alpha2
              fieldsType: FieldsV1
              fieldsV1:
                f:spec:
                  .: {}
                  f:description: {}
                  f:jqExpression: {}
                  f:method: {}
                  f:params: {}
                  f:provider: {}
                  f:sampleSize: {}
                  f:type: {}
                  f:units: {}
                  f:urlTemplate: {}
              manager: kubectl-create
              operation: Update
              time: "2021-06-15T12:21:06Z"
            name: mean-latency
            namespace: iter8-seldon
            resourceVersion: "19709"
            selfLink: /apis/iter8.tools/v2alpha2/namespaces/iter8-seldon/metrics/mean-latency
            uid: 05bf8e06-348f-4367-b7d3-37c83e548f2e
          spec:
            description: Mean latency
            jqExpression: .data.result[0].value[1] | tonumber
            method: GET
            params:
            - name: query
              value: |
                (sum(increase(seldon_api_executor_client_requests_seconds_sum{seldon_deployment_id='$sid',predictor_name='$predictor',kubernetes_namespace='$ns'}[${elapsedTime}s])) or on() vector(0)) / (sum(increase(seldon_api_executor_client_requests_seconds_count{seldon_deployment_id='$sid',predictor_name='$predictor',kubernetes_namespace='$ns'}[${elapsedTime}s])) or on() vector(0))
            provider: prometheus
            sampleSize: iter8-seldon/request-count
            type: Gauge
            units: milliseconds
            urlTemplate: http://seldon-core-analytics-prometheus-seldon.seldon-system/api/v1/query
        name: iter8-seldon/mean-latency
      - metricObj:
          apiVersion: iter8.tools/v2alpha2
          kind: Metric
          metadata:
            creationTimestamp: "2021-06-15T12:21:06Z"
            generation: 1
            managedFields:
            - apiVersion: iter8.tools/v2alpha2
              fieldsType: FieldsV1
              fieldsV1:
                f:spec:
                  .: {}
                  f:description: {}
                  f:jqExpression: {}
                  f:method: {}
                  f:params: {}
                  f:provider: {}
                  f:sampleSize: {}
                  f:type: {}
                  f:units: {}
                  f:urlTemplate: {}
              manager: kubectl-create
              operation: Update
              time: "2021-06-15T12:21:06Z"
            name: 95th-percentile-tail-latency
            namespace: iter8-seldon
            resourceVersion: "19705"
            selfLink: /apis/iter8.tools/v2alpha2/namespaces/iter8-seldon/metrics/95th-percentile-tail-latency
            uid: e33054b8-8cb3-422f-9b24-698f38728759
          spec:
            description: 95th percentile tail latency
            jqExpression: .data.result[0].value[1] | tonumber
            method: GET
            params:
            - name: query
              value: |
                histogram_quantile(0.95, sum(rate(seldon_api_executor_client_requests_seconds_bucket{seldon_deployment_id='$sid',predictor_name='$predictor',kubernetes_namespace='$ns'}[${elapsedTime}s])) by (le))
            provider: prometheus
            sampleSize: iter8-seldon/request-count
            type: Gauge
            units: milliseconds
            urlTemplate: http://seldon-core-analytics-prometheus-seldon.seldon-system/api/v1/query
        name: iter8-seldon/95th-percentile-tail-latency
      - metricObj:
          apiVersion: iter8.tools/v2alpha2
          kind: Metric
          metadata:
            creationTimestamp: "2021-06-15T12:21:06Z"
            generation: 1
            managedFields:
            - apiVersion: iter8.tools/v2alpha2
              fieldsType: FieldsV1
              fieldsV1:
                f:spec:
                  .: {}
                  f:description: {}
                  f:jqExpression: {}
                  f:method: {}
                  f:params: {}
                  f:provider: {}
                  f:sampleSize: {}
                  f:type: {}
                  f:urlTemplate: {}
              manager: kubectl-create
              operation: Update
              time: "2021-06-15T12:21:06Z"
            name: error-rate
            namespace: iter8-seldon
            resourceVersion: "19707"
            selfLink: /apis/iter8.tools/v2alpha2/namespaces/iter8-seldon/metrics/error-rate
            uid: 2e5bec2f-bad2-4375-927d-05e2897d2280
          spec:
            description: Fraction of requests with error responses
            jqExpression: .data.result[0].value[1] | tonumber
            method: GET
            params:
            - name: query
              value: |
                (sum(increase(seldon_api_executor_server_requests_seconds_count{code!='200',seldon_deployment_id='$sid',predictor_name='$predictor',kubernetes_namespace='$ns'}[${elapsedTime}s])) or on() vector(0)) / (sum(increase(seldon_api_executor_server_requests_seconds_count{seldon_deployment_id='$sid',predictor_name='$predictor',kubernetes_namespace='$ns'}[${elapsedTime}s])) or on() vector(0))
            provider: prometheus
            sampleSize: iter8-seldon/request-count
            type: Gauge
            urlTemplate: http://seldon-core-analytics-prometheus-seldon.seldon-system/api/v1/query
        name: iter8-seldon/error-rate
      stage: Completed
      startTime: "2021-06-15T15:20:20Z"
      versionRecommendedForPromotion: iris-v2


## Cleanup


```python
!kubectl delete -f fortio.yaml
!kubectl delete -f experiment.yaml
!kubectl delete -f ../../metrics.yaml
!kubectl delete -f abtest.yaml
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
    namespace "ns-production" deleted
    seldondeployment.machinelearning.seldon.io "iris" deleted



```python

```
