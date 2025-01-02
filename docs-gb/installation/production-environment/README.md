---
description: Install Core 2 in a production Kubernetes environment.
---

## Prerequisites

* Set up and connect to a Kubernetes cluster running version 1.27 or later. For instructions on connecting to your Kubernetes cluster, refer to the documentation provided by your cloud provider.
* Install [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl), the Kubernetes command-line tool.
* Install [Helm](https://helm.sh/docs/intro/install/), the package manager for Kubernetes.
  

To use Seldon Core 2 in a production environment:
1. [Create namespaces](seldon-core-2.md#creating-namespaces)
2. [Install Seldon Core 2](seldon-core-2.md#installing-seldon-core-2)

Seldon publishes the [Helm charts](https://github.com/SeldonIO/helm-charts) that are required to install Seldon Core 2. For more information see about the Helm charts and the related dependencies, [Helm charts](/docs-gb/installation/README.md#helm-charts) and [Dependencies](/docs-gb/installation/README.md#seldon-core-2-dependencies).

## Creating Namespaces

*   Create a namespace to contain the main components of Seldon Core 2. For example, create the namespace `seldon-mesh`:

    ```bash
    kubectl create ns seldon-mesh || echo "Namespace seldon-mesh already exists"
    ```
*   Create a namespace to contain the components related to request logging. For example, create the namespace `seldon-logs`:

    ```bash
    kubectl create ns seldon-logs || echo "Namespace seldon-logs already exists"
    ```

## Installing Seldon Core 2

1.  Add and update the Helm charts `seldon-charts` to the repository.

    ```bash
    helm repo add seldon-charts https://seldonio.github.io/helm-charts/
    helm repo update seldon-charts
    ```
2.  Install custom resource definitions for Seldon Core 2.

    ```bash
    helm upgrade seldon-core-v2-crds seldon-charts/seldon-core-v2-crds \
    --version 2.8.5 \
    --namespace default \
    --install 
    ```
3.  Create a YAML file to specify the initial configuration for Seldon Core 2 operator. For example, create the `components-values.yaml` file. Use your preferred text editor to create and save the file with the following content:

    ```yaml
    controller:
      clusterwide: true

    dataflow:
      resources:
        cpu: 500m

    envoy:
      service:
        type: ClusterIP

    kafka:
      bootstrap: seldon-kafka-bootstrap.seldon-mesh:9092
      topics:
        numPartitions: 4

    opentelemetry:
      enable: false

    scheduler:
      service:
        type: ClusterIP

    serverConfig:
      mlserver:
        resources:
          cpu: 1
          memory: 2Gi

      triton:
        resources:
          cpu: 1
          memory: 2Gi

    serviceGRPCPrefix: "http2-"
    ```
    This configuration installs Seldon Core 2 operator across an entire Kubernetes cluster.
    You can configure the installation to deploy the Seldon Core 2 operator in a specific namespace. To do this, set `clusterwide` to `false` in the `components-values.yaml` file.

4.  Change to the directory that contains the `components-values.yaml` file and then install Seldon Core 2 operator in the namespace `seldon-system`.

    ```bash
     helm upgrade seldon-core-v2-components seldon-charts/seldon-core-v2-setup \
     --version 2.8.5 \
     -f components-values.yaml \
     --namespace seldon-system \
     --install
    ```
    To install Seldon Core 2 operator in a specific namespace `seldon`.

    ```bash
     helm upgrade seldon-core-v2-components seldon-charts/seldon-core-v2-setup \
     --version 2.8.5 \
     -f components-values.yaml \
     --namespace seldon\
     --install
    ```
5.  Install Seldon Core 2 runtimes in the namespace `seldon-mesh`.

    ```bash
    helm upgrade seldon-core-v2-runtime seldon-charts/seldon-core-v2-runtime \
    --version 2.8.5 \
    --namespace seldon-mesh \
    --install
    ```
6. Install Seldon Core 2 servers in the namespace `seldon-mesh`.

    ```bash
     helm upgrade seldon-core-v2-servers seldon-charts/seldon-core-v2-servers \
     --version 2.8.5 \
     --namespace seldon-mesh \
     --install
    ```
7. Check Seldon Core 2 operator, runtimes, servers, and CRDS are installed in the namespace `seldon-mesh`:
    ```bash
     kubectl get pods -n seldon-mesh
    ```
    The output should be similar to this:
    ```
    NAME                                            READY   STATUS             RESTARTS      AGE
    hodometer-749d7c6875-4d4vw                      1/1     Running            0             4m33s
    mlserver-0                                      3/3     Running            0             4m10s
    seldon-dataflow-engine-7b98c76d67-v2ztq         0/1     CrashLoopBackOff   5 (49s ago)   4m33s
    seldon-envoy-bb99f6c6b-4mpjd                    1/1     Running            0             4m33s
    seldon-modelgateway-5c76c7695b-bhfj5            1/1     Running            0             4m34s
    seldon-pipelinegateway-584c7d95c-bs8c9          1/1     Running            0             4m34s
    seldon-scheduler-0                              1/1     Running            0             4m34s
    seldon-v2-controller-manager-5dd676c7b7-xq5sm   1/1     Running            0             4m52s
    triton-0                                        2/3     Running            0             4m10s
    ```
{% hint style="info" %}
**Note**: The status of the Pod that begins with the name `seldon-dataflow-engine` is not running because Kafka is not still integrated with Seldon Core 2.
{% endhint %}

## Next steps
You can integrate Seldon Core 2 with Kafka that is [self-hosted](/docs-gb/installation/learning-environment/self-hosted-kafka.md) or a [managed Kafka](/docs-gb/installation/production-environment/managed-kafka.md).

## Additional Resources

* [Seldon Enterprise Documentation](https://docs.seldon.ai/seldon-enterprise-platform)
* [GKE Documentation](https://cloud.google.com/kubernetes-engine/docs)
* [AWS Documentation](https://docs.aws.amazon.com)
* [Azure Documentation](https://learn.microsoft.com/en-us/azure)
