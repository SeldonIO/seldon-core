Seldon Core 2 provides a state of the art solution for machine learning inference. 

## Prerequisites

* Set up and connect to a Kubernetes cluster running version 1.23 or later. For instructions on connecting to your Kubernetes cluster, refer to the documentation provided by your cloud provider.
* Install [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl), the Kubernetes command-line tool.
* Install [Helm](https://helm.sh/docs/intro/install/), the package manager for Kubernetes.

To use Seldon Core 2 in a production environment:
1. [Create namespaces](seldon-core-2.md#creating-namespaces)
2. [Install Seldon Core 2](seldon-core-2.md#installing-seldon-core-2)

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
    --version 2.8.3 \
    --namespace default \
    --install 
    ```
3.  Install Seldon Core 2 operator in the namespace `seldon-mesh`.

    ```bash
     helm upgrade seldon-core-v2-setup seldon-charts/seldon-core-v2-setup \
     --version 2.8.3 \
     --namespace seldon-mesh --set controller.clusterwide=true \
     --install
    ```
    This configuration installs Seldon Core 2 operator across an entire Kubernetes cluster. To perform cluster-wide operations, create `ClusterRoles` and ensure your user has the necessary permissions during deployment. With cluster-wide operations, you can create `SeldonRuntimes` in any namespace.

    You can configure the installation to deploy the Seldon Core 2 operator in a specific namespace so that it control resources in the provided namespace. To do this, set `controller.clusterwide` to `false`.
4.  Install Seldon Core 2 runtimes in the namespace `seldon-mesh`.

    ```bash
    helm upgrade seldon-core-v2-runtime seldon-charts/seldon-core-v2-runtime \
    --version 2.8.3 \
    --namespace seldon-mesh \
    --install
    ```
5. Install Seldon Core 2 servers in the namespace `seldon-mesh`.

    ```bash
     helm upgrade seldon-core-v2-servers seldon-charts/seldon-core-v2-servers \
     --version 2.8.3 \
     --namespace seldon-mesh \
     --install
    ```
6. Check Seldon Core 2 operator, runtimes, servers, and CRDS are installed in the namespace `seldon-mesh`:
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

* Seldon Core Documentation
* Seldon Enterprise Documentation
* GKE Documentation
* AWS Documentation
* Azure Documentation