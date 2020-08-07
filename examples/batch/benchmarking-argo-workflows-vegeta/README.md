## Benchmarking with Argo Worfklows & Vegeta

In this notebook we will dive into how you can run bench marking with batch processing with Argo Workflows, Seldon Core and Vegeta.

Dependencies:

* Seldon core installed as per the docs with Istio as an ingress 
* Argo Workfklows installed in cluster (and argo CLI for commands)


## Setup

### Install Seldon Core
Use the notebook to [set-up Seldon Core with Ambassador or Istio Ingress](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html).

Note: If running with KIND you need to make sure do follow [these steps](https://github.com/argoproj/argo/issues/2376#issuecomment-595593237) as workaround to the `/.../docker.sock` known issue.


### Install Argo Workflows
You can follow the instructions from the official [Argo Workflows Documentation](https://github.com/argoproj/argo#quickstart).

You also need to make sure that argo has permissions to create seldon deployments - for this you can just create a default-admin rolebinding as follows:


```python
!kubectl create rolebinding default-admin --clusterrole=admin --serviceaccount=default:default
```

    rolebinding.rbac.authorization.k8s.io/default-admin created


### Create Benchmark Argo Workflow

In order to create a benchmark, we created a simple argo workflow template so you can leverage the power of the helm charts.

Before we dive into the contents of the full helm chart, let's first give it a try with some of the settings.

We will run a batch job that will set up a Seldon Deployment with 1 replicas and 4 cpus (with 100 max workers) to send requests.


```python
!helm template seldon-benchmark-workflow helm-charts/seldon-benchmark-workflow/ \
    --set workflow.name=seldon-benchmark-process \
    --set seldonDeployment.name=sklearn \
    --set seldonDeployment.replicas=1 \
    --set seldonDeployment.serverWorkers=1 \
    --set seldonDeployment.serverThreads=10 \
    --set seldonDeployment.apiType=rest \
    --set benchmark.cpus=4 \
    --set benchmark.maxWorkers=100 \
    --set benchmark.duration=30s \
    --set benchmark.rate=0 \
    --set benchmark.data='\{"data": {"ndarray": [[0\,1\,2\,3]]\}\}' \
    | argo submit -
```

    Name:                seldon-benchmark-process
    Namespace:           default
    ServiceAccount:      default
    Status:              Pending
    Created:             Fri Aug 07 17:55:06 +0100 (now)



```python
!argo list
```

    NAME                       STATUS    AGE   DURATION   PRIORITY
    seldon-benchmark-process   Running   3s    3s         0



```python
!argo get seldon-benchmark-process
```

    Name:                seldon-benchmark-process
    Namespace:           default
    ServiceAccount:      default
    Status:              Succeeded
    Created:             Fri Aug 07 17:55:06 +0100 (1 minute ago)
    Started:             Fri Aug 07 17:55:06 +0100 (1 minute ago)
    Finished:            Fri Aug 07 17:56:11 +0100 (3 seconds ago)
    Duration:            1 minute 5 seconds
    
    [39mSTEP[0m                                                             PODNAME                              DURATION  MESSAGE
     [32m✔[0m seldon-benchmark-process (seldon-benchmark-process)                                                          
     ├---[32m✔[0m create-seldon-resource (create-seldon-resource-template)  seldon-benchmark-process-3980407503  2s        
     ├---[32m✔[0m wait-seldon-resource (wait-seldon-resource-template)      seldon-benchmark-process-2136965893  27s       
     └---[32m✔[0m run-benchmark (run-benchmark-template)                    seldon-benchmark-process-780051119   32s       



```python
!argo logs -w seldon-benchmark-process 
```

    [35mcreate-seldon-resource[0m:	time="2020-08-07T16:55:07.490Z" level=info msg="Starting Workflow Executor" version=v2.9.3
    [35mcreate-seldon-resource[0m:	time="2020-08-07T16:55:07.497Z" level=info msg="Creating a docker executor"
    [35mcreate-seldon-resource[0m:	time="2020-08-07T16:55:07.499Z" level=info msg="Executor (version: v2.9.3, build_date: 2020-07-18T19:11:19Z) initialized (pod: default/seldon-benchmark-process-3980407503) with template:\n{\"name\":\"create-seldon-resource-template\",\"arguments\":{},\"inputs\":{},\"outputs\":{},\"metadata\":{},\"resource\":{\"action\":\"create\",\"manifest\":\"apiVersion: machinelearning.seldon.io/v1\\nkind: SeldonDeployment\\nmetadata:\\n  name: \\\"sklearn\\\"\\n  namespace: default\\n  ownerReferences:\\n  - apiVersion: argoproj.io/v1alpha1\\n    blockOwnerDeletion: true\\n    kind: Workflow\\n    name: \\\"seldon-benchmark-process\\\"\\n    uid: \\\"b56998f0-2f6c-4f76-89b3-8b8b24ee9ec4\\\"\\nspec:\\n  name: \\\"sklearn\\\"\\n  transport: rest\\n  predictors:\\n    - componentSpecs:\\n      - spec:\\n        containers:\\n        - name: classifier\\n          env:\\n          - name: GUNICORN_THREADS\\n            value: 10\\n          - name: GUNICORN_WORKERS\\n            value: 1\\n          resources:\\n            requests:\\n              cpu: 50m\\n              memory: 100Mi\\n            limits:\\n              cpu: 50m\\n              memory: 1000Mi\\n      graph:\\n        children: []\\n        implementation: SKLEARN_SERVER\\n        modelUri: gs://seldon-models/sklearn/iris\\n        name: classifier\\n      name: default\\n      replicas: 1\\n\"}}"
    [35mcreate-seldon-resource[0m:	time="2020-08-07T16:55:07.499Z" level=info msg="Loading manifest to /tmp/manifest.yaml"
    [35mcreate-seldon-resource[0m:	time="2020-08-07T16:55:07.499Z" level=info msg="kubectl create -f /tmp/manifest.yaml -o json"
    [35mcreate-seldon-resource[0m:	time="2020-08-07T16:55:08.210Z" level=info msg=default/SeldonDeployment.machinelearning.seldon.io/sklearn
    [35mcreate-seldon-resource[0m:	time="2020-08-07T16:55:08.211Z" level=info msg="No output parameters"
    [32mwait-seldon-resource[0m:	Waiting for deployment "sklearn-default-0-classifier" rollout to finish: 0 of 1 updated replicas are available...
    [32mwait-seldon-resource[0m:	deployment "sklearn-default-0-classifier" successfully rolled out
    [35mrun-benchmark[0m:	{"latencies":{"total":3009995580656,"mean":273785299,"50th":223315257,"90th":262145632,"95th":271603494,"99th":2137159641,"max":5772979420,"min":17831002},"bytes_in":{"total":1363256,"mean":124},"bytes_out":{"total":373796,"mean":34},"earliest":"2020-08-07T16:55:39.319895683Z","latest":"2020-08-07T16:56:09.320730933Z","end":"2020-08-07T16:56:09.523609512Z","duration":30000835250,"wait":202878579,"requests":10994,"rate":366.4564639079507,"throughput":363.9949730103768,"success":1,"status_codes":{"200":10994},"errors":[]}



```python
import json
wf_logs = !argo logs -w seldon-benchmark-process 
wf_bench = wf_logs[-1]
wf_json_str = wf_bench[24:]
results = json.loads(wf_json_str)

print("Latencies:")
print("\tmean:", results["latencies"]["mean"] / 1e6, "ms")
print("\t50th:", results["latencies"]["50th"] / 1e6, "ms")
print("\t90th:", results["latencies"]["90th"] / 1e6, "ms")
print("\t95th:", results["latencies"]["95th"] / 1e6, "ms")
print("\t99th:", results["latencies"]["99th"] / 1e6, "ms")
print("")
print("Throughput:", str(results["throughput"]) + "/s")
print("Errors:", len(results["errors"]) > 0)
```

    Latencies:
    	mean: 273.785299 ms
    	50th: 223.315257 ms
    	90th: 262.145632 ms
    	95th: 271.603494 ms
    	99th: 2137.159641 ms
    
    Throughput: 363.9949730103768/s
    Errors: False



```python
!argo delete seldon-benchmark-process
```

    Workflow 'seldon-benchmark-process' deleted


## Create GRPC benchmark with GHZ and Argo Workflows 


```python
!helm template seldon-benchmark-workflow helm-charts/seldon-benchmark-workflow/ \
    --set workflow.name=seldon-benchmark-process \
    --set seldonDeployment.name=sklearn \
    --set seldonDeployment.replicas=1 \
    --set seldonDeployment.serverWorkers=1 \
    --set seldonDeployment.serverThreads=10 \
    --set seldonDeployment.apiType=grpc \
    --set benchmark.cpus=4 \
    --set benchmark.maxWorkers=100 \
    --set benchmark.duration="30s" \
    --set benchmark.rate=0 \
    --set benchmark.data='\{"data": {"ndarray": [[0\,1\,2\,3]]\}\}' \
    | argo submit -
```

    Name:                seldon-benchmark-process
    Namespace:           default
    ServiceAccount:      default
    Status:              Pending
    Created:             Fri Aug 07 17:56:34 +0100 (now)



```python
!argo list
```

    NAME                       STATUS    AGE   DURATION   PRIORITY
    seldon-benchmark-process   Running   3s    3s         0



```python
!argo get seldon-benchmark-process
```

    Name:                seldon-benchmark-process
    Namespace:           default
    ServiceAccount:      default
    Status:              Succeeded
    Created:             Fri Aug 07 17:56:34 +0100 (1 minute ago)
    Started:             Fri Aug 07 17:56:34 +0100 (1 minute ago)
    Finished:            Fri Aug 07 17:57:36 +0100 (6 seconds ago)
    Duration:            1 minute 2 seconds
    
    [39mSTEP[0m                                                             PODNAME                              DURATION  MESSAGE
     [32m✔[0m seldon-benchmark-process (seldon-benchmark-process)                                                          
     ├---[32m✔[0m create-seldon-resource (create-seldon-resource-template)  seldon-benchmark-process-3980407503  1s        
     ├---[32m✔[0m wait-seldon-resource (wait-seldon-resource-template)      seldon-benchmark-process-2136965893  25s       
     └---[32m✔[0m run-benchmark (run-benchmark-template)                    seldon-benchmark-process-780051119   31s       



```python
!argo logs -w seldon-benchmark-process 
```

    [35mcreate-seldon-resource[0m:	time="2020-08-07T16:56:35.211Z" level=info msg="Starting Workflow Executor" version=v2.9.3
    [35mcreate-seldon-resource[0m:	time="2020-08-07T16:56:35.214Z" level=info msg="Creating a docker executor"
    [35mcreate-seldon-resource[0m:	time="2020-08-07T16:56:35.214Z" level=info msg="Executor (version: v2.9.3, build_date: 2020-07-18T19:11:19Z) initialized (pod: default/seldon-benchmark-process-3980407503) with template:\n{\"name\":\"create-seldon-resource-template\",\"arguments\":{},\"inputs\":{},\"outputs\":{},\"metadata\":{},\"resource\":{\"action\":\"create\",\"manifest\":\"apiVersion: machinelearning.seldon.io/v1\\nkind: SeldonDeployment\\nmetadata:\\n  name: \\\"sklearn\\\"\\n  namespace: default\\n  ownerReferences:\\n  - apiVersion: argoproj.io/v1alpha1\\n    blockOwnerDeletion: true\\n    kind: Workflow\\n    name: \\\"seldon-benchmark-process\\\"\\n    uid: \\\"8164ae2c-9578-47d1-8007-356c5c9c4d65\\\"\\nspec:\\n  name: \\\"sklearn\\\"\\n  transport: grpc\\n  predictors:\\n    - componentSpecs:\\n      - spec:\\n        containers:\\n        - name: classifier\\n          env:\\n          - name: GUNICORN_THREADS\\n            value: 10\\n          - name: GUNICORN_WORKERS\\n            value: 1\\n          resources:\\n            requests:\\n              cpu: 50m\\n              memory: 100Mi\\n            limits:\\n              cpu: 50m\\n              memory: 1000Mi\\n      graph:\\n        children: []\\n        implementation: SKLEARN_SERVER\\n        modelUri: gs://seldon-models/sklearn/iris\\n        name: classifier\\n      name: default\\n      replicas: 1\\n\"}}"
    [35mcreate-seldon-resource[0m:	time="2020-08-07T16:56:35.214Z" level=info msg="Loading manifest to /tmp/manifest.yaml"
    [35mcreate-seldon-resource[0m:	time="2020-08-07T16:56:35.214Z" level=info msg="kubectl create -f /tmp/manifest.yaml -o json"
    [35mcreate-seldon-resource[0m:	time="2020-08-07T16:56:35.772Z" level=info msg=default/SeldonDeployment.machinelearning.seldon.io/sklearn
    [35mcreate-seldon-resource[0m:	time="2020-08-07T16:56:35.772Z" level=info msg="No output parameters"
    [32mwait-seldon-resource[0m:	Waiting for deployment "sklearn-default-0-classifier" rollout to finish: 0 of 1 updated replicas are available...
    [32mwait-seldon-resource[0m:	deployment "sklearn-default-0-classifier" successfully rolled out
    [35mrun-benchmark[0m:	{"date":"2020-08-07T16:57:35Z","endReason":"timeout","options":{"host":"istio-ingressgateway.istio-system.svc.cluster.local:80","proto":"/proto/prediction.proto","import-paths":["/proto","."],"call":"seldon.protos.Seldon/Predict","insecure":true,"total":2147483647,"concurrency":50,"connections":1,"duration":30000000000,"timeout":20000000000,"dial-timeout":10000000000,"data":{"data":{"ndarray":[[0,1,2,3]]}},"binary":false,"metadata":{"namespace":"default","seldon":"sklearn"},"CPUs":4},"count":15426,"total":30000789586,"average":96978401,"fastest":25564101,"slowest":213825809,"rps":514.1864668521462,"errorDistribution":{"rpc error: code = Unavailable desc = transport is closing":50},"statusCodeDistribution":{"OK":15376,"Unavailable":50},"latencyDistribution":[{"percentage":10,"latency":74147803},{"percentage":25,"latency":81290000},{"percentage":50,"latency":95194883},{"percentage":75,"latency":108884099},{"percentage":90,"latency":123530079},{"percentage":95,"latency":133409105},{"percentage":99,"latency":153339606}]}



```python
import json
wf_logs = !argo logs -w seldon-benchmark-process 
wf_bench = wf_logs[-1]
wf_json_str = wf_bench[24:]
results = json.loads(wf_json_str)

print("Latencies:")
print("\tmean:", results["average"] / 1e6, "ms")
print("\t50th:", results["latencyDistribution"][-5]["latency"] / 1e6, "ms")
print("\t90th:", results["latencyDistribution"][-3]["latency"] / 1e6, "ms")
print("\t95th:", results["latencyDistribution"][-2]["latency"] / 1e6, "ms")
print("\t99th:", results["latencyDistribution"][-1]["latency"] / 1e6, "ms")
print("")
print("Rate:", str(results["rps"]) + "/s")
print("Errors:", results["statusCodeDistribution"].get("Unavailable", 0) > 0)
print("Errors:", results["statusCodeDistribution"])
```

    Latencies:
    	mean: 96.978401 ms
    	50th: 95.194883 ms
    	90th: 123.530079 ms
    	95th: 133.409105 ms
    	99th: 153.339606 ms
    
    rps: 514.1864668521462/s
    Errors: True
    Errors: {'OK': 15376, 'Unavailable': 50}



```python
!argo delete seldon-benchmark-process
```

    Workflow 'seldon-benchmark-process' deleted



```python

```
