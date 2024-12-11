---
description: >-
  Learn about installing Istio ingress controller in a Kubernetes cluster running Seldon Core 2.
---

An ingress controller functions as a reverse proxy and load balancer, implementing a Kubernetes Ingress. It adds an abstraction layer for traffic routing by receiving traffic from outside the Kubernetes platform and load balancing it to Pods running within the Kubernetes cluster. 

Seldon Core 2 works seamlessly with any service mesh or ingress controller, offering flexibility in your deployment setup. This guide provides detailed instructions for installing and configuring Istio with Seldon Core 2.

# Istio

Istio implements the Kubernetes ingress resource to expose a service and make it accessible from outside the cluster. You can install Istio in either a self-hosted Kubernetes cluster or a managed Kubernetes service provided by a cloud provider that is running the Seldon Enterprise Platform.

## Prerequisites

* Install[ Seldon Core 2](/docs-gb/installation/production-environment/README.md).
* Ensure that you install a version of Istio that is compatible with your Kubernetes cluster version. For detailed information on supported versions, refer to the [Istio Compatibility Matrix](https://istio.io/latest/docs/releases/supported-releases/#support-status-of-istio-releases).

## Installing Istio ingress controller

Installing Istio ingress controller in a Kubernetes cluster running Seldon Enterprise Platform involves these tasks:

1. [Install Istio](istio.md#install-istio)
2. [Install Istio Ingress Gateway](istio.md#install-istio-ingress-gateway)
3. [Expose Seldon mesh service](istio.md#expose-seldon-mesh-service)

### Install Istio

1.  Download the Istio installation package for the version you want to use. In the following command replace `<version>` with the version of Istio that you downloaded:

    ```
    curl -L https://istio.io/downloadIstio | sh -
    cd istio-<version>
    export PATH=$PWD/bin:$PATH
    ```
2.  Install the Istio Custom Resource Definitions (CRDs) and Istio components in your cluster using the `istioctl` command line tool:

    ```
    istioctl install --set profile=default -y
    ```
3.  Create a namespace where you want to enable Istio automatic sidecar injection. For example in the namespace `seldon-mesh`:

    ```
    kubectl label namespace seldon-mesh istio-injection=enabled
    ```

### Install Istio Ingress Gateway

1.  Verify that Istio Ingress Gateway is installed:

    ```
    kubectl get svc istio-ingressgateway -n istio-system
    ```

    This should return details of the Istio Ingress Gateway, including the external IP address.
2.  Create a YAML file to specify Gateway resource in the `istio-system` namespace to expose your application. For example, create the `istio-seldon-gateway.yaml` file. Use your preferred text editor to create and save the file with the following content:

    ```
     apiVersion: networking.istio.io/v1alpha3
     kind: Gateway
     metadata:
       name: my-gateway
       namespace: seldon-mesh
     spec:
       selector:
         istio: ingressgateway # Use Istio's default ingress gateway
       servers:
       - port:
           number: 80
           name: http
           protocol: HTTP
         hosts:
         - "*"
    ```
3.  Change to the directory that contains `istio-seldon-gateway.yaml` file and apply the configuration:

    ```
    kubectl apply -f istio-seldon-gateway.yaml
    ```

    When the configuration is applied, you should see this:

    ```
    gateway.networking.istio.io/seldon-gateway created
    ```
4.  Find the IP address of the Seldon Core 2 instance running with Istio:

    ```
    ISTIO_INGRESS=$(kubectl get svc seldon-mesh -n seldon-mesh -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    
    echo "Seldon Core 2: http://$ISTIO_INGRESS"

    ```

    {% hint style="info" %}
    Make a note of the IP address that is displayed in the output. This is the IP address that you require to try out the [Kubernetes Examples](/docs-gb/examples/k8s-examples.md).
    {% endhint %}

### Expose Seldon mesh service

It is important to expose `seldon-service` service to enable communication between deployed machine learning models and external clients or services. The Seldon Core 2 inference API is exposed through the `seldon-mesh` service in the `seldon-mesh` namespace. If you install Core 2 in multiple namespaces, you need to expose the `seldon-mesh` service in each of namespace.

1.  Verify if the `seldon-mesh` service is running for example, in the namespace `seldon`.

    ```bash
    kubectl get svc -n seldon-mesh
    ```

    When the services are running you should see something similar to this:

    ```bash
    mlserver-0               ClusterIP      None             <none>          9000/TCP,9500/TCP,9005/TCP                                                                  43m
    seldon-mesh              LoadBalancer   34.118.225.130   34.90.213.15    80:32228/TCP,9003:31265/TCP                                                                 45m
    seldon-pipelinegateway   ClusterIP      None             <none>          9010/TCP,9011/TCP                                                                           45m
    seldon-scheduler         LoadBalancer   34.118.225.138   35.204.34.162   9002:32099/TCP,9004:32100/TCP,9044:30342/TCP,9005:30473/TCP,9055:32732/TCP,9008:32716/TCP   45m
    triton-0                 ClusterIP      None             <none>          9000/TCP,9500/TCP,9005/TCP 
    ```
2.  Create a YAML file to create a VirtualService named `iris-route` in the namespace `seldon-mesh`. For example, create the `seldon-mesh-vs.yaml` file. Use your preferred text editor to create and save the file with the following content:

    ```yaml
    apiVersion: networking.istio.io/v1alpha3
    kind: VirtualService
    metadata:
      name: iris-route
      namespace: seldon-mesh
    spec:
      gateways:
        - istio-system/seldon-gateway
      hosts:
        - "*"
      http:
        - match:
          - uri:
              prefix: /v2
        name: iris-http
        route:
        - destination:
          host: seldon-mesh.seldon-mesh.svc.cluster.local
        headers:
      ```
3.  Create a virtual service to expose the `seldon-mesh` service.

    ```
    kubectl apply -f seldon-mesh-vs.yaml
    ```

    When the virtual service is created, you should see this:

    ```
    virtualservice.networking.istio.io/iris-route created
    ```
#### Optional: Enable HTTPS/TLS

To secure your Ingress with HTTPS, you can configure TLS settings in the `Gateway` resource using a certificate and key. This involves additional steps like creating Kubernetes secrets for your certificates.

#### Additional Resources

* [Istio Documentation](https://istio.io/latest/docs/tasks/traffic-management/ingress/secure-ingress/)
* [GKE Ingress Guide](https://cloud.google.com/kubernetes-engine/docs/concepts/ingress)
* [AWS Documentation](https://docs.aws.amazon.com/eks/latest/userguide/what-is-eks.html)
