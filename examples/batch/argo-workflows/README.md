## Batch processing with Argo Worfklows

In this notebook we will dive into how you can run batch processing with Argo Workflows and Seldon Core.

Dependencies:

* Seldon core installed as per the docs with an ingress
* Argo Workfklows installed in cluster (and argo CLI for commands)


## Seldon Core Batch with Object Store

In some cases we may want to read the data from an object source.

In this case we will show how you can read from an object store, in this case minio.

The workflow will look as follows:

![](assets/seldon-batch.jpg)

For this we will assume you have installed the Minio (mc) CLI - we will use a Minio client in the cluster but you can use another object store provider like S3, Google Cloud, Azure, etc.

### Set up Minio in your cluster


```bash
%%bash 
helm install minio stable/minio \
    --set accessKey=minioadmin \
    --set secretKey=minioadmin \
    --set image.tag=RELEASE.2020-04-15T19-42-18Z
```

    NAME: minio
    LAST DEPLOYED: Thu Apr 30 10:57:00 2020
    NAMESPACE: default
    STATUS: deployed
    REVISION: 1
    TEST SUITE: None
    NOTES:
    Minio can be accessed via port 9000 on the following DNS name from within your cluster:
    minio.default.svc.cluster.local
    
    To access Minio from localhost, run the below commands:
    
      1. export POD_NAME=$(kubectl get pods --namespace default -l "release=minio" -o jsonpath="{.items[0].metadata.name}")
    
      2. kubectl port-forward $POD_NAME 9000 --namespace default
    
    Read more about port forwarding here: http://kubernetes.io/docs/user-guide/kubectl/kubectl_port-forward/
    
    You can now access Minio server on http://localhost:9000. Follow the below steps to connect to Minio server with mc client:
    
      1. Download the Minio mc client - https://docs.minio.io/docs/minio-client-quickstart-guide
    
      2. mc config host add minio-local http://localhost:9000 minioadmin minioadmin S3v4
    
      3. mc ls minio-local
    
    Alternately, you can use your browser or the Minio SDK to access the server - https://docs.minio.io/categories/17


### Forward the Minio port so you can access it

You can do this by runnning the following command in your terminal:
```
kubectl port-forward svc/minio 9000:9000
    ```
    
### Configure local minio client


```python
!mc config host add minio-local http://localhost:9000 minioadmin minioadmin
```

    [m[32mAdded `minio-local` successfully.[0m
    [0m

### Create some input for our model

We will create a file that will contain the inputs that will be sent to our model


```python
with open("assets/input-data.txt", "w") as f:
    for i in range(10000):
        f.write('{"data":{"ndarray":[[1, 2, 3, 4]]}, "meta": { "puid": "' + str(i) + '"}}\n')
```

### Check the contents of the file


```python
!wc -l assets/input-data.txt
!head assets/input-data.txt
```

    10000 assets/input-data.txt
    {"data":{"ndarray":[[1, 2, 3, 4]]}, "meta": { "puid": "0"}}
    {"data":{"ndarray":[[1, 2, 3, 4]]}, "meta": { "puid": "1"}}
    {"data":{"ndarray":[[1, 2, 3, 4]]}, "meta": { "puid": "2"}}
    {"data":{"ndarray":[[1, 2, 3, 4]]}, "meta": { "puid": "3"}}
    {"data":{"ndarray":[[1, 2, 3, 4]]}, "meta": { "puid": "4"}}
    {"data":{"ndarray":[[1, 2, 3, 4]]}, "meta": { "puid": "5"}}
    {"data":{"ndarray":[[1, 2, 3, 4]]}, "meta": { "puid": "6"}}
    {"data":{"ndarray":[[1, 2, 3, 4]]}, "meta": { "puid": "7"}}
    {"data":{"ndarray":[[1, 2, 3, 4]]}, "meta": { "puid": "8"}}
    {"data":{"ndarray":[[1, 2, 3, 4]]}, "meta": { "puid": "9"}}


### Upload the file to our minio


```python
!mc mb minio-local/data
!mc cp assets/input-data.txt minio-local/data/
```

    [m[32;1mBucket created successfully `minio-local/data`.[0m
    ...-data.txt:  614.15 KiB / 614.15 KiB ┃▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓┃ 9.74 MiB/s 0s[0m[0m[m[32;1m

### Create Argo Workflow

In order to create our argo workflow we have made it simple so you can leverage the power of the helm charts.

Before we dive into the contents of the full helm chart, let's first give it a try with some of the settings.

We will run a batch job that will set up a Seldon Deployment with 10 replicas and 100 batch client workers to send requests.


```python
!ls helm-charts/
```

    seldon-batch-workflow



```python
!helm template seldon-batch-workflow helm-charts/seldon-batch-workflow \
    --set workflow.name=seldon-batch-process \
    --set seldonDeployment.create=false \
    --set seldonDeployment.name="seldon-batch" \
    --set seldonDeployment.replicas=10 \
    --set batchWorker.workers=100 \
    --set minio.inputDataPath="s3://data/input-data.txt" \
    --set minio.outputDataPath="s3://data/output-data-{{workflow.uid}}.txt" \
    | argo submit -
```

    Name:                seldon-batch-process
    Namespace:           default
    ServiceAccount:      default
    Status:              Pending
    Created:             Mon May 18 21:55:31 +0100 (now)



```python
!argo list
```

    NAME                   STATUS    AGE   DURATION   PRIORITY
    seldon-batch-process   Running   46s   46s        0



```python
!argo get seldon-batch-process
```

    Name:                seldon-batch-process
    Namespace:           default
    ServiceAccount:      default
    Status:              Running
    Created:             Mon May 18 21:55:31 +0100 (53 seconds ago)
    Started:             Mon May 18 21:55:31 +0100 (53 seconds ago)
    Duration:            53 seconds
    
    [39mSTEP[0m                                                             PODNAME                          DURATION  MESSAGE
     [36m●[0m seldon-batch-process (seldon-batch-process)                                                              
     ├---[32m✔[0m create-seldon-resource (create-seldon-resource-template)  seldon-batch-process-3626514072  1s        
     ├---[32m✔[0m wait-seldon-resource (wait-seldon-resource-template)      seldon-batch-process-2052519094  30s       
     └---[36m●[0m process-batch-inputs (process-batch-inputs-template)      seldon-batch-process-50851621    20s       



```python
!argo logs -w seldon-batch-process 
```

    [35mcreate-seldon-resource[0m:	time="2020-05-18T20:52:21Z" level=info msg="Starting Workflow Executor" version=vv2.7.4+50b209c.dirty
    [35mcreate-seldon-resource[0m:	time="2020-05-18T20:52:21Z" level=info msg="Creating a docker executor"
    [35mcreate-seldon-resource[0m:	time="2020-05-18T20:52:21Z" level=info msg="Executor (version: vv2.7.4+50b209c.dirty, build_date: 2020-04-16T16:37:57Z) initialized (pod: default/seldon-batch-process-3626514072) with template:\n{\"name\":\"create-seldon-resource-template\",\"arguments\":{},\"inputs\":{},\"outputs\":{},\"metadata\":{},\"resource\":{\"action\":\"create\",\"manifest\":\"apiVersion: machinelearning.seldon.io/v1\\nkind: SeldonDeployment\\nmetadata:\\n  name: \\\"seldon-844299e9-abf7-4742-9d4d-eeb405433569\\\"\\n  namespace: default\\n  ownerReferences:\\n  - apiVersion: argoproj.io/v1alpha1\\n    blockOwnerDeletion: true\\n    kind: Workflow\\n    name: \\\"seldon-batch-process\\\"\\n    uid: \\\"844299e9-abf7-4742-9d4d-eeb405433569\\\"\\nspec:\\n  name: \\\"seldon-844299e9-abf7-4742-9d4d-eeb405433569\\\"\\n  predictors:\\n    - graph:\\n        children: []\\n        implementation: SKLEARN_SERVER\\n        modelUri: gs://seldon-models/sklearn/iris\\n        name: classifier\\n      name: default\\n      replicas: 1\\n        \\n\"}}"
    [35mcreate-seldon-resource[0m:	time="2020-05-18T20:52:21Z" level=info msg="Loading manifest to /tmp/manifest.yaml"
    [35mcreate-seldon-resource[0m:	time="2020-05-18T20:52:21Z" level=info msg="kubectl create -f /tmp/manifest.yaml -o json"
    [35mcreate-seldon-resource[0m:	time="2020-05-18T20:52:22Z" level=info msg=default/SeldonDeployment.machinelearning.seldon.io/seldon-844299e9-abf7-4742-9d4d-eeb405433569
    [35mcreate-seldon-resource[0m:	time="2020-05-18T20:52:22Z" level=info msg="No output parameters"
    [32mwait-seldon-resource[0m:	Waiting for deployment "seldon-70a32d330b04187c79768e37cda5d600" rollout to finish: 0 of 1 updated replicas are available...
    [32mwait-seldon-resource[0m:	deployment "seldon-70a32d330b04187c79768e37cda5d600" successfully rolled out


## Check output in object store

We can now visualise the output that we obtained in the object store.

First we can check that the file is present:


```python
import json
wf_arr = !argo get seldon-batch-process -o json
wf = json.loads("".join(wf_arr))
WF_ID = wf["metadata"]["uid"]
print(f"Workflow ID is {WF_ID}")
```

    Workflow ID is 844299e9-abf7-4742-9d4d-eeb405433569



```python
!mc ls minio-local/data/output-data-"$WF_ID".txt
```

    [m[32m[2020-05-18 21:53:16 BST] [0m[33m 1.3MiB [0m[1moutput-data-844299e9-abf7-4742-9d4d-eeb405433569.txt[0m
    [0m

Now we can output the contents of the file created using the `mc head` command.


```python
!mc cp minio-local/data/output-data-"$WF_ID".txt assets/output-data.txt
!head assets/output-data.txt
```

    ...33569.txt:  1.31 MiB / 1.31 MiB ┃▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓┃ 26.12 MiB/s 0s[0m[0m[m[32;1m{"data":{"names":["t:0","t:1","t:2"],"ndarray":[[0.0006985194531162841,0.003668039039435755,0.9956334415074478]]},"meta":{"puid":"4"}}
    {"data":{"names":["t:0","t:1","t:2"],"ndarray":[[0.0006985194531162841,0.003668039039435755,0.9956334415074478]]},"meta":{"puid":"0"}}
    {"data":{"names":["t:0","t:1","t:2"],"ndarray":[[0.0006985194531162841,0.003668039039435755,0.9956334415074478]]},"meta":{"puid":"3"}}
    {"data":{"names":["t:0","t:1","t:2"],"ndarray":[[0.0006985194531162841,0.003668039039435755,0.9956334415074478]]},"meta":{"puid":"1"}}
    {"data":{"names":["t:0","t:1","t:2"],"ndarray":[[0.0006985194531162841,0.003668039039435755,0.9956334415074478]]},"meta":{"puid":"100"}}
    {"data":{"names":["t:0","t:1","t:2"],"ndarray":[[0.0006985194531162841,0.003668039039435755,0.9956334415074478]]},"meta":{"puid":"6"}}
    {"data":{"names":["t:0","t:1","t:2"],"ndarray":[[0.0006985194531162841,0.003668039039435755,0.9956334415074478]]},"meta":{"puid":"7"}}
    {"data":{"names":["t:0","t:1","t:2"],"ndarray":[[0.0006985194531162841,0.003668039039435755,0.9956334415074478]]},"meta":{"puid":"9"}}
    {"data":{"names":["t:0","t:1","t:2"],"ndarray":[[0.0006985194531162841,0.003668039039435755,0.9956334415074478]]},"meta":{"puid":"8"}}
    {"data":{"names":["t:0","t:1","t:2"],"ndarray":[[0.0006985194531162841,0.003668039039435755,0.9956334415074478]]},"meta":{"puid":"106"}}



```python
!argo delete seldon-batch-process
```

    Workflow 'seldon-batch-process' deleted



```python

```
