## Batch processing with Argo Worfklows

In this notebook we will dive into how you can run batch processing with Argo Workflows and Seldon Core.

Dependencies:

* Seldon core installed as per the docs with an ingress
* Argo Workfklows installed in cluster (and argo CLI for commands)


## Argo Workflows Example

Let's try an argo workflows example to see intuitively how it works. 

In this case we will trigger a workflow with 3 steps (first one will execute and the other two jobs are dependent on that).

The example below will basically run a workflow in the following order:

![](assets/argo-example.jpg)


```python
mkdir -p assets
```


```python
%%writefile assets/argo-example.yaml
---
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  name: argo-basic-example
spec:
  entrypoint: hello-hello-hello
  # This spec contains two templates: hello-hello-hello and whalesay
  templates:
  - name: hello-hello-hello
    # Instead of just running a container
    # This template has a sequence of steps
    steps:
    - - name: hello1            # hello1 is run before the following steps
        template: whalesay
        arguments:
          parameters:
          - name: message
            value: "hello1"
    - - name: hello2a           # double dash => run after previous step
        template: whalesay
        arguments:
          parameters:
          - name: message
            value: "hello2a"
      - name: hello2b           # single dash => run in parallel with previous step
        template: whalesay
        arguments:
          parameters:
          - name: message
            value: "hello2b"
  # This is the same template as from the previous example
  - name: whalesay
    inputs:
      parameters:
      - name: message
    container:
      image: docker/whalesay
      command: [cowsay]
      args: ["{{inputs.parameters.message}}"]
        
```

    Overwriting assets/argo-example.yaml



```python
!argo submit assets/argo-example.yaml
```

    Name:                argo-basic-example
    Namespace:           default
    ServiceAccount:      default
    Status:              Pending
    Created:             Sat May 02 12:15:51 +0100 (now)



```python
!argo list
```

    NAME                 STATUS    AGE   DURATION   PRIORITY
    argo-basic-example   Running   2s    2s         0



```python
!argo get argo-basic-example
```

    Name:                argo-basic-example
    Namespace:           default
    ServiceAccount:      default
    Status:              Succeeded
    Created:             Sat May 02 12:15:51 +0100 (47 seconds ago)
    Started:             Sat May 02 12:15:51 +0100 (47 seconds ago)
    Finished:            Sat May 02 12:16:38 +0100 (now)
    Duration:            47 seconds
    
    [39mSTEP[0m                                       PODNAME                        DURATION  MESSAGE
     [32m✔[0m argo-basic-example (hello-hello-hello)                                           
     ├---[32m✔[0m hello1 (whalesay)                   argo-basic-example-3663002391  39s       
     └-·-[32m✔[0m hello2a (whalesay)                  argo-basic-example-3531035296  3s        
       └-[32m✔[0m hello2b (whalesay)                  argo-basic-example-3581368153  5s        



```python
!argo logs -w argo-basic-example
```

    [37mhello1[0m:	 ________ 
    [37mhello1[0m:	< hello1 >
    [37mhello1[0m:	 -------- 
    [37mhello1[0m:	    \
    [37mhello1[0m:	     \
    [37mhello1[0m:	      \     
    [37mhello1[0m:	                    ##        .            
    [37mhello1[0m:	              ## ## ##       ==            
    [37mhello1[0m:	           ## ## ## ##      ===            
    [37mhello1[0m:	       /""""""""""""""""___/ ===        
    [37mhello1[0m:	  ~~~ {~~ ~~~~ ~~~ ~~~~ ~~ ~ /  ===- ~~~   
    [37mhello1[0m:	       \______ o          __/            
    [37mhello1[0m:	        \    \        __/             
    [37mhello1[0m:	          \____\______/   
    [37mhello2a[0m:	 _________ 
    [37mhello2a[0m:	< hello2a >
    [37mhello2a[0m:	 --------- 
    [37mhello2a[0m:	    \
    [37mhello2a[0m:	     \
    [37mhello2a[0m:	      \     
    [37mhello2a[0m:	                    ##        .            
    [37mhello2a[0m:	              ## ## ##       ==            
    [37mhello2a[0m:	           ## ## ## ##      ===            
    [37mhello2a[0m:	       /""""""""""""""""___/ ===        
    [37mhello2a[0m:	  ~~~ {~~ ~~~~ ~~~ ~~~~ ~~ ~ /  ===- ~~~   
    [37mhello2a[0m:	       \______ o          __/            
    [37mhello2a[0m:	        \    \        __/             
    [37mhello2a[0m:	          \____\______/   
    [34mhello2b[0m:	 _________ 
    [34mhello2b[0m:	< hello2b >
    [34mhello2b[0m:	 --------- 
    [34mhello2b[0m:	    \
    [34mhello2b[0m:	     \
    [34mhello2b[0m:	      \     
    [34mhello2b[0m:	                    ##        .            
    [34mhello2b[0m:	              ## ## ##       ==            
    [34mhello2b[0m:	           ## ## ## ##      ===            
    [34mhello2b[0m:	       /""""""""""""""""___/ ===        
    [34mhello2b[0m:	  ~~~ {~~ ~~~~ ~~~ ~~~~ ~~ ~ /  ===- ~~~   
    [34mhello2b[0m:	       \______ o          __/            
    [34mhello2b[0m:	        \    \        __/             
    [34mhello2b[0m:	          \____\______/   



```python
!argo delete argo-basic-example
```

    Workflow 'argo-basic-example' deleted


## Seldon Core Batch 
Now we can leverage this functionality by using seldon core batch.

The structure of this job will be the following:

![](assets/seldon-batch-simple.jpg)

THe file below denotes the structure of the three steps in this workflow:


```python
%%writefile assets/seldon-batch.yaml
---
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  name: seldon-batch-simple
spec:
  entrypoint: seldon-batch-process
  templates:
  - name: seldon-batch-process
    steps:
    - - name: create-seldon-resource            
        template: create-seldon-resource-template
    - - name: wait-seldon-resource
        template: wait-seldon-resource-template
    - - name: process-batch-inputs
        template: process-batch-inputs-template
            
  - name: create-seldon-resource-template
    resource:
      action: create
      manifest: |
        apiVersion: machinelearning.seldon.io/v1
        kind: SeldonDeployment
        metadata:
          name: "{{workflow.uid}}"
          ownerReferences:
          - apiVersion: argoproj.io/v1alpha1
            blockOwnerDeletion: true
            kind: Workflow
            name: "{{workflow.name}}"
            uid: "{{workflow.uid}}"
        spec:
          name: "{{workflow.uid}}"
          predictors:
            - graph:
                children: []
                implementation: SKLEARN_SERVER
                modelUri: gs://seldon-models/sklearn/iris
                name: classifier
              name: default
              replicas: 1
                
  - name: wait-seldon-resource-template
    script:
      image: seldonio/core-builder:0.14
      command: [bash]
      source: |
        sleep 5
        kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id="{{workflow.uid}}" -o jsonpath='{.items[0].metadata.name}')
        
  - name: process-batch-inputs-template
    script:
      image: seldonio/seldon-core-s2i-python3:1.1.1-SNAPSHOT
      command: [python]
      source: |
        from seldon_core.seldon_client import SeldonClient
        import numpy as np
        import time
        sc = SeldonClient(
            gateway_endpoint="istio-ingressgateway.istio-system.svc.cluster.local",
            deployment_name="{{workflow.uid}}",
            namespace="default")
        for i in range(10):
            data = np.array([[i, i, i, i]])
            output = sc.predict(data=data)
            print(output.response)
            
```

    Overwriting assets/seldon-batch.yaml



```python
!argo submit assets/seldon-batch.yaml
```

    Name:                seldon-batch-simple
    Namespace:           default
    ServiceAccount:      default
    Status:              Pending
    Created:             Sat May 02 12:14:04 +0100 (now)



```python
!argo list
```

    NAME                  STATUS    AGE   DURATION   PRIORITY
    seldon-batch-simple   Running   8s    8s         0



```python
!argo get seldon-batch-simple
```

    Name:                seldon-batch-simple
    Namespace:           default
    ServiceAccount:      default
    Status:              Running
    Created:             Sat May 02 12:14:04 +0100 (9 seconds ago)
    Started:             Sat May 02 12:14:04 +0100 (9 seconds ago)
    Duration:            9 seconds
    
    [39mSTEP[0m                                                             PODNAME                         DURATION  MESSAGE
     [36m●[0m seldon-batch-simple (seldon-batch-process)                                                              
     ├---[32m✔[0m create-seldon-resource (create-seldon-resource-template)  seldon-batch-simple-3724798319  2s        
     └---[36m●[0m wait-seldon-resource (wait-seldon-resource-template)      seldon-batch-simple-4145437349  7s        



```python
!argo logs -w seldon-batch-simple
```

    [35mcreate-seldon-resource[0m:	time="2020-05-02T11:14:05Z" level=info msg="Starting Workflow Executor" version=vv2.7.4+50b209c.dirty
    [35mcreate-seldon-resource[0m:	time="2020-05-02T11:14:05Z" level=info msg="Creating a docker executor"
    [35mcreate-seldon-resource[0m:	time="2020-05-02T11:14:05Z" level=info msg="Executor (version: vv2.7.4+50b209c.dirty, build_date: 2020-04-16T16:37:57Z) initialized (pod: default/seldon-batch-simple-3724798319) with template:\n{\"name\":\"create-seldon-resource-template\",\"arguments\":{},\"inputs\":{},\"outputs\":{},\"metadata\":{},\"resource\":{\"action\":\"create\",\"manifest\":\"apiVersion: machinelearning.seldon.io/v1\\nkind: SeldonDeployment\\nmetadata:\\n  name: \\\"b754cbcc-e2f2-49ee-978b-643699dee396\\\"\\n  ownerReferences:\\n  - apiVersion: argoproj.io/v1alpha1\\n    blockOwnerDeletion: true\\n    kind: Workflow\\n    name: \\\"seldon-batch-simple\\\"\\n    uid: \\\"b754cbcc-e2f2-49ee-978b-643699dee396\\\"\\nspec:\\n  name: \\\"b754cbcc-e2f2-49ee-978b-643699dee396\\\"\\n  predictors:\\n    - graph:\\n        children: []\\n        implementation: SKLEARN_SERVER\\n        modelUri: gs://seldon-models/sklearn/iris\\n        name: classifier\\n      name: default\\n      replicas: 1\\n        \\n\"}}"
    [35mcreate-seldon-resource[0m:	time="2020-05-02T11:14:05Z" level=info msg="Loading manifest to /tmp/manifest.yaml"
    [35mcreate-seldon-resource[0m:	time="2020-05-02T11:14:05Z" level=info msg="kubectl create -f /tmp/manifest.yaml -o json"
    [35mcreate-seldon-resource[0m:	time="2020-05-02T11:14:06Z" level=info msg=default/SeldonDeployment.machinelearning.seldon.io/b754cbcc-e2f2-49ee-978b-643699dee396
    [35mcreate-seldon-resource[0m:	time="2020-05-02T11:14:06Z" level=info msg="No output parameters"
    [32mwait-seldon-resource[0m:	Waiting for deployment "b754cbcc-e2f2-49ee-978b-643699dee396-default-0-classifier" rollout to finish: 0 of 1 updated replicas are available...
    [32mwait-seldon-resource[0m:	deployment "b754cbcc-e2f2-49ee-978b-643699dee396-default-0-classifier" successfully rolled out
    [39mprocess-batch-inputs[0m:	{'data': {'names': ['t:0', 't:1', 't:2'], 'tensor': {'shape': [1, 3], 'values': [0.3664487684438811, 0.48528762951761806, 0.14826360203850078]}}, 'meta': {}}
    [39mprocess-batch-inputs[0m:	{'data': {'names': ['t:0', 't:1', 't:2'], 'tensor': {'shape': [1, 3], 'values': [0.2075509473561692, 0.2443463805811625, 0.5481026720626684]}}, 'meta': {}}
    [39mprocess-batch-inputs[0m:	{'data': {'names': ['t:0', 't:1', 't:2'], 'tensor': {'shape': [1, 3], 'values': [0.06995304386311439, 0.04864300564562103, 0.8814039504912645]}}, 'meta': {}}
    [39mprocess-batch-inputs[0m:	{'data': {'names': ['t:0', 't:1', 't:2'], 'tensor': {'shape': [1, 3], 'values': [0.01859472366015777, 0.006956450489196832, 0.9744488258506454]}}, 'meta': {}}
    [39mprocess-batch-inputs[0m:	{'data': {'names': ['t:0', 't:1', 't:2'], 'tensor': {'shape': [1, 3], 'values': [0.004653433216061866, 0.0009398331072469446, 0.9944067336766912]}}, 'meta': {}}
    [39mprocess-batch-inputs[0m:	{'data': {'names': ['t:0', 't:1', 't:2'], 'tensor': {'shape': [1, 3], 'values': [0.0011463235173706913, 0.0001256712307515923, 0.9987280052518777]}}, 'meta': {}}
    [39mprocess-batch-inputs[0m:	{'data': {'names': ['t:0', 't:1', 't:2'], 'tensor': {'shape': [1, 3], 'values': [0.0002812079399700444, 1.6767055911301234e-05, 0.9997020250041186]}}, 'meta': {}}
    [39mprocess-batch-inputs[0m:	{'data': {'names': ['t:0', 't:1', 't:2'], 'tensor': {'shape': [1, 3], 'values': [6.890870086562858e-05, 2.235872103887171e-06, 0.9999288554270304]}}, 'meta': {}}
    [39mprocess-batch-inputs[0m:	{'data': {'names': ['t:0', 't:1', 't:2'], 'tensor': {'shape': [1, 3], 'values': [1.6881012306246608e-05, 2.981118341014688e-07, 0.9999828208758597]}}, 'meta': {}}
    [39mprocess-batch-inputs[0m:	{'data': {'names': ['t:0', 't:1', 't:2'], 'tensor': {'shape': [1, 3], 'values': [4.135155685795574e-06, 3.974630518304911e-08, 0.9999958250980091]}}, 'meta': {}}



```python
outputs = !(argo logs -w seldon-batch-simple --no-color | grep "process-batch-inputs" | cut -c 23-)
for o in outputs:
    print(o)
```

    {'data': {'names': ['t:0', 't:1', 't:2'], 'tensor': {'shape': [1, 3], 'values': [0.3664487684438811, 0.48528762951761806, 0.14826360203850078]}}, 'meta': {}}
    {'data': {'names': ['t:0', 't:1', 't:2'], 'tensor': {'shape': [1, 3], 'values': [0.2075509473561692, 0.2443463805811625, 0.5481026720626684]}}, 'meta': {}}
    {'data': {'names': ['t:0', 't:1', 't:2'], 'tensor': {'shape': [1, 3], 'values': [0.06995304386311439, 0.04864300564562103, 0.8814039504912645]}}, 'meta': {}}
    {'data': {'names': ['t:0', 't:1', 't:2'], 'tensor': {'shape': [1, 3], 'values': [0.01859472366015777, 0.006956450489196832, 0.9744488258506454]}}, 'meta': {}}
    {'data': {'names': ['t:0', 't:1', 't:2'], 'tensor': {'shape': [1, 3], 'values': [0.004653433216061866, 0.0009398331072469446, 0.9944067336766912]}}, 'meta': {}}
    {'data': {'names': ['t:0', 't:1', 't:2'], 'tensor': {'shape': [1, 3], 'values': [0.0011463235173706913, 0.0001256712307515923, 0.9987280052518777]}}, 'meta': {}}
    {'data': {'names': ['t:0', 't:1', 't:2'], 'tensor': {'shape': [1, 3], 'values': [0.0002812079399700444, 1.6767055911301234e-05, 0.9997020250041186]}}, 'meta': {}}
    {'data': {'names': ['t:0', 't:1', 't:2'], 'tensor': {'shape': [1, 3], 'values': [6.890870086562858e-05, 2.235872103887171e-06, 0.9999288554270304]}}, 'meta': {}}
    {'data': {'names': ['t:0', 't:1', 't:2'], 'tensor': {'shape': [1, 3], 'values': [1.6881012306246608e-05, 2.981118341014688e-07, 0.9999828208758597]}}, 'meta': {}}
    {'data': {'names': ['t:0', 't:1', 't:2'], 'tensor': {'shape': [1, 3], 'values': [4.135155685795574e-06, 3.974630518304911e-08, 0.9999958250980091]}}, 'meta': {}}



```python
!argo delete seldon-batch-simple
```

    Workflow 'seldon-batch-simple' deleted


## Seldon Core Batch with Object Store

In some cases we may want to read the data from an object source.

In this case we will show how you can read from an object store, in this case minio.

The workflow will look as follows:

![](assets/seldon-batch-store.jpg)

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

### Configure bucket for Argo Artefact passing


```python
%%writefile assets/argo-config.yaml
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: workflow-controller-configmap
data:
  config: |
    artifactRepository:
      s3:
        bucket: argo-artifacts
        endpoint: minio.default.svc.cluster.local:9000
        insecure: true
        accessKeySecret:
          name: minio
          key: accesskey
        secretKeySecret:
          name: minio
          key: secretkey
```

    Overwriting assets/argo-config.yaml


### Make sure the configmap is in the same location as your argo controller


```python
!kubectl apply -n argo -f assets/argo-config.yaml
```

    configmap/workflow-controller-configmap unchanged


### Create the bucket for artifact passing


```python
!mc mb minio-local/argo-artifacts
```

    [m[32;1mBucket created successfully `minio-local/argo-artifacts`.[0m
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
    --set seldonDeployment.create=true \
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
