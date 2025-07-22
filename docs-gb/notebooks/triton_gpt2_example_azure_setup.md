# Setup Azure Kubernetes Infrastructure

In this notebook we will:
- Login to Azure account
- [Create AKS cluster with](#Create-AKS-Cluster-and-NodePools)
  - **GPU enabled Spot VM nodepool** for running ML elastic training
  - **CPU VM nodepool** for running typical workloads
- [Azure Storage Account for hosting model data](#Create-Storage-Account-for-training-data)
- Deploy Kubernetes Components
  - [Install **Azure Blob CSI Driver**](#Install-Kubernetes-Blob-CSI-Driver) to map Blob storage to container as persistent volumes
  - [Create Kubernetes **PersistentVolume** and PersistentVolumeClaim](#Create-Persistent-Volume-for-Azure-Blob)

---

## Define Variables
Set variables required for the project:

```python
subscription_id = "<xxxx-xxxx-xxxx-xxxx>"  # fill in
resource_group = "seldon"  # feel free to replace or use this default
region = "eastus2"  # feel free to replace or use this default

storage_account_name = "modeltestsgpt"  # fill in
storage_container_name = "gpt2tf"

aks_name = "modeltests"  # feel free to replace or use this default
aks_gpupool = "gpunodes"  # feel free to replace or use this default
aks_cpupool = "cpunodes"  # feel free to replace or use this default
aks_gpu_sku = "Standard_NC6s_v3"  # feel free to replace or use this default
aks_cpu_sku = "Standard_F8s_v2"
```

---

## Azure account login
If you are not already logged in to an Azure account, the command below will initiate a login. This will pop up a browser where you can select your login. (If no web browser is available or if the web browser fails to open, use device code flow with `az login --use-device-code` or login in WSL command prompt and proceed to notebook)

```bash
az login -o table
```

```bash
az account set --subscription "$subscription_id"
```

```bash
az account show
```

---

## Create Resource Group
Azure encourages the use of groups to organize all the Azure components you deploy. That way it is easier to find them but also we can delete a number of resources simply by deleting the group.

```bash
az group create -l $region -n $resource_group
```

---

## Create AKS Cluster and NodePools
Below, we create the AKS cluster with default 1 system node (to save time, in production use more nodes as per best practices) in the resource group we created earlier. This step can take 5 or more minutes.

```bash
time az aks create --resource-group $resource_group \
    --name $aks_name \
    --node-vm-size Standard_D8s_v3  \
    --node-count 1 \
    --location $region  \
    --kubernetes-version 1.18.17 \
    --node-osdisk-type Ephemeral \
    --generate-ssh-keys
```

---

## Connect to AKS Cluster
To configure kubectl to connect to Kubernetes cluster, run the following command:

```bash
az aks get-credentials --resource-group $resource_group --name $aks_name
```

Let's verify connection by listing the nodes:

```bash
kubectl get nodes
```

Example output:
```
NAME                                STATUS   ROLES   AGE     VERSION
aks-agentpool-28613018-vmss000000   Ready    agent   28d     v1.19.9
aks-agentpool-28613018-vmss000001   Ready    agent   28d     v1.19.9
aks-agentpool-28613018-vmss000002   Ready    agent   28d     v1.19.9
aks-cpunodes-28613018-vmss000000    Ready    agent   28d     v1.19.9
aks-cpunodes-28613018-vmss000001    Ready    agent   28d     v1.19.9
aks-gpunodes-28613018-vmss000001    Ready    agent   5h27m   v1.19.9
```

---

Taint System node with `CriticalAddonsOnly` taint so it is available only for system workloads:

```bash
kubectl taint nodes -l kubernetes.azure.com/mode=system CriticalAddonsOnly=true:NoSchedule --overwrite
```

---

## Create GPU enabled and CPU Node Pools
To create GPU enabled nodepool, will use fully configured AKS image that contains the NVIDIA device plugin for Kubernetes, see [Use the AKS specialized GPU image (preview)](https://docs.microsoft.com/en-us/azure/aks/gpu-cluster#use-the-aks-specialized-gpu-image-preview). Creating nodepools could take five or more minutes.

```bash
time az feature register --name GPUDedicatedVHDPreview --namespace Microsoft.ContainerService
time az feature list -o table --query "[?contains(name, 'Microsoft.ContainerService/GPUDedicatedVHDPreview')].{Name:name,State:properties.state}"
time az provider register --namespace Microsoft.ContainerService
time az extension add --name aks-preview
```

---

## Create GPU NodePool with GPU taint
For more information on Azure Nodepools: https://docs.microsoft.com/en-us/azure/aks/use-multiple-node-pools

```bash
time az aks nodepool add \
    --resource-group $resource_group \
    --cluster-name $aks_name \
    --name $aks_gpupool \
    --node-taints nvidia.com=gpu:NoSchedule \
    --node-count 1 \
    --node-vm-size  $aks_gpu_sku \
    --aks-custom-headers UseGPUDedicatedVHD=true,usegen2vm=true
```

---

## Verify GPU is available on Kubernetes Node
Now use the kubectl describe node command to confirm that the GPUs are schedulable. Under the Capacity section, for Standard_NC12 sku the GPU should list as `nvidia.com/gpu: 2`

```bash
kubectl describe node -l accelerator=nvidia | grep nvidia -A 5 -B 5
```

---

## Create CPU NodePool for running regular workloads

```bash
time az aks nodepool add \
  --resource-group $resource_group \
  --cluster-name $aks_name \
  --name $aks_cpupool \
  --enable-cluster-autoscaler \
  --node-osdisk-type Ephemeral \
  --min-count 1 \
  --max-count 3 \
  --node-vm-size $aks_cpu_sku  \
  --node-osdisk-size 128
```

---

## Verify Taints on the Kubernetes nodes
Verify that system pool and have the Taints `CriticalAddonsOnly` and `sku=gpu` respectively:

```bash
kubectl get nodes -o json | jq '.items[].spec.taints'
```

---

# Create Storage Account for training data
In this section of the notebook, we'll create an Azure blob storage that we'll use throughout the tutorial. This object store will be used to store input images and save checkpoints. Use `az cli` to create the account.

```bash
time az storage account create -n $storage_account_name -g $resource_group --query 'provisioningState'
```

Grab the keys of the storage account that was just created. We would need them for binding Kubernetes Persistent Volume. The --quote '[0].value' part of the command simply means to select the value of the zero-th indexed of the set of keys.

```bash
az storage account keys list --account-name $storage_account_name -g $resource_group --query '[0].value' -o tsv
```

The stdout from the command above is stored in a string array of 1. Select the element in the array.

```python
storage_account_key = key[0]
```

```bash
# create storage container
az storage container create \
    --account-name $storage_account_name \
    --account-key $storage_account_key \
    --name $storage_container_name
```

---

# Install Kubernetes Blob CSI Driver
[Azure Blob Storage CSI driver for Kubernetes](https://github.com/kubernetes-sigs/blob-csi-driver) allows Kubernetes to access Azure Storage. We will deploy it using Helm3 package manager as described in the docs https://github.com/kubernetes-sigs/blob-csi-driver/tree/master/charts

```bash
az aks get-credentials --resource-group $resource_group --name $aks_name
helm repo add blob-csi-driver https://raw.githubusercontent.com/kubernetes-sigs/blob-csi-driver/master/charts
helm install blob-csi-driver blob-csi-driver/blob-csi-driver --namespace kube-system --version v1.1.0
```

```bash
kubectl -n kube-system get pods -l "app.kubernetes.io/instance=blob-csi-driver"
```

---

## Create Persistent Volume for Azure Blob
For more details on creating `PersistentVolume` using CSI driver refer to https://github.com/kubernetes-sigs/blob-csi-driver/blob/master/deploy/example/e2e_usage.md

```bash
# Create secret to access storage account
kubectl create secret generic azure-blobsecret --from-literal azurestorageaccountname=$storage_account_name --from-literal azurestorageaccountkey="$storage_account_key" --type=Opaque
```

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

Create the file `azure-blobfuse-pv.yaml`:

```yaml
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

```bash
# Create PersistentVolume and PersistentVolumeClaim for container mounts
kubectl apply -f azure-blobfuse-pv.yaml
```

```bash
# Verify PVC is bound
kubectl get pv,pvc
```

Example output:
```
NAME                          CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS        CLAIM                 STORAGECLASS   REASON   AGE
persistentvolume/pv-blob      10Gi       RWX            Retain           Terminating   default/pvc-blob                              113m
persistentvolume/pv-gptblob   10Gi       RWX            Retain           Bound         default/pvc-gptblob                           18s

NAME                                STATUS        VOLUME       CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/pvc-blob      Terminating   pv-blob      10Gi       RWX                           113m
persistentvolumeclaim/pvc-gptblob   Bound         pv-gptblob   10Gi       RWX                           17s
```

---

In the end of this step you will have AKS cluster and Storage account in resource group. AKS cluster will have cpu and gpu nodepools in addition to system nodepool.

