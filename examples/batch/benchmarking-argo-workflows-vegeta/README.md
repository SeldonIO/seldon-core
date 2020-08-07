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
!helm template seldon-batch-workflow helm-charts/seldon-batch-workflow/ \
    --set workflow.name=seldon-batch-process \
    --set seldonDeployment.name=sklearn \
    --set seldonDeployment.replicas=1 \
    --set seldonDeployment.serverWorkers=1 \
    --set seldonDeployment.serverThreads=10 \
    --set benchmark.cpus=4 \
    | argo submit -
```

    Name:                seldon-batch-process
    Namespace:           default
    ServiceAccount:      default
    Status:              Pending
    Created:             Thu Aug 06 14:54:09 +0100 (now)



```python
!argo list
```

    NAME                   STATUS      AGE   DURATION   PRIORITY
    seldon-batch-process   Succeeded   1m    1m         0



```python
!argo get seldon-batch-process
```

    Name:                seldon-batch-process
    Namespace:           default
    ServiceAccount:      default
    Status:              Succeeded
    Created:             Thu Aug 06 14:40:51 +0100 (1 minute ago)
    Started:             Thu Aug 06 14:40:51 +0100 (1 minute ago)
    Finished:            Thu Aug 06 14:41:57 +0100 (7 seconds ago)
    Duration:            1 minute 6 seconds
    
    [39mSTEP[0m                                                             PODNAME                          DURATION  MESSAGE
     [32m✔[0m seldon-batch-process (seldon-batch-process)                                                              
     ├---[32m✔[0m create-seldon-resource (create-seldon-resource-template)  seldon-batch-process-3626514072  1s        
     ├---[32m✔[0m wait-seldon-resource (wait-seldon-resource-template)      seldon-batch-process-2052519094  27s       
     └---[32m✔[0m run-benchmark (run-benchmark-template)                    seldon-batch-process-244800534   33s       



```python
!argo logs -w seldon-batch-process 
```

    [35mcreate-seldon-resource[0m:	time="2020-08-06T13:54:10.329Z" level=info msg="Starting Workflow Executor" version=v2.9.3
    [35mcreate-seldon-resource[0m:	time="2020-08-06T13:54:10.333Z" level=info msg="Creating a docker executor"
    [35mcreate-seldon-resource[0m:	time="2020-08-06T13:54:10.333Z" level=info msg="Executor (version: v2.9.3, build_date: 2020-07-18T19:11:19Z) initialized (pod: default/seldon-batch-process-3626514072) with template:\n{\"name\":\"create-seldon-resource-template\",\"arguments\":{},\"inputs\":{},\"outputs\":{},\"metadata\":{},\"resource\":{\"action\":\"create\",\"manifest\":\"apiVersion: machinelearning.seldon.io/v1\\nkind: SeldonDeployment\\nmetadata:\\n  name: \\\"sklearn\\\"\\n  namespace: default\\n  ownerReferences:\\n  - apiVersion: argoproj.io/v1alpha1\\n    blockOwnerDeletion: true\\n    kind: Workflow\\n    name: \\\"seldon-batch-process\\\"\\n    uid: \\\"853b463e-cd3f-42f8-b99a-0f82a83cf6f4\\\"\\nspec:\\n  name: \\\"sklearn\\\"\\n  predictors:\\n    - componentSpecs:\\n      - spec:\\n        containers:\\n        - name: classifier\\n          env:\\n          - name: GUNICORN_THREADS\\n            value: 10\\n          - name: GUNICORN_WORKERS\\n            value: 1\\n          resources:\\n            requests:\\n              cpu: 50m\\n              memory: 100Mi\\n            limits:\\n              cpu: 50m\\n              memory: 1000Mi\\n      graph:\\n        children: []\\n        implementation: SKLEARN_SERVER\\n        modelUri: gs://seldon-models/sklearn/iris\\n        name: classifier\\n      name: default\\n      replicas: 1\\n\"}}"
    [35mcreate-seldon-resource[0m:	time="2020-08-06T13:54:10.333Z" level=info msg="Loading manifest to /tmp/manifest.yaml"
    [35mcreate-seldon-resource[0m:	time="2020-08-06T13:54:10.333Z" level=info msg="kubectl create -f /tmp/manifest.yaml -o json"
    [35mcreate-seldon-resource[0m:	time="2020-08-06T13:54:10.945Z" level=info msg=default/SeldonDeployment.machinelearning.seldon.io/sklearn
    [35mcreate-seldon-resource[0m:	time="2020-08-06T13:54:10.945Z" level=info msg="No output parameters"
    [32mwait-seldon-resource[0m:	Waiting for deployment "sklearn-default-0-classifier" rollout to finish: 0 of 1 updated replicas are available...
    [32mwait-seldon-resource[0m:	deployment "sklearn-default-0-classifier" successfully rolled out
    [35mrun-benchmark[0m:	{"latencies":{"total":3013824928755,"mean":263055331,"50th":261546418,"90th":289796292,"95th":297608972,"99th":324027882,"max":1364625689,"min":22613200},"bytes_in":{"total":1420668,"mean":124},"bytes_out":{"total":538479,"mean":47},"earliest":"2020-08-06T13:54:42.340084846Z","latest":"2020-08-06T13:55:12.339980392Z","end":"2020-08-06T13:55:12.624864691Z","duration":29999895546,"wait":284884299,"requests":11457,"rate":381.9013297040498,"throughput":378.3088422183642,"success":1,"status_codes":{"200":11457},"errors":[]}



```python
def print_vegeta_results(results):
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


```python
import json
wf_logs = !argo logs -w seldon-batch-process 
wf_bench = wf_logs[-1]
wf_json_str = wf_bench[24:]
wf_json = json.loads(wf_json_str)
print_vegeta_results(wf_json)
```

    Latencies:
    	mean: 263.055331 ms
    	50th: 261.546418 ms
    	90th: 289.796292 ms
    	95th: 297.608972 ms
    	99th: 324.027882 ms
    
    Throughput: 378.3088422183642/s
    Errors: False



```python
!argo delete seldon-batch-process
```

    Workflow 'seldon-batch-process' deleted



```python

```
