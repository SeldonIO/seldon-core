# Install Locally

This guide runs through how to set up and install Seldon Core in a Kubernetes cluster running on your local machine. By the end, you'll have Seldon Core up and running and be ready to start deploying machine learning models.

## Prerequisites

In order to install Seldon Core locally you'll need the following tools:

{% hint style="warning" %}
Depending on permissions for your local machine and the directory you're working in, some tools might require root access
{% endhint %}

### Docker or Podman

[Docker](https://www.docker.com/) and [Podman](https://podman.io/) are container engines. Kind needs a container engine (like docker or podman) to actually run the containers inside your clusters. You only need one of either Docker or Podman. Note that Docker is no longer free for individual use at large companies:

- Install Docker for [Linux](https://docs.docker.com/engine/install/ubuntu/), [Mac](https://docs.docker.com/desktop/mac/install/), [Windows](https://docs.docker.com/desktop/windows/install/)
- Or [Install Podman](https://podman.io/getting-started/installation)

{% hint style="info" %}
If using Podman remember to set `alias docker=podman`
{% endhint %}

### Kind

[Kind](https://kind.sigs.k8s.io/) is a tool for running Kubernetes clusters locally. We'll use it to create a cluster on your machine so that you can install Seldon Core in to it. If don't already have [kind](https://kind.sigs.k8s.io/) installed on your machine, you'll need to follow their installation guide:

- [Install Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)

### Kubectl

[`kubectl`](https://kubernetes.io/docs/reference/kubectl/overview/) is the Kubernetes command-line tool. It allows you to run commands against Kubernetes clusters, which we'll need to do as part of setting up Seldon Core.

- [Install kubectl on Linux](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux)
- [Install kubectl on macOS](https://kubernetes.io/docs/tasks/tools/install-kubectl-macos)
- [Install kubectl on Windows](https://kubernetes.io/docs/tasks/tools/install-kubectl-windows)

### Helm

[Helm](https://helm.sh/) is a package manager that makes it easy to find, share and use software built for Kubernetes. If you don't already have Helm installed locally, you can install it here:

- [Install Helm](https://helm.sh/docs/intro/install/)

## Set Up Kind

Once kind is installed on your system you can create a new Kubernetes cluster by running:

{% tabs %}

{% tab title="Istio" %}

```bash
kind create cluster --name seldon
```

{% endtab %}

{% tab title="Ambassador" %}

```bash
cat <<EOF | kind create cluster --name seldon --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
EOF
```

{% endtab %}

{% endtabs %}

After `kind` has created your cluster, you can configure `kubectl` to use the cluster by setting the context:

```bash
kubectl cluster-info --context kind-seldon
```

From now on, all commands run using `kubectl` will be directed at your `kind` cluster.

{% hint style="info" %}
Kind prefixes your cluster names with `kind-` so your cluster context is `kind-seldon` and not just `seldon`
{% endhint %}

## Install Cluster Ingress

`Ingress` is a Kubernetes object that provides routing rules for your cluster. It manages the incoming traffic and routes it to the services running inside the cluster.

Seldon Core supports using either [Istio](https://istio.io/) or [Ambassador](https://www.getambassador.io/) to manage incoming traffic. Seldon Core automatically creates the objects and rules required to route traffic to your deployed machine learning models.

{% tabs %}

{% tab title="Istio" %}

Istio is an open source service mesh. If the term *service mesh* is unfamiliar to you, it's worth reading [a little more about Istio](https://istio.io/latest/about/service-mesh/).

**Download Istio**

For Linux and macOS, the easiest way to download Istio is using the following command:

```bash
curl -L https://istio.io/downloadIstio | sh -
```

Move to the Istio package directory. For example, if the package is `istio-1.11.4`:

```bash
cd istio-1.11.4
```

Add the istioctl client to your path (Linux or macOS):

```bash
export PATH=$PWD/bin:$PATH
```

**Install Istio**

Istio provides a command line tool `istioctl` to make the installation process easy. The `demo` [configuration profile](https://istio.io/latest/docs/setup/additional-setup/config-profiles/) has a good set of defaults that will work on your local cluster.

```bash
istioctl install --set profile=demo -y
```

The namespace label `istio-injection=enabled` instructs Istio to automatically inject proxies alongside anything we deploy in that namespace. We'll set it up for our `default` namespace:

```bash
kubectl label namespace default istio-injection=enabled
```

**Create Istio Gateway**

In order for Seldon Core to use Istio's features to manage cluster traffic, we need to create an [Istio Gateway](https://istio.io/latest/docs/tasks/traffic-management/ingress/ingress-control/) by running the following command:

{% hint style="warning" %}
You will need to copy the entire command from the code block below
{% endhint %}

```yaml
kubectl apply -f - << END
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: seldon-gateway
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway # use istio default controller
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
    hosts:
    - "*"
END
```

For custom configuration and more details on installing Seldon Core with Istio please see the [Istio Ingress](../ingress/istio.md) page.

{% endtab %}

{% tab title="Ambassador" %}

[Ambassador](https://www.getambassador.io/) is a Kubernetes ingress controller and API gateway. It routes incoming traffic to the underlying Kubernetes workloads through configuration. Install Ambassador following their docs.

{% endtab %}

{% endtabs %}

## Install Seldon Core

Before we install Seldon Core, we'll create a new namespace `seldon-system` for the operator to run in:

```bash
kubectl create namespace seldon-system
```

We're now ready to install Seldon Core in our cluster. Run the following command for your choice of Ingress:

{% tabs %}

{% tab title="Istio" %}

```bash
helm install seldon-core seldon-core-operator \
    --repo https://storage.googleapis.com/seldon-charts \
    --set usageMetrics.enabled=true \
    --set istio.enabled=true \
    --namespace seldon-system
```

{% endtab %}

{% tab title="Ambassador" %}

```bash
helm install seldon-core seldon-core-operator \
    --repo https://storage.googleapis.com/seldon-charts \
    --set usageMetrics.enabled=true \
    --set ambassador.enabled=true \
    --namespace seldon-system
```

{% endtab %}

{% endtabs %}

You can check that your Seldon Controller is running by doing:

```bash
kubectl get pods -n seldon-system
```

You should see a `seldon-controller-manager` pod with `STATUS=Running`.

## Local Port Forwarding

Because your kubernetes cluster is running locally, we need to forward a port on your local machine to one in the cluster for us to be able to access it externally. You can do this by running:

{% tabs %}

{% tab title="Istio" %}

```bash
kubectl port-forward -n istio-system svc/istio-ingressgateway 8080:80
```

{% endtab %}

{% tab title="Ambassador" %}

```bash
kubectl port-forward -n ambassador svc/ambassador 8080:80
```

{% endtab %}

{% endtabs %}

This will forward any traffic from port 8080 on your local machine to port 80 inside your cluster.

You have now successfully installed Seldon Core on a local cluster and are ready to [start deploying models](../workflow/github-readme.md) as production microservices.
