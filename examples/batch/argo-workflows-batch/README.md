# Batch processing with Argo Worfklows

In this notebook we will dive into how you can run batch processing with Argo Workflows and Seldon Core.

Dependencies:

* Seldon core installed as per the docs with an ingress
* Minio running in your cluster to use as local (s3) object storage
* Argo Workfklows installed in cluster (and argo CLI for commands)

### Setup

#### Install Seldon Core
Use the notebook to [set-up Seldon Core with Ambassador or Istio Ingress](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html).

Note: If running with KIND you need to make sure do follow [these steps](https://github.com/argoproj/argo-workflows/issues/2376#issuecomment-595593237) as workaround to the `/.../docker.sock` known issue.

#### Set up Minio in your cluster
Use the notebook to [set-up Minio in your cluster](https://docs.seldon.io/projects/seldon-core/en/latest/examples/minio_setup.html).

#### Create rclone configuration
In this example, our workflow stages responsible for pulling / pushing data to in-cluster MinIO S3 storage will use `rclone` CLI.
In order to configure the CLI we will create a following secret:


```python
%%writefile rclone-config.yaml
apiVersion: v1
kind: Secret
metadata:
  name: rclone-config-secret
type: Opaque
stringData:
  rclone.conf: |
    [cluster-minio]
    type = s3
    provider = minio
    env_auth = false
    access_key_id = minioadmin
    secret_access_key = minioadmin
    endpoint = http://minio.minio-system.svc.cluster.local:9000
```

    Overwriting rclone-config.yaml



```python
!kubectl apply -n default -f rclone-config.yaml
```

    secret/rclone-config-secret created


#### Install Argo Workflows
You can follow the instructions from the official [Argo Workflows Documentation](https://github.com/argoproj/argo#quickstart).

You also need to make sure that argo has permissions to create seldon deployments - for this you can create a role:


```python
%%writefile role.yaml
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
```

    Overwriting role.yaml



```python
!!kubectl apply -n default -f role.yaml
```




    ['role.rbac.authorization.k8s.io/workflow created']



A service account:


```python
!kubectl create -n default serviceaccount workflow
```

    serviceaccount/workflow created


And a binding


```python
!kubectl create rolebinding workflow -n default --role=workflow --serviceaccount=default:workflow
```

    rolebinding.rbac.authorization.k8s.io/workflow created


### Create some input for our model

We will create a file that will contain the inputs that will be sent to our model


```python
mkdir -p assets/
```


```python
import random
import os
random.seed(0)
with open("assets/input-data.txt", "w") as f:
    for _ in range(10000):
        data = [random.random() for _ in range(4)]
        data = "[[" + ", ".join(str(x) for x in data) + "]]\n"
        f.write(data)
```

#### Check the contents of the file


```python
!wc -l assets/input-data.txt
!head assets/input-data.txt
```

    10000 assets/input-data.txt
    [[0.8444218515250481, 0.7579544029403025, 0.420571580830845, 0.25891675029296335]]
    [[0.5112747213686085, 0.4049341374504143, 0.7837985890347726, 0.30331272607892745]]
    [[0.4765969541523558, 0.5833820394550312, 0.9081128851953352, 0.5046868558173903]]
    [[0.28183784439970383, 0.7558042041572239, 0.6183689966753316, 0.25050634136244054]]
    [[0.9097462559682401, 0.9827854760376531, 0.8102172359965896, 0.9021659504395827]]
    [[0.3101475693193326, 0.7298317482601286, 0.8988382879679935, 0.6839839319154413]]
    [[0.47214271545271336, 0.1007012080683658, 0.4341718354537837, 0.6108869734438016]]
    [[0.9130110532378982, 0.9666063677707588, 0.47700977655271704, 0.8653099277716401]]
    [[0.2604923103919594, 0.8050278270130223, 0.5486993038355893, 0.014041700164018955]]
    [[0.7197046864039541, 0.39882354222426875, 0.824844977148233, 0.6681532012318508]]


#### Upload the file to our minio


```python
!mc mb minio-seldon/data
!mc cp assets/input-data.txt minio-seldon/data/
```

    [m[32;1mBucket created successfully `minio-seldon/data`.[0m
    ...-data.txt:  820.96 KiB / 820.96 KiB ┃▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓┃ 71.44 MiB/s 0s[0m[0m[m[32;1m

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
    ServiceAccount:      workflow
    Status:              Pending
    Created:             Fri Jan 15 11:44:56 +0000 (now)
    Progress:            



```python
!argo list -n default
```

    NAME                   STATUS    AGE   DURATION   PRIORITY
    seldon-batch-process   Running   10s   10s        0



```python
!argo get -n default seldon-batch-process
```

    Name:                seldon-batch-process
    Namespace:           default
    ServiceAccount:      workflow
    Status:              Succeeded
    Conditions:          
     Completed           True
    Created:             Fri Jan 15 11:44:56 +0000 (2 minutes ago)
    Started:             Fri Jan 15 11:44:56 +0000 (2 minutes ago)
    Finished:            Fri Jan 15 11:47:00 +0000 (36 seconds ago)
    Duration:            2 minutes 4 seconds
    Progress:            6/6
    ResourcesDuration:   2m18s*(1 cpu),2m18s*(100Mi memory)
    
    [39mSTEP[0m                           TEMPLATE                         PODNAME                          DURATION  MESSAGE
     [32m✔[0m seldon-batch-process        seldon-batch-process                                                          
     ├───[32m✔[0m create-seldon-resource  create-seldon-resource-template  seldon-batch-process-3626514072  2s          
     ├───[32m✔[0m wait-seldon-resource    wait-seldon-resource-template    seldon-batch-process-2052519094  31s         
     ├───[32m✔[0m download-object-store   download-object-store-template   seldon-batch-process-1257652469  4s          
     ├───[32m✔[0m process-batch-inputs    process-batch-inputs-template    seldon-batch-process-2033515954  33s         
     ├───[32m✔[0m upload-object-store     upload-object-store-template     seldon-batch-process-2123074048  3s          
     └───[32m✔[0m delete-seldon-resource  delete-seldon-resource-template  seldon-batch-process-2070809024  9s          



```python
!argo -n default logs seldon-batch-process
```

    [32mseldon-batch-process-3626514072: time="2021-01-15T11:44:57.620Z" level=info msg="Starting Workflow Executor" version=v2.12.3[0m
    [32mseldon-batch-process-3626514072: time="2021-01-15T11:44:57.622Z" level=info msg="Creating a K8sAPI executor"[0m
    [32mseldon-batch-process-3626514072: time="2021-01-15T11:44:57.622Z" level=info msg="Executor (version: v2.12.3, build_date: 2021-01-05T00:54:54Z) initialized (pod: default/seldon-batch-process-3626514072) with template:\n{\"name\":\"create-seldon-resource-template\",\"arguments\":{},\"inputs\":{},\"outputs\":{},\"metadata\":{\"annotations\":{\"sidecar.istio.io/inject\":\"false\"}},\"resource\":{\"action\":\"create\",\"manifest\":\"apiVersion: machinelearning.seldon.io/v1\\nkind: SeldonDeployment\\nmetadata:\\n  name: \\\"sklearn\\\"\\n  namespace: default\\n  ownerReferences:\\n  - apiVersion: argoproj.io/v1alpha1\\n    blockOwnerDeletion: true\\n    kind: Workflow\\n    name: \\\"seldon-batch-process\\\"\\n    uid: \\\"511f64a2-0699-42eb-897a-c0a57b24072c\\\"\\nspec:\\n  name: \\\"sklearn\\\"\\n  predictors:\\n    - componentSpecs:\\n      - spec:\\n        containers:\\n        - name: classifier\\n          env:\\n          - name: GUNICORN_THREADS\\n            value: 10\\n          - name: GUNICORN_WORKERS\\n            value: 1\\n          resources:\\n            requests:\\n              cpu: 50m\\n              memory: 100Mi\\n            limits:\\n              cpu: 50m\\n              memory: 1000Mi\\n      graph:\\n        children: []\\n        implementation: SKLEARN_SERVER\\n        modelUri: gs://seldon-models/sklearn/iris\\n        name: classifier\\n      name: default\\n      replicas: 10\\n\"}}"[0m
    [32mseldon-batch-process-3626514072: time="2021-01-15T11:44:57.622Z" level=info msg="Loading manifest to /tmp/manifest.yaml"[0m
    [32mseldon-batch-process-3626514072: time="2021-01-15T11:44:57.622Z" level=info msg="kubectl create -f /tmp/manifest.yaml -o json"[0m
    [32mseldon-batch-process-3626514072: time="2021-01-15T11:44:58.044Z" level=info msg=default/SeldonDeployment.machinelearning.seldon.io/sklearn[0m
    [32mseldon-batch-process-3626514072: time="2021-01-15T11:44:58.044Z" level=info msg="Starting SIGUSR2 signal monitor"[0m
    [32mseldon-batch-process-3626514072: time="2021-01-15T11:44:58.045Z" level=info msg="No output parameters"[0m
    [33mseldon-batch-process-2052519094: Waiting for deployment "sklearn-default-0-classifier" rollout to finish: 0 of 10 updated replicas are available...[0m
    [33mseldon-batch-process-2052519094: Waiting for deployment "sklearn-default-0-classifier" rollout to finish: 1 of 10 updated replicas are available...[0m
    [33mseldon-batch-process-2052519094: Waiting for deployment "sklearn-default-0-classifier" rollout to finish: 2 of 10 updated replicas are available...[0m
    [33mseldon-batch-process-2052519094: Waiting for deployment "sklearn-default-0-classifier" rollout to finish: 3 of 10 updated replicas are available...[0m
    [33mseldon-batch-process-2052519094: Waiting for deployment "sklearn-default-0-classifier" rollout to finish: 4 of 10 updated replicas are available...[0m
    [33mseldon-batch-process-2052519094: Waiting for deployment "sklearn-default-0-classifier" rollout to finish: 5 of 10 updated replicas are available...[0m
    [33mseldon-batch-process-2052519094: Waiting for deployment "sklearn-default-0-classifier" rollout to finish: 6 of 10 updated replicas are available...[0m
    [33mseldon-batch-process-2052519094: Waiting for deployment "sklearn-default-0-classifier" rollout to finish: 7 of 10 updated replicas are available...[0m
    [33mseldon-batch-process-2052519094: Waiting for deployment "sklearn-default-0-classifier" rollout to finish: 8 of 10 updated replicas are available...[0m
    [33mseldon-batch-process-2052519094: Waiting for deployment "sklearn-default-0-classifier" rollout to finish: 9 of 10 updated replicas are available...[0m
    [33mseldon-batch-process-2052519094: deployment "sklearn-default-0-classifier" successfully rolled out[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:00,306 - batch_processor.py:167 - INFO:  Processed instances: 100[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:00,321 - batch_processor.py:167 - INFO:  Processed instances: 200[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:00,411 - batch_processor.py:167 - INFO:  Processed instances: 300[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:00,464 - batch_processor.py:167 - INFO:  Processed instances: 400[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:00,768 - batch_processor.py:167 - INFO:  Processed instances: 500[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:01,055 - batch_processor.py:167 - INFO:  Processed instances: 600[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:01,439 - batch_processor.py:167 - INFO:  Processed instances: 700[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:01,757 - batch_processor.py:167 - INFO:  Processed instances: 800[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:02,025 - batch_processor.py:167 - INFO:  Processed instances: 900[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:02,303 - batch_processor.py:167 - INFO:  Processed instances: 1000[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:02,563 - batch_processor.py:167 - INFO:  Processed instances: 1100[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:02,864 - batch_processor.py:167 - INFO:  Processed instances: 1200[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:03,151 - batch_processor.py:167 - INFO:  Processed instances: 1300[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:03,447 - batch_processor.py:167 - INFO:  Processed instances: 1400[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:03,848 - batch_processor.py:167 - INFO:  Processed instances: 1500[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:04,244 - batch_processor.py:167 - INFO:  Processed instances: 1600[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:04,547 - batch_processor.py:167 - INFO:  Processed instances: 1700[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:04,783 - batch_processor.py:167 - INFO:  Processed instances: 1800[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:05,080 - batch_processor.py:167 - INFO:  Processed instances: 1900[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:05,478 - batch_processor.py:167 - INFO:  Processed instances: 2000[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:06,074 - batch_processor.py:167 - INFO:  Processed instances: 2100[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:06,438 - batch_processor.py:167 - INFO:  Processed instances: 2200[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:06,817 - batch_processor.py:167 - INFO:  Processed instances: 2300[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:07,101 - batch_processor.py:167 - INFO:  Processed instances: 2400[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:07,312 - batch_processor.py:167 - INFO:  Processed instances: 2500[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:07,538 - batch_processor.py:167 - INFO:  Processed instances: 2600[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:07,766 - batch_processor.py:167 - INFO:  Processed instances: 2700[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:07,979 - batch_processor.py:167 - INFO:  Processed instances: 2800[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:08,265 - batch_processor.py:167 - INFO:  Processed instances: 2900[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:08,565 - batch_processor.py:167 - INFO:  Processed instances: 3000[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:08,820 - batch_processor.py:167 - INFO:  Processed instances: 3100[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:09,069 - batch_processor.py:167 - INFO:  Processed instances: 3200[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:09,326 - batch_processor.py:167 - INFO:  Processed instances: 3300[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:09,566 - batch_processor.py:167 - INFO:  Processed instances: 3400[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:09,786 - batch_processor.py:167 - INFO:  Processed instances: 3500[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:09,993 - batch_processor.py:167 - INFO:  Processed instances: 3600[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:10,209 - batch_processor.py:167 - INFO:  Processed instances: 3700[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:10,487 - batch_processor.py:167 - INFO:  Processed instances: 3800[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:10,757 - batch_processor.py:167 - INFO:  Processed instances: 3900[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:10,967 - batch_processor.py:167 - INFO:  Processed instances: 4000[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:11,221 - batch_processor.py:167 - INFO:  Processed instances: 4100[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:11,449 - batch_processor.py:167 - INFO:  Processed instances: 4200[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:11,705 - batch_processor.py:167 - INFO:  Processed instances: 4300[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:11,914 - batch_processor.py:167 - INFO:  Processed instances: 4400[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:12,130 - batch_processor.py:167 - INFO:  Processed instances: 4500[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:12,345 - batch_processor.py:167 - INFO:  Processed instances: 4600[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:12,621 - batch_processor.py:167 - INFO:  Processed instances: 4700[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:12,962 - batch_processor.py:167 - INFO:  Processed instances: 4800[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:13,219 - batch_processor.py:167 - INFO:  Processed instances: 4900[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:13,463 - batch_processor.py:167 - INFO:  Processed instances: 5000[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:13,730 - batch_processor.py:167 - INFO:  Processed instances: 5100[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:13,966 - batch_processor.py:167 - INFO:  Processed instances: 5200[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:14,246 - batch_processor.py:167 - INFO:  Processed instances: 5300[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:14,479 - batch_processor.py:167 - INFO:  Processed instances: 5400[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:14,764 - batch_processor.py:167 - INFO:  Processed instances: 5500[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:15,041 - batch_processor.py:167 - INFO:  Processed instances: 5600[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:15,319 - batch_processor.py:167 - INFO:  Processed instances: 5700[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:15,561 - batch_processor.py:167 - INFO:  Processed instances: 5800[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:15,795 - batch_processor.py:167 - INFO:  Processed instances: 5900[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:16,038 - batch_processor.py:167 - INFO:  Processed instances: 6000[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:16,284 - batch_processor.py:167 - INFO:  Processed instances: 6100[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:16,565 - batch_processor.py:167 - INFO:  Processed instances: 6200[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:16,862 - batch_processor.py:167 - INFO:  Processed instances: 6300[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:17,145 - batch_processor.py:167 - INFO:  Processed instances: 6400[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:17,416 - batch_processor.py:167 - INFO:  Processed instances: 6500[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:17,669 - batch_processor.py:167 - INFO:  Processed instances: 6600[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:17,881 - batch_processor.py:167 - INFO:  Processed instances: 6700[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:18,096 - batch_processor.py:167 - INFO:  Processed instances: 6800[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:18,357 - batch_processor.py:167 - INFO:  Processed instances: 6900[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:18,609 - batch_processor.py:167 - INFO:  Processed instances: 7000[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:18,996 - batch_processor.py:167 - INFO:  Processed instances: 7100[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:19,426 - batch_processor.py:167 - INFO:  Processed instances: 7200[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:19,804 - batch_processor.py:167 - INFO:  Processed instances: 7300[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:20,139 - batch_processor.py:167 - INFO:  Processed instances: 7400[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:20,528 - batch_processor.py:167 - INFO:  Processed instances: 7500[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:21,047 - batch_processor.py:167 - INFO:  Processed instances: 7600[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:21,439 - batch_processor.py:167 - INFO:  Processed instances: 7700[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:22,050 - batch_processor.py:167 - INFO:  Processed instances: 7800[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:22,435 - batch_processor.py:167 - INFO:  Processed instances: 7900[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:22,812 - batch_processor.py:167 - INFO:  Processed instances: 8000[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:23,226 - batch_processor.py:167 - INFO:  Processed instances: 8100[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:23,547 - batch_processor.py:167 - INFO:  Processed instances: 8200[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:23,898 - batch_processor.py:167 - INFO:  Processed instances: 8300[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:24,308 - batch_processor.py:167 - INFO:  Processed instances: 8400[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:24,672 - batch_processor.py:167 - INFO:  Processed instances: 8500[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:24,986 - batch_processor.py:167 - INFO:  Processed instances: 8600[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:25,284 - batch_processor.py:167 - INFO:  Processed instances: 8700[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:25,545 - batch_processor.py:167 - INFO:  Processed instances: 8800[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:25,851 - batch_processor.py:167 - INFO:  Processed instances: 8900[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:26,124 - batch_processor.py:167 - INFO:  Processed instances: 9000[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:26,445 - batch_processor.py:167 - INFO:  Processed instances: 9100[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:26,817 - batch_processor.py:167 - INFO:  Processed instances: 9200[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:27,165 - batch_processor.py:167 - INFO:  Processed instances: 9300[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:27,483 - batch_processor.py:167 - INFO:  Processed instances: 9400[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:27,794 - batch_processor.py:167 - INFO:  Processed instances: 9500[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:28,099 - batch_processor.py:167 - INFO:  Processed instances: 9600[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:28,479 - batch_processor.py:167 - INFO:  Processed instances: 9700[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:28,912 - batch_processor.py:167 - INFO:  Processed instances: 9800[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:29,465 - batch_processor.py:167 - INFO:  Processed instances: 9900[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:30,012 - batch_processor.py:167 - INFO:  Processed instances: 10000[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:30,012 - batch_processor.py:168 - INFO:  Total processed instances: 10000[0m
    [33mseldon-batch-process-2033515954: 2021-01-15 11:46:30,012 - batch_processor.py:116 - INFO:  Elapsed time: 30.899641513824463[0m
    [35mseldon-batch-process-2070809024: seldondeployment.machinelearning.seldon.io "sklearn" deleted[0m


### Check output in object store

We can now visualise the output that we obtained in the object store.

First we can check that the file is present:


```python
import json
wf_arr = !argo get -n default seldon-batch-process -o json
wf = json.loads("".join(wf_arr))
WF_UID = wf["metadata"]["uid"]
print(f"Workflow UID is {WF_UID}")
```

    Workflow UID is 511f64a2-0699-42eb-897a-c0a57b24072c



```python
!mc ls minio-seldon/data/output-data-"$WF_UID".txt
```

    [m[32m[2021-01-15 11:46:42 GMT] [0m[33m 3.4MiB [0moutput-data-511f64a2-0699-42eb-897a-c0a57b24072c.txt
    [0m

Now we can output the contents of the file created using the `mc head` command.


```python
!mc cp minio-seldon/data/output-data-"$WF_UID".txt assets/output-data.txt
!head assets/output-data.txt
```

    ...4072c.txt:  3.36 MiB / 3.36 MiB ┃▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓┃ 192.59 MiB/s 0s[0m[0m[m[32;1m{"data": {"names": ["t:0", "t:1", "t:2"], "ndarray": [[0.1859090109477526, 0.46433848375587844, 0.349752505296369]]}, "meta": {"requestPath": {"classifier": "seldonio/sklearnserver:1.6.0-dev"}, "tags": {"tags": {"batch_id": "3c4000b8-5727-11eb-91c1-6e88dc41eb63", "batch_index": 1.0, "batch_instance_id": "3c40e1e0-5727-11eb-9fe5-6e88dc41eb63"}}}}
    {"data": {"names": ["t:0", "t:1", "t:2"], "ndarray": [[0.1679456497678022, 0.42318259169768935, 0.4088717585345084]]}, "meta": {"requestPath": {"classifier": "seldonio/sklearnserver:1.6.0-dev"}, "tags": {"tags": {"batch_id": "3c4000b8-5727-11eb-91c1-6e88dc41eb63", "batch_index": 22.0, "batch_instance_id": "3c42efb2-5727-11eb-9fe5-6e88dc41eb63"}}}}
    {"data": {"names": ["t:0", "t:1", "t:2"], "ndarray": [[0.5329356306409886, 0.2531124742231082, 0.21395189513590318]]}, "meta": {"requestPath": {"classifier": "seldonio/sklearnserver:1.6.0-dev"}, "tags": {"tags": {"batch_id": "3c4000b8-5727-11eb-91c1-6e88dc41eb63", "batch_index": 25.0, "batch_instance_id": "3c43dac6-5727-11eb-9fe5-6e88dc41eb63"}}}}
    {"data": {"names": ["t:0", "t:1", "t:2"], "ndarray": [[0.5057216294927378, 0.37562353221834527, 0.11865483828891676]]}, "meta": {"requestPath": {"classifier": "seldonio/sklearnserver:1.6.0-dev"}, "tags": {"tags": {"batch_id": "3c4000b8-5727-11eb-91c1-6e88dc41eb63", "batch_index": 20.0, "batch_instance_id": "3c4294a4-5727-11eb-9fe5-6e88dc41eb63"}}}}
    {"data": {"names": ["t:0", "t:1", "t:2"], "ndarray": [[0.16020781530738484, 0.49084414063547427, 0.3489480440571409]]}, "meta": {"requestPath": {"classifier": "seldonio/sklearnserver:1.6.0-dev"}, "tags": {"tags": {"batch_id": "3c4000b8-5727-11eb-91c1-6e88dc41eb63", "batch_index": 24.0, "batch_instance_id": "3c439a34-5727-11eb-9fe5-6e88dc41eb63"}}}}
    {"data": {"names": ["t:0", "t:1", "t:2"], "ndarray": [[0.49551509682202705, 0.4192462053867995, 0.08523869779117352]]}, "meta": {"requestPath": {"classifier": "seldonio/sklearnserver:1.6.0-dev"}, "tags": {"tags": {"batch_id": "3c4000b8-5727-11eb-91c1-6e88dc41eb63", "batch_index": 0.0, "batch_instance_id": "3c40c6d8-5727-11eb-9fe5-6e88dc41eb63"}}}}
    {"data": {"names": ["t:0", "t:1", "t:2"], "ndarray": [[0.17817271417040353, 0.4160568279837039, 0.4057704578458926]]}, "meta": {"requestPath": {"classifier": "seldonio/sklearnserver:1.6.0-dev"}, "tags": {"tags": {"batch_id": "3c4000b8-5727-11eb-91c1-6e88dc41eb63", "batch_index": 6.0, "batch_instance_id": "3c41aa58-5727-11eb-9fe5-6e88dc41eb63"}}}}
    {"data": {"names": ["t:0", "t:1", "t:2"], "ndarray": [[0.31086648314817084, 0.43371070280306884, 0.25542281404876027]]}, "meta": {"requestPath": {"classifier": "seldonio/sklearnserver:1.6.0-dev"}, "tags": {"tags": {"batch_id": "3c4000b8-5727-11eb-91c1-6e88dc41eb63", "batch_index": 27.0, "batch_instance_id": "3c44420e-5727-11eb-9fe5-6e88dc41eb63"}}}}
    {"data": {"names": ["t:0", "t:1", "t:2"], "ndarray": [[0.4381942165350952, 0.39483980719426687, 0.16696597627063794]]}, "meta": {"requestPath": {"classifier": "seldonio/sklearnserver:1.6.0-dev"}, "tags": {"tags": {"batch_id": "3c4000b8-5727-11eb-91c1-6e88dc41eb63", "batch_index": 34.0, "batch_instance_id": "3c448ff2-5727-11eb-9fe5-6e88dc41eb63"}}}}
    {"data": {"names": ["t:0", "t:1", "t:2"], "ndarray": [[0.2975075875929912, 0.25439317776178244, 0.44809923464522644]]}, "meta": {"requestPath": {"classifier": "seldonio/sklearnserver:1.6.0-dev"}, "tags": {"tags": {"batch_id": "3c4000b8-5727-11eb-91c1-6e88dc41eb63", "batch_index": 4.0, "batch_instance_id": "3c41837a-5727-11eb-9fe5-6e88dc41eb63"}}}}



```python
!argo delete -n default seldon-batch-process
```

    Workflow 'seldon-batch-process' deleted
