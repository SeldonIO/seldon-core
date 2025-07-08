# Batch processing with Argo Worfklows and HDFS

In this notebook we will dive into how you can run batch processing with Argo Workflows and Seldon Core.

Dependencies:

* Seldon core installed as per the docs with an ingress
* HDFS namenode/datanode accessible from your cluster (here in-cluster installation for demo)
* Argo Workfklows installed in cluster (and argo CLI for commands)
* Python `hdfscli` for interacting with the installed `hdfs` instance

## Setup

### Install Seldon Core
Use the notebook to [set-up Seldon Core with Ambassador or Istio Ingress](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html).

Note: If running with KIND you need to make sure do follow [these steps](https://github.com/argoproj/argo-workflows/issues/2376#issuecomment-595593237) as workaround to the `/.../docker.sock` known issue:
```bash
kubectl patch -n argo configmap workflow-controller-configmap \
    --type merge -p '{"data": {"config": "containerRuntimeExecutor: k8sapi"}}'
```    

### Install HDFS
For this example we will need a running `hdfs` storage. We can use these [helm charts](https://artifacthub.io/packages/helm/gradiant/hdfs) from Gradiant.

```bash
helm repo add gradiant https://gradiant.github.io/charts/
kubectl create namespace hdfs-system || echo "namespace hdfs-system already exists"
helm install hdfs gradiant/hdfs --namespace hdfs-system
```

Once installation is complete, run in separate terminal a `port-forward` command for us to be able to push/pull batch data.
```bash
kubectl port-forward -n hdfs-system svc/hdfs-httpfs 14000:14000
```


### Install and configure hdfscli 
In this example we will be using [hdfscli](https://pypi.org/project/hdfs/) Python library for interacting with HDFS.
It supports both the WebHDFS (and HttpFS) API as well as Kerberos authentication (not covered by the example).

You can install it with
```bash
pip install hdfs==2.5.8
```

To be able to put `input-data.txt` for our batch job into hdfs we need to configure the client


```python
%%writefile hdfscli.cfg
[global]
default.alias = batch

[batch.alias]
url = http://localhost:14000
user = hdfs
```

    Overwriting hdfscli.cfg


### Install Argo Workflows
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
!kubectl apply -f role.yaml
```

    role.rbac.authorization.k8s.io/workflow unchanged


A service account:


```python
!kubectl create serviceaccount workflow
```

    serviceaccount/workflow created


And a binding


```python
!kubectl create rolebinding workflow --role=workflow --serviceaccount=seldon:workflow
```

    rolebinding.rbac.authorization.k8s.io/workflow created


## Create Seldon Deployment

For purpose of this batch example we will assume that Seldon Deployment is created independently from the workflow logic


```python
%%writefile deployment.yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: sklearn
  namespace: seldon
spec:
  name: iris
  predictors:
  - graph:
      children: []
      implementation: SKLEARN_SERVER
      modelUri: gs://seldon-models/v1.19.0-dev/sklearn/iris
      name: classifier
      logger:
        mode: all
    name: default
    replicas: 3
```

    Overwriting deployment.yaml



```python
!kubectl apply -f deployment.yaml
```

    seldondeployment.machinelearning.seldon.io/sklearn configured



```python
!kubectl -n seldon rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=sklearn -o jsonpath='{.items[0].metadata.name}')
```

    deployment "sklearn-default-0-classifier" successfully rolled out


## Create Input Data


```python
import os
import random

random.seed(0)
with open("input-data.txt", "w") as f:
    for _ in range(10000):
        data = [random.random() for _ in range(4)]
        data = "[[" + ", ".join(str(x) for x in data) + "]]\n"
        f.write(data)
```


```bash
%%bash
HDFSCLI_CONFIG=./hdfscli.cfg hdfscli upload input-data.txt /batch-data/input-data.txt
```

## Prepare HDFS config / client image

For connecting to the `hdfs` from inside the cluster we will use the same `hdfscli` tool as we used above to put data in there.

We will configure `hdfscli` using `hdfscli.cfg` file stored inside kubernetes secret:


```python
%%writefile hdfs-config.yaml
apiVersion: v1
kind: Secret
metadata:
  name: seldon-hdfscli-secret-file
type: Opaque
stringData:
  hdfscli.cfg: |
    [global]
    default.alias = batch

    [batch.alias]
    url = http://hdfs-httpfs.hdfs-system.svc.cluster.local:14000
    user = hdfs
```

    Overwriting hdfs-config.yaml



```python
!kubectl apply -f hdfs-config.yaml
```

    secret/seldon-hdfscli-secret-file configured


For the client image we will use a following minimal Dockerfile


```python
%%writefile Dockerfile
FROM python:3.8
RUN pip install hdfs==2.5.8
ENV HDFSCLI_CONFIG /etc/hdfs/hdfscli.cfg
```

    Overwriting Dockerfile


That is build and published as `seldonio/hdfscli:1.6.0-dev`

## Create Workflow

This simple workflow will consist of three stages:
- download-input-data
- process-batch-inputs
- upload-output-data


```python
%%writefile workflow.yaml
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  name: sklearn-batch-job
  namespace: seldon

  labels:
    deployment-name: sklearn
    deployment-kind: SeldonDeployment

spec:
  volumeClaimTemplates:
  - metadata:
      name: seldon-job-pvc
      namespace: seldon
      ownerReferences:
      - apiVersion: argoproj.io/v1alpha1
        blockOwnerDeletion: true
        kind: Workflow
        name: '{{workflow.name}}'
        uid: '{{workflow.uid}}'
    spec:
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 1Gi

  volumes:
  - name: config
    secret:
      secretName: seldon-hdfscli-secret-file

  arguments:
    parameters:
    - name: batch_deployment_name
      value: sklearn
    - name: batch_namespace
      value: seldon

    - name: input_path
      value: /batch-data/input-data.txt
    - name: output_path
      value: /batch-data/output-data-{{workflow.name}}.txt

    - name: batch_gateway_type
      value: istio
    - name: batch_gateway_endpoint
      value: istio-ingressgateway.istio-system.svc.cluster.local
    - name: batch_transport_protocol
      value: rest
    - name: workers
      value: "10"
    - name: retries
      value: "3"
    - name: data_type
      value: data
    - name: payload_type
      value: ndarray

  entrypoint: seldon-batch-process

  templates:
  - name: seldon-batch-process
    steps:
    - - arguments: {}
        name: download-input-data
        template: download-input-data
    - - arguments: {}
        name: process-batch-inputs
        template: process-batch-inputs
    - - arguments: {}
        name: upload-output-data
        template: upload-output-data

  - name: download-input-data
    script:
      image: seldonio/hdfscli:1.6.0-dev
      volumeMounts:
      - mountPath: /assets
        name: seldon-job-pvc

      - mountPath: /etc/hdfs
        name: config
        readOnly: true

      env:
      - name: INPUT_DATA_PATH
        value: '{{workflow.parameters.input_path}}'

      - name: HDFSCLI_CONFIG
        value: /etc/hdfs/hdfscli.cfg

      command: [sh]
      source: |
        hdfscli download ${INPUT_DATA_PATH} /assets/input-data.txt

  - name: process-batch-inputs
    container:
      image: seldonio/seldon-core-s2i-python37:1.19.0-dev

      volumeMounts:
      - mountPath: /assets
        name: seldon-job-pvc

      env:
      - name: SELDON_BATCH_DEPLOYMENT_NAME
        value: '{{workflow.parameters.batch_deployment_name}}'
      - name: SELDON_BATCH_NAMESPACE
        value: '{{workflow.parameters.batch_namespace}}'
      - name: SELDON_BATCH_GATEWAY_TYPE
        value: '{{workflow.parameters.batch_gateway_type}}'
      - name: SELDON_BATCH_HOST
        value: '{{workflow.parameters.batch_gateway_endpoint}}'
      - name: SELDON_BATCH_TRANSPORT
        value: '{{workflow.parameters.batch_transport_protocol}}'
      - name: SELDON_BATCH_DATA_TYPE
        value: '{{workflow.parameters.data_type}}'
      - name: SELDON_BATCH_PAYLOAD_TYPE
        value: '{{workflow.parameters.payload_type}}'
      - name: SELDON_BATCH_WORKERS
        value: '{{workflow.parameters.workers}}'
      - name: SELDON_BATCH_RETRIES
        value: '{{workflow.parameters.retries}}'
      - name: SELDON_BATCH_INPUT_DATA_PATH
        value: /assets/input-data.txt
      - name: SELDON_BATCH_OUTPUT_DATA_PATH
        value: /assets/output-data.txt

      command: [seldon-batch-processor]
      args: [--benchmark]


  - name: upload-output-data
    script:
      image: seldonio/hdfscli:1.6.0-dev
      volumeMounts:
      - mountPath: /assets
        name: seldon-job-pvc

      - mountPath: /etc/hdfs
        name: config
        readOnly: true

      env:
      - name: OUTPUT_DATA_PATH
        value: '{{workflow.parameters.output_path}}'

      - name: HDFSCLI_CONFIG
        value: /etc/hdfs/hdfscli.cfg

      command: [sh]
      source: |
        hdfscli upload /assets/output-data.txt ${OUTPUT_DATA_PATH}
```

    Overwriting workflow.yaml



```python
!argo submit --serviceaccount workflow workflow.yaml
```

    Name:                sklearn-batch-job
    Namespace:           seldon
    ServiceAccount:      workflow
    Status:              Pending
    Created:             Thu Jan 14 18:36:52 +0000 (now)
    Progress:            
    Parameters:          
      batch_deployment_name: sklearn
      batch_namespace:   seldon
      input_path:        /batch-data/input-data.txt
      output_path:       /batch-data/output-data-{{workflow.name}}.txt
      batch_gateway_type: istio
      batch_gateway_endpoint: istio-ingressgateway.istio-system.svc.cluster.local
      batch_transport_protocol: rest
      workers:           10
      retries:           3
      data_type:         data
      payload_type:      ndarray



```python
!argo list
```

    NAME                STATUS    AGE   DURATION   PRIORITY
    sklearn-batch-job   Running   1s    1s         0



```python
!argo get sklearn-batch-job
```

    Name:                sklearn-batch-job
    Namespace:           seldon
    ServiceAccount:      workflow
    Status:              Running
    Created:             Thu Jan 14 18:36:52 +0000 (39 seconds ago)
    Started:             Thu Jan 14 18:36:52 +0000 (39 seconds ago)
    Duration:            39 seconds
    Progress:            1/2
    ResourcesDuration:   1s*(100Mi memory),1s*(1 cpu)
    Parameters:          
      batch_deployment_name: sklearn
      batch_namespace:   seldon
      input_path:        /batch-data/input-data.txt
      output_path:       /batch-data/output-data-{{workflow.name}}.txt
      batch_gateway_type: istio
      batch_gateway_endpoint: istio-ingressgateway.istio-system.svc.cluster.local
      batch_transport_protocol: rest
      workers:           10
      retries:           3
      data_type:         data
      payload_type:      ndarray
    
    [39mSTEP[0m                         TEMPLATE              PODNAME                       DURATION  MESSAGE
     [36m●[0m sklearn-batch-job         seldon-batch-process                                            
     ├───[32m✔[0m download-input-data   download-input-data   sklearn-batch-job-2227322232  6s          
     └───[36m●[0m process-batch-inputs  process-batch-inputs  sklearn-batch-job-2877616693  29s         



```python
!argo logs sklearn-batch-job
```

    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:05,000 - batch_processor.py:167 - INFO:  Processed instances: 100[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:05,417 - batch_processor.py:167 - INFO:  Processed instances: 200[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:06,213 - batch_processor.py:167 - INFO:  Processed instances: 300[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:06,642 - batch_processor.py:167 - INFO:  Processed instances: 400[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:06,974 - batch_processor.py:167 - INFO:  Processed instances: 500[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:07,278 - batch_processor.py:167 - INFO:  Processed instances: 600[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:07,628 - batch_processor.py:167 - INFO:  Processed instances: 700[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:08,378 - batch_processor.py:167 - INFO:  Processed instances: 800[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:09,003 - batch_processor.py:167 - INFO:  Processed instances: 900[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:09,337 - batch_processor.py:167 - INFO:  Processed instances: 1000[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:09,697 - batch_processor.py:167 - INFO:  Processed instances: 1100[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:10,014 - batch_processor.py:167 - INFO:  Processed instances: 1200[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:10,349 - batch_processor.py:167 - INFO:  Processed instances: 1300[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:10,843 - batch_processor.py:167 - INFO:  Processed instances: 1400[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:11,207 - batch_processor.py:167 - INFO:  Processed instances: 1500[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:11,562 - batch_processor.py:167 - INFO:  Processed instances: 1600[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:11,975 - batch_processor.py:167 - INFO:  Processed instances: 1700[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:12,350 - batch_processor.py:167 - INFO:  Processed instances: 1800[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:12,783 - batch_processor.py:167 - INFO:  Processed instances: 1900[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:13,139 - batch_processor.py:167 - INFO:  Processed instances: 2000[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:13,563 - batch_processor.py:167 - INFO:  Processed instances: 2100[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:13,928 - batch_processor.py:167 - INFO:  Processed instances: 2200[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:14,352 - batch_processor.py:167 - INFO:  Processed instances: 2300[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:14,699 - batch_processor.py:167 - INFO:  Processed instances: 2400[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:15,042 - batch_processor.py:167 - INFO:  Processed instances: 2500[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:15,701 - batch_processor.py:167 - INFO:  Processed instances: 2600[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:16,124 - batch_processor.py:167 - INFO:  Processed instances: 2700[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:16,748 - batch_processor.py:167 - INFO:  Processed instances: 2800[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:17,300 - batch_processor.py:167 - INFO:  Processed instances: 2900[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:17,904 - batch_processor.py:167 - INFO:  Processed instances: 3000[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:18,454 - batch_processor.py:167 - INFO:  Processed instances: 3100[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:18,823 - batch_processor.py:167 - INFO:  Processed instances: 3200[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:19,236 - batch_processor.py:167 - INFO:  Processed instances: 3300[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:19,586 - batch_processor.py:167 - INFO:  Processed instances: 3400[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:20,317 - batch_processor.py:167 - INFO:  Processed instances: 3500[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:20,948 - batch_processor.py:167 - INFO:  Processed instances: 3600[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:21,356 - batch_processor.py:167 - INFO:  Processed instances: 3700[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:21,851 - batch_processor.py:167 - INFO:  Processed instances: 3800[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:22,205 - batch_processor.py:167 - INFO:  Processed instances: 3900[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:22,553 - batch_processor.py:167 - INFO:  Processed instances: 4000[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:23,051 - batch_processor.py:167 - INFO:  Processed instances: 4100[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:23,557 - batch_processor.py:167 - INFO:  Processed instances: 4200[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:24,016 - batch_processor.py:167 - INFO:  Processed instances: 4300[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:24,350 - batch_processor.py:167 - INFO:  Processed instances: 4400[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:24,883 - batch_processor.py:167 - INFO:  Processed instances: 4500[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:25,295 - batch_processor.py:167 - INFO:  Processed instances: 4600[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:25,669 - batch_processor.py:167 - INFO:  Processed instances: 4700[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:26,055 - batch_processor.py:167 - INFO:  Processed instances: 4800[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:26,795 - batch_processor.py:167 - INFO:  Processed instances: 4900[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:27,462 - batch_processor.py:167 - INFO:  Processed instances: 5000[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:27,887 - batch_processor.py:167 - INFO:  Processed instances: 5100[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:28,332 - batch_processor.py:167 - INFO:  Processed instances: 5200[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:28,742 - batch_processor.py:167 - INFO:  Processed instances: 5300[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:29,069 - batch_processor.py:167 - INFO:  Processed instances: 5400[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:29,443 - batch_processor.py:167 - INFO:  Processed instances: 5500[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:29,840 - batch_processor.py:167 - INFO:  Processed instances: 5600[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:30,235 - batch_processor.py:167 - INFO:  Processed instances: 5700[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:30,578 - batch_processor.py:167 - INFO:  Processed instances: 5800[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:31,024 - batch_processor.py:167 - INFO:  Processed instances: 5900[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:31,381 - batch_processor.py:167 - INFO:  Processed instances: 6000[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:31,847 - batch_processor.py:167 - INFO:  Processed instances: 6100[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:32,239 - batch_processor.py:167 - INFO:  Processed instances: 6200[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:32,603 - batch_processor.py:167 - INFO:  Processed instances: 6300[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:33,080 - batch_processor.py:167 - INFO:  Processed instances: 6400[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:33,567 - batch_processor.py:167 - INFO:  Processed instances: 6500[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:34,043 - batch_processor.py:167 - INFO:  Processed instances: 6600[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:34,444 - batch_processor.py:167 - INFO:  Processed instances: 6700[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:34,812 - batch_processor.py:167 - INFO:  Processed instances: 6800[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:35,148 - batch_processor.py:167 - INFO:  Processed instances: 6900[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:35,519 - batch_processor.py:167 - INFO:  Processed instances: 7000[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:35,873 - batch_processor.py:167 - INFO:  Processed instances: 7100[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:36,278 - batch_processor.py:167 - INFO:  Processed instances: 7200[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:36,694 - batch_processor.py:167 - INFO:  Processed instances: 7300[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:37,061 - batch_processor.py:167 - INFO:  Processed instances: 7400[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:37,509 - batch_processor.py:167 - INFO:  Processed instances: 7500[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:37,865 - batch_processor.py:167 - INFO:  Processed instances: 7600[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:38,211 - batch_processor.py:167 - INFO:  Processed instances: 7700[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:38,590 - batch_processor.py:167 - INFO:  Processed instances: 7800[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:39,028 - batch_processor.py:167 - INFO:  Processed instances: 7900[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:39,419 - batch_processor.py:167 - INFO:  Processed instances: 8000[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:39,910 - batch_processor.py:167 - INFO:  Processed instances: 8100[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:40,532 - batch_processor.py:167 - INFO:  Processed instances: 8200[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:41,022 - batch_processor.py:167 - INFO:  Processed instances: 8300[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:41,436 - batch_processor.py:167 - INFO:  Processed instances: 8400[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:41,800 - batch_processor.py:167 - INFO:  Processed instances: 8500[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:42,238 - batch_processor.py:167 - INFO:  Processed instances: 8600[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:42,704 - batch_processor.py:167 - INFO:  Processed instances: 8700[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:43,079 - batch_processor.py:167 - INFO:  Processed instances: 8800[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:43,712 - batch_processor.py:167 - INFO:  Processed instances: 8900[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:44,075 - batch_processor.py:167 - INFO:  Processed instances: 9000[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:44,459 - batch_processor.py:167 - INFO:  Processed instances: 9100[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:44,806 - batch_processor.py:167 - INFO:  Processed instances: 9200[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:45,344 - batch_processor.py:167 - INFO:  Processed instances: 9300[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:45,764 - batch_processor.py:167 - INFO:  Processed instances: 9400[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:46,110 - batch_processor.py:167 - INFO:  Processed instances: 9500[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:46,547 - batch_processor.py:167 - INFO:  Processed instances: 9600[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:46,987 - batch_processor.py:167 - INFO:  Processed instances: 9700[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:47,371 - batch_processor.py:167 - INFO:  Processed instances: 9800[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:47,905 - batch_processor.py:167 - INFO:  Processed instances: 9900[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:48,289 - batch_processor.py:167 - INFO:  Processed instances: 10000[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:48,290 - batch_processor.py:168 - INFO:  Total processed instances: 10000[0m
    [35msklearn-batch-job-2877616693: 2021-01-14 18:37:48,290 - batch_processor.py:116 - INFO:  Elapsed time: 43.7087140083313[0m


## Pull output-data from hdfs


```bash
%%bash
HDFSCLI_CONFIG=./hdfscli.cfg hdfscli download /batch-data/output-data-sklearn-batch-job.txt output-data.txt
```


```python
!head output-data.txt
```

    {"data": {"names": ["t:0", "t:1", "t:2"], "ndarray": [[0.49551509682202705, 0.4192462053867995, 0.08523869779117352]]}, "meta": {"requestPath": {"classifier": "seldonio/sklearnserver:1.6.0-dev"}, "tags": {"tags": {"batch_id": "409a3f56-5697-11eb-be58-06ee7f9820ec", "batch_index": 0.0, "batch_instance_id": "409ad222-5697-11eb-90de-06ee7f9820ec"}}}}
    {"data": {"names": ["t:0", "t:1", "t:2"], "ndarray": [[0.14889581912569078, 0.40048258722097885, 0.45062159365333043]]}, "meta": {"requestPath": {"classifier": "seldonio/sklearnserver:1.6.0-dev"}, "tags": {"tags": {"batch_id": "409a3f56-5697-11eb-be58-06ee7f9820ec", "batch_index": 10.0, "batch_instance_id": "409d56e6-5697-11eb-90de-06ee7f9820ec"}}}}
    {"data": {"names": ["t:0", "t:1", "t:2"], "ndarray": [[0.1859090109477526, 0.46433848375587844, 0.349752505296369]]}, "meta": {"requestPath": {"classifier": "seldonio/sklearnserver:1.6.0-dev"}, "tags": {"tags": {"batch_id": "409a3f56-5697-11eb-be58-06ee7f9820ec", "batch_index": 1.0, "batch_instance_id": "409ad68c-5697-11eb-90de-06ee7f9820ec"}}}}
    {"data": {"names": ["t:0", "t:1", "t:2"], "ndarray": [[0.35453094061556073, 0.3866773326679568, 0.2587917267164825]]}, "meta": {"requestPath": {"classifier": "seldonio/sklearnserver:1.6.0-dev"}, "tags": {"tags": {"batch_id": "409a3f56-5697-11eb-be58-06ee7f9820ec", "batch_index": 3.0, "batch_instance_id": "409bb106-5697-11eb-90de-06ee7f9820ec"}}}}
    {"data": {"names": ["t:0", "t:1", "t:2"], "ndarray": [[0.14218706541271167, 0.2726759160836421, 0.5851370185036463]]}, "meta": {"requestPath": {"classifier": "seldonio/sklearnserver:1.6.0-dev"}, "tags": {"tags": {"batch_id": "409a3f56-5697-11eb-be58-06ee7f9820ec", "batch_index": 13.0, "batch_instance_id": "409dabc8-5697-11eb-90de-06ee7f9820ec"}}}}
    {"data": {"names": ["t:0", "t:1", "t:2"], "ndarray": [[0.15720251854631545, 0.3840752321558323, 0.45872224929785227]]}, "meta": {"requestPath": {"classifier": "seldonio/sklearnserver:1.6.0-dev"}, "tags": {"tags": {"batch_id": "409a3f56-5697-11eb-be58-06ee7f9820ec", "batch_index": 2.0, "batch_instance_id": "409b7362-5697-11eb-90de-06ee7f9820ec"}}}}
    {"data": {"names": ["t:0", "t:1", "t:2"], "ndarray": [[0.1808891729172985, 0.32704139903027096, 0.49206942805243054]]}, "meta": {"requestPath": {"classifier": "seldonio/sklearnserver:1.6.0-dev"}, "tags": {"tags": {"batch_id": "409a3f56-5697-11eb-be58-06ee7f9820ec", "batch_index": 14.0, "batch_instance_id": "409dac86-5697-11eb-90de-06ee7f9820ec"}}}}
    {"data": {"names": ["t:0", "t:1", "t:2"], "ndarray": [[0.14218974549047703, 0.41059890080264444, 0.4472113537068785]]}, "meta": {"requestPath": {"classifier": "seldonio/sklearnserver:1.6.0-dev"}, "tags": {"tags": {"batch_id": "409a3f56-5697-11eb-be58-06ee7f9820ec", "batch_index": 15.0, "batch_instance_id": "409de20a-5697-11eb-90de-06ee7f9820ec"}}}}
    {"data": {"names": ["t:0", "t:1", "t:2"], "ndarray": [[0.2643002677975754, 0.44720843507174224, 0.28849129713068233]]}, "meta": {"requestPath": {"classifier": "seldonio/sklearnserver:1.6.0-dev"}, "tags": {"tags": {"batch_id": "409a3f56-5697-11eb-be58-06ee7f9820ec", "batch_index": 16.0, "batch_instance_id": "409dfe98-5697-11eb-90de-06ee7f9820ec"}}}}
    {"data": {"names": ["t:0", "t:1", "t:2"], "ndarray": [[0.2975075875929912, 0.25439317776178244, 0.44809923464522644]]}, "meta": {"requestPath": {"classifier": "seldonio/sklearnserver:1.6.0-dev"}, "tags": {"tags": {"batch_id": "409a3f56-5697-11eb-be58-06ee7f9820ec", "batch_index": 4.0, "batch_instance_id": "409c3a2c-5697-11eb-90de-06ee7f9820ec"}}}}



```python
!kubectl delete -f deployment.yaml
```

    seldondeployment.machinelearning.seldon.io "sklearn" deleted



```python
!argo delete sklearn-batch-job
```

    Workflow 'sklearn-batch-job' deleted

