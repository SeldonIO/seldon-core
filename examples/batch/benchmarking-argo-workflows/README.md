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



```python
def get_results(results, print_results=True):
    final = {}
    if "average" in results:
        final["mean"] = results["average"] / 1e6
        if results.get("latencyDistribution", False):
            final["50th"] = results["latencyDistribution"][-5]["latency"] / 1e6
            final["90th"] = results["latencyDistribution"][-3]["latency"] / 1e6
            final["95th"] = results["latencyDistribution"][-2]["latency"] / 1e6
            final["99th"] = results["latencyDistribution"][-1]["latency"] / 1e6
        final["rate"] = results["rps"]
        final["errors"] = results["statusCodeDistribution"]
    else:
        final["mean"] = results["latencies"]["mean"] / 1e6
        final["50th"] = results["latencies"]["50th"] / 1e6
        final["90th"] = results["latencies"]["90th"] / 1e6
        final["95th"] = results["latencies"]["95th"] / 1e6
        final["99th"] = results["latencies"]["99th"] / 1e6
        final["rate"] = results["throughput"]
        final["errors"] = results["errors"]
    if print_results:    
        print("Latencies:")
        print("\tmean:", final["mean"], "ms")
        print("\t50th:", final["50th"], "ms")
        print("\t90th:", final["90th"], "ms")
        print("\t95th:", final["95th"], "ms")
        print("\t99th:", final["99th"], "ms")
        print("")
        print("Rate:", str(final["rate"]) + "/s")
        print("Errors:", final["errors"])
    return final
```

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
    Created:             Fri Aug 07 18:09:40 +0100 (now)



```python
!argo list
```

    NAME                       STATUS      AGE   DURATION   PRIORITY
    seldon-benchmark-process   Succeeded   2m    1m         0



```python
!argo get seldon-benchmark-process
```

    Name:                seldon-benchmark-process
    Namespace:           default
    ServiceAccount:      default
    Status:              Succeeded
    Created:             Fri Aug 07 18:09:40 +0100 (2 minutes ago)
    Started:             Fri Aug 07 18:09:40 +0100 (2 minutes ago)
    Finished:            Fri Aug 07 18:11:09 +0100 (51 seconds ago)
    Duration:            1 minute 29 seconds
    
    [39mSTEP[0m                                                             PODNAME                              DURATION  MESSAGE
     [32m✔[0m seldon-benchmark-process (seldon-benchmark-process)                                                          
     ├---[32m✔[0m create-seldon-resource (create-seldon-resource-template)  seldon-benchmark-process-3980407503  2s        
     ├---[32m✔[0m wait-seldon-resource (wait-seldon-resource-template)      seldon-benchmark-process-2136965893  49s       
     └---[32m✔[0m run-benchmark (run-benchmark-template)                    seldon-benchmark-process-780051119   32s       



```python
!argo logs -w seldon-benchmark-process 
```

    [35mcreate-seldon-resource[0m:	time="2020-08-07T17:09:41.804Z" level=info msg="Starting Workflow Executor" version=v2.9.3
    [35mcreate-seldon-resource[0m:	time="2020-08-07T17:09:41.809Z" level=info msg="Creating a docker executor"
    [35mcreate-seldon-resource[0m:	time="2020-08-07T17:09:41.809Z" level=info msg="Executor (version: v2.9.3, build_date: 2020-07-18T19:11:19Z) initialized (pod: default/seldon-benchmark-process-3980407503) with template:\n{\"name\":\"create-seldon-resource-template\",\"arguments\":{},\"inputs\":{},\"outputs\":{},\"metadata\":{},\"resource\":{\"action\":\"create\",\"manifest\":\"apiVersion: machinelearning.seldon.io/v1\\nkind: SeldonDeployment\\nmetadata:\\n  name: \\\"sklearn\\\"\\n  namespace: default\\n  ownerReferences:\\n  - apiVersion: argoproj.io/v1alpha1\\n    blockOwnerDeletion: true\\n    kind: Workflow\\n    name: \\\"seldon-benchmark-process\\\"\\n    uid: \\\"e0364966-b2c1-4ee7-a7cf-421952ba3c7a\\\"\\nspec:\\n  annotations:\\n    seldon.io/executor: \\\"false\\\"\\n  name: \\\"sklearn\\\"\\n  transport: rest\\n  predictors:\\n    - componentSpecs:\\n      - spec:\\n        containers:\\n        - name: classifier\\n          env:\\n          - name: GUNICORN_THREADS\\n            value: 10\\n          - name: GUNICORN_WORKERS\\n            value: 1\\n          resources:\\n            requests:\\n              cpu: 50m\\n              memory: 100Mi\\n            limits:\\n              cpu: 50m\\n              memory: 1000Mi\\n      graph:\\n        children: []\\n        implementation: SKLEARN_SERVER\\n        modelUri: gs://seldon-models/sklearn/iris\\n        name: classifier\\n      name: default\\n      replicas: 1\\n\"}}"
    [35mcreate-seldon-resource[0m:	time="2020-08-07T17:09:41.809Z" level=info msg="Loading manifest to /tmp/manifest.yaml"
    [35mcreate-seldon-resource[0m:	time="2020-08-07T17:09:41.810Z" level=info msg="kubectl create -f /tmp/manifest.yaml -o json"
    [35mcreate-seldon-resource[0m:	time="2020-08-07T17:09:42.456Z" level=info msg=default/SeldonDeployment.machinelearning.seldon.io/sklearn
    [35mcreate-seldon-resource[0m:	time="2020-08-07T17:09:42.457Z" level=info msg="No output parameters"
    [32mwait-seldon-resource[0m:	Waiting for deployment "sklearn-default-0-classifier" rollout to finish: 0 of 1 updated replicas are available...
    [32mwait-seldon-resource[0m:	deployment "sklearn-default-0-classifier" successfully rolled out
    [35mrun-benchmark[0m:	{"latencies":{"total":3011298973622,"mean":339033885,"50th":272840630,"90th":339539236,"95th":368299307,"99th":4982426813,"max":5597505277,"min":206244298},"bytes_in":{"total":3081764,"mean":346.9673496960144},"bytes_out":{"total":301988,"mean":34},"earliest":"2020-08-07T17:10:37.117884325Z","latest":"2020-08-07T17:11:07.118729145Z","end":"2020-08-07T17:11:07.366654843Z","duration":30000844820,"wait":247925698,"requests":8882,"rate":296.05832946673667,"throughput":293.63176909007353,"success":1,"status_codes":{"200":8882},"errors":[]}



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
    	mean: 339.033885 ms
    	50th: 272.84063 ms
    	90th: 339.539236 ms
    	95th: 368.299307 ms
    	99th: 4982.426813 ms
    
    Throughput: 293.63176909007353/s
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
    --set benchmark.duration="120s" \
    --set benchmark.rate=0 \
    --set benchmark.data='\{"data": {"ndarray": [[0\,1\,2\,3]]\}\}' \
    | argo submit -
```

    Name:                seldon-benchmark-process
    Namespace:           default
    ServiceAccount:      default
    Status:              Pending
    Created:             Fri Aug 07 18:22:38 +0100 (now)



```python
!argo list
```

    NAME                       STATUS      AGE   DURATION   PRIORITY
    seldon-benchmark-process   Succeeded   4m    2m         0



```python
!argo get seldon-benchmark-process
```

    Name:                seldon-benchmark-process
    Namespace:           default
    ServiceAccount:      default
    Status:              Succeeded
    Created:             Fri Aug 07 18:22:38 +0100 (4 minutes ago)
    Started:             Fri Aug 07 18:22:38 +0100 (4 minutes ago)
    Finished:            Fri Aug 07 18:25:11 +0100 (1 minute ago)
    Duration:            2 minutes 33 seconds
    
    [39mSTEP[0m                                                             PODNAME                              DURATION  MESSAGE
     [32m✔[0m seldon-benchmark-process (seldon-benchmark-process)                                                          
     ├---[32m✔[0m create-seldon-resource (create-seldon-resource-template)  seldon-benchmark-process-3980407503  2s        
     ├---[32m✔[0m wait-seldon-resource (wait-seldon-resource-template)      seldon-benchmark-process-2136965893  26s       
     └---[32m✔[0m run-benchmark (run-benchmark-template)                    seldon-benchmark-process-780051119   2m        



```python
!argo logs -w seldon-benchmark-process 
```

    [35mcreate-seldon-resource[0m:	time="2020-08-07T17:22:39.446Z" level=info msg="Starting Workflow Executor" version=v2.9.3
    [35mcreate-seldon-resource[0m:	time="2020-08-07T17:22:39.450Z" level=info msg="Creating a docker executor"
    [35mcreate-seldon-resource[0m:	time="2020-08-07T17:22:39.450Z" level=info msg="Executor (version: v2.9.3, build_date: 2020-07-18T19:11:19Z) initialized (pod: default/seldon-benchmark-process-3980407503) with template:\n{\"name\":\"create-seldon-resource-template\",\"arguments\":{},\"inputs\":{},\"outputs\":{},\"metadata\":{},\"resource\":{\"action\":\"create\",\"manifest\":\"apiVersion: machinelearning.seldon.io/v1\\nkind: SeldonDeployment\\nmetadata:\\n  name: \\\"sklearn\\\"\\n  namespace: default\\n  ownerReferences:\\n  - apiVersion: argoproj.io/v1alpha1\\n    blockOwnerDeletion: true\\n    kind: Workflow\\n    name: \\\"seldon-benchmark-process\\\"\\n    uid: \\\"e472d69d-44ed-4a45-86b3-d4b64146002b\\\"\\nspec:\\n  name: \\\"sklearn\\\"\\n  transport: grpc\\n  predictors:\\n    - componentSpecs:\\n      - spec:\\n        containers:\\n        - name: classifier\\n          env:\\n          - name: GUNICORN_THREADS\\n            value: 10\\n          - name: GUNICORN_WORKERS\\n            value: 1\\n      graph:\\n        children: []\\n        implementation: SKLEARN_SERVER\\n        modelUri: gs://seldon-models/sklearn/iris\\n        name: classifier\\n      name: default\\n      replicas: 1\\n\"}}"
    [35mcreate-seldon-resource[0m:	time="2020-08-07T17:22:39.450Z" level=info msg="Loading manifest to /tmp/manifest.yaml"
    [35mcreate-seldon-resource[0m:	time="2020-08-07T17:22:39.450Z" level=info msg="kubectl create -f /tmp/manifest.yaml -o json"
    [35mcreate-seldon-resource[0m:	time="2020-08-07T17:22:40.060Z" level=info msg=default/SeldonDeployment.machinelearning.seldon.io/sklearn
    [35mcreate-seldon-resource[0m:	time="2020-08-07T17:22:40.060Z" level=info msg="No output parameters"
    [32mwait-seldon-resource[0m:	Waiting for deployment "sklearn-default-0-classifier" rollout to finish: 0 of 1 updated replicas are available...
    [32mwait-seldon-resource[0m:	deployment "sklearn-default-0-classifier" successfully rolled out
    [35mrun-benchmark[0m:	{"date":"2020-08-07T17:25:09Z","endReason":"timeout","options":{"host":"istio-ingressgateway.istio-system.svc.cluster.local:80","proto":"/proto/prediction.proto","import-paths":["/proto","."],"call":"seldon.protos.Seldon/Predict","insecure":true,"total":2147483647,"concurrency":50,"connections":1,"duration":120000000000,"timeout":20000000000,"dial-timeout":10000000000,"data":{"data":{"ndarray":[[0,1,2,3]]}},"binary":false,"metadata":{"namespace":"default","seldon":"sklearn"},"CPUs":4},"count":88874,"total":120001033613,"average":67376309,"fastest":21863600,"slowest":148816057,"rps":740.6102874631579,"errorDistribution":{"rpc error: code = Unavailable desc = transport is closing":50},"statusCodeDistribution":{"OK":88824,"Unavailable":50},"latencyDistribution":[{"percentage":10,"latency":54583101},{"percentage":25,"latency":59326600},{"percentage":50,"latency":65257398},{"percentage":75,"latency":73167799},{"percentage":90,"latency":82939600},{"percentage":95,"latency":89598800},{"percentage":99,"latency":101463001}]}



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
    	mean: 67.376309 ms
    	50th: 65.257398 ms
    	90th: 82.9396 ms
    	95th: 89.5988 ms
    	99th: 101.463001 ms
    
    Rate: 740.6102874631579/s
    Errors: True
    Errors: {'OK': 88824, 'Unavailable': 50}



```python
!argo delete seldon-benchmark-process
```

    Workflow 'seldon-benchmark-process' deleted


## Run a set of tests

We can now leverage the helm charts we created above to run a grid search on a set of parameters.


```python
import itertools as it
import json
import time

grid_opts = {
    "A-replicas": [1, 3],
    "B-serverWorkers": [1, 4],
    "C-serverThreads": [50, 200],
    "D-apiType": ["rest", "grpc"],
    "E-cpus": [1, 4],
    "F-maxWorkers": [100, 300],
    "G-useEngine": ["true", "false"],
}

allNames = sorted(grid_opts)
combinations = it.product(*(grid_opts[Name] for Name in allNames))
all_results = []
for curr_values in combinations:
    print("VALUES:", curr_values)
    replicas, server_workers, server_threads, api_type, cpus, max_wokers, use_engine = curr_values

    # For some reason python vars don't work with multiline helm charts
    %env REPLICAS=$replicas
    %env SERVER_WORKERS=$server_workers
    %env SERVER_THREADS=$server_threads
    %env API_TYPE=$api_type
    %env CPUS=$cpus
    %env MAX_WORKERS=$max_wokers
    %env USE_ENGINE=$use_engine
    
    !helm template seldon-benchmark-workflow helm-charts/seldon-benchmark-workflow/ \
        --set workflow.name=seldon-benchmark-process \
        --set seldonDeployment.name=sklearn \
        --set seldonDeployment.replicas=$REPLICAS \
        --set seldonDeployment.serverWorkers=$SERVER_WORKERS \
        --set seldonDeployment.serverThreads=$SERVER_THREADS \
        --set seldonDeployment.apiType=$API_TYPE \
        --set seldonDeployment.useEngine=\"$USE_ENGINE\" \
        --set benchmark.cpus=$CPUS \
        --set benchmark.maxWorkers=$MAX_WORKERS \
        --set benchmark.duration=120s \
        --set benchmark.rate=0 \
        --set benchmark.data='\{"data": {"ndarray": [[0\,1\,2\,3]]\}\}' \
        | argo submit --wait -
    
    !argo wait seldon-benchmark-process 
    
    wf_logs = !argo logs -w seldon-benchmark-process 
    wf_bench = wf_logs[-1]
    wf_json_str = wf_bench[24:]
    results = json.loads(wf_json_str)
    
    result = get_results(results)
    result["replicas"] = replicas
    result["server_workers"] = server_workers
    result["server_threads"] = server_threads
    result["apiType"] = api_type
    result["cpus"] = cpus
    result["max_wokers"] = max_wokers
    result["use_engine"] = use_engine
    all_results.append(result)
    
    !argo delete seldon-benchmark-process
    time.sleep(1)
    print("\n\n")
    
```

## Deeper Analysis
Now that we have all the parameters, we can do a deeper analysis


```python
import pandas as pd
df = pd.DataFrame.from_dict(results)
df.head()
```




<div>
<style scoped>
    .dataframe tbody tr th:only-of-type {
        vertical-align: middle;
    }

    .dataframe tbody tr th {
        vertical-align: top;
    }

    .dataframe thead th {
        text-align: right;
    }
</style>
<table border="1" class="dataframe">
  <thead>
    <tr style="text-align: right;">
      <th></th>
      <th>replicas</th>
      <th>server_workers</th>
      <th>server_threads</th>
      <th>apiType</th>
      <th>cpus</th>
      <th>max_wokers</th>
      <th>use_engine</th>
      <th>mean</th>
      <th>50th</th>
      <th>90th</th>
      <th>95th</th>
      <th>99th</th>
      <th>rate</th>
      <th>errors</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <th>0</th>
      <td>1</td>
      <td>1</td>
      <td>50</td>
      <td>rest</td>
      <td>1</td>
      <td>200</td>
      <td>true</td>
      <td>489.269344</td>
      <td>455.617128</td>
      <td>612.294382</td>
      <td>672.510108</td>
      <td>832.322767</td>
      <td>407.879172</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>1</th>
      <td>1</td>
      <td>1</td>
      <td>50</td>
      <td>rest</td>
      <td>1</td>
      <td>200</td>
      <td>false</td>
      <td>529.767457</td>
      <td>514.151876</td>
      <td>591.278115</td>
      <td>621.463805</td>
      <td>749.348556</td>
      <td>376.649458</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>2</th>
      <td>1</td>
      <td>1</td>
      <td>50</td>
      <td>rest</td>
      <td>4</td>
      <td>200</td>
      <td>true</td>
      <td>547.618426</td>
      <td>526.472215</td>
      <td>661.947413</td>
      <td>720.039676</td>
      <td>863.596098</td>
      <td>364.363839</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>3</th>
      <td>1</td>
      <td>1</td>
      <td>50</td>
      <td>rest</td>
      <td>4</td>
      <td>200</td>
      <td>false</td>
      <td>593.880113</td>
      <td>602.945695</td>
      <td>737.993290</td>
      <td>770.777543</td>
      <td>1003.510371</td>
      <td>336.075411</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>4</th>
      <td>1</td>
      <td>1</td>
      <td>50</td>
      <td>grpc</td>
      <td>1</td>
      <td>200</td>
      <td>true</td>
      <td>95.322943</td>
      <td>97.896699</td>
      <td>117.221999</td>
      <td>125.852400</td>
      <td>141.615501</td>
      <td>523.628160</td>
      <td>{'OK': 62790, 'Unavailable': 50}</td>
    </tr>
  </tbody>
</table>
</div>



### GRPC as expected outperforms REST


```python
df.sort_values("rate", ascending=False)
```




<div>
<style scoped>
    .dataframe tbody tr th:only-of-type {
        vertical-align: middle;
    }

    .dataframe tbody tr th {
        vertical-align: top;
    }

    .dataframe thead th {
        text-align: right;
    }
</style>
<table border="1" class="dataframe">
  <thead>
    <tr style="text-align: right;">
      <th></th>
      <th>replicas</th>
      <th>server_workers</th>
      <th>server_threads</th>
      <th>apiType</th>
      <th>cpus</th>
      <th>max_wokers</th>
      <th>use_engine</th>
      <th>mean</th>
      <th>50th</th>
      <th>90th</th>
      <th>95th</th>
      <th>99th</th>
      <th>rate</th>
      <th>errors</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <th>60</th>
      <td>3</td>
      <td>4</td>
      <td>200</td>
      <td>grpc</td>
      <td>1</td>
      <td>200</td>
      <td>true</td>
      <td>31.389861</td>
      <td>23.769589</td>
      <td>71.583795</td>
      <td>78.881398</td>
      <td>91.312797</td>
      <td>1586.593680</td>
      <td>{'OK': 190361, 'Unavailable': 48}</td>
    </tr>
    <tr>
      <th>52</th>
      <td>3</td>
      <td>4</td>
      <td>50</td>
      <td>grpc</td>
      <td>1</td>
      <td>200</td>
      <td>true</td>
      <td>31.398451</td>
      <td>26.313000</td>
      <td>64.841515</td>
      <td>73.035800</td>
      <td>88.744198</td>
      <td>1586.555365</td>
      <td>{'OK': 190333, 'Unavailable': 71}</td>
    </tr>
    <tr>
      <th>45</th>
      <td>3</td>
      <td>1</td>
      <td>200</td>
      <td>grpc</td>
      <td>1</td>
      <td>200</td>
      <td>false</td>
      <td>32.191240</td>
      <td>30.448302</td>
      <td>60.616301</td>
      <td>68.724406</td>
      <td>91.484308</td>
      <td>1547.003054</td>
      <td>{'OK': 185606, 'Unavailable': 49}</td>
    </tr>
    <tr>
      <th>61</th>
      <td>3</td>
      <td>4</td>
      <td>200</td>
      <td>grpc</td>
      <td>1</td>
      <td>200</td>
      <td>false</td>
      <td>32.727674</td>
      <td>28.483400</td>
      <td>63.750796</td>
      <td>72.597310</td>
      <td>90.693812</td>
      <td>1521.590875</td>
      <td>{'OK': 182555, 'Unavailable': 49}</td>
    </tr>
    <tr>
      <th>55</th>
      <td>3</td>
      <td>4</td>
      <td>50</td>
      <td>grpc</td>
      <td>4</td>
      <td>200</td>
      <td>false</td>
      <td>33.629848</td>
      <td>29.610701</td>
      <td>67.065895</td>
      <td>77.773100</td>
      <td>97.296599</td>
      <td>1479.320474</td>
      <td>{'OK': 177471, 'Unavailable': 50}</td>
    </tr>
    <tr>
      <th>...</th>
      <td>...</td>
      <td>...</td>
      <td>...</td>
      <td>...</td>
      <td>...</td>
      <td>...</td>
      <td>...</td>
      <td>...</td>
      <td>...</td>
      <td>...</td>
      <td>...</td>
      <td>...</td>
      <td>...</td>
      <td>...</td>
    </tr>
    <tr>
      <th>10</th>
      <td>1</td>
      <td>1</td>
      <td>200</td>
      <td>rest</td>
      <td>4</td>
      <td>200</td>
      <td>true</td>
      <td>571.452398</td>
      <td>556.699256</td>
      <td>693.093315</td>
      <td>751.197598</td>
      <td>1024.233714</td>
      <td>348.889260</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>11</th>
      <td>1</td>
      <td>1</td>
      <td>200</td>
      <td>rest</td>
      <td>4</td>
      <td>200</td>
      <td>false</td>
      <td>587.900216</td>
      <td>556.869872</td>
      <td>723.744376</td>
      <td>774.244702</td>
      <td>939.994423</td>
      <td>339.396160</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>3</th>
      <td>1</td>
      <td>1</td>
      <td>50</td>
      <td>rest</td>
      <td>4</td>
      <td>200</td>
      <td>false</td>
      <td>593.880113</td>
      <td>602.945695</td>
      <td>737.993290</td>
      <td>770.777543</td>
      <td>1003.510371</td>
      <td>336.075411</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>8</th>
      <td>1</td>
      <td>1</td>
      <td>200</td>
      <td>rest</td>
      <td>1</td>
      <td>200</td>
      <td>true</td>
      <td>633.043624</td>
      <td>617.853285</td>
      <td>741.229073</td>
      <td>776.560578</td>
      <td>1846.623159</td>
      <td>314.908167</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>9</th>
      <td>1</td>
      <td>1</td>
      <td>200</td>
      <td>rest</td>
      <td>1</td>
      <td>200</td>
      <td>false</td>
      <td>641.530606</td>
      <td>653.922529</td>
      <td>802.558303</td>
      <td>847.414484</td>
      <td>1570.484029</td>
      <td>310.839312</td>
      <td>[]</td>
    </tr>
  </tbody>
</table>
<p>64 rows × 14 columns</p>
</div>



### Deeper dive REST
As expected replicas has the biggest impact. It seems the parameters on the benchmark worker don't seem to affect throughput.


```python
df[df["apiType"]=="rest"].sort_values("rate", ascending=False)
```




<div>
<style scoped>
    .dataframe tbody tr th:only-of-type {
        vertical-align: middle;
    }

    .dataframe tbody tr th {
        vertical-align: top;
    }

    .dataframe thead th {
        text-align: right;
    }
</style>
<table border="1" class="dataframe">
  <thead>
    <tr style="text-align: right;">
      <th></th>
      <th>replicas</th>
      <th>server_workers</th>
      <th>server_threads</th>
      <th>apiType</th>
      <th>cpus</th>
      <th>max_wokers</th>
      <th>use_engine</th>
      <th>mean</th>
      <th>50th</th>
      <th>90th</th>
      <th>95th</th>
      <th>99th</th>
      <th>rate</th>
      <th>errors</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <th>41</th>
      <td>3</td>
      <td>1</td>
      <td>200</td>
      <td>rest</td>
      <td>1</td>
      <td>200</td>
      <td>false</td>
      <td>201.167546</td>
      <td>8.844305</td>
      <td>629.250888</td>
      <td>690.807158</td>
      <td>809.635194</td>
      <td>992.298652</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>48</th>
      <td>3</td>
      <td>4</td>
      <td>50</td>
      <td>rest</td>
      <td>1</td>
      <td>200</td>
      <td>true</td>
      <td>208.429576</td>
      <td>11.377699</td>
      <td>655.466848</td>
      <td>685.265506</td>
      <td>758.664504</td>
      <td>957.846772</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>51</th>
      <td>3</td>
      <td>4</td>
      <td>50</td>
      <td>rest</td>
      <td>4</td>
      <td>200</td>
      <td>false</td>
      <td>211.228526</td>
      <td>13.592301</td>
      <td>641.484819</td>
      <td>675.713639</td>
      <td>795.682869</td>
      <td>945.090980</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>59</th>
      <td>3</td>
      <td>4</td>
      <td>200</td>
      <td>rest</td>
      <td>4</td>
      <td>200</td>
      <td>false</td>
      <td>214.358834</td>
      <td>13.573121</td>
      <td>670.449768</td>
      <td>690.048496</td>
      <td>722.537613</td>
      <td>930.694079</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>57</th>
      <td>3</td>
      <td>4</td>
      <td>200</td>
      <td>rest</td>
      <td>1</td>
      <td>200</td>
      <td>false</td>
      <td>216.646320</td>
      <td>9.336961</td>
      <td>684.733598</td>
      <td>704.485018</td>
      <td>733.636276</td>
      <td>921.350903</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>40</th>
      <td>3</td>
      <td>1</td>
      <td>200</td>
      <td>rest</td>
      <td>1</td>
      <td>200</td>
      <td>true</td>
      <td>217.722397</td>
      <td>16.593757</td>
      <td>657.144743</td>
      <td>695.158232</td>
      <td>745.726065</td>
      <td>916.803160</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>32</th>
      <td>3</td>
      <td>1</td>
      <td>50</td>
      <td>rest</td>
      <td>1</td>
      <td>200</td>
      <td>true</td>
      <td>218.817952</td>
      <td>10.808913</td>
      <td>689.809571</td>
      <td>757.737985</td>
      <td>867.650689</td>
      <td>912.589694</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>56</th>
      <td>3</td>
      <td>4</td>
      <td>200</td>
      <td>rest</td>
      <td>1</td>
      <td>200</td>
      <td>true</td>
      <td>221.031876</td>
      <td>9.197338</td>
      <td>690.217169</td>
      <td>711.800471</td>
      <td>742.657817</td>
      <td>903.072311</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>50</th>
      <td>3</td>
      <td>4</td>
      <td>50</td>
      <td>rest</td>
      <td>4</td>
      <td>200</td>
      <td>true</td>
      <td>221.263249</td>
      <td>16.583482</td>
      <td>688.637696</td>
      <td>711.870214</td>
      <td>781.197685</td>
      <td>902.315850</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>58</th>
      <td>3</td>
      <td>4</td>
      <td>200</td>
      <td>rest</td>
      <td>4</td>
      <td>200</td>
      <td>true</td>
      <td>221.566956</td>
      <td>11.037262</td>
      <td>685.417461</td>
      <td>713.923684</td>
      <td>771.814053</td>
      <td>901.132352</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>35</th>
      <td>3</td>
      <td>1</td>
      <td>50</td>
      <td>rest</td>
      <td>4</td>
      <td>200</td>
      <td>false</td>
      <td>225.719114</td>
      <td>15.998348</td>
      <td>704.701196</td>
      <td>741.890962</td>
      <td>852.664830</td>
      <td>884.187996</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>33</th>
      <td>3</td>
      <td>1</td>
      <td>50</td>
      <td>rest</td>
      <td>1</td>
      <td>200</td>
      <td>false</td>
      <td>229.653366</td>
      <td>9.844413</td>
      <td>725.066803</td>
      <td>775.186525</td>
      <td>857.762245</td>
      <td>869.461119</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>42</th>
      <td>3</td>
      <td>1</td>
      <td>200</td>
      <td>rest</td>
      <td>4</td>
      <td>200</td>
      <td>true</td>
      <td>231.016536</td>
      <td>15.829218</td>
      <td>737.382688</td>
      <td>788.027859</td>
      <td>885.482116</td>
      <td>863.960992</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>49</th>
      <td>3</td>
      <td>4</td>
      <td>50</td>
      <td>rest</td>
      <td>1</td>
      <td>200</td>
      <td>false</td>
      <td>231.986927</td>
      <td>11.193407</td>
      <td>702.083677</td>
      <td>769.889421</td>
      <td>901.360146</td>
      <td>860.495277</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>43</th>
      <td>3</td>
      <td>1</td>
      <td>200</td>
      <td>rest</td>
      <td>4</td>
      <td>200</td>
      <td>false</td>
      <td>239.150794</td>
      <td>14.147647</td>
      <td>722.982655</td>
      <td>789.211063</td>
      <td>929.436195</td>
      <td>834.381347</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>34</th>
      <td>3</td>
      <td>1</td>
      <td>50</td>
      <td>rest</td>
      <td>4</td>
      <td>200</td>
      <td>true</td>
      <td>240.088790</td>
      <td>121.078205</td>
      <td>707.862815</td>
      <td>771.405571</td>
      <td>965.932529</td>
      <td>831.402721</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>26</th>
      <td>1</td>
      <td>4</td>
      <td>200</td>
      <td>rest</td>
      <td>4</td>
      <td>200</td>
      <td>true</td>
      <td>413.608259</td>
      <td>409.729690</td>
      <td>442.576049</td>
      <td>460.804621</td>
      <td>502.762769</td>
      <td>482.699096</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>17</th>
      <td>1</td>
      <td>4</td>
      <td>50</td>
      <td>rest</td>
      <td>1</td>
      <td>200</td>
      <td>false</td>
      <td>429.042835</td>
      <td>412.423403</td>
      <td>500.170846</td>
      <td>522.423418</td>
      <td>586.685379</td>
      <td>465.431891</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>27</th>
      <td>1</td>
      <td>4</td>
      <td>200</td>
      <td>rest</td>
      <td>4</td>
      <td>200</td>
      <td>false</td>
      <td>432.609142</td>
      <td>426.606234</td>
      <td>488.443435</td>
      <td>512.393140</td>
      <td>556.238288</td>
      <td>461.578501</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>25</th>
      <td>1</td>
      <td>4</td>
      <td>200</td>
      <td>rest</td>
      <td>1</td>
      <td>200</td>
      <td>false</td>
      <td>463.422714</td>
      <td>450.181537</td>
      <td>551.644801</td>
      <td>602.270942</td>
      <td>670.647806</td>
      <td>430.891782</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>16</th>
      <td>1</td>
      <td>4</td>
      <td>50</td>
      <td>rest</td>
      <td>1</td>
      <td>200</td>
      <td>true</td>
      <td>475.510231</td>
      <td>456.056479</td>
      <td>583.716159</td>
      <td>650.365364</td>
      <td>746.791628</td>
      <td>419.975983</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>19</th>
      <td>1</td>
      <td>4</td>
      <td>50</td>
      <td>rest</td>
      <td>4</td>
      <td>200</td>
      <td>false</td>
      <td>481.143061</td>
      <td>450.734477</td>
      <td>602.026223</td>
      <td>689.302618</td>
      <td>863.072782</td>
      <td>414.795159</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>18</th>
      <td>1</td>
      <td>4</td>
      <td>50</td>
      <td>rest</td>
      <td>4</td>
      <td>200</td>
      <td>true</td>
      <td>488.185779</td>
      <td>436.842244</td>
      <td>628.922397</td>
      <td>735.512654</td>
      <td>1068.474298</td>
      <td>408.992844</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>0</th>
      <td>1</td>
      <td>1</td>
      <td>50</td>
      <td>rest</td>
      <td>1</td>
      <td>200</td>
      <td>true</td>
      <td>489.269344</td>
      <td>455.617128</td>
      <td>612.294382</td>
      <td>672.510108</td>
      <td>832.322767</td>
      <td>407.879172</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>24</th>
      <td>1</td>
      <td>4</td>
      <td>200</td>
      <td>rest</td>
      <td>1</td>
      <td>200</td>
      <td>true</td>
      <td>514.472545</td>
      <td>488.358257</td>
      <td>591.629431</td>
      <td>631.392813</td>
      <td>1517.062374</td>
      <td>387.882855</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>1</th>
      <td>1</td>
      <td>1</td>
      <td>50</td>
      <td>rest</td>
      <td>1</td>
      <td>200</td>
      <td>false</td>
      <td>529.767457</td>
      <td>514.151876</td>
      <td>591.278115</td>
      <td>621.463805</td>
      <td>749.348556</td>
      <td>376.649458</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>2</th>
      <td>1</td>
      <td>1</td>
      <td>50</td>
      <td>rest</td>
      <td>4</td>
      <td>200</td>
      <td>true</td>
      <td>547.618426</td>
      <td>526.472215</td>
      <td>661.947413</td>
      <td>720.039676</td>
      <td>863.596098</td>
      <td>364.363839</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>10</th>
      <td>1</td>
      <td>1</td>
      <td>200</td>
      <td>rest</td>
      <td>4</td>
      <td>200</td>
      <td>true</td>
      <td>571.452398</td>
      <td>556.699256</td>
      <td>693.093315</td>
      <td>751.197598</td>
      <td>1024.233714</td>
      <td>348.889260</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>11</th>
      <td>1</td>
      <td>1</td>
      <td>200</td>
      <td>rest</td>
      <td>4</td>
      <td>200</td>
      <td>false</td>
      <td>587.900216</td>
      <td>556.869872</td>
      <td>723.744376</td>
      <td>774.244702</td>
      <td>939.994423</td>
      <td>339.396160</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>3</th>
      <td>1</td>
      <td>1</td>
      <td>50</td>
      <td>rest</td>
      <td>4</td>
      <td>200</td>
      <td>false</td>
      <td>593.880113</td>
      <td>602.945695</td>
      <td>737.993290</td>
      <td>770.777543</td>
      <td>1003.510371</td>
      <td>336.075411</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>8</th>
      <td>1</td>
      <td>1</td>
      <td>200</td>
      <td>rest</td>
      <td>1</td>
      <td>200</td>
      <td>true</td>
      <td>633.043624</td>
      <td>617.853285</td>
      <td>741.229073</td>
      <td>776.560578</td>
      <td>1846.623159</td>
      <td>314.908167</td>
      <td>[]</td>
    </tr>
    <tr>
      <th>9</th>
      <td>1</td>
      <td>1</td>
      <td>200</td>
      <td>rest</td>
      <td>1</td>
      <td>200</td>
      <td>false</td>
      <td>641.530606</td>
      <td>653.922529</td>
      <td>802.558303</td>
      <td>847.414484</td>
      <td>1570.484029</td>
      <td>310.839312</td>
      <td>[]</td>
    </tr>
  </tbody>
</table>
</div>



### Deep dive on GRPC


```python
df[df["apiType"]=="grpc"].sort_values("rate", ascending=False)
```




<div>
<style scoped>
    .dataframe tbody tr th:only-of-type {
        vertical-align: middle;
    }

    .dataframe tbody tr th {
        vertical-align: top;
    }

    .dataframe thead th {
        text-align: right;
    }
</style>
<table border="1" class="dataframe">
  <thead>
    <tr style="text-align: right;">
      <th></th>
      <th>replicas</th>
      <th>server_workers</th>
      <th>server_threads</th>
      <th>apiType</th>
      <th>cpus</th>
      <th>max_wokers</th>
      <th>use_engine</th>
      <th>mean</th>
      <th>50th</th>
      <th>90th</th>
      <th>95th</th>
      <th>99th</th>
      <th>rate</th>
      <th>errors</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <th>60</th>
      <td>3</td>
      <td>4</td>
      <td>200</td>
      <td>grpc</td>
      <td>1</td>
      <td>200</td>
      <td>true</td>
      <td>31.389861</td>
      <td>23.769589</td>
      <td>71.583795</td>
      <td>78.881398</td>
      <td>91.312797</td>
      <td>1586.593680</td>
      <td>{'OK': 190361, 'Unavailable': 48}</td>
    </tr>
    <tr>
      <th>52</th>
      <td>3</td>
      <td>4</td>
      <td>50</td>
      <td>grpc</td>
      <td>1</td>
      <td>200</td>
      <td>true</td>
      <td>31.398451</td>
      <td>26.313000</td>
      <td>64.841515</td>
      <td>73.035800</td>
      <td>88.744198</td>
      <td>1586.555365</td>
      <td>{'OK': 190333, 'Unavailable': 71}</td>
    </tr>
    <tr>
      <th>45</th>
      <td>3</td>
      <td>1</td>
      <td>200</td>
      <td>grpc</td>
      <td>1</td>
      <td>200</td>
      <td>false</td>
      <td>32.191240</td>
      <td>30.448302</td>
      <td>60.616301</td>
      <td>68.724406</td>
      <td>91.484308</td>
      <td>1547.003054</td>
      <td>{'OK': 185606, 'Unavailable': 49}</td>
    </tr>
    <tr>
      <th>61</th>
      <td>3</td>
      <td>4</td>
      <td>200</td>
      <td>grpc</td>
      <td>1</td>
      <td>200</td>
      <td>false</td>
      <td>32.727674</td>
      <td>28.483400</td>
      <td>63.750796</td>
      <td>72.597310</td>
      <td>90.693812</td>
      <td>1521.590875</td>
      <td>{'OK': 182555, 'Unavailable': 49}</td>
    </tr>
    <tr>
      <th>55</th>
      <td>3</td>
      <td>4</td>
      <td>50</td>
      <td>grpc</td>
      <td>4</td>
      <td>200</td>
      <td>false</td>
      <td>33.629848</td>
      <td>29.610701</td>
      <td>67.065895</td>
      <td>77.773100</td>
      <td>97.296599</td>
      <td>1479.320474</td>
      <td>{'OK': 177471, 'Unavailable': 50}</td>
    </tr>
    <tr>
      <th>47</th>
      <td>3</td>
      <td>1</td>
      <td>200</td>
      <td>grpc</td>
      <td>4</td>
      <td>200</td>
      <td>false</td>
      <td>33.861023</td>
      <td>30.207400</td>
      <td>70.272698</td>
      <td>83.485103</td>
      <td>105.639301</td>
      <td>1469.503585</td>
      <td>{'OK': 176302, 'Unavailable': 50}</td>
    </tr>
    <tr>
      <th>62</th>
      <td>3</td>
      <td>4</td>
      <td>200</td>
      <td>grpc</td>
      <td>4</td>
      <td>200</td>
      <td>true</td>
      <td>34.746801</td>
      <td>31.896585</td>
      <td>72.732796</td>
      <td>84.032763</td>
      <td>99.433090</td>
      <td>1432.045405</td>
      <td>{'OK': 171799, 'Unavailable': 50}</td>
    </tr>
    <tr>
      <th>54</th>
      <td>3</td>
      <td>4</td>
      <td>50</td>
      <td>grpc</td>
      <td>4</td>
      <td>200</td>
      <td>true</td>
      <td>34.786883</td>
      <td>32.141197</td>
      <td>72.554313</td>
      <td>82.649702</td>
      <td>95.049705</td>
      <td>1430.209225</td>
      <td>{'OK': 171578, 'Unavailable': 49}</td>
    </tr>
    <tr>
      <th>37</th>
      <td>3</td>
      <td>1</td>
      <td>50</td>
      <td>grpc</td>
      <td>1</td>
      <td>200</td>
      <td>false</td>
      <td>35.149376</td>
      <td>35.153187</td>
      <td>62.842800</td>
      <td>72.791800</td>
      <td>94.240299</td>
      <td>1416.745392</td>
      <td>{'OK': 169973, 'Unavailable': 50}</td>
    </tr>
    <tr>
      <th>36</th>
      <td>3</td>
      <td>1</td>
      <td>50</td>
      <td>grpc</td>
      <td>1</td>
      <td>200</td>
      <td>true</td>
      <td>35.167657</td>
      <td>31.859300</td>
      <td>65.644895</td>
      <td>76.240799</td>
      <td>98.925899</td>
      <td>1415.967279</td>
      <td>{'OK': 169880, 'Unavailable': 48}</td>
    </tr>
    <tr>
      <th>46</th>
      <td>3</td>
      <td>1</td>
      <td>200</td>
      <td>grpc</td>
      <td>4</td>
      <td>200</td>
      <td>true</td>
      <td>35.286173</td>
      <td>24.988500</td>
      <td>83.079301</td>
      <td>94.264796</td>
      <td>111.448895</td>
      <td>1410.595798</td>
      <td>{'OK': 169202, 'Unavailable': 71}</td>
    </tr>
    <tr>
      <th>53</th>
      <td>3</td>
      <td>4</td>
      <td>50</td>
      <td>grpc</td>
      <td>1</td>
      <td>200</td>
      <td>false</td>
      <td>35.543940</td>
      <td>30.528900</td>
      <td>69.449895</td>
      <td>82.465882</td>
      <td>100.381195</td>
      <td>1400.945365</td>
      <td>{'OK': 168074, 'Unavailable': 50}</td>
    </tr>
    <tr>
      <th>63</th>
      <td>3</td>
      <td>4</td>
      <td>200</td>
      <td>grpc</td>
      <td>4</td>
      <td>200</td>
      <td>false</td>
      <td>35.706181</td>
      <td>30.175300</td>
      <td>76.121701</td>
      <td>85.842385</td>
      <td>99.072701</td>
      <td>1393.469861</td>
      <td>{'OK': 167180, 'Unavailable': 49}</td>
    </tr>
    <tr>
      <th>39</th>
      <td>3</td>
      <td>1</td>
      <td>50</td>
      <td>grpc</td>
      <td>4</td>
      <td>200</td>
      <td>false</td>
      <td>36.026804</td>
      <td>33.541192</td>
      <td>69.942798</td>
      <td>81.321704</td>
      <td>108.528901</td>
      <td>1381.482907</td>
      <td>{'OK': 165711, 'Unavailable': 69}</td>
    </tr>
    <tr>
      <th>38</th>
      <td>3</td>
      <td>1</td>
      <td>50</td>
      <td>grpc</td>
      <td>4</td>
      <td>200</td>
      <td>true</td>
      <td>36.325718</td>
      <td>35.598498</td>
      <td>73.211997</td>
      <td>82.948302</td>
      <td>102.248397</td>
      <td>1369.739820</td>
      <td>{'OK': 164333, 'Unavailable': 49}</td>
    </tr>
    <tr>
      <th>44</th>
      <td>3</td>
      <td>1</td>
      <td>200</td>
      <td>grpc</td>
      <td>1</td>
      <td>200</td>
      <td>true</td>
      <td>37.326561</td>
      <td>35.609388</td>
      <td>70.522598</td>
      <td>79.731401</td>
      <td>101.297400</td>
      <td>1334.058278</td>
      <td>{'OK': 160053, 'Unavailable': 50}</td>
    </tr>
    <tr>
      <th>29</th>
      <td>1</td>
      <td>4</td>
      <td>200</td>
      <td>grpc</td>
      <td>1</td>
      <td>200</td>
      <td>false</td>
      <td>63.240129</td>
      <td>61.519201</td>
      <td>72.905000</td>
      <td>77.140700</td>
      <td>89.520499</td>
      <td>789.347786</td>
      <td>{'OK': 94678, 'Unavailable': 50}</td>
    </tr>
    <tr>
      <th>28</th>
      <td>1</td>
      <td>4</td>
      <td>200</td>
      <td>grpc</td>
      <td>1</td>
      <td>200</td>
      <td>true</td>
      <td>63.537119</td>
      <td>61.855200</td>
      <td>74.299100</td>
      <td>79.876601</td>
      <td>97.179900</td>
      <td>785.631011</td>
      <td>{'OK': 94233, 'Unavailable': 50}</td>
    </tr>
    <tr>
      <th>20</th>
      <td>1</td>
      <td>4</td>
      <td>50</td>
      <td>grpc</td>
      <td>1</td>
      <td>200</td>
      <td>true</td>
      <td>65.711577</td>
      <td>64.220500</td>
      <td>78.085300</td>
      <td>83.563600</td>
      <td>94.907700</td>
      <td>759.690398</td>
      <td>{'OK': 91119, 'Unavailable': 50}</td>
    </tr>
    <tr>
      <th>21</th>
      <td>1</td>
      <td>4</td>
      <td>50</td>
      <td>grpc</td>
      <td>1</td>
      <td>200</td>
      <td>false</td>
      <td>66.898143</td>
      <td>63.420800</td>
      <td>83.837100</td>
      <td>92.332400</td>
      <td>108.138499</td>
      <td>746.209307</td>
      <td>{'OK': 89501, 'Unavailable': 50}</td>
    </tr>
    <tr>
      <th>30</th>
      <td>1</td>
      <td>4</td>
      <td>200</td>
      <td>grpc</td>
      <td>4</td>
      <td>200</td>
      <td>true</td>
      <td>67.211609</td>
      <td>65.504200</td>
      <td>79.989899</td>
      <td>86.808200</td>
      <td>106.460500</td>
      <td>742.433252</td>
      <td>{'OK': 89044, 'Unavailable': 50}</td>
    </tr>
    <tr>
      <th>7</th>
      <td>1</td>
      <td>1</td>
      <td>50</td>
      <td>grpc</td>
      <td>4</td>
      <td>200</td>
      <td>false</td>
      <td>67.770632</td>
      <td>62.168504</td>
      <td>88.674303</td>
      <td>102.537000</td>
      <td>120.848185</td>
      <td>736.385539</td>
      <td>{'OK': 88318, 'Unavailable': 49}</td>
    </tr>
    <tr>
      <th>31</th>
      <td>1</td>
      <td>4</td>
      <td>200</td>
      <td>grpc</td>
      <td>4</td>
      <td>200</td>
      <td>false</td>
      <td>70.577834</td>
      <td>68.972899</td>
      <td>84.869200</td>
      <td>89.875600</td>
      <td>102.761897</td>
      <td>707.046156</td>
      <td>{'OK': 84796, 'Unavailable': 50}</td>
    </tr>
    <tr>
      <th>22</th>
      <td>1</td>
      <td>4</td>
      <td>50</td>
      <td>grpc</td>
      <td>4</td>
      <td>200</td>
      <td>true</td>
      <td>70.818411</td>
      <td>67.591600</td>
      <td>87.914104</td>
      <td>97.004000</td>
      <td>115.388900</td>
      <td>704.647865</td>
      <td>{'OK': 84514, 'Unavailable': 50}</td>
    </tr>
    <tr>
      <th>15</th>
      <td>1</td>
      <td>1</td>
      <td>200</td>
      <td>grpc</td>
      <td>4</td>
      <td>200</td>
      <td>false</td>
      <td>71.571627</td>
      <td>69.348700</td>
      <td>91.609598</td>
      <td>98.471998</td>
      <td>111.237797</td>
      <td>697.252435</td>
      <td>{'OK': 83622, 'Unavailable': 50}</td>
    </tr>
    <tr>
      <th>23</th>
      <td>1</td>
      <td>4</td>
      <td>50</td>
      <td>grpc</td>
      <td>4</td>
      <td>200</td>
      <td>false</td>
      <td>73.853780</td>
      <td>70.604701</td>
      <td>91.031400</td>
      <td>98.064600</td>
      <td>116.658902</td>
      <td>675.704389</td>
      <td>{'OK': 81035, 'Unavailable': 50}</td>
    </tr>
    <tr>
      <th>14</th>
      <td>1</td>
      <td>1</td>
      <td>200</td>
      <td>grpc</td>
      <td>4</td>
      <td>200</td>
      <td>true</td>
      <td>89.662500</td>
      <td>87.678702</td>
      <td>107.762199</td>
      <td>118.226099</td>
      <td>146.838610</td>
      <td>556.478774</td>
      <td>{'OK': 66728, 'Unavailable': 50}</td>
    </tr>
    <tr>
      <th>6</th>
      <td>1</td>
      <td>1</td>
      <td>50</td>
      <td>grpc</td>
      <td>4</td>
      <td>200</td>
      <td>true</td>
      <td>90.655025</td>
      <td>91.964500</td>
      <td>108.453597</td>
      <td>116.581800</td>
      <td>148.048199</td>
      <td>550.406903</td>
      <td>{'OK': 66003, 'Unavailable': 50}</td>
    </tr>
    <tr>
      <th>5</th>
      <td>1</td>
      <td>1</td>
      <td>50</td>
      <td>grpc</td>
      <td>1</td>
      <td>200</td>
      <td>false</td>
      <td>92.930400</td>
      <td>93.020601</td>
      <td>113.056104</td>
      <td>122.476104</td>
      <td>150.119004</td>
      <td>537.076992</td>
      <td>{'OK': 64405, 'Unavailable': 50}</td>
    </tr>
    <tr>
      <th>12</th>
      <td>1</td>
      <td>1</td>
      <td>200</td>
      <td>grpc</td>
      <td>1</td>
      <td>200</td>
      <td>true</td>
      <td>94.695951</td>
      <td>94.988002</td>
      <td>111.319799</td>
      <td>118.210000</td>
      <td>134.270997</td>
      <td>527.054914</td>
      <td>{'OK': 63202, 'Unavailable': 50}</td>
    </tr>
    <tr>
      <th>4</th>
      <td>1</td>
      <td>1</td>
      <td>50</td>
      <td>grpc</td>
      <td>1</td>
      <td>200</td>
      <td>true</td>
      <td>95.322943</td>
      <td>97.896699</td>
      <td>117.221999</td>
      <td>125.852400</td>
      <td>141.615501</td>
      <td>523.628160</td>
      <td>{'OK': 62790, 'Unavailable': 50}</td>
    </tr>
    <tr>
      <th>13</th>
      <td>1</td>
      <td>1</td>
      <td>200</td>
      <td>grpc</td>
      <td>1</td>
      <td>200</td>
      <td>false</td>
      <td>96.016296</td>
      <td>97.410200</td>
      <td>113.779899</td>
      <td>120.184499</td>
      <td>136.929395</td>
      <td>519.810588</td>
      <td>{'OK': 62332, 'Unavailable': 50}</td>
    </tr>
  </tbody>
</table>
</div>




```python

```
