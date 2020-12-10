# Seldon Core Release 0.3.0

A summary of the main contributions to the [Seldon Core release 0.3.0](https://github.com/SeldonIO/seldon-core/releases/tag/v0.3.0).

## Istio Ingress Support
Seldon core can now be used in conjunction with [istio](https://istio.io/). Istio provides an [ingress gateway](https://istio.io/docs/tasks/traffic-management/ingress/) which Seldon Core can automatically wire up new deployments to. 

Ensure when you install the seldon-core operator via Helm that you enable istio. For example:

```bash 
helm install seldon-core-operator --name seldon-core --set istio.enabled=true --repo https://storage.googleapis.com/seldon-charts --set usageMetrics.enabled=true
```

The seldon operator can now automatically add ingress API routes to the gateway to allow access to your running models.

The update also adds a new `traffic` field to the predictor specification. This will allow you to create a SeldonDeployment for canary and A/B testing use cases by providing a set amount of traffic (interpreted as a percentage) to each predictor using istio routing. An example that shows how to accomplish this for canary rollout is shown [here](../examples/istio_canary.html)

## Seldon Operator Update

One core part of the Seldon platform is the Operator which consists of the Custom Resource that allows users to define `SeldonDeployments` and the controller that understands and acts upon users creating these custom resources in their clusters. The controller has been rewritten in Go to provide better support for CRDs that the original Java implementation did not have.

Installation of Seldon Core is now accomplished with the single Helm install:

```bash
helm install seldon-core-operator --name seldon-core --repo https://storage.googleapis.com/seldon-charts
```

One consequence of the release is that volume mounts inside user Pods can be written in the usual manner now in our SeldonDeployment CRD. Whereas before you needed to use a more verbose representation due to the Java operator using the official Kubernetes proto buffer definitions:

```yaml
volumes:
- name: data
  volumeSource:
    hostPath:
      path: /data
      type: Directory
```

You can now use the standard way if you use volumes inside your PodSpec in a SeldonDeployment.

```yaml
volumes:
- name: data
    hostPath:
      path: /data
      type: Directory
```

Another change is that following standard practice our Operator now only has clusterwide permissions and will need to be installed by a sys-admin once for your cluster and we don't allow namespaced scoped installs of the controller.

## Separate Operator and Ingress Installation

As part of the operator change we have separated out the install of Seldon to focus on installing only the operator. Users can then install the ingress functionality they wish to use as a separate step. At present we support [Ambassador](../ingress/ambassador.html) and [Istio](../ingress/istio.html). For both these ingress technologies the Seldon controller can be utilised to automatically wire up ingress API endpoints inside Ambassador or an Istio ingress gateway. For the new installation instructions see [here](../workflow/install.html).

## Custom JSON Payloads

From customer requests we have extended our `SeldonMessage` payloads to allow arbitrary JSON in a field `jsonData`.
In our python wrapper support for the `jsonData` field provides the user defined predict function with a 'dict' when the `jsonData` appears in the incoming payload. If a 'dict' is returned from the predict call the data is returned via the 'jsonData' field. An example payload could be:

```JSON
{
 "jsonData" :
  {
    "custom": [1, 2, 3],
    "meta":"foo"
  }
}
```
			
## More Example Integrations

Our range of example has expanded to include:

 * [AWS EKS MNIST Example](../examples/aws_eks_deep_mnist.html)
 * [SKLearn Spacy NLP Example](../examples/sklearn_spacy_text_classifier_example.html)
 * [End to end kubeflow pipeline with Seldon](../examples/kubeflow_seldon_e2e_pipeline.html)

## Deprecation warning
As we move to stabilise the core functionality there are a few things we will be removing support for:

 * Ksonnet installation. As [Ksonnet](https://ksonnet.io/) as a project as been stopped we have stopped updates to our ksonnet installation. It will remain in the github repository master branch for the present time. Please contact us if you have a hard dependency on its usage.
 * Seldon OAuth Gateway. Our example OAuth will be deprecated in a future release and users should migrate to Cloud Native alternatives such as Ambassador or Istio. If you have particular ingress products you want us to support please get in contact.



