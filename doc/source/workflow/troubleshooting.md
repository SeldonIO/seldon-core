# Troubleshooting Guide

If your Seldon Deployment does not seem to be running here are some tips to
diagnose the issue.

## My model does not seem to be running

Check whether the Seldon Deployment is running:

```bash
kubectl get sdep
```

If it exists, check its status, for a Seldon deployment called `<name>`:

```bash
kubectl get sdep <name> -o jsonpath='{.status}'
```

This might look like:

```bash
>kubectl get sdep

NAME      AGE
mymodel   1m

>kubectl get sdep mymodel -o jsonpath='{.status}'
map[predictorStatus:[map[name:mymodel-mymodel-7cd068f replicas:1 replicasAvailable:1]] state:Available]
```

If you have the `jq` tool installed you can get a nicer output with:

```bash
>kubectl get sdep mymodel -o json | jq .status
{
  "predictorStatus": [
    {
      "name": "mymodel-mymodel-7cd068f",
      "replicas": 1,
      "replicasAvailable": 1
    }
  ],
  "state": "Available"
}
```

For a model with invalid json/yaml an example is shown below:

```bash
>kubectl get sdep seldon-model -o json | jq .status
{
  "description": "Cannot find field: imagename in message k8s.io.api.core.v1.Container",
  "state": "Failed"
}
```

## Check all events on the SeldonDeployment

```bash
kubectl describe sdep mysdep
```

This will show each event from the operator including create, update, delete
and error events.

## My Seldon Deployment remains in "creating" state

Check if the pods are running successfully.

## I get 404s when calling the Ambassador endpoint

If your model is running and you are using Ambassador for ingress and are
having problems check the diagnostics page of Ambassador.
See [here](https://www.getambassador.io/docs/edge-stack/latest/topics/running/debugging/).
You can then find out what path your model can be found under to ensure the URL
you are using is correct.

If your ambassador isn't running at all then check the pod logs with `kubectl logs <pod_name>`.
Note that if ambassador is installed with cluster-wide scope then its rbac
should also not be namespaced, otherwise there will be a permissions error.

## I get 500s when calling my model over the API

Check the logs of your running model pods.

## My Seldon Deployment is not listed

Check the logs of the Seldon Operator.
This is the pod which handles the Seldon Deployment graphs sent to Kubernetes.
On a default installation, you can find the operator pod on the `seldon-system`
namespace.
The pod will be labelled as `control-plane=seldon-controller-manager`, so to
get the logs you can run:

```bash
kubectl logs -n seldon-system -l control-plane=seldon-controller-manager
```

### Invalid memory address

On some cases, you will see an error message on the operator logs like the
following:

```
panic: runtime error: invalid memory address or nil pointer dereference
```

This error can be caused by empty or unexpected values in the
`SeldonDeployment` spec.
The main cause is usually a misconfiguration of the mutating webhook.
To fix it, you can try to [re-install Seldon Core](./install.md) in your
cluster.

## I have tried the above and I'm still confused

- Contact our [Slack Community](https://join.slack.com/t/seldondev/shared_invite/enQtMzA2Mzk1Mzg0NjczLTJlNjQ1NTE5Y2MzMWIwMGUzYjNmZGFjZjUxODU5Y2EyMDY0M2U3ZmRiYTBkOTRjMzZhZjA4NjJkNDkxZTA2YmU)
- Create an [issue on Seldon Core's Github repo](https://github.com/SeldonIO/seldon-core/issues).
  Please make sure to add any diagnostics from the above suggestions to help us
  diagnose your issue.
