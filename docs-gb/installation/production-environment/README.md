---
description: Install Core 2 in a production Kubernetes environment.
---

## Prerequisites

* Set up and connect to a Kubernetes cluster running version 1.27 or later. For instructions on connecting to your Kubernetes cluster, refer to the documentation provided by your cloud provider.
* Install [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl), the Kubernetes command-line tool.
* Install [Helm](https://helm.sh/docs/intro/install/), the package manager for Kubernetes.
  

To use Seldon Core 2 in a production environment:
1. [Create namespaces](README.md#creating-namespaces)
2. [Install Seldon Core 2](README.md#installing-seldon-core-2)

Seldon publishes the [Helm charts](https://github.com/SeldonIO/helm-charts) that are required to install Seldon Core 2. For more information about the Helm charts and the related dependencies, see [Helm charts](/docs-gb/installation/README.md#helm-charts) and [Dependencies](/docs-gb/installation/README.md#seldon-core-2-dependencies).

## Creating Namespaces

*   Create a namespace to contain the main components of Seldon Core 2. For example, create the namespace `seldon-mesh`:

    ```bash
    kubectl create ns seldon-mesh || echo "Namespace seldon-mesh already exists"
    ```
*   Create a namespace to contain the components related to monitoring. For example, create the namespace `seldon-monitoring`:

    ```bash
    kubectl create ns seldon-monitoring || echo "Namespace seldon-monitoring already exists"
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
    --namespace default \
    --install 
    ```
3.  Install Seldon Core 2 operator in the `seldon-mesh` namespace.

    ```bash
     helm upgrade seldon-core-v2-setup seldon-charts/seldon-core-v2-setup \
     --namespace seldon-mesh --set controller.clusterwide=true \
     --install
    ```
    This configuration installs the Seldon Core 2 operator across an entire Kubernetes cluster. To perform cluster-wide operations, create `ClusterRoles` and ensure your user has the necessary permissions during deployment. With cluster-wide operations, you can create `SeldonRuntimes` in any namespace.

    With cluster-wide installation, you can specify the namespaces to watch by setting `controller.watchNamespaces` to a comma-separated list of namespaces (e.g., `{ns1, ns2}`). This allows the Seldon Core 2 operator to monitor and manage resources in those namespaces.

    You can also install multiple operators in different namespaces, and configure them to watch a disjoint set of namespaces. For example, you can install two operators in `op-ns1` and `op-ns2`, and configure them to watch `ns1, ns2` and `ns3, ns4`, respectively, using the following commands:

    ```bash
    for ns in op-ns1 op-ns2 ns1 ns2 ns3 ns4; do kubectl create ns "$ns"; done
    ```

    We now install the first operator in `op-ns1` and configure it to watch `ns1` and `ns2`:

    ```bash
    helm upgrade seldon-core-v2-setup seldon-charts/seldon-core-v2-setup \
    --namespace op-ns1 \
    --set controller.clusterwide=true \
    --set "controller.watchNamespaces={ns1,ns2}" \
    --install
    ```

    Next, we install the second operator in `op-ns2` and configure it to watch `ns3` and `ns4`:

    ```bash
    helm upgrade seldon-core-v2-setup seldon-charts/seldon-core-v2-setup \
    --namespace op-ns2 \
    --set controller.clusterwide=true \
    --set "controller.watchNamespaces={ns3,ns4}" \
    --set controller.skipClusterRoleCreation=true \
    --install
    ```

    Note that the second operator is installed with `skipClusterRoleCreation=true` to avoid re-creating the `ClusterRole` and `ClusterRoleBinding` that were created by the first operator. 

    Finally, you can configure the installation to deploy the Seldon Core 2 operator in a specific namespace so that it control resources in the provided namespace. To do this, set `controller.clusterwide` to `false`.

4.  Install Seldon Core 2 runtimes in the namespace `seldon-mesh`.

    ```bash
    helm upgrade seldon-core-v2-runtime seldon-charts/seldon-core-v2-runtime \
    --namespace seldon-mesh \
    --install
    ```
5. Install Seldon Core 2 servers in the namespace `seldon-mesh`. Two example servers named `mlserver-0`, and `triton-0` are installed so that you can load the models to these servers after installation.

    ```bash
     helm upgrade seldon-core-v2-servers seldon-charts/seldon-core-v2-servers \
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
**Note**: Pods with names starting with `seldon-dataflow-engine`, `seldon-pipelinegateway`, and `seldon-modelgateway` may generate log errors until they successfully connect to Kafka. This occurs because Kafka is not yet fully integrated with Seldon Core 2.
{% endhint %}

## Next steps
You can integrate Seldon Core 2 with Kafka that is [self-hosted](/docs-gb/installation/learning-environment/self-hosted-kafka.md) or a [managed Kafka](/docs-gb/installation/production-environment/managed-kafka.md).

## Additional Resources

* [Seldon Enterprise Documentation](https://docs.seldon.ai/seldon-enterprise-platform)
* [GKE Documentation](https://cloud.google.com/kubernetes-engine/docs)
* [AWS Documentation](https://docs.aws.amazon.com)
* [Azure Documentation](https://learn.microsoft.com/en-us/azure)
