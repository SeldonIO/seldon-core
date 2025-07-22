# Pipeline example with OpenVINO inference execution engine 

This notebook illustrates how you can serve ensemble of models using [OpenVINO prediction model](https://github.com/SeldonIO/seldon-core/tree/master/examples/models/openvino_imagenet_ensemble/resources/model).
The demo includes optimized ResNet50 and DenseNet169 models by OpenVINO model optimizer. 
They have [reduced precision](https://www.edge-ai-vision.com/2019/02/introducing-int8-quantization-for-fast-cpu-inference-using-openvino/) of graph operations from FP32 to INT8. It significantly improves the execution peformance with minimal impact on the accuracy. The gain is particularly visible with the latest Casade Lake CPU with [VNNI](https://www.intel.com/content/www/us/en/developer/articles/guide/deep-learning-with-avx512-and-dl-boost.html#inpage-nav-1) extension.

![pipeline](pipeline1.png)

## Setup Seldon Core

Use the setup notebook to [Setup Cluster](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Setup-Cluster) with [Ambassador Ingress](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Ambassador) and [Install Seldon Core](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Install-Seldon-Core). Instructions [also online](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html).


```python
!kubectl create namespace seldon
```


```python
!kubectl config set-context $(kubectl config current-context) --namespace=seldon
```

## Deploy Seldon pipeline with Intel OpenVINO models ensemble

 * Ingest compressed JPEG binary and transform to TensorFlow Proto payload
 * Ensemble two OpenVINO optimized models for ImageNet classification: ResNet50, DenseNet169
 * Return result in human readable text




```python
!pygmentize seldon_ov_predict_ensemble.yaml
```


```python
!kubectl apply -f seldon_ov_predict_ensemble.yaml
```


```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=openvino-model -o jsonpath='{.items[0].metadata.name}')
```

### Using the exemplary grpc client

Install client dependencies: seldon-core and grpcio packages


```python
!pip install -q seldon-core grpcio
```


```python
!python seldon_grpc_client.py --debug --ambassador localhost:8003
```

For more extensive test see the client help.

You can change the default test-input file including labeled list of images to calculate accuracy based on complete imagenet dataset. Follow the format from file `input_images.txt` - path to the image and imagenet class in every line.


```python
!python seldon_grpc_client.py --help
```

## Examining the logs

You can use Seldon containers logs to get additional details about the execution:



```python
!kubectl logs $(kubectl get pods -l seldon-app=openvino-model-openvino -o jsonpath='{.items[0].metadata.name}') prediction1 --tail=10
```


```python
!kubectl logs $(kubectl get pods -l seldon-app=openvino-model-openvino -o jsonpath='{.items[0].metadata.name}') prediction2 --tail=10
```


```python
!kubectl logs $(kubectl get pods -l seldon-app=openvino-model-openvino -o jsonpath='{.items[0].metadata.name}') imagenet-itransformer --tail=10
```

## Performance consideration

In production environment with a shared workloads, you might consider contraining the CPU resources for individual pipeline components. You might restrict the assigned capacity using [Kubernetes capabilities](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/). This configuration can be added to seldon pipeline definition.

Another option for tuning the resource allocation is adding environment variable `OMP_NUM_THREADS`. It can indicate how many threads will be used by OpenVINO execution engine and how many CPU cores can be consumed. The recommended value is equal to the number of allocated CPU physical cores.

In the tests using GKE service in Google Cloud on nodes with 32 SkyLake vCPU assigned, the following configuration was set on prediction components. It achieved the optimal latency and throughput:
```
"resources": {
  "requests": {
     "cpu": "1"
  },
  "limits": {
     "cpu": "32"
  }
}

"env": [
  {
    "name": "KMP_AFFINITY",
    "value": "granularity=fine,verbose,compact,1,0"
  },
  {
    "name": "KMP_BLOCKTIME",
    "value": "1"
  },
  {
    "name": "OMP_NUM_THREADS",
    "value": "8"
  }
]
```


```python
!kubectl delete -f seldon_ov_predict_ensemble.json
```
