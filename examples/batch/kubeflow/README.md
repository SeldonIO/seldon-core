## Batch processing with Kubeflow Pipelines
In this notebook we will dive into how you can run batch processing with Kubeflow Pipelines and Seldon Core.




```bash
%%bash
export PIPELINE_VERSION=0.5.1
kubectl apply -k github.com/kubeflow/pipelines/manifests/kustomize/cluster-scoped-resources?ref=$PIPELINE_VERSION
kubectl wait --for condition=established --timeout=60s crd/applications.app.k8s.io
kubectl apply -k github.com/kubeflow/pipelines/manifests/kustomize/env/dev?ref=$PIPELINE_VERSION
```


```
pip install kfp
```


```
mkdir -p assets/
```


```
%%writefile assets/seldon-batch-pipeline.py

import kfp.dsl as dsl
import yaml
from kubernetes import client as k8s

@dsl.pipeline(
  name='SeldonBatch',
  description='A batch processing pipeline for seldon models'
)
def nlp_pipeline(
        deployment_name="seldon-batch",
        namespace="kubeflow",
        seldon_server="SKLEARN_SERVER",
        model_path="gs://seldon-models/sklearn/iris",
        gateway_endpoint="istio-ingressgateway.istio-system.svc.cluster.local",
        retries=3,
        replicas=10,
        workers=100,
        input_path="data/input-data.txt",
        output_path="data/output-data.txt"):
    """
    Pipeline 
    """
    
    vop = dsl.VolumeOp(
      name='seldon-batch-pvc',
      resource_name="seldon-batch-pvc",
      modes=dsl.VOLUME_MODE_RWO,
      size="2Mi"
    )
    
    seldon_deployment_yaml = f"""
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: "{deployment_name}"
  namespace: "{namespace}"
spec:
  name: "{deployment_name}"
  predictors:
    - graph:
        children: []
        implementation: "{seldon_server}"
        modelUri: "{model_path}"
        name: classifier
      name: default
      replicas: "{replicas}"
    """
    
    deploy_step = dsl.ResourceOp(
        name="deploy_seldon",
        action="create",
        k8s_resource=yaml.safe_load(seldon_deployment_yaml))
    
    wait_for_ready = dsl.ContainerOp(
        name="wait_seldon",
        image="bitnami/kubectl:1.17",
        command="bash",
        arguments=[
            "-c",
            "sleep 10 && kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id={deployment_name} -o jsonpath='{.items[0].metadata.name')"   
        ])
    
    download_from_object_store = dsl.ContainerOp(
        name="download-from-object-store",
        image="minio/mc:RELEASE.2020-04-17T08-55-48Z",
        command="sh",
        arguments=[
            "-c",
            f"mc config host add minio-local minioadmin minioadmin && mc cp minio-local/{input_path} /assets/input-data.txt"   
        ],
        pvolumes={ "/assets", vop.volume })
    

    batch_process_step = dsl.ContainerOp(
        name='data_downloader',
        image='seldonio/seldon-core-s2i-python37:1.1.1-rc',
        command="seldon-batch-processor",
        arguments=[
            "--deployment-name", deployment_name,
            "--namespace", namespace,
            "--host", gateway_endpoint,
            "--retries", retries,
            "--input-data-path", input_path,
            "--output-data-path", output_path
        ],
        pvolumes={ "/assets", vop.volume }
    )
    
    upload_to_object_store = dsl.ContainerOp(
        name="upload-to-object-store",
        image="minio/mc:RELEASE.2020-04-17T08-55-48Z",
        command="sh",
        arguments=[
            "-c",
            f"mc config host add minio-local minioadmin minioadmin && mc cp /assets/output-data.txt minio-local/{output_path}"   
        ],
        pvolumes={ "/assets", vop.volume })
    
    
    wait_for_ready.after(deploy_step)
    download_from_object_store.after(wait_for_ready)
    batch_process_step.after(download_from_object_store)
    upload_to_object_store.after(batch_process_step)

if __name__ == '__main__':
  import kfp.compiler as compiler
  compiler.Compiler().compile(nlp_pipeline, __file__ + '.tar.gz')

```

    Overwriting assets/seldon-batch-pipeline.py



```
!python assets/seldon-batch-pipeline.py
```

    /home/alejandro/miniconda3/lib/python3.7/site-packages/kfp/components/_data_passing.py:168: UserWarning: Missing type name was inferred as "Integer" based on the value "3".
      warnings.warn('Missing type name was inferred as "{}" based on the value "{}".'.format(type_name, str(value)))
    /home/alejandro/miniconda3/lib/python3.7/site-packages/kfp/components/_data_passing.py:168: UserWarning: Missing type name was inferred as "Integer" based on the value "10".
      warnings.warn('Missing type name was inferred as "{}" based on the value "{}".'.format(type_name, str(value)))
    /home/alejandro/miniconda3/lib/python3.7/site-packages/kfp/components/_data_passing.py:168: UserWarning: Missing type name was inferred as "Integer" based on the value "100".
      warnings.warn('Missing type name was inferred as "{}" based on the value "{}".'.format(type_name, str(value)))
    Traceback (most recent call last):
      File "assets/seldon-batch-pipeline.py", line 108, in <module>
        compiler.Compiler().compile(nlp_pipeline, __file__ + '.tar.gz')
      File "/home/alejandro/miniconda3/lib/python3.7/site-packages/kfp/compiler/compiler.py", line 926, in compile
        allow_telemetry=allow_telemetry)
      File "/home/alejandro/miniconda3/lib/python3.7/site-packages/kfp/compiler/compiler.py", line 983, in _create_and_write_workflow
        allow_telemetry)
      File "/home/alejandro/miniconda3/lib/python3.7/site-packages/kfp/compiler/compiler.py", line 806, in _create_workflow
        pipeline_func(*args_list)
      File "assets/seldon-batch-pipeline.py", line 72, in nlp_pipeline
        pvolumes={ "/assets", vop.volume })
    TypeError: unhashable type: 'PipelineVolume'



```
!ls assets/
```

    seldon-batch-pipeline.py  seldon-batch-pipeline.py.tar.gz  seldon-batch.jpg



```

```
