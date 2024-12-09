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
2.  Install Custom resource definitions for Seldon Core 2.

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
## Additional Resources

* Seldon Core Documentation
* Seldon Enterprise Documentation
* GKE Documentation
* AWS Documentation
* Azure Documentation