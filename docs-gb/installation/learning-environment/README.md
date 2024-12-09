---
description: Installing Seldon Core 2 in a learning environment.
---

# Learning Environment

You can install Seldon Core 2 on your local computer that is running a Kubernetes cluster using [kind](https://kubernetes.io/docs/tasks/tools/#kind).

You could also install Selcon Core 2 locally if you have installed [Docker Compose](https://docs.docker.com/compose/install/) and `make` utility on your Linux systems. Clone the Seldon core repository:
`git clone https://github.com/SeldonIO/seldon-core --branch=v2`,change to the `seldon-core` directory, and run `make deploy-local`.


{% hint style="info" %}
**Note**: These instructions guide you through installing the Seldon Core 2 on a local Kubernetes cluster, focusing on ease of learning. Ensure your kind cluster is running on hardware with at least 32GB of RAM. For installing Seldon Core 2 in a production environment, see[ cluster requirements](../production-environment/#cluster-requirements).
{% endhint %}


## Prerequisites

* Install a Kubernetes cluster that is running version 1.23 or later.
* Install [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl), the Kubernetes command-line tool.
* Install [Helm](https://helm.sh/docs/intro/install/), the package manager for Kubernetes or [Ansible](https://docs.ansible.com/ansible/latest/installation_guide/intro_installation.html#installing-and-upgrading-ansible), the automation tool used for provisioning, configuration management, and application deployment.  

## Installing Seldon Core 2

{% tabs %}

{% tab title="Helm" %}
1. Create a namespace to contain the main components of Seldon Core 2. For example, create the namespace `seldon-mesh`:

    ```bash
    kubectl create ns seldon-mesh || echo "Namespace seldon-mesh already exists"
    ```
2.  Add and update the Helm charts `seldon-charts` to the repository.

    ```bash
    helm repo add seldon-charts https://seldonio.github.io/helm-charts/
    helm repo update seldon-charts
    ```
3.  Install Custom resource definitions for Seldon Core 2.

    ```bash
    helm upgrade seldon-core-v2-crds seldon-charts/seldon-core-v2-crds \
    --version 2.8.3 \
    --namespace default \
    --install 
    ```
4.  Install Seldon Core 2 operator in the namespace `seldon-mesh`.

    ```bash
     helm upgrade seldon-core-v2-setup seldon-charts/seldon-core-v2-setup \
     --version 2.8.3 \
     --namespace seldon-mesh --set controller.clusterwide=true \
     --install
    ```
    This configuration installs Seldon Core 2 operator across an entire Kubernetes cluster. To perform cluster-wide operations, create `ClusterRoles` and ensure your user has the necessary permissions during deployment. With cluster-wide operations, you can create `SeldonRuntimes` in any namespace.

    You can configure the installation to deploy the Seldon Core 2 operator in a specific namespace so that it control resources in the provided namespace. To do this, set `controller.clusterwide` to `false`.

5.  Install Seldon Core 2 runtimes in the namespace `seldon-mesh`.

    ```bash
    helm upgrade seldon-core-v2-runtime seldon-charts/seldon-core-v2-runtime \
    --version 2.8.3 \
    --namespace seldon-mesh \
    --install
    ```
6. Install Seldon Core 2 servers in the namespace `seldon-mesh`.

    ```bash
     helm upgrade seldon-core-v2-servers seldon-charts/seldon-core-v2-servers \
     --version 2.8.3 \
     --namespace seldon-mesh \
     --install
    ```
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

This creates a kind cluster and install ecosystem dependencies such kafka,
prometheus, opentelemetry, and jager as well as all the seldon-specific components.
The seldon components are installed using helm-charts from the current git
checkout (`../k8s/helm-charts/`).

Internally this runs, in order, the following playbooks (described in more detail
in the sections below):
- kind-cluster.yaml
- setup-ecosystem.yaml
- setup-seldon.yaml

You may pass any of the additonal variables which are configurable for those playbooks to `seldon-all`. 

For example:

```bash
ansible-playbook playbooks/seldon-all.yaml -e seldon_mesh_namespace=my-seldon-mesh -e install_prometheus=no -e @playbooks/vars/set-custom-images.yaml
```

Running the playbooks individually, gives you more control over what and when it runs. For example, if you want to install into an existing k8s cluster.

### Multiple commands

1. Create a kind cluster.

    ```bash
    ansible-playbook playbooks/kind-cluster.yaml
    ```
1. Setup ecosystem.

    ```bash
    ansible-playbook playbooks/setup-ecosystem.yaml
    ```
    Seldon runs by default in `seldon-mesh` namespace and a Jaeger pod and  and OpenTelemetry collector are installed in the namespace. 
    To install in a different namespace, `<mynamespace>`:

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

   
## Next

To explore the inference Pipeline feature of Seldon Core 2, you need to complete the installation of other components in the following order:

1. [Integrating with Kafka](self-hosted-kafka.md)
2. [Installing a Service mesh](../production-environment/ingress-controller/istio.md)


