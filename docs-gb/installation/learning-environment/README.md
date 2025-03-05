---
description: Install Seldon Core 2 in a local learning environment.
---

# Learning Environment

You can install Seldon Core 2 on your local computer that is running a Kubernetes cluster using [kind](https://kubernetes.io/docs/tasks/tools/#kind).

Seldon publishes the [Helm charts](https://github.com/SeldonIO/helm-charts) that are required to install Seldon Core 2. For more information about the Helm charts and the related dependencies,see [Helm charts](/docs-gb/installation/README.md#helm-charts) and [Dependencies](/docs-gb/installation/README.md#seldon-core-2-dependencies).

{% hint style="info" %}
**Note**: These instructions guide you through installing Seldon Core 2 on a local Kubernetes cluster, focusing on ease of learning. Ensure your [kind](https://kubernetes.io/docs/tasks/tools/#kind) cluster is running on hardware with at least 32GB of RAM and a load balancer such as MetalLB is configured.
{% endhint %}


## Prerequisites
* Install a Kubernetes cluster that is running version 1.27 or later. 
* Install [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl), the Kubernetes command-line tool.
* Install [Helm](https://helm.sh/docs/intro/install/), the package manager for Kubernetes or [Ansible](https://docs.ansible.com/ansible/latest/installation_guide/intro_installation.html#installing-and-upgrading-ansible), the automation tool used for provisioning, configuration management, and application deployment.


{% hint style="info" %}
**Note**: Ansible automates provisioning, configuration management, and handles all dependencies required for Seldon Core 2.
With Helm, you need to configure and manage the dependencies yourself.
{% endhint %}

## Installing Seldon Core 2

{% tabs %}

{% tab title="Helm" %}
1. Create a namespace to contain the main components of Seldon Core 2. For example, create the `seldon-mesh` namespace.

    ```bash
    kubectl create ns seldon-mesh || echo "Namespace seldon-mesh already exists"
    ```
2.  Add and update the Helm charts, `seldon-charts`, to the repository.

    ```bash
    helm repo add seldon-charts https://seldonio.github.io/helm-charts/
    helm repo update seldon-charts
    ```
3.  Install Custom resource definitions for Seldon Core 2.

    ```bash
    helm upgrade seldon-core-v2-crds seldon-charts/seldon-core-v2-crds \
    --namespace default \
    --install 
    ```
4.  Install Seldon Core 2 operator in the `seldon-mesh` namespace.

    ```bash
     helm upgrade seldon-core-v2-setup seldon-charts/seldon-core-v2-setup \
     --namespace seldon-mesh --set controller.clusterwide=true \
     --install
    ```
    This configuration installs the Seldon Core 2 operator across an entire Kubernetes cluster. To perform cluster-wide operations, create `ClusterRoles` and ensure your user has the necessary permissions during deployment. With cluster-wide operations, you can create `SeldonRuntimes` in any namespace.

    You can configure the installation to deploy the Seldon Core 2 operator in a specific namespace so that it control resources in the provided namespace. To do this, set `controller.clusterwide` to `false`.

5.  Install Seldon Core 2 runtimes in the `seldon-mesh` namespace.

    ```bash
    helm upgrade seldon-core-v2-runtime seldon-charts/seldon-core-v2-runtime \
    --namespace seldon-mesh \
    --install
    ```
6. Install Seldon Core 2 servers in the `seldon-mesh` namespace. Two example servers named `mlserver-0`, and `triton-0` are installed so that you can load the models to these servers after installation.

    ```bash
     helm upgrade seldon-core-v2-servers seldon-charts/seldon-core-v2-servers \
     --namespace seldon-mesh \
     --install
    ```
7. Check Seldon Core 2 operator, runtimes, servers, and CRDS are installed in the `seldon-mesh` namespace. It might take a couple of minutes for all the Pods to be ready. To check the status of the Pods in real time use this command: `kubectl get pods -w -n seldon-mesh`. 

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
    
{% endtab %}

{% tab title="Ansible" %}

{% hint style="info" %}
**Note**: For more information about configurations, see the supported versions of [Python libraries](https://github.com/SeldonIO/seldon-core/tree/v2/ansible#installing-ansible), and [customization options](https://github.com/SeldonIO/seldon-core/tree/v2/ansible#customizing-ansible-installation)
{% endhint %}

You can install Seldon Core 2 and its components using Ansible in one of the following methods:
* [Single command](#single-command)
* [Multiple commands](#multiple-commands)

### Single command

To install Seldon Core 2 into a new local kind Kubernetes cluster, you can use the `seldon-all` playbook with a single command:

```bash
ansible-playbook playbooks/seldon-all.yaml
```

This creates a kind cluster and installs ecosystem dependencies such kafka,
Prometheus, OpenTelemetry, and Jaeger as well as all the seldon-specific components.
The seldon components are installed using helm-charts from the current git
checkout (`../k8s/helm-charts/`).

Internally this runs, in order, the following playbooks:
- kind-cluster.yaml
- setup-ecosystem.yaml
- setup-seldon.yaml

You may pass any of the additonal variables which are configurable for those playbooks to `seldon-all`. 

For example:

```bash
ansible-playbook playbooks/seldon-all.yaml -e seldon_mesh_namespace=my-seldon-mesh -e install_prometheus=no -e @playbooks/vars/set-custom-images.yaml
```

Running the playbooks individually gives you more control over what and when it runs. For example, if you want to install into an existing k8s cluster.

### Multiple commands

1. Create a kind cluster.

    ```bash
    ansible-playbook playbooks/kind-cluster.yaml
    ```
1. Setup ecosystem.

    ```bash
    ansible-playbook playbooks/setup-ecosystem.yaml
    ```
    Seldon runs by default in the `seldon-mesh` namespace and a Jaeger pod and OpenTelemetry collector are installed in the namespace. 
    To install in a different `<mynamespace>` namespace:

    ```bash
    ansible-playbook playbooks/setup-ecosystem.yaml -e seldon_mesh_namespace=<mynamespace>
    ```
1. Install Seldon Core 2 in the `ansible/` folder.

    ```bash
    ansible-playbook playbooks/setup-seldon.yaml
    ```
    To install in a different namespace, `<mynamespace>`.

    ```bash
    ansible-playbook playbooks/setup-seldon.yaml -e seldon_mesh_namespace=<mynamespace>
    ```
{% endtab %}

{% endtabs %}

   
## Next Steps

If you installed Seldon Core 2 using Helm, you need to complete the installation of other components in the following order:

1. [Integrating with Kafka](self-hosted-kafka.md)
2. [Installing a Service mesh](../production-environment/ingress-controller/istio.md)


