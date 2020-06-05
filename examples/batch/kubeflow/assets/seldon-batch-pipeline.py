import kfp.dsl as dsl
import yaml
from kubernetes import client as k8s


@dsl.pipeline(
    name="SeldonBatch", description="A batch processing pipeline for seldon models"
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
    input_path="s3://data/input-data.txt",
    output_path="s3://data/output-data.txt",
):
    """
    Pipeline 
    """

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
        k8s_resource=yaml.safe_load(seldon_deployment_yaml),
    )

    batch_process_step = dsl.ContainerOp(
        name="data_downloader",
        image="seldonio/seldon-core-s2i-python37:1.1.1-SNAPSHOT",
        command="seldon-batch-processor",
        arguments=[
            "--deployment-name",
            deployment_name,
            "--namespace",
            namespace,
            "--host",
            gateway_endpoint,
            "--retries",
            retries,
            "--input-data-path",
            input_path,
            "--output-data-path",
            output_path,
        ],
    )

    batch_process_step.after(deploy_step)


if __name__ == "__main__":
    import kfp.compiler as compiler

    compiler.Compiler().compile(nlp_pipeline, __file__ + ".tar.gz")
