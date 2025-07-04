# Setup Azure Kubernetes Infrastructure
In this notebook we will 
- Login to Aure account
- [Create AKS cluster with](#Create-AKS-Cluster-and-NodePools)
  - **GPU enabled Spot VM nodepool** for running ML elastic training
  - **CPU VM nodepool** for running typical workloads 
- [Azure Storage Account for hosting model data](#Create-Storage-Account-for-training-data)
- Deploy Kubernetes Components
  - [Install **Azure Blob CSI Driver**](#Install-Kubernetes-Blob-CSI-Driver) to map Blob storage to container as persistent volumes
  - [Create Kubernetes **PersistentVolume** and PersistentVolumeClaim](#Create-Persistent-Volume-for-Azure-Blob)

## Define Variables
Set variables required for the project


```python
subscription_id = "<xxxx-xxxx-xxxx-xxxx>"  # fill in
resource_group = "seldon"  # feel free to replace or use this default
region = "eastus2"  # ffeel free to replace or use this default

storage_account_name = "modeltestsgpt"  # fill in
storage_container_name = "gpt2tf"

aks_name = "modeltests"  # feel free to replace or use this default
aks_gpupool = "gpunodes"  # feel free to replace or use this default
aks_cpupool = "cpunodes"  # feel free to replace or use this default
aks_gpu_sku = "Standard_NC6s_v3"  # feel free to replace or use this default
aks_cpu_sku = "Standard_F8s_v2"
```

## Azure account login
If you are not already logged in to an Azure account, the command below will initiate a login. This will pop up a browser where you can select your login. (if no web browser is available or if the web browser fails to open, use device code flow with `az login --use-device-code` or login in WSL command  prompt and proceed to notebook)


```bash
%%bash
az login -o table

```


```python
!az account set --subscription "$subscription_id"
```


```python
!az account show
```

## Create Resource Group
Azure encourages the use of groups to organize all the Azure components you deploy. That way it is easier to find them but also we can delete a number of resources simply by deleting the group.


```python
!az group create -l {region} -n {resource_group}
```

## Create AKS Cluster and NodePools <a id="aks"/>
Below, we create the AKS cluster with default 1 system node (to save time, in production use more nodes as per best practices) in the resource group we created earlier. This step can take 5 or more minutes.



```python
%%time
!az aks create --resource-group {resource_group} \
    --name {aks_name} \
    --node-vm-size Standard_D8s_v3  \
    --node-count 1 \
    --location {region}  \
    --kubernetes-version 1.18.17 \
    --node-osdisk-type Ephemeral \    
    --generate-ssh-keys
```

## Connect to AKS Cluster
To configure kubectl to connect to Kubernetes cluster, run the following command


```python
!az aks get-credentials --resource-group {resource_group} --name {aks_name}
```

Let's verify connection by listing the nodes.


```python
!kubectl get nodes
```

    NAME                                STATUS   ROLES   AGE     VERSION
    aks-agentpool-28613018-vmss000000   Ready    agent   28d     v1.19.9
    aks-agentpool-28613018-vmss000001   Ready    agent   28d     v1.19.9
    aks-agentpool-28613018-vmss000002   Ready    agent   28d     v1.19.9
    aks-cpunodes-28613018-vmss000000    Ready    agent   28d     v1.19.9
    aks-cpunodes-28613018-vmss000001    Ready    agent   28d     v1.19.9
    aks-gpunodes-28613018-vmss000001    Ready    agent   5h27m   v1.19.9


Taint System node with `CriticalAddonsOnly` taint so it is available only for system workloads


```python
!kubectl taint nodes -l kubernetes.azure.com/mode=system CriticalAddonsOnly=true:NoSchedule --overwrite
```

## Create GPU enabled and CPU Node Pools
To create GPU enabled nodepool, will use fully configured AKS image that contains the NVIDIA device plugin for Kubenetes, see [Use the AKS specialized GPU image (preview)](https://docs.microsoft.com/en-us/azure/aks/gpu-cluster#use-the-aks-specialized-gpu-image-preview). Creating nodepools could take five or more minutes.


```python
%%time
!az feature register --name GPUDedicatedVHDPreview --namespace Microsoft.ContainerService
!az feature list -o table --query "[?contains(name, 'Microsoft.ContainerService/GPUDedicatedVHDPreview')].{Name:name,State:properties.state}"
!az provider register --namespace Microsoft.ContainerService
!az extension add --name aks-preview
```

## Create  GPU NodePool with GPU taint
For more information on Azure Nodepools https://docs.microsoft.com/en-us/azure/aks/use-multiple-node-pools 


```python
%%time
print ({aks_gpu_sku})
!az aks nodepool add \
    --resource-group {resource_group} \
    --cluster-name {aks_name} \
    --name {aks_gpupool} \
    --node-taints nvidia.com=gpu:NoSchedule \
    --node-count 1 \
    --node-vm-size  {aks_gpu_sku} \
    --aks-custom-headers UseGPUDedicatedVHD=true,usegen2vm=true
```

    {
    [33mThe behavior of this command has been altered by the following extension: aks-preview[0m
    [91mNode pool gpunodes already exists, please try a different name, use 'aks nodepool list' to get current list of node pool[0m
    [0mCPU times: user 275 ms, sys: 79 ms, total: 354 ms
    Wall time: 5.38 s


## Verify GPU is available on Kubernetes Node
Now use the kubectl describe node command to confirm that the GPUs are schedulable. Under the Capacity section, for Standard_NC12 sku the GPU should list as `nvidia.com/gpu: 2`


```python
!kubectl describe node -l accelerator=nvidia | grep nvidia -A 5 -B 5
```

    Name:               aks-gpunodes-28613018-vmss000001
    Roles:              agent
    Labels:             accelerator=nvidia
                        agentpool=gpunodes
                        beta.kubernetes.io/arch=amd64
                        beta.kubernetes.io/instance-type=Standard_NC12
                        beta.kubernetes.io/os=linux
                        failure-domain.beta.kubernetes.io/region=eastus2
    --
      cpu:                            12
      ephemeral-storage:              129900528Ki
      hugepages-1Gi:                  0
      hugepages-2Mi:                  0
      memory:                         115387540Ki
      nvidia.com/gpu:                 2
      pods:                           30
    Allocatable:
      attachable-volumes-azure-disk:  48
      cpu:                            11780m
      ephemeral-storage:              119716326407
      hugepages-1Gi:                  0
      hugepages-2Mi:                  0
      memory:                         105854100Ki
      nvidia.com/gpu:                 2
      pods:                           30
    System Info:
      Machine ID:                 db67bd967e1441febad873ba49d35adc
      System UUID:                f39ce4bc-11c6-8643-8a8a-dfb4998a0524
      Boot ID:                    eb926e42-d4e7-4760-b124-9b09c0e56c57
    --
      memory                         275Mi (0%)  850Mi (0%)
      ephemeral-storage              0 (0%)      0 (0%)
      hugepages-1Gi                  0 (0%)      0 (0%)
      hugepages-2Mi                  0 (0%)      0 (0%)
      attachable-volumes-azure-disk  0           0
      nvidia.com/gpu                 0           0
    Events:                          <none>


## Create CPU NodePool for running regular workloads


```python
%%time
!az aks nodepool add \
  --resource-group {resource_group} \
    --cluster-name {aks_name} \
    --name {aks_cpupool} \
    --enable-cluster-autoscaler \
    --node-osdisk-type Ephemeral \
    --min-count 1 \
    --max-count 3 \
    --node-vm-size {aks_cpu_sku}  \
    --node-osdisk-size 128 
```

    [33mThe behavior of this command has been altered by the following extension: aks-preview[0m
    {
      "agentPoolType": "VirtualMachineScaleSets",
      "availabilityZones": null,
      "count": 3,
      "enableAutoScaling": true,
      "enableEncryptionAtHost": false,
      "enableFips": false,
      "enableNodePublicIp": false,
      "gpuInstanceProfile": null,
      "id": "/subscriptions/xxxx-xxxx-xxxx-xxxx-xxxxxx/resourcegroups/seldon/providers/Microsoft.ContainerService/managedClusters/modeltests/agentPools/cpunodes",
      "kubeletConfig": null,
      "kubeletDiskType": "OS",
      "linuxOsConfig": null,
      "maxCount": 3,
      "maxPods": 30,
      "minCount": 1,
      "mode": "User",
      "name": "cpunodes",
      "nodeImageVersion": "AKSUbuntu-1804gen2containerd-2021.05.08",
      "nodeLabels": null,
      "nodePublicIpPrefixId": null,
      "nodeTaints": null,
      "orchestratorVersion": "1.19.9",
      "osDiskSizeGb": 128,
      "osDiskType": "Ephemeral",
      "osSku": "Ubuntu",
      "osType": "Linux",
      "podSubnetId": null,
      "powerState": {
        "code": "Running"
      },
      "provisioningState": "Succeeded",
      "proximityPlacementGroupId": null,
      "resourceGroup": "seldon",
      "scaleSetEvictionPolicy": null,
      "scaleSetPriority": null,
      "spotMaxPrice": null,
      "tags": null,
      "type": "Microsoft.ContainerService/managedClusters/agentPools",
      "upgradeSettings": {
        "maxSurge": null
      },
      "vmSize": "Standard_F8s_v2",
      "vnetSubnetId": "/subscriptions/xxxxx-xxxx-xxxx-xxxxx-xxxxxx/resourceGroups/seldon/providers/Microsoft.Network/virtualNetworks/seldon-vnet/subnets/default"
    }
    [K[0mCPU times: user 4.17 s, sys: 1.51 s, total: 5.68 s
    Wall time: 2min 36s


## Verify Taints on the Kubernetes nodes
Verify that system pool and have the Taints `CriticalAddonsOnly` and `sku=gpu` respectively   



```python
!kubectl get nodes -o json | jq '.items[].spec.taints'
```

    [1;39m[
      [1;39m{
        [0m[34;1m"effect"[0m[1;39m: [0m[0;32m"NoSchedule"[0m[1;39m,
        [0m[34;1m"key"[0m[1;39m: [0m[0;32m"CriticalAddonsOnly"[0m[1;39m,
        [0m[34;1m"value"[0m[1;39m: [0m[0;32m"true"[0m[1;39m
      [1;39m}[0m[1;39m
    [1;39m][0m
    [1;39m[
      [1;39m{
        [0m[34;1m"effect"[0m[1;39m: [0m[0;32m"NoSchedule"[0m[1;39m,
        [0m[34;1m"key"[0m[1;39m: [0m[0;32m"CriticalAddonsOnly"[0m[1;39m,
        [0m[34;1m"value"[0m[1;39m: [0m[0;32m"true"[0m[1;39m
      [1;39m}[0m[1;39m
    [1;39m][0m
    [1;39m[
      [1;39m{
        [0m[34;1m"effect"[0m[1;39m: [0m[0;32m"NoSchedule"[0m[1;39m,
        [0m[34;1m"key"[0m[1;39m: [0m[0;32m"CriticalAddonsOnly"[0m[1;39m,
        [0m[34;1m"value"[0m[1;39m: [0m[0;32m"true"[0m[1;39m
      [1;39m}[0m[1;39m
    [1;39m][0m
    [1;30mnull[0m
    [1;30mnull[0m
    [1;39m[
      [1;39m{
        [0m[34;1m"effect"[0m[1;39m: [0m[0;32m"NoSchedule"[0m[1;39m,
        [0m[34;1m"key"[0m[1;39m: [0m[0;32m"sku"[0m[1;39m,
        [0m[34;1m"value"[0m[1;39m: [0m[0;32m"gpu"[0m[1;39m
      [1;39m}[0m[1;39m
    [1;39m][0m


# Create Storage Account for training data <a id="storageaccount"/>
In this section of the notebook, we'll create an Azure blob storage that we'll use throughout the tutorial. This object store will be used to store input images and save checkpoints. Use `az cli` to create the account


```python
%%time
!az storage account create -n {storage_account_name} -g {resource_group} --query 'provisioningState'
```

    "Succeeded"
    [K[0mCPU times: user 674 ms, sys: 214 ms, total: 888 ms
    Wall time: 22 s


Grab the keys of the storage account that was just created.We would need them for binding Kubernetes Persistent Volume. The --quote '[0].value' part of the command simply means to select the value of the zero-th indexed of the set of keys.


```python
key = !az storage account keys list --account-name {storage_account_name} -g {resource_group} --query '[0].value' -o tsv
```


The stdout from the command above is stored in a string array of 1. Select the element in the array.


```python
storage_account_key = key[0]
```


```python
# create storage container

!az storage container create \
    --account-name {storage_account_name} \
    --account-key {storage_account_key} \
    --name {storage_container_name}
```

    {
      "created": true
    }
    [0m

# Install Kubernetes Blob CSI Driver <a id="csidriver"/>
[Azure Blob Storage CSI driver for Kubernetes](https://github.com/kubernetes-sigs/blob-csi-driver) allows Kubernetes to access Azure Storage. We will deploy it using Helm3 package manager as described in the docs https://github.com/kubernetes-sigs/blob-csi-driver/tree/master/charts


```python
!az aks get-credentials --resource-group {resource_group} --name {aks_name}
```


```python
!helm repo add blob-csi-driver https://raw.githubusercontent.com/kubernetes-sigs/blob-csi-driver/master/charts
!helm install blob-csi-driver blob-csi-driver/blob-csi-driver --namespace kube-system --version v1.1.0
```

    "blob-csi-driver" already exists with the same configuration, skipping
    W0527 23:11:20.183604   13719 warnings.go:70] storage.k8s.io/v1beta1 CSIDriver is deprecated in v1.19+, unavailable in v1.22+; use storage.k8s.io/v1 CSIDriver
    W0527 23:11:20.506450   13719 warnings.go:70] storage.k8s.io/v1beta1 CSIDriver is deprecated in v1.19+, unavailable in v1.22+; use storage.k8s.io/v1 CSIDriver
    NAME: blob-csi-driver
    LAST DEPLOYED: Thu May 27 23:11:19 2021
    NAMESPACE: kube-system
    STATUS: deployed
    REVISION: 1
    TEST SUITE: None
    NOTES:
    The Azure Blob Storage CSI driver is getting deployed to your cluster.
    
    To check Azure Blob Storage CSI driver pods status, please run:
    
      kubectl --namespace=kube-system get pods --selector="release=blob-csi-driver" --watch



```python
!kubectl -n kube-system get pods -l "app.kubernetes.io/instance=blob-csi-driver"
```

    NAME                                   READY   STATUS    RESTARTS   AGE
    csi-blob-controller-7b9db4967c-fbsm2   4/4     Running   0          22s
    csi-blob-controller-7b9db4967c-hdglw   4/4     Running   0          22s
    csi-blob-node-7tgl8                    3/3     Running   0          22s
    csi-blob-node-89rkn                    3/3     Running   0          22s
    csi-blob-node-nnhfh                    3/3     Running   0          22s
    csi-blob-node-pb584                    3/3     Running   0          22s
    csi-blob-node-q6z6t                    3/3     Running   0          22s
    csi-blob-node-tq4mh                    3/3     Running   0          22s


## Create Persistent Volume for Azure Blob <a id="pv"/>
For more details on creating   `PersistentVolume` using CSI driver refer to https://github.com/kubernetes-sigs/blob-csi-driver/blob/master/deploy/example/e2e_usage.md


```python
# Create secret to access storage account
!kubectl create secret generic azure-blobsecret --from-literal azurestorageaccountname={storage_account_name} --from-literal azurestorageaccountkey="{storage_account_key}" --type=Opaque
```

    secret/azure-blobsecret created


Persistent Volume YAML definition is in `azure-blobfules-pv.yaml` with fields pointing to secret created above and containername we created in storage account:
```yaml
  csi:
    driver: blob.csi.azure.com
    readOnly: false
    volumeHandle: trainingdata  # make sure this volumeid is unique in the cluster
    volumeAttributes:
      containerName: workerdata # !! Modify if changed in Notebook
    nodeStageSecretRef:
      name: azure-blobsecret
     
```


```python
%%writefile azure-blobfuse-pv.yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: pv-gptblob
  
spec:
  capacity:
    storage: 10Gi
  accessModes:
    - ReadWriteMany
  persistentVolumeReclaimPolicy: Retain  # "Delete" is not supported in static provisioning
  csi:
    driver: blob.csi.azure.com
    readOnly: false
    volumeHandle: trainingdata  # make sure this volumeid is unique in the cluster
    volumeAttributes:
      containerName: gpt2onnx # Modify if changed in Notebook
    nodeStageSecretRef:
      name: azure-blobsecret
      namespace: default
  mountOptions:
    - -o uid=8888     # user in  Pod security context
    - -o allow_other    
    
---
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: pvc-gptblob
 
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 10Gi
  volumeName: pv-gptblob
  storageClassName: ""

```

    Overwriting azure-blobfuse-pv.yaml



```python
# Create PersistentVolume and PersistenVollumeClaim for container mounts
!kubectl apply -f  azure-blobfuse-pv.yaml
```

    persistentvolume/pv-gptblob created
    persistentvolumeclaim/pvc-gptblob created



```python
# Verify PVC is bound
!kubectl get pv,pvc
```

    NAME                          CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS        CLAIM                 STORAGECLASS   REASON   AGE
    persistentvolume/pv-blob      10Gi       RWX            Retain           Terminating   default/pvc-blob                              113m
    persistentvolume/pv-gptblob   10Gi       RWX            Retain           Bound         default/pvc-gptblob                           18s
    
    NAME                                STATUS        VOLUME       CAPACITY   ACCESS MODES   STORAGECLASS   AGE
    persistentvolumeclaim/pvc-blob      Terminating   pv-blob      10Gi       RWX                           113m
    persistentvolumeclaim/pvc-gptblob   Bound         pv-gptblob   10Gi       RWX                           17s


In the end of this step you will have AKS cluster and Storage account in resource group. ALK cluster will have cpu and gpu nodepools in addition to system nodepool.

