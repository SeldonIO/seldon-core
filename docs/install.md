# Install Seldon-Core

To install seldon-core on a Kubernetes cluster you have several choices:

 * Decide on which package manager to use, we support:
   * Helm
   * Ksonnet
 * Decide on how you wish APIs to be exposed, we support:
   * Ambassador reverse proxy
   * Seldon's builtin OAuth API Gateway
 * Decide on whether you wish to contribute anonymous usage metrics. We encourage you to allow anonymous usage metrics to help us improve the project by understanding the deployment environments. More details can be found [here](/docs/developer/readme.md#usage-reporting)
  * Does your kubernetes cluster have RBAC enabled?
    * If not then disable seldon RBAC setup

Follow one of the methods below:

## With Helm

 * [Install Helm](https://docs.helm.sh)
 * [Optionally, Install Ambassador](https://www.getambassador.io)
 * Install Seldon CRD. Set:
    * ```usage_metrics.enabled``` as appropriate.

```
helm install seldon-core-crd --name seldon-core-crd --repo https://storage.googleapis.com/seldon-charts \
     --set usage_metrics.enabled=true
```
 * Install seldon-core components. Set
    * ```apife.enabled``` : (default true) set to ```false``` if you have installed Ambassador.
    * ```rbac.enabled``` : (default true) set to ```false``` if running an old Kubernetes cluster without RBAC.
```    
helm install seldon-core --name seldon-core --repo https://storage.googleapis.com/seldon-charts \
     --set apife.enabled=<true|false> \
     --set rbac.enabled=<true|false>
```

Notes

 * You can use ```--namespace``` to install seldon-core to a particular namespace

## With Ksonnet

 * [install Ksonnet](https://ksonnet.io/)
 * [Optionally, Install Ambassador](https://www.getambassador.io) 
 * Create a seldon ksonnet app
 ```
 ks init my-ml-deployment --api-spec=version:v1.8.0
 ```
 * Install seldon-core. Set:
   * ```withApife``` set to ```false``` if you are using Ambassador
   * ```withRbac``` set to ```true``` if your cluster has RBAC enabled
```
cd my-ml-deployment && \
    ks registry add seldon-core github.com/SeldonIO/seldon-core/tree/master/seldon-core && \
    ks pkg install seldon-core/seldon-core@master && \
    ks generate seldon-core seldon-core \
       --withApife=<true|false> \
       --withRbac=<true|false> 
```
 * Launch components onto cluster
 ```
 ks apply default
 ```
Notes

 * You can use ```--namespace``` to install seldon-core to a particular namespace

## Other Options

### Install with kubeflow

  * [Install Seldon as part of kubeflow.](https://github.com/kubeflow/kubeflow/blob/master/user_guide.md)