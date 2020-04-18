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
  generateName: steps-
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

    Name:                steps-9tgj9
    Namespace:           default
    ServiceAccount:      default
    Status:              Pending
    Created:             Fri Apr 17 08:27:14 +0100 (8 hours ago)



```python
!argo list
```

    NAME          STATUS      AGE   DURATION   PRIORITY
    steps-9tgj9   Succeeded   8h    3m         0



```python
output=!argo list | grep steps
WF_NAME=output[0].split()[0]
print(WF_NAME)
```

    steps-9tgj9



```python
!argo get $WF_NAME
```

    Name:                steps-9tgj9
    Namespace:           default
    ServiceAccount:      default
    Status:              Succeeded
    Created:             Fri Apr 17 08:27:14 +0100 (8 hours ago)
    Started:             Fri Apr 17 08:27:14 +0100 (8 hours ago)
    Finished:            Fri Apr 17 08:30:48 +0100 (8 hours ago)
    Duration:            3 minutes 34 seconds
    
    [39mSTEP[0m                                PODNAME                 DURATION  MESSAGE
     [32m✔[0m steps-9tgj9 (hello-hello-hello)                                    
     ├---[32m✔[0m hello1 (whalesay)            steps-9tgj9-3240403473  3m        
     └-·-[32m✔[0m hello2a (whalesay)           steps-9tgj9-3510808138  3s        
       └-[32m✔[0m hello2b (whalesay)           steps-9tgj9-3494030519  5s        



```python
!argo logs -w $WF_NAME
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
!argo delete $WF_NAME
```

    Workflow 'steps-9tgj9' deleted


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
  generateName: seldon-batch-
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
        kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id="{{workflow.uid}}" -o jsonpath='{.items[0].metadata.name}')
        
  - name: process-batch-inputs-template
    script:
      image: seldonio/seldon-core-s2i-python3:1.1.1-SNAPSHOT
      command: [python]
      source: |
        from seldon_core.seldon_client import SeldonClient
        import numpy as np
        import time
        time.sleep(10)
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

    Name:                seldon-batch-wxbr5
    Namespace:           default
    ServiceAccount:      default
    Status:              Pending
    Created:             Sat Apr 18 12:36:57 +0100 (now)



```python
!argo list
```

    NAME                 STATUS      AGE   DURATION   PRIORITY
    seldon-batch-wxbr5   Running     2s    2s         0
    seldon-batch-kslgh   Succeeded   2m    42s        0



```python
output=!argo list | grep seldon-batch
WF_NAME=output[0].split()[0]
print(WF_NAME)
```

    seldon-batch-kslgh



```python
!argo get $WF_NAME
```

    Name:                seldon-batch-wxbr5
    Namespace:           default
    ServiceAccount:      default
    Status:              Running
    Created:             Sat Apr 18 12:36:57 +0100 (3 seconds ago)
    Started:             Sat Apr 18 12:36:57 +0100 (3 seconds ago)
    Duration:            3 seconds
    
    [39mSTEP[0m                                                             PODNAME                        DURATION  MESSAGE
     [36m●[0m seldon-batch-wxbr5 (seldon-batch-process)                                                              
     └---[36m●[0m create-seldon-resource (create-seldon-resource-template)  seldon-batch-wxbr5-2588046603  3s        



```python
!argo logs -w $WF_NAME
```

    [35mcreate-seldon-resource[0m:	time="2020-04-18T11:36:58Z" level=info msg="Starting Workflow Executor" version=vv2.7.4+50b209c.dirty
    [35mcreate-seldon-resource[0m:	time="2020-04-18T11:36:58Z" level=info msg="Creating a docker executor"
    [35mcreate-seldon-resource[0m:	time="2020-04-18T11:36:58Z" level=info msg="Executor (version: vv2.7.4+50b209c.dirty, build_date: 2020-04-16T16:37:57Z) initialized (pod: default/seldon-batch-wxbr5-2588046603) with template:\n{\"name\":\"create-seldon-resource-template\",\"arguments\":{},\"inputs\":{},\"outputs\":{},\"metadata\":{},\"resource\":{\"action\":\"create\",\"manifest\":\"apiVersion: machinelearning.seldon.io/v1\\nkind: SeldonDeployment\\nmetadata:\\n  name: \\\"b83971e5-e0e9-488f-9bd0-4c57dc97c79c\\\"\\n  ownerReferences:\\n  - apiVersion: argoproj.io/v1alpha1\\n    blockOwnerDeletion: true\\n    kind: Workflow\\n    name: \\\"seldon-batch-wxbr5\\\"\\n    uid: \\\"b83971e5-e0e9-488f-9bd0-4c57dc97c79c\\\"\\nspec:\\n  name: \\\"b83971e5-e0e9-488f-9bd0-4c57dc97c79c\\\"\\n  predictors:\\n    - graph:\\n        children: []\\n        implementation: SKLEARN_SERVER\\n        modelUri: gs://seldon-models/sklearn/iris\\n        name: classifier\\n      name: default\\n      replicas: 1\\n\"}}"
    [35mcreate-seldon-resource[0m:	time="2020-04-18T11:36:58Z" level=info msg="Loading manifest to /tmp/manifest.yaml"
    [35mcreate-seldon-resource[0m:	time="2020-04-18T11:36:58Z" level=info msg="kubectl create -f /tmp/manifest.yaml -o json"
    [35mcreate-seldon-resource[0m:	time="2020-04-18T11:36:59Z" level=info msg=default/SeldonDeployment.machinelearning.seldon.io/b83971e5-e0e9-488f-9bd0-4c57dc97c79c
    [35mcreate-seldon-resource[0m:	time="2020-04-18T11:36:59Z" level=info msg="No output parameters"
    [32mwait-seldon-resource[0m:	Waiting for deployment "b83971e5-e0e9-488f-9bd0-4c57dc97c79c-default-0-classifier" rollout to finish: 0 of 1 updated replicas are available...
    [32mwait-seldon-resource[0m:	deployment "b83971e5-e0e9-488f-9bd0-4c57dc97c79c-default-0-classifier" successfully rolled out
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
outputs = !(argo logs -w $WF_NAME --no-color | grep "process-batch-inputs" | cut -c 23-)
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
!argo delete $WF_NAME
```

    Workflow 'seldon-batch-kslgh' deleted


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
    LAST DEPLOYED: Sat Apr 18 15:59:30 2020
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
kubectl port-forward -n minio-system svc/minio 9000:9000
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
    for i in range(10):
        f.write(f"[[{i}, {i}, {i}, {i}]]\n")
```

### Check the contents of the file


```python
!cat assets/input-data.txt
```

    [[0, 0, 0, 0]]
    [[1, 1, 1, 1]]
    [[2, 2, 2, 2]]
    [[3, 3, 3, 3]]
    [[4, 4, 4, 4]]
    [[5, 5, 5, 5]]
    [[6, 6, 6, 6]]
    [[7, 7, 7, 7]]
    [[8, 8, 8, 8]]
    [[9, 9, 9, 9]]


### Upload the file to our minio


```python
!mc mb minio-local/data
!mc cp assets/input-data.txt minio-local/data/
```

    [m[32;1mBucket created successfully `minio-local/data`.[0m
    ...-data.txt:  150 B / 150 B ┃▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓┃ 2.07 KiB/s 0s[0m[0m[m[32;1m

### Create Job to Execute


```python
%%writefile assets/seldon-batch-store.yaml
---
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  generateName: seldon-batch-object-store-
spec:
  entrypoint: seldon-batch-process
  templates:
  - name: seldon-batch-process
    steps:
    - - name: create-seldon-resource            
        template: create-seldon-resource-template
    - - name: wait-seldon-resource
        template: wait-seldon-resource-template
    - - name: download-object-store
        template: download-object-store-template
    - - name: process-batch-inputs
        template: process-batch-inputs-template
        arguments:
          artifacts:
          - name: input-data
            from: "{{steps.download-object-store.outputs.artifacts.input-data}}"
    - - name: upload-object-store
        template: upload-object-store-template
        arguments:
          artifacts:
          - name: output-data
            from: "{{steps.process-batch-inputs.outputs.artifacts.output-data}}"
            
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
        sleep 3
        kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id="{{workflow.uid}}" -o jsonpath='{.items[0].metadata.name}')
                     
  - name: download-object-store-template
    script:
      image: minio/mc:RELEASE.2020-04-17T08-55-48Z
      command: [sh]
      source: |
        mc config host add minio-local http://minio.default.svc.cluster.local:9000 minioadmin minioadmin
        mc cp minio-local/data/input-data.txt /assets/input-data.txt
    outputs:
      artifacts:
      - name: input-data
        path: /assets/input-data.txt
            
  - name: process-batch-inputs-template
    inputs:
      artifacts:
      - name: input-data
        path: /assets/input-data.txt
    outputs:
      artifacts:
      - name: output-data
        path: /assets/output-data.txt
    script:
      image: seldonio/seldon-core-s2i-python3:1.1.1-SNAPSHOT
      command: [python]
      source: |
        from seldon_core.seldon_client import SeldonClient
        import numpy as np
        import time
        import json
        time.sleep(10)
        sc = SeldonClient(
            gateway_endpoint="istio-ingressgateway.istio-system.svc.cluster.local",
            deployment_name="{{workflow.uid}}",
            namespace="default")
        input_file = open("/assets/input-data.txt", "r")
        output_file = open("/assets/output-data.txt", "w")
        print("SENDING DATA")
        for d in input_file:
            arr = json.loads(d)
            data = np.array(arr)
            output = sc.predict(data=data)
            output_file.write(f"{output.response}\n")
        print("DONE SENDING DATA")
            
  - name: upload-object-store-template
    inputs:
      artifacts:
      - name: output-data
        path: /assets/output-data.txt
    script:
      image: minio/mc:RELEASE.2020-04-17T08-55-48Z
      command: [sh]
      source: |
        mc config host add minio-local http://minio.default.svc.cluster.local:9000 minioadmin minioadmin
        mc cp /assets/output-data.txt minio-local/data/output-data-{{workflow.name}}.txt
        
```

    Overwriting assets/seldon-batch-store.yaml


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

### Create Argo Workflow


```python
!argo submit assets/seldon-batch-store.yaml
```

    Name:                seldon-batch-object-store-bbzth
    Namespace:           default
    ServiceAccount:      default
    Status:              Pending
    Created:             Sat Apr 18 16:48:55 +0100 (now)



```python
!argo list
```

    NAME                              STATUS    AGE   DURATION   PRIORITY
    seldon-batch-object-store-bbzth   Running   2s    2s         0



```python
output=!argo list | grep seldon-batch
WF_NAME=output[0].split()[0]
print(WF_NAME)
```

    seldon-batch-object-store-bbzth



```python
!argo get $WF_NAME
```

    Name:                seldon-batch-object-store-bbzth
    Namespace:           default
    ServiceAccount:      default
    Status:              Succeeded
    Created:             Sat Apr 18 16:48:55 +0100 (54 seconds ago)
    Started:             Sat Apr 18 16:48:55 +0100 (54 seconds ago)
    Finished:            Sat Apr 18 16:49:48 +0100 (1 second ago)
    Duration:            53 seconds
    
    [39mSTEP[0m                                                             PODNAME                                     DURATION  MESSAGE
     [32m✔[0m seldon-batch-object-store-bbzth (seldon-batch-process)                                                              
     ├---[32m✔[0m create-seldon-resource (create-seldon-resource-template)  seldon-batch-object-store-bbzth-3919184201  2s        
     ├---[32m✔[0m wait-seldon-resource (wait-seldon-resource-template)      seldon-batch-object-store-bbzth-2771248891  26s       
     ├---[32m✔[0m download-object-store (download-object-store-template)    seldon-batch-object-store-bbzth-507765538   3s        
     ├---[32m✔[0m process-batch-inputs (process-batch-inputs-template)      seldon-batch-object-store-bbzth-3607615647  13s       
     └---[32m✔[0m upload-object-store (upload-object-store-template)        seldon-batch-object-store-bbzth-163395403   3s        



```python
!argo logs -w $WF_NAME
```

    [35mcreate-seldon-resource[0m:	time="2020-04-18T15:48:56Z" level=info msg="Starting Workflow Executor" version=vv2.7.4+50b209c.dirty
    [35mcreate-seldon-resource[0m:	time="2020-04-18T15:48:56Z" level=info msg="Creating a docker executor"
    [35mcreate-seldon-resource[0m:	time="2020-04-18T15:48:56Z" level=info msg="Executor (version: vv2.7.4+50b209c.dirty, build_date: 2020-04-16T16:37:57Z) initialized (pod: default/seldon-batch-object-store-bbzth-3919184201) with template:\n{\"name\":\"create-seldon-resource-template\",\"arguments\":{},\"inputs\":{},\"outputs\":{},\"metadata\":{},\"resource\":{\"action\":\"create\",\"manifest\":\"apiVersion: machinelearning.seldon.io/v1\\nkind: SeldonDeployment\\nmetadata:\\n  name: \\\"eacde3ac-debe-4080-b5a3-86ca8adc8314\\\"\\n  ownerReferences:\\n  - apiVersion: argoproj.io/v1alpha1\\n    blockOwnerDeletion: true\\n    kind: Workflow\\n    name: \\\"seldon-batch-object-store-bbzth\\\"\\n    uid: \\\"eacde3ac-debe-4080-b5a3-86ca8adc8314\\\"\\nspec:\\n  name: \\\"eacde3ac-debe-4080-b5a3-86ca8adc8314\\\"\\n  predictors:\\n    - graph:\\n        children: []\\n        implementation: SKLEARN_SERVER\\n        modelUri: gs://seldon-models/sklearn/iris\\n        name: classifier\\n      name: default\\n      replicas: 1\\n        \\n\"}}"
    [35mcreate-seldon-resource[0m:	time="2020-04-18T15:48:56Z" level=info msg="Loading manifest to /tmp/manifest.yaml"
    [35mcreate-seldon-resource[0m:	time="2020-04-18T15:48:56Z" level=info msg="kubectl create -f /tmp/manifest.yaml -o json"
    [35mcreate-seldon-resource[0m:	time="2020-04-18T15:48:57Z" level=info msg=default/SeldonDeployment.machinelearning.seldon.io/eacde3ac-debe-4080-b5a3-86ca8adc8314
    [35mcreate-seldon-resource[0m:	time="2020-04-18T15:48:57Z" level=info msg="No output parameters"
    [32mwait-seldon-resource[0m:	Waiting for deployment "eacde3ac-debe-4080-b5a3-86ca8adc8314-default-0-classifier" rollout to finish: 0 of 1 updated replicas are available...
    [32mwait-seldon-resource[0m:	deployment "eacde3ac-debe-4080-b5a3-86ca8adc8314-default-0-classifier" successfully rolled out
    [34mdownload-object-store[0m:	Added `minio-local` successfully.
    [34mdownload-object-store[0m:	`minio-local/data/input-data.txt` -> `/assets/input-data.txt`
    [34mdownload-object-store[0m:	Total: 0 B, Transferred: 150 B, Speed: 14.81 KiB/s
    [39mprocess-batch-inputs[0m:	SENDING DATA
    [39mprocess-batch-inputs[0m:	DONE SENDING DATA
    [31mupload-object-store[0m:	Added `minio-local` successfully.
    [31mupload-object-store[0m:	`/assets/output-data.txt` -> `minio-local/data/output-data-seldon-batch-object-store-bbzth.txt`
    [31mupload-object-store[0m:	Total: 0 B, Transferred: 1.57 KiB, Speed: 54.89 KiB/s


## Check output in object store

We can now visualise the output that we obtained in the object store.

First we can check that the file is present:


```python
!mc ls minio-local/data/output-data-"$WF_NAME".txt
```

    [m[32m[2020-04-18 16:49:46 BST] [0m[33m 1.6KiB [0m[1moutput-data-seldon-batch-object-store-bbzth.txt[0m
    [0m

Now we can output the contents of the file created using the `mc head` command.


```python
!mc head minio-local/data/output-data-"$WF_NAME".txt
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
!argo delete $WF_NAME
```

    Workflow 'seldon-batch-object-store-bbzth' deleted



```python

```
