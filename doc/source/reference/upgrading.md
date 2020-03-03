# Upgrading Notes and Breaking Changes


## Upgrading from 1.0 to 1.1

### Breaking Changes

#### Deployment Naming and Rolling Updates

The deployments created by Seldon Core have been changed to follow a fixed scheme. It will now be:

```
<seldondeployment name>-<predictor name>-<podspec idx>
```

So for example:

```
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: seldon-model
spec:
  name: test-deployment
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier_rest:1.3
          name: classifier
    graph:
      children: []
      endpoint:
        type: REST
      name: classifier
      type: MODEL
    name: example
    replicas: 1
```

For the above resource, one Deployment will be created with name:

```
seldon-model-example-0
```

This will change how rolling updates are done. Now any change to the first PodSpec above will be updated via a rolling update as expected.

#### New Golang Service Orchestrator

A new Golang service orchestrator will replace the existing Java version. You can keep the existing Java service orchestrator if needed via an [install setting in helm or kustomize](../graph/svcorch.hml).

The new executor does not allow mixed protocols. You can use REST or gRPC components but not mix them in the same inference graph.


#### Misc Breaking Changes:

 * Ambassador retries has been removed from the previous hardwired value of 3. Retries is now available via an [annotation for Ambassador](../ingress/ambassador.html).


## Upgrading to 0.2.8 from previous versions.

### Helm

The helm charts to install Seldon Core have changed. There is now a single Helm chart `seldon-core-operator` that installs the CRD and its controller. Ingress options are now separate and you need to choose between the available options which are at present:

 * Ambassador - via its official Helm chart
 * Istio

For more details see the [install docs](../workflow/install.md).

The Helm chart `seldon-core-operator` will require clusterwide RBAC and should be installed by a cluster admin.

### Ksonnet

Ksonnet is now deprecated. You should convert to using Helm to install Seldon Core.
