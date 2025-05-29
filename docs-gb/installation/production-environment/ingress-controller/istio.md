---
description: Learn how to install and configure Istio ingress controller for Seldon Core 2 in Kubernetes, including service mesh integration, traffic routing, and load balancing. This comprehensive guide covers Istio installation, ingress gateway setup, virtual service configuration, and exposing ML model services for production deployment.
---

An ingress controller functions as a reverse proxy and load balancer, implementing a Kubernetes Ingress. It adds an abstraction layer for traffic routing by receiving traffic from outside the Kubernetes platform and load balancing it to Pods running within the Kubernetes cluster. 

Seldon Core 2 works seamlessly with any service mesh or ingress controller, offering flexibility in your deployment setup. This guide provides detailed instructions for installing and configuring Istio with Seldon Core 2.

# Istio

Istio implements the Kubernetes ingress resource to expose a service and make it accessible from outside the cluster. You can install Istio in either a self-hosted Kubernetes cluster or a managed Kubernetes service provided by a cloud provider that is running the Seldon Core 2.

## Prerequisites

* Install[ Seldon Core 2](/docs-gb/installation/production-environment/README.md).
* Ensure that you install a version of Istio that is compatible with your Kubernetes cluster version. For detailed information on supported versions, refer to the [Istio Compatibility Matrix](https://istio.io/latest/docs/releases/supported-releases/#support-status-of-istio-releases).

## Installing Istio ingress controller

Installing Istio ingress controller in a Kubernetes cluster running Seldon Core 2 involves these tasks:

1. [Install Istio](istio.md#install-istio)
2. [Install Istio Ingress Gateway](istio.md#install-istio-ingress-gateway)
3. [Expose Seldon mesh service](istio.md#expose-seldon-mesh-service)

### Install Istio

1.  Add the Istio Helm charts repository and update it:

    ```
    helm repo add istio https://istio-release.storage.googleapis.com/charts
    helm repo update
    ```
2.  Create the `istio-system` namespace where Istio components are installed:

    ```
    kubectl create namespace istio-system
    ```
3.  Install the base component:

    ```
    helm install istio-base istio/base -n istio-system
    ```
4. Install Istiod, the Istio control plane:

    ```
    helm install istiod istio/istiod -n istio-system --wait
    ```    

### Install Istio Ingress Gateway

1. Install Istio Ingress Gateway:

    ```
    helm install istio-ingressgateway istio/gateway -n istio-system
    ```

1.  Verify that Istio Ingress Gateway is installed:

    ```
    kubectl get svc istio-ingressgateway -n istio-system
    ```

    This should return details of the Istio Ingress Gateway, including the external IP address.

2. Verify that all Istio Pods are running:

    ```
    kubectl get pods -n istio-system
    ```
    The output is similar to:

    ```
    NAME                          READY   STATUS    RESTARTS   AGE
    istiod-xxxxxxx-xxxxx          1/1     Running   0          2m
    istio-ingressgateway-xxxxx    1/1     Running   0          2m
    ```
3. Inject Envoy sidecars into application Pods in the namespace `seldon-mesh`:

    ```
    kubectl label namespace seldon-mesh istio-injection=enabled
    ```    
4.  Verify that the injection happens to the Pods in the namespace `seldon-mesh`:

    ```
    kubectl get namespace seldon-mesh --show-labels
    ```
5.  Find the IP address of the Seldon Core 2 instance running with Istio:

    ```
    ISTIO_INGRESS=$(kubectl get svc seldon-mesh -n seldon-mesh -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    
    echo "Seldon Core 2: http://$ISTIO_INGRESS"

    ```

    {% hint style="info" %}
    Make a note of the IP address that is displayed in the output. This is the IP address that you require to [test the installation](/docs-gb/installation/test-installation.md).
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
        - name: iris-http
          match:
            - uri:
                prefix: /v2
          route:
            - destination:
                host: seldon-mesh.seldon-mesh.svc.cluster.local
    ```

3.  Create a virtual service to expose the `seldon-mesh` service.

    ```
    kubectl apply -f seldon-mesh-vs.yaml
    ```

    When the virtual service is created, you should see this:

    ```
    virtualservice.networking.istio.io/iris-route created
    ```


### Next Steps
[Verify the installation](/docs-gb/installation/test-installation.md)

#### Additional Resources

* [Istio Documentation](https://istio.io/latest/docs/tasks/traffic-management/ingress/secure-ingress/)
* [GKE Ingress Guide](https://cloud.google.com/kubernetes-engine/docs/concepts/ingress)
* [AWS Documentation](https://docs.aws.amazon.com/eks/latest/userguide/what-is-eks.html)
