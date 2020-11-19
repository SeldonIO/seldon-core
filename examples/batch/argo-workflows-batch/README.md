# Batch processing with Argo Worfklows

In this notebook we will dive into how you can run batch processing with Argo Workflows and Seldon Core.

Dependencies:

* Seldon core installed as per the docs with an ingress
* Minio running in your cluster to use as local (s3) object storage
* Argo Workfklows installed in cluster (and argo CLI for commands)


### Setup

#### Install Seldon Core
Use the notebook to [set-up Seldon Core with Ambassador or Istio Ingress](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html).

ote: If running with KIND you need to make sure do follow [these steps](https://github.com/argoproj/argo/issues/2376#issuecomment-595593237) as workaround to the `/.../docker.sock` known issue.

### Set up Minio in your cluster
se the notebook to [set-up Minio in your cluster](https://docs.seldon.io/projects/seldon-core/en/latest/examples/minio_setup.html).

### Copy the Minio Secret to namespace

We need to re-use the minio secret for the batch job, so this can be done by just copying the minio secret created in the `minio-system`

The command below just copies the secred with the name "minio" from the minio-system namespace to the default namespace.


```python
!kubectl get secret minio -n minio-system -o json | jq '{apiVersion,data,kind,metadata,type} | .metadata |= {"annotations", "name"}' | kubectl apply -n default -f -
```

    secret/minio created


#### Install Argo Workflows
You can follow the instructions from the official [Argo Workflows Documentation](https://github.com/argoproj/argo#quickstart).

You also need to make sure that argo has permissions to create seldon deployments - for this you can a role:


```python
!kubectl create -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: workflow
rules:
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - "*"
- apiGroups:
  - "apps"
  resources:
  - deployments
  verbs:
  - "*"
- apiGroups:
  - ""
  resources:
  - pods/log
  verbs:
  - "*"
- apiGroups:
  - machinelearning.seldon.io
  resources:
  - "*"
  verbs:
  - "*"
EOF
```

A service account:
```python
!kubectl create serviceaccount workflow
```

And a binding:
```python
!kubectl create rolebinding workflow --role=workflow --serviceaccount=default:workflow
```

### Create some input for our model

We will create a file that will contain the inputs that will be sent to our model


```python
mkdir -p assets/
```


```python
with open("assets/input-data.txt", "w") as f:
    for i in range(10000):
        f.write('[[1, 2, 3, 4]]\n')
```

#### Check the contents of the file


```python
!wc -l assets/input-data.txt
!head assets/input-data.txt
```

    10000 assets/input-data.txt
    [[1, 2, 3, 4]]
    [[1, 2, 3, 4]]
    [[1, 2, 3, 4]]
    [[1, 2, 3, 4]]
    [[1, 2, 3, 4]]
    [[1, 2, 3, 4]]
    [[1, 2, 3, 4]]
    [[1, 2, 3, 4]]
    [[1, 2, 3, 4]]
    [[1, 2, 3, 4]]


#### Upload the file to our minio


```python
!mc mb minio-seldon/data
!mc cp assets/input-data.txt minio-seldon/data/
```

    [m[32;1mBucket created successfully `minio-seldon/data`.[0m
    ...-data.txt:  146.48 KiB / 146.48 KiB ┃▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓┃ 2.14 MiB/s 0s[0m[0m[m[32;1m

#### Create Argo Workflow

In order to create our argo workflow we have made it simple so you can leverage the power of the helm charts.

Before we dive into the contents of the full helm chart, let's first give it a try with some of the settings.

We will run a batch job that will set up a Seldon Deployment with 10 replicas and 100 batch client workers to send requests.


```python
!helm template seldon-batch-workflow helm-charts/seldon-batch-workflow/ \
    --set workflow.name=seldon-batch-process \
    --set seldonDeployment.name=sklearn \
    --set seldonDeployment.replicas=10 \
    --set seldonDeployment.serverWorkers=1 \
    --set seldonDeployment.serverThreads=10 \
    --set batchWorker.workers=100 \
    --set batchWorker.payloadType=ndarray \
    --set batchWorker.dataType=data \
    | argo submit --serviceaccount workflow -
```

    Name:                seldon-batch-process
    Namespace:           default
    ServiceAccount:      default
    Status:              Pending
    Created:             Thu Aug 06 08:21:47 +0100 (now)



```python
!argo list
```

    NAME                   STATUS    AGE   DURATION   PRIORITY
    seldon-batch-process   Running   2s    2s         0



```python
!argo get seldon-batch-process
```

    Name:                seldon-batch-process
    Namespace:           default
    ServiceAccount:      default
    Status:              Succeeded
    Created:             Thu Aug 06 08:03:31 +0100 (1 minute ago)
    Started:             Thu Aug 06 08:03:31 +0100 (1 minute ago)
    Finished:            Thu Aug 06 08:04:54 +0100 (26 seconds ago)
    Duration:            1 minute 23 seconds
    
    [39mSTEP[0m                                                             PODNAME                          DURATION  MESSAGE
     [32m✔[0m seldon-batch-process (seldon-batch-process)                                                              
     ├---[32m✔[0m create-seldon-resource (create-seldon-resource-template)  seldon-batch-process-3626514072  2s        
     ├---[32m✔[0m wait-seldon-resource (wait-seldon-resource-template)      seldon-batch-process-2052519094  28s       
     ├---[32m✔[0m download-object-store (download-object-store-template)    seldon-batch-process-1257652469  2s        
     ├---[32m✔[0m process-batch-inputs (process-batch-inputs-template)      seldon-batch-process-2033515954  39s       
     └---[32m✔[0m upload-object-store (upload-object-store-template)        seldon-batch-process-2123074048  3s        



```python
!argo logs -w seldon-batch-process || argo logs seldon-batch-process # The 2nd command is for argo 2.8+
```

    [35mcreate-seldon-resource[0m:	time="2020-08-06T07:21:48.400Z" level=info msg="Starting Workflow Executor" version=v2.9.3
    [35mcreate-seldon-resource[0m:	time="2020-08-06T07:21:48.404Z" level=info msg="Creating a docker executor"
    [35mcreate-seldon-resource[0m:	time="2020-08-06T07:21:48.404Z" level=info msg="Executor (version: v2.9.3, build_date: 2020-07-18T19:11:19Z) initialized (pod: default/seldon-batch-process-3626514072) with template:\n{\"name\":\"create-seldon-resource-template\",\"arguments\":{},\"inputs\":{},\"outputs\":{},\"metadata\":{},\"resource\":{\"action\":\"create\",\"manifest\":\"apiVersion: machinelearning.seldon.io/v1\\nkind: SeldonDeployment\\nmetadata:\\n  name: \\\"sklearn\\\"\\n  namespace: default\\n  ownerReferences:\\n  - apiVersion: argoproj.io/v1alpha1\\n    blockOwnerDeletion: true\\n    kind: Workflow\\n    name: \\\"seldon-batch-process\\\"\\n    uid: \\\"401c8bc0-0ff0-4f7b-94ba-347df5c786f9\\\"\\nspec:\\n  name: \\\"sklearn\\\"\\n  predictors:\\n    - componentSpecs:\\n      - spec:\\n        containers:\\n        - name: classifier\\n          env:\\n          - name: GUNICORN_THREADS\\n            value: 10\\n          - name: GUNICORN_WORKERS\\n            value: 1\\n          resources:\\n            requests:\\n              cpu: 50m\\n              memory: 100Mi\\n            limits:\\n              cpu: 50m\\n              memory: 1000Mi\\n      graph:\\n        children: []\\n        implementation: SKLEARN_SERVER\\n        modelUri: gs://seldon-models/sklearn/iris\\n        name: classifier\\n      name: default\\n      replicas: 10\\n        \\n\"}}"
    [35mcreate-seldon-resource[0m:	time="2020-08-06T07:21:48.404Z" level=info msg="Loading manifest to /tmp/manifest.yaml"
    [35mcreate-seldon-resource[0m:	time="2020-08-06T07:21:48.405Z" level=info msg="kubectl create -f /tmp/manifest.yaml -o json"
    [35mcreate-seldon-resource[0m:	time="2020-08-06T07:21:48.954Z" level=info msg=default/SeldonDeployment.machinelearning.seldon.io/sklearn
    [35mcreate-seldon-resource[0m:	time="2020-08-06T07:21:48.954Z" level=info msg="No output parameters"
    [32mwait-seldon-resource[0m:	Waiting for deployment "sklearn-default-0-classifier" rollout to finish: 0 of 10 updated replicas are available...
    [32mwait-seldon-resource[0m:	Waiting for deployment "sklearn-default-0-classifier" rollout to finish: 1 of 10 updated replicas are available...
    [32mwait-seldon-resource[0m:	Waiting for deployment "sklearn-default-0-classifier" rollout to finish: 2 of 10 updated replicas are available...
    [32mwait-seldon-resource[0m:	Waiting for deployment "sklearn-default-0-classifier" rollout to finish: 3 of 10 updated replicas are available...
    [32mwait-seldon-resource[0m:	Waiting for deployment "sklearn-default-0-classifier" rollout to finish: 4 of 10 updated replicas are available...
    [32mwait-seldon-resource[0m:	Waiting for deployment "sklearn-default-0-classifier" rollout to finish: 5 of 10 updated replicas are available...
    [32mwait-seldon-resource[0m:	Waiting for deployment "sklearn-default-0-classifier" rollout to finish: 6 of 10 updated replicas are available...
    [32mwait-seldon-resource[0m:	Waiting for deployment "sklearn-default-0-classifier" rollout to finish: 7 of 10 updated replicas are available...
    [32mwait-seldon-resource[0m:	Waiting for deployment "sklearn-default-0-classifier" rollout to finish: 8 of 10 updated replicas are available...
    [32mwait-seldon-resource[0m:	Waiting for deployment "sklearn-default-0-classifier" rollout to finish: 9 of 10 updated replicas are available...
    [32mwait-seldon-resource[0m:	deployment "sklearn-default-0-classifier" successfully rolled out
    [34mdownload-object-store[0m:	Added `minio-local` successfully.
    [34mdownload-object-store[0m:	`minio-local/data/input-data.txt` -> `/assets/input-data.txt`
    [34mdownload-object-store[0m:	Total: 0 B, Transferred: 146.48 KiB, Speed: 31.81 MiB/s
    [39mprocess-batch-inputs[0m:	Elapsed time: 35.089903831481934
    [31mupload-object-store[0m:	Added `minio-local` successfully.
    [31mupload-object-store[0m:	`/assets/output-data.txt` -> `minio-local/data/output-data-401c8bc0-0ff0-4f7b-94ba-347df5c786f9.txt`
    [31mupload-object-store[0m:	Total: 0 B, Transferred: 2.75 MiB, Speed: 105.34 MiB/s


### Check output in object store

We can now visualise the output that we obtained in the object store.

First we can check that the file is present:


```python
import json
wf_arr = !argo get seldon-batch-process -o json
wf = json.loads("".join(wf_arr))
WF_ID = wf["metadata"]["uid"]
print(f"Workflow ID is {WF_ID}")
```

    Workflow ID is 401c8bc0-0ff0-4f7b-94ba-347df5c786f9



```python
!mc ls minio-seldon/data/output-data-"$WF_ID".txt
```

    [m[32m[2020-08-06 08:23:07 BST] [0m[33m 2.7MiB [0m[1moutput-data-401c8bc0-0ff0-4f7b-94ba-347df5c786f9.txt[0m
    [0m

Now we can output the contents of the file created using the `mc head` command.


```python
!mc cp minio-seldon/data/output-data-"$WF_ID".txt assets/output-data.txt
!head assets/output-data.txt
```

    ...786f9.txt:  2.75 MiB / 2.75 MiB ┃▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓┃ 26.55 MiB/s 0s[0m[0m[m[32;1m{"data": {"names": ["t:0", "t:1", "t:2"], "ndarray": [[0.0006985194531162841, 0.003668039039435755, 0.9956334415074478]]}, "meta": {"tags": {"tags": {"batch_id": "95e6e8d0-d7b5-11ea-b00e-ea443eed4c19", "batch_index": 2.0, "batch_instance_id": "95e7df56-d7b5-11ea-b5f2-ea443eed4c19"}}}}
    {"data": {"names": ["t:0", "t:1", "t:2"], "ndarray": [[0.0006985194531162841, 0.003668039039435755, 0.9956334415074478]]}, "meta": {"tags": {"tags": {"batch_id": "95e6e8d0-d7b5-11ea-b00e-ea443eed4c19", "batch_index": 0.0, "batch_instance_id": "95e77c3c-d7b5-11ea-b5f2-ea443eed4c19"}}}}
    {"data": {"names": ["t:0", "t:1", "t:2"], "ndarray": [[0.0006985194531162841, 0.003668039039435755, 0.9956334415074478]]}, "meta": {"tags": {"tags": {"batch_id": "95e6e8d0-d7b5-11ea-b00e-ea443eed4c19", "batch_index": 1.0, "batch_instance_id": "95e787ae-d7b5-11ea-b5f2-ea443eed4c19"}}}}
    {"data": {"names": ["t:0", "t:1", "t:2"], "ndarray": [[0.0006985194531162841, 0.003668039039435755, 0.9956334415074478]]}, "meta": {"tags": {"tags": {"batch_id": "95e6e8d0-d7b5-11ea-b00e-ea443eed4c19", "batch_index": 3.0, "batch_instance_id": "95e80990-d7b5-11ea-b5f2-ea443eed4c19"}}}}
    {"data": {"names": ["t:0", "t:1", "t:2"], "ndarray": [[0.0006985194531162841, 0.003668039039435755, 0.9956334415074478]]}, "meta": {"tags": {"tags": {"batch_id": "95e6e8d0-d7b5-11ea-b00e-ea443eed4c19", "batch_index": 4.0, "batch_instance_id": "95e83cf8-d7b5-11ea-b5f2-ea443eed4c19"}}}}
    {"data": {"names": ["t:0", "t:1", "t:2"], "ndarray": [[0.0006985194531162841, 0.003668039039435755, 0.9956334415074478]]}, "meta": {"tags": {"tags": {"batch_id": "95e6e8d0-d7b5-11ea-b00e-ea443eed4c19", "batch_index": 6.0, "batch_instance_id": "95e85990-d7b5-11ea-b5f2-ea443eed4c19"}}}}
    {"data": {"names": ["t:0", "t:1", "t:2"], "ndarray": [[0.0006985194531162841, 0.003668039039435755, 0.9956334415074478]]}, "meta": {"tags": {"tags": {"batch_id": "95e6e8d0-d7b5-11ea-b00e-ea443eed4c19", "batch_index": 8.0, "batch_instance_id": "95e85e40-d7b5-11ea-b5f2-ea443eed4c19"}}}}
    {"data": {"names": ["t:0", "t:1", "t:2"], "ndarray": [[0.0006985194531162841, 0.003668039039435755, 0.9956334415074478]]}, "meta": {"tags": {"tags": {"batch_id": "95e6e8d0-d7b5-11ea-b00e-ea443eed4c19", "batch_index": 7.0, "batch_instance_id": "95e85c1a-d7b5-11ea-b5f2-ea443eed4c19"}}}}
    {"data": {"names": ["t:0", "t:1", "t:2"], "ndarray": [[0.0006985194531162841, 0.003668039039435755, 0.9956334415074478]]}, "meta": {"tags": {"tags": {"batch_id": "95e6e8d0-d7b5-11ea-b00e-ea443eed4c19", "batch_index": 10.0, "batch_instance_id": "95e864c6-d7b5-11ea-b5f2-ea443eed4c19"}}}}
    {"data": {"names": ["t:0", "t:1", "t:2"], "ndarray": [[0.0006985194531162841, 0.003668039039435755, 0.9956334415074478]]}, "meta": {"tags": {"tags": {"batch_id": "95e6e8d0-d7b5-11ea-b00e-ea443eed4c19", "batch_index": 5.0, "batch_instance_id": "95e83f8c-d7b5-11ea-b5f2-ea443eed4c19"}}}}



```python
!argo delete seldon-batch-process
```

    Workflow 'seldon-batch-process' deleted



```python

```
