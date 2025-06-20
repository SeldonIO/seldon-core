# Install on Google Cloud Platform

This guide runs through how to set up and install Seldon Core in a Kubernetes cluster running on Google Cloud. By the end, you'll have Seldon Core up and running and be ready to start deploying machine learning models.

## Prerequisites

### Google Cloud CLI

You will need the Google Cloud SDK in order to retrieve your cluster authentication credentials. It can also be used to create clusters and other resources for you:

- [Install Google Cloud SDK](https://cloud.google.com/sdk/docs/install)

### Google Kubernetes Engine (GKE) Cluster

If you haven't already created a Kubernetes cluster on GKE, you can follow this quickstart guide to get set up with your first cluster. Ensure you create 3 nodes (`--num-nodes` passed to `gcloud container clusters create`) to be able to run a minimal example:

- [Create GKE Cluster](https://cloud.google.com/kubernetes-engine/docs/quickstart)

{% hint style="warning" %}
Your GKE cluster needs to be a `Standard` type cluster. GKE Autopilot clusters currently don't support installation of ingress controllers.
{% endhint %}

### Kubectl

[`kubectl`](https://kubernetes.io/docs/reference/kubectl/overview/) is the Kubernetes command-line tool. It allows you to run commands against Kubernetes clusters, which we'll need to do as part of setting up Seldon Core.

- [Install kubectl on Linux](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux)
- [Install kubectl on macOS](https://kubernetes.io/docs/tasks/tools/install-kubectl-macos)
- [Install kubectl on Windows](https://kubernetes.io/docs/tasks/tools/install-kubectl-windows)

### Helm

[Helm](https://helm.sh/) is a package manager that makes it easy to find, share and use software built for Kubernetes. If you don't already have Helm installed locally, you can install it here:

- [Install Helm](https://helm.sh/docs/intro/install/)

## Connect to Your Cluster

You can connect to your cluster by running the following `gcloud` command:

```bash
gcloud container clusters get-credentials CLUSTER_NAME
```

This will configure `kubectl` to use your Google Cloud Kubernetes cluster. Don't forget to replace `CLUSTER_NAME` with whatever you called your cluster when you created it. If you've forgotten, you can run `gcloud container clusters list`.

{% hint style="info" %}
If you get authentication errors while running the command above, try running `gcloud auth login` to check you are correctly logged in.
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

## Accessing your models

Congratulations! Seldon Core is now fully installed and operational. Before you move on to deploying models, make a note of your cluster IP and port:

{% tabs %}

{% tab title="Istio" %}

```bash
export INGRESS_HOST=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
export INGRESS_PORT=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="http2")].port}')
export INGRESS_URL=$INGRESS_HOST:$INGRESS_PORT
echo $INGRESS_URL
```

This is the public address you will use to access models running in your cluster.

{% endtab %}

{% tab title="Ambassador" %}

```bash
export INGRESS_HOST=$(kubectl -n ambassador get service ambassador -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
export INGRESS_PORT=$(kubectl -n ambassador get service ambassador -o jsonpath='{.spec.ports[?(@.name=="http")].port}')
export INGRESS_URL=$INGRESS_HOST:$INGRESS_PORT
echo $INGRESS_URL
```

This is the public address you will use to access models running in your cluster.

{% endtab %}

{% endtabs %}

You are now ready to [deploy models to your cluster](../workflow/github-readme.md).
