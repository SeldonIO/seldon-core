import kfp.dsl as dsl
import yaml
from kubernetes import client as k8s


@dsl.pipeline(
    name="SeldonBatch", description="A batch processing pipeline for seldon models"
)
def nlp_pipeline(
    namespace="kubeflow",
    seldon_server="SKLEARN_SERVER",
    model_path="gs://seldon-models/v1.13.0-dev/sklearn/iris",
    gateway_endpoint="istio-ingressgateway.istio-system.svc.cluster.local",
    retries=3,
    replicas=10,
    workers=100,
    input_path="data/input-data.txt",
    output_path="data/output-data.txt",
):
    """
    Pipeline 
    """

    vop = dsl.VolumeOp(
        name="seldon-batch-pvc",
        resource_name="seldon-batch-pvc",
        modes=dsl.VOLUME_MODE_RWO,
        size="2Mi",
    )

    seldon_deployment_yaml = f"""
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: "{{{{workflow.name}}}}"
  namespace: "{namespace}"
spec:
  name: "{{{{workflow.name}}}}"
  predictors:
  - graph:
      children: []
      implementation: "{seldon_server}"
      modelUri: "{model_path}"
      name: classifier
    name: default
    """

    deploy_step = dsl.ResourceOp(
        name="deploy_seldon",
        action="create",
        k8s_resource=yaml.safe_load(seldon_deployment_yaml),
    )

    scale_and_wait = dsl.ContainerOp(
        name="scale_and_wait_seldon",
        image="bitnami/kubectl:1.17",
        command="bash",
        arguments=[
            "-c",
            f"sleep 10 && kubectl scale --namespace {namespace} --replicas={replicas} sdep/{{{{workflow.name}}}} && sleep 2 && kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id={{{{workflow.name}}}} -o jsonpath='{{.items[0].metadata.name'}})",
        ],
    )

    download_from_object_store = dsl.ContainerOp(
        name="download-from-object-store",
        image="minio/mc:RELEASE.2020-04-17T08-55-48Z",
        command="sh",
        arguments=[
            "-c",
            f"mc config host add minio-local http://minio.default.svc.cluster.local:9000 minioadmin minioadmin && mc cp minio-local/{input_path} /assets/input-data.txt",
        ],
        pvolumes={"/assets": vop.volume},
    )

    batch_process_step = dsl.ContainerOp(
        name="data_downloader",
        image="seldonio/seldon-core-s2i-python37:1.1.1-rc",
        command="seldon-batch-processor",
        arguments=[
            "--deployment-name",
            "{{workflow.name}}",
            "--namespace",
            namespace,
            "--host",
            gateway_endpoint,
            "--retries",
            retries,
            "--input-data-path",
            "/assets/input-data.txt",
            "--output-data-path",
            "/assets/output-data.txt",
            "--benchmark",
        ],
        pvolumes={"/assets": vop.volume},
    )

    upload_to_object_store = dsl.ContainerOp(
        name="upload-to-object-store",
        image="minio/mc:RELEASE.2020-04-17T08-55-48Z",
        command="sh",
        arguments=[
            "-c",
            f"mc config host add minio-local http://minio.default.svc.cluster.local:9000 minioadmin minioadmin && mc cp /assets/output-data.txt minio-local/{output_path}",
        ],
        pvolumes={"/assets": vop.volume},
    )

    delete_step = dsl.ResourceOp(
        name="delete_seldon",
        action="delete",
        k8s_resource=yaml.safe_load(seldon_deployment_yaml),
    )

    scale_and_wait.after(deploy_step)
    download_from_object_store.after(scale_and_wait)
    batch_process_step.after(download_from_object_store)
    upload_to_object_store.after(batch_process_step)
    delete_step.after(upload_to_object_store)


if __name__ == "__main__":
    import kfp.compiler as compiler

    compiler.Compiler().compile(nlp_pipeline, __file__ + ".tar.gz")
