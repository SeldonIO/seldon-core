# Benchmarking with Argo Worfklows & Vegeta

In this notebook we will dive into how you can run bench marking with batch processing with Argo Workflows, Seldon Core and Vegeta.

Dependencies:

* Seldon core installed as per the docs with Istio as an ingress 
* Argo Workfklows installed in cluster (and argo CLI for commands)


## Setup

### Install Seldon Core
Use the notebook to [set-up Seldon Core with Ambassador or Istio Ingress](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html).

Note: If running with KIND you need to make sure do follow [these steps](https://github.com/argoproj/argo-workflows/issues/2376#issuecomment-595593237) as workaround to the `/.../docker.sock` known issue.


### Install Argo Workflows
You can follow the instructions from the official [Argo Workflows Documentation](https://github.com/argoproj/argo#quickstart).

Download the right CLi for your environment following the documentation (https://github.com/argoproj/argo-workflows/releases/tag/v3.0.8)

You also need to make sure that argo has permissions to create seldon deployments - for this you can just create a default-admin rolebinding as follows:

Set up the RBAC so the argo workflow is able to create seldon deployments.

Set up the configmap in order for it to work in KIND and other environments where Docker may not be thr main runtime (see https://github.com/argoproj/argo-workflows/issues/5243#issuecomment-792993742)

### Create Benchmark Argo Workflow

In order to create a benchmark, we created a simple argo workflow template so you can leverage the power of the helm charts.

Before we dive into the contents of the full helm chart, let's first give it a try with some of the settings.

We will run a batch job that will set up a Seldon Deployment with 1 replicas and 4 cpus (with 100 max workers) to send requests.


```python
!helm template seldon-benchmark-workflow ../../../helm-charts/seldon-benchmark-workflow/ \
    --set workflow.namespace=argo \
    --set workflow.name=seldon-benchmark-process \
    --set workflow.parallelism=2 \
    --set seldonDeployment.name=sklearn \
    --set seldonDeployment.replicas="1" \
    --set seldonDeployment.serverWorkers="5" \
    --set seldonDeployment.serverThreads=1 \
    --set seldonDeployment.modelUri="gs://seldon-models/v1.19.0-dev/sklearn/iris" \
    --set seldonDeployment.server="SKLEARN_SERVER" \
    --set seldonDeployment.apiType="rest|grpc" \
    --set seldonDeployment.requests.cpu="2000Mi" \
    --set seldonDeployment.limits.cpu="2000Mi" \
    --set seldonDeployment.disableOrchestrator="true|false" \
    --set benchmark.cpu="2" \
    --set benchmark.concurrency="1" \
    --set benchmark.duration="30s" \
    --set benchmark.rate=0 \
    --set benchmark.data='\{"data": {"ndarray": [[0\,1\,2\,3]]\}\}' \
    | argo submit -
```

    Name:                seldon-benchmark-process
    Namespace:           argo
    ServiceAccount:      default
    Status:              Pending
    Created:             Mon Jun 28 18:38:12 +0100 (now)
    Progress:            



```python
!argo list -n argo
```

    NAME                       STATUS    AGE   DURATION   PRIORITY
    seldon-benchmark-process   Running   20s   20s        0



```python
!argo logs -f seldon-benchmark-process -n argo
```

    [33mseldon-benchmark-process-635956972: [{"name": "sklearn-0", "replicas": "1", "serverWorkers": "5", "serverThreads": "1", "modelUri": "gs://seldon-models/sklearn/iris", "image": "", "server": "SKLEARN_SERVER", "apiType": "rest", "requestsCpu": "2000Mi", "requestsMemory": "100Mi", "limitsCpu": "2000Mi", "limitsMemory": "1000Mi", "benchmarkCpu": "2", "concurrency": "1", "duration": "30s", "rate": "0", "disableOrchestrator": "true", "params": "{\"name\": \"sklearn-0\", \"replicas\": \"1\", \"serverWorkers\": \"5\", \"serverThreads\": \"1\", \"modelUri\": \"gs://seldon-models/sklearn/iris\", \"image\": \"\", \"server\": \"SKLEARN_SERVER\", \"apiType\": \"rest\", \"requestsCpu\": \"2000Mi\", \"requestsMemory\": \"100Mi\", \"limitsCpu\": \"2000Mi\", \"limitsMemory\": \"1000Mi\", \"benchmarkCpu\": \"2\", \"concurrency\": \"1\", \"duration\": \"30s\", \"rate\": \"0\", \"disableOrchestrator\": \"true\"}"}][0m
    [31mseldon-benchmark-process-2323867814: time="2021-06-28T17:27:52.287Z" level=info msg="Starting Workflow Executor" version="{v3.0.3 2021-05-11T21:14:20Z 02071057c082cf295ab8da68f1b2027ff8762b5a v3.0.3 clean go1.15.7 gc linux/amd64}"[0m
    [31mseldon-benchmark-process-2323867814: time="2021-06-28T17:27:52.289Z" level=info msg="Creating a docker executor"[0m
    [31mseldon-benchmark-process-2323867814: time="2021-06-28T17:27:52.289Z" level=info msg="Executor (version: v3.0.3, build_date: 2021-05-11T21:14:20Z) initialized (pod: argo/seldon-benchmark-process-2323867814) with template:\n{\"name\":\"create-seldon-resource-template\",\"inputs\":{\"parameters\":[{\"name\":\"inparam\",\"value\":\"sklearn-0\"},{\"name\":\"replicas\",\"value\":\"1\"},{\"name\":\"serverWorkers\",\"value\":\"5\"},{\"name\":\"serverThreads\",\"value\":\"1\"},{\"name\":\"modelUri\",\"value\":\"gs://seldon-models/sklearn/iris\"},{\"name\":\"image\",\"value\":\"\"},{\"name\":\"server\",\"value\":\"SKLEARN_SERVER\"},{\"name\":\"apiType\",\"value\":\"rest\"},{\"name\":\"requestsCpu\",\"value\":\"2000Mi\"},{\"name\":\"requestsMemory\",\"value\":\"100Mi\"},{\"name\":\"limitsCpu\",\"value\":\"2000Mi\"},{\"name\":\"limitsMemory\",\"value\":\"1000Mi\"},{\"name\":\"benchmarkCpu\",\"value\":\"2\"},{\"name\":\"concurrency\",\"value\":\"1\"},{\"name\":\"duration\",\"value\":\"30s\"},{\"name\":\"rate\",\"value\":\"0\"},{\"name\":\"params\",\"value\":\"{\\\\\\\"name\\\\\\\": \\\\\\\"sklearn-0\\\\\\\", \\\\\\\"replicas\\\\\\\": \\\\\\\"1\\\\\\\", \\\\\\\"serverWorkers\\\\\\\": \\\\\\\"5\\\\\\\", \\\\\\\"serverThreads\\\\\\\": \\\\\\\"1\\\\\\\", \\\\\\\"modelUri\\\\\\\": \\\\\\\"gs://seldon-models/sklearn/iris\\\\\\\", \\\\\\\"image\\\\\\\": \\\\\\\"\\\\\\\", \\\\\\\"server\\\\\\\": \\\\\\\"SKLEARN_SERVER\\\\\\\", \\\\\\\"apiType\\\\\\\": \\\\\\\"rest\\\\\\\", \\\\\\\"requestsCpu\\\\\\\": \\\\\\\"2000Mi\\\\\\\", \\\\\\\"requestsMemory\\\\\\\": \\\\\\\"100Mi\\\\\\\", \\\\\\\"limitsCpu\\\\\\\": \\\\\\\"2000Mi\\\\\\\", \\\\\\\"limitsMemory\\\\\\\": \\\\\\\"1000Mi\\\\\\\", \\\\\\\"benchmarkCpu\\\\\\\": \\\\\\\"2\\\\\\\", \\\\\\\"concurrency\\\\\\\": \\\\\\\"1\\\\\\\", \\\\\\\"duration\\\\\\\": \\\\\\\"30s\\\\\\\", \\\\\\\"rate\\\\\\\": \\\\\\\"0\\\\\\\", \\\\\\\"disableOrchestrator\\\\\\\": \\\\\\\"true\\\\\\\"}\"},{\"name\":\"disableOrchestrator\",\"value\":\"true\"}]},\"outputs\":{},\"metadata\":{},\"resource\":{\"action\":\"create\",\"manifest\":\"apiVersion: machinelearning.seldon.io/v1\\nkind: SeldonDeployment\\nmetadata:\\n  name: \\\"sklearn-0\\\"\\n  namespace: argo\\n  ownerReferences:\\n  - apiVersion: argoproj.io/v1alpha1\\n    blockOwnerDeletion: true\\n    kind: Workflow\\n    name: \\\"seldon-benchmark-process\\\"\\n    uid: \\\"8be89da7-cb46-40c5-98ac-c17dd9d99d12\\\"\\nspec:\\n  name: \\\"sklearn-0\\\"\\n  transport: \\\"rest\\\"\\n  predictors:\\n    - annotations:\\n        seldonio/no-engine: \\\"true\\\"\\n      componentSpecs:\\n      - spec:\\n          containers:\\n          - name: classifier\\n            env:\\n            - name: GUNICORN_THREADS\\n              value: \\\"1\\\"\\n            - name: GUNICORN_WORKERS\\n              value: \\\"5\\\"\\n      graph:\\n        children: []\\n        implementation: SKLEARN_SERVER\\n        modelUri: gs://seldon-models/sklearn/iris\\n        name: classifier\\n      name: default\\n      replicas: 1\\n\"}}"[0m
    [31mseldon-benchmark-process-2323867814: time="2021-06-28T17:27:52.289Z" level=info msg="Loading manifest to /tmp/manifest.yaml"[0m
    [31mseldon-benchmark-process-2323867814: time="2021-06-28T17:27:52.289Z" level=info msg="kubectl create -f /tmp/manifest.yaml -o json"[0m
    [31mseldon-benchmark-process-2323867814: time="2021-06-28T17:27:52.808Z" level=info msg=argo/SeldonDeployment.machinelearning.seldon.io/sklearn-0[0m
    [31mseldon-benchmark-process-2323867814: time="2021-06-28T17:27:52.808Z" level=info msg="Starting SIGUSR2 signal monitor"[0m
    [31mseldon-benchmark-process-2323867814: time="2021-06-28T17:27:52.808Z" level=info msg="No output parameters"[0m
    [32mseldon-benchmark-process-2388065488: Waiting for deployment "sklearn-0-default-0-classifier" rollout to finish: 0 of 1 updated replicas are available...[0m
    [32mseldon-benchmark-process-2388065488: deployment "sklearn-0-default-0-classifier" successfully rolled out[0m
    [31mseldon-benchmark-process-3406592537: {"latencies":{"total":29976397800,"mean":4636720,"50th":4085862,"90th":6830905,"95th":7974040,"99th":12617402,"max":37090400,"min":2523500},"bytes_in":{"total":1183095,"mean":183},"bytes_out":{"total":219810,"mean":34},"earliest":"2021-06-28T17:28:30.4200499Z","latest":"2021-06-28T17:29:00.4217614Z","end":"2021-06-28T17:29:00.4249416Z","duration":30001711500,"wait":3180200,"requests":6465,"rate":215.48770642634838,"throughput":215.46486701700042,"success":1,"status_codes":{"200":6465},"errors":[],"params":{"name":"sklearn-0","replicas":"1","serverWorkers":"5","serverThreads":"1","modelUri":"gs://seldon-models/sklearn/iris","image":"","server":"SKLEARN_SERVER","apiType":"rest","requestsCpu":"2000Mi","requestsMemory":"100Mi","limitsCpu":"2000Mi","limitsMemory":"1000Mi","benchmarkCpu":"2","concurrency":"1","duration":"30s","rate":"0","disableOrchestrator":"true"}}[0m
    [37mseldon-benchmark-process-1122797932: seldondeployment.machinelearning.seldon.io "sklearn-0" deleted[0m



```python
!argo get seldon-benchmark-process -n argo
```

    Name:                seldon-benchmark-process
    Namespace:           argo
    ServiceAccount:      default
    Status:              Running
    Conditions:          
     PodRunning          False
    Created:             Mon Jun 28 18:38:12 +0100 (4 minutes ago)
    Started:             Mon Jun 28 18:38:12 +0100 (4 minutes ago)
    Duration:            4 minutes 53 seconds
    Progress:            14/15
    ResourcesDuration:   4m41s*(1 cpu),4m41s*(100Mi memory)
    
    [39mSTEP[0m                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           TEMPLATE                               PODNAME                              DURATION  MESSAGE
     [36m●[0m seldon-benchmark-process                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    seldon-benchmark-process                                                                                                   
     ├───[32m✔[0m generate-parameters                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     generate-parameters-template           seldon-benchmark-process-635956972   3s                                             
     └─┬─[32m✔[0m run-benchmark-iteration(0:apiType:rest,benchmarkCpu:2,concurrency:1,disableOrchestrator:true,duration:30s,image:,limitsCpu:2000Mi,limitsMemory:1000Mi,modelUri:gs://seldon-models/sklearn/iris,name:sklearn-0,params:{\"name\": \"sklearn-0\", \"replicas\": \"1\", \"serverWorkers\": \"5\", \"serverThreads\": \"1\", \"modelUri\": \"gs://seldon-models/sklearn/iris\", \"image\": \"\", \"server\": \"SKLEARN_SERVER\", \"apiType\": \"rest\", \"requestsCpu\": \"2000Mi\", \"requestsMemory\": \"100Mi\", \"limitsCpu\": \"2000Mi\", \"limitsMemory\": \"1000Mi\", \"benchmarkCpu\": \"2\", \"concurrency\": \"1\", \"duration\": \"30s\", \"rate\": \"0\", \"disableOrchestrator\": \"true\"},rate:0,replicas:1,requestsCpu:2000Mi,requestsMemory:100Mi,server:SKLEARN_SERVER,serverThreads:1,serverWorkers:5)    run-benchmark-iteration-step-template                                                                                      
       │ ├───[32m✔[0m create-seldon-resource                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              create-seldon-resource-template        seldon-benchmark-process-2323867814  1s                                             
       │ ├───[32m✔[0m wait-seldon-resource                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                wait-seldon-resource-template          seldon-benchmark-process-2388065488  15s                                            
       │ ├─┬─[39m○[0m run-benchmark-grpc                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  run-benchmark-template-grpc                                                           when 'rest == grpc' evaluated false  
       │ │ └─[32m✔[0m run-benchmark-rest                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  run-benchmark-template-rest            seldon-benchmark-process-3406592537  31s                                            
       │ └───[32m✔[0m delete-seldon-resource                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              delete-seldon-resource-template        seldon-benchmark-process-1122797932  3s                                             
       ├─[32m✔[0m run-benchmark-iteration(1:apiType:rest,benchmarkCpu:2,concurrency:1,disableOrchestrator:false,duration:30s,image:,limitsCpu:2000Mi,limitsMemory:1000Mi,modelUri:gs://seldon-models/sklearn/iris,name:sklearn-1,params:{\"name\": \"sklearn-1\", \"replicas\": \"1\", \"serverWorkers\": \"5\", \"serverThreads\": \"1\", \"modelUri\": \"gs://seldon-models/sklearn/iris\", \"image\": \"\", \"server\": \"SKLEARN_SERVER\", \"apiType\": \"rest\", \"requestsCpu\": \"2000Mi\", \"requestsMemory\": \"100Mi\", \"limitsCpu\": \"2000Mi\", \"limitsMemory\": \"1000Mi\", \"benchmarkCpu\": \"2\", \"concurrency\": \"1\", \"duration\": \"30s\", \"rate\": \"0\", \"disableOrchestrator\": \"false\"},rate:0,replicas:1,requestsCpu:2000Mi,requestsMemory:100Mi,server:SKLEARN_SERVER,serverThreads:1,serverWorkers:5)  run-benchmark-iteration-step-template                                                                                      
       │ ├───[32m✔[0m create-seldon-resource                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              create-seldon-resource-template        seldon-benchmark-process-1282498819  2s                                             
       │ ├───[32m✔[0m wait-seldon-resource                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                wait-seldon-resource-template          seldon-benchmark-process-634593      17s                                            
       │ ├─┬─[39m○[0m run-benchmark-grpc                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  run-benchmark-template-grpc                                                           when 'rest == grpc' evaluated false  
       │ │ └─[32m✔[0m run-benchmark-rest                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  run-benchmark-template-rest            seldon-benchmark-process-692934928   32s                                            
       │ └───[32m✔[0m delete-seldon-resource                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              delete-seldon-resource-template        seldon-benchmark-process-3935684561  2s                                             
       ├─[32m✔[0m run-benchmark-iteration(2:apiType:grpc,benchmarkCpu:2,concurrency:1,disableOrchestrator:true,duration:30s,image:,limitsCpu:2000Mi,limitsMemory:1000Mi,modelUri:gs://seldon-models/sklearn/iris,name:sklearn-2,params:{\"name\": \"sklearn-2\", \"replicas\": \"1\", \"serverWorkers\": \"5\", \"serverThreads\": \"1\", \"modelUri\": \"gs://seldon-models/sklearn/iris\", \"image\": \"\", \"server\": \"SKLEARN_SERVER\", \"apiType\": \"grpc\", \"requestsCpu\": \"2000Mi\", \"requestsMemory\": \"100Mi\", \"limitsCpu\": \"2000Mi\", \"limitsMemory\": \"1000Mi\", \"benchmarkCpu\": \"2\", \"concurrency\": \"1\", \"duration\": \"30s\", \"rate\": \"0\", \"disableOrchestrator\": \"true\"},rate:0,replicas:1,requestsCpu:2000Mi,requestsMemory:100Mi,server:SKLEARN_SERVER,serverThreads:1,serverWorkers:5)    run-benchmark-iteration-step-template                                                                                      
       │ ├───[32m✔[0m create-seldon-resource                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              create-seldon-resource-template        seldon-benchmark-process-637309828   1s                                             
       │ ├───[32m✔[0m wait-seldon-resource                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                wait-seldon-resource-template          seldon-benchmark-process-2284480586  19s                                            
       │ ├─┬─[32m✔[0m run-benchmark-grpc                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  run-benchmark-template-grpc            seldon-benchmark-process-2580139445  32s                                            
       │ │ └─[39m○[0m run-benchmark-rest                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  run-benchmark-template-rest                                                           when 'grpc == rest' evaluated false  
       │ └───[32m✔[0m delete-seldon-resource                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              delete-seldon-resource-template        seldon-benchmark-process-757645342   3s                                             
       └─[36m●[0m run-benchmark-iteration(3:apiType:grpc,benchmarkCpu:2,concurrency:1,disableOrchestrator:false,duration:30s,image:,limitsCpu:2000Mi,limitsMemory:1000Mi,modelUri:gs://seldon-models/sklearn/iris,name:sklearn-3,params:{\"name\": \"sklearn-3\", \"replicas\": \"1\", \"serverWorkers\": \"5\", \"serverThreads\": \"1\", \"modelUri\": \"gs://seldon-models/sklearn/iris\", \"image\": \"\", \"server\": \"SKLEARN_SERVER\", \"apiType\": \"grpc\", \"requestsCpu\": \"2000Mi\", \"requestsMemory\": \"100Mi\", \"limitsCpu\": \"2000Mi\", \"limitsMemory\": \"1000Mi\", \"benchmarkCpu\": \"2\", \"concurrency\": \"1\", \"duration\": \"30s\", \"rate\": \"0\", \"disableOrchestrator\": \"false\"},rate:0,replicas:1,requestsCpu:2000Mi,requestsMemory:100Mi,server:SKLEARN_SERVER,serverThreads:1,serverWorkers:5)  run-benchmark-iteration-step-template                                                                                      
         ├───[32m✔[0m create-seldon-resource                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              create-seldon-resource-template        seldon-benchmark-process-1376808213  1s                                             
         └───[33m◷[0m wait-seldon-resource                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                wait-seldon-resource-template          seldon-benchmark-process-3668579423  7s                                             


## Process the results

We can now print the results in a consumable format.

## Deeper Analysis
Now that we have all the parameters, we can do a deeper analysis


```python

```




    True




```python
import sys

sys.path.append("../../../testing/scripts")
import pandas as pd
from seldon_e2e_utils import bench_results_from_output_logs

results = bench_results_from_output_logs("seldon-benchmark-process", namespace="argo")
df = pd.DataFrame.from_dict(results)
```


```python
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
      <th>mean</th>
      <th>50th</th>
      <th>90th</th>
      <th>95th</th>
      <th>99th</th>
      <th>throughputAchieved</th>
      <th>success</th>
      <th>errors</th>
      <th>name</th>
      <th>replicas</th>
      <th>...</th>
      <th>apiType</th>
      <th>requestsCpu</th>
      <th>requestsMemory</th>
      <th>limitsCpu</th>
      <th>limitsMemory</th>
      <th>benchmarkCpu</th>
      <th>concurrency</th>
      <th>duration</th>
      <th>rate</th>
      <th>disableOrchestrator</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <th>0</th>
      <td>4.573302</td>
      <td>4.018635</td>
      <td>6.225710</td>
      <td>7.480878</td>
      <td>13.893386</td>
      <td>218.518811</td>
      <td>6559</td>
      <td>0</td>
      <td>sklearn-0</td>
      <td>1</td>
      <td>...</td>
      <td>rest</td>
      <td>2000Mi</td>
      <td>100Mi</td>
      <td>2000Mi</td>
      <td>1000Mi</td>
      <td>2</td>
      <td>1</td>
      <td>30s</td>
      <td>0</td>
      <td>true</td>
    </tr>
    <tr>
      <th>1</th>
      <td>4.565145</td>
      <td>3.939032</td>
      <td>6.785393</td>
      <td>7.928704</td>
      <td>13.315820</td>
      <td>218.892806</td>
      <td>6568</td>
      <td>0</td>
      <td>sklearn-1</td>
      <td>1</td>
      <td>...</td>
      <td>rest</td>
      <td>2000Mi</td>
      <td>100Mi</td>
      <td>2000Mi</td>
      <td>1000Mi</td>
      <td>2</td>
      <td>1</td>
      <td>30s</td>
      <td>0</td>
      <td>false</td>
    </tr>
    <tr>
      <th>2</th>
      <td>3.747319</td>
      <td>3.212300</td>
      <td>5.651600</td>
      <td>6.858700</td>
      <td>9.191800</td>
      <td>258.595746</td>
      <td>7757</td>
      <td>1</td>
      <td>sklearn-2</td>
      <td>1</td>
      <td>...</td>
      <td>grpc</td>
      <td>2000Mi</td>
      <td>100Mi</td>
      <td>2000Mi</td>
      <td>1000Mi</td>
      <td>2</td>
      <td>1</td>
      <td>30s</td>
      <td>0</td>
      <td>true</td>
    </tr>
    <tr>
      <th>3</th>
      <td>4.271879</td>
      <td>3.855800</td>
      <td>6.495800</td>
      <td>7.353500</td>
      <td>8.980500</td>
      <td>226.930063</td>
      <td>6807</td>
      <td>1</td>
      <td>sklearn-3</td>
      <td>1</td>
      <td>...</td>
      <td>grpc</td>
      <td>2000Mi</td>
      <td>100Mi</td>
      <td>2000Mi</td>
      <td>1000Mi</td>
      <td>2</td>
      <td>1</td>
      <td>30s</td>
      <td>0</td>
      <td>false</td>
    </tr>
  </tbody>
</table>
<p>4 rows × 25 columns</p>
</div>




```python
!argo delete seldon-benchmark-process -n argo || echo "Argo workflow already deleted or not exists"
```

    Workflow 'seldon-benchmark-process' deleted



```python

```
