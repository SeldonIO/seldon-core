# Upgrading Seldon Core

This page provides with instructions on how to upgrade from previous versions. Each of these have to be run sequentially if it's expected for previous running models to be upgraded without disruption (i.e if you are running version 0.4.2, you will first have to upgrade to 0.5.2 and then to 1.1, etc).

If you were running our Openshift 0.4.2 certified operator and are looking to upgrade to our 1.1 certified operator, you will also need to follow the "upgrading process" steps in the "Upgrading to 0.5.2 from previous versions" section.

Make sure you also [read the CHANGELOG](./changelog.html) to see the detailed features and bug-fixes in each version.

## Upgrading to 1.3

### Breaking Changes

The version of sklearn used by the default sklearn server will be 0.23.2. To use a different version you will need to follow the steps described in the [sklearn server documentation](../servers/sklearn.html).

## Upgrading to 1.2.1

*[NOTE]* 1.2.0 has issue where all Seldon Deployments are marked as "NotReady" as there is a [bug caused by a volumeName update](https://github.com/SeldonIO/seldon-core/issues/2017). This can be resolved by following the 1.2.0 volume patch [as outlined by this example](../examples/patch_1_2.html). It is recommended to upgrade to version 1.2.1 directly instead.

All seldon-managed pods will be subject to a rolling update as part of this upgrade.

### New Features / Breaking Changes

 * The helm value `createResources` has been renamed `managerCreateResources`.
 * To allow CRDs to be created by the manager. If `managerCreateResources` is true then extra RBAC to `create` CRDs is added from the previous versions RBAC which was to just list and get.
 * If upgrading the analytics helm chart then a `kubectl delete deployment -n seldon-system -l app=grafana` should be [run first](https://github.com/SeldonIO/seldon-core/pull/1917)
 * All the prepackaged model servers are now created with RedHat UBI images. One consequence of this is that they will all run as non-root as it best practice.
 
### Request Logger

The values.yaml for the seldon-core-operator helm chart has changed. The field `defaultRequestLoggerEndpointPrefix` is replaced by:

```
  requestLogger:
    defaultEndpoint: 'http://default-broker'
```

This default value will find a broker in the model's namespace. To point to another namespace it would be `default-broker.anothernamespace`.

## Upgrading to 1.1 from previous versions

As we moved to 1.x+ there are several breaking changes that need to be considered. These are outlined below

### New Features / Breaking Changes

#### Deployment Naming and Rolling Updates

The deployments created by Seldon Core have been changed to follow a fixed scheme. It will now be:

```
<seldondeployment name>-<predictor name>-<podspec idx>-<container names>
```

So for example:

```
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: rest-seldon
spec:
  name: restseldon
  protocol: seldon
  transport: rest
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier_rest:1.3
          name: classifier
    graph:
      name: classifier
      type: MODEL
    name: model
    replicas: 1
```

For the above resource, one Deployment will be created with name:

```
rest-seldon-model-0-classifier
```

This will change how rolling updates are done. Now any change to the first PodSpec above will be updated via a rolling update as expected if the names of the containers are not changed. If however you changed "classifier" to "classifier2" you would get a new deployment created which would replace the old deployment when running.


#### New Service Orchestrator

From version 1.1 Seldon Core comes with a new service orchestrator written in Go which replaces the previous Java engine. Some breaking changes are present:

 * Metadata fields in the Seldon Protocol are no longer added. Any custom metata data will need to be added and exposed to Prometheus metrics by the individual components in the graph
 * All components in the graph must either be REST or gRPC and only the given protocol is exposed externally.
 * The metric names placed in Prometheus have changed to include the `executor` name rather than `engine` : see the [analytics docs](../analytics/analytics.html)

The new service orchestrator comes with several advantages including ability to handle Tensorflow REST and gRPC protocols and full metrics and tracing support for both REST and gRPC.

For those wishing to use the deprecated Java engine service orchestrator see [the service orchestrator docs](../graph/svcorch.md) for details.

#### Ambassador Retries

Ambassador retries has been removed from the previous hardwired value of 3. Retries is now available via an [annotation for Ambassador](../ingress/ambassador.html).


### Python Wrapper Tag Update

The Python Wrapper was using naming convention in the format 0.1 ... 0.18. In this release we have renamed the version of the Python Wrapper tag to match the same convention as the Executor, Operator, etc. This means that the Python Wrapper tag for this release is 1.1, and the snapshot would be 1.1.1-SNAPSHOT

### Dated SNAPSHOTS

Whenever a new PR was merged to master, we have set up our CI to build a "SNAPSHOT" version, which would contain the Docker images for that specific development / master-branch code.

Previously, we always had the SNAPSHOT tag being overriden with the latest. This didn't allow us to know what version someone may be trying out when using master, so we wanted to introduce a way to actually get unique tags for every image that gets landed into master.

Now every time that a PR is landed to master, a new "dated" SNAPSHOT version is created, which pushes images with the tag `"<next-version>-SNAPSHOT_<timestamp>"`. A new branch is also created with the name `"v<next-version>-SNAPSHOT_<timestamp>"`, which contains the respective helm charts, and allows for the specific version (as outlined by the version in `version.txt`) to be installed.

You can follow the instructions in the [installation page](../workflow/install.md) to install the snapshot version.

### Wrapper compatibility table

To verify if Seldon Core v1.0 and v.1.1 is compatible with older s2i wrapper versions we conducted a simple test with a one-node model.
The model has been deployed both with REST and GRPC API with both new orchestrator and the deprecated Java engine (v1.0 only with Java Engine).
Test verifies if model can successfully serve inference requests.

**NOTE:** Full support of custom metrics and tags with new orchestrator is only available from Python wrapper version 0.19.
If you need to use older version of Python wrapper you can continue to use Java engine as described above until the next release.


| Language Wrapper |     Version   | API Type | New Orchestrator  | Deprecated Java engine | Notes                                   |
|------------------|---------------|----------|-------------------|------------------------|-----------------------------------------|
| Python           | 0.19          | both     | yes               | yes                    | full support of custom metrics and tags |
| Python           | 0.11 ... 0.18 | both     | yes               | yes                    |                .                        |
| Python           | 0.10          | REST     | no                | yes                    |                .                        |
| Python           | 0.10          | GRPC     | yes               | yes                    |                .                        |
| Python           | < 0.10        | GRPC     | ?                 | ?                      |                .                        |
| Java             | 0.2 & 0.1     | REST     | yes               | yes                    |   minor difference in request format    |
| Java             | 0.2 & 0.1     | GRPC     | yes               | yes                    |                .                        |



Example of request format difference with Java wrapper deployed with REST API:

1. Using new orchestrator:
```bash
curl -s -X POST \
    -d 'json={"data": {"names": ["a", "b"], "ndarray": [[1.0, 2.0]]}}' \
    localhost:8003/seldon/seldon/compat-rest-java-02-executor/api/v1.0/predictions
```

2. Using deprecated Java engine:
```bash
curl -s -X POST -H 'Content-Type: application/json' \
    -d '{"data": {"names": ["a", "b"], "ndarray": [[1.0, 2.0]]}}' \
    localhost:8003/seldon/seldon/compat-rest-java-02-engine/api/v1.0/predictions
```


## Upgrading to 0.5.2 from previous versions

This version included significant improvements and features, including the addition of pre-packaged model servers, fixing several critical bugs.

This was the version that also dropped support for kubernetes 1.11, and added changes in the mutating and validating webhooks that you need to make sure are transitioned as outlined below.

### Upgrading process

In order to upgrade, the main requirement is to make sure that the Kubernetes cluster is updated to 1.12 or higher.

Once this is done, it's necessary to delete the old webhooks. This can be done with the following commands (you need to make sure that the commands are executed in the namespace in which seldon core was installed).

## Upgrading to 0.2.8 from previous versions.

### Upgrading process

#### Installation process now with Helm

The helm charts to install Seldon Core have changed. There is now a single Helm chart `seldon-core-operator` that installs the CRD and its controller. Ingress options are now separate and you need to choose between the available options which are at present:

 * Ambassador - via its official Helm chart
 * Istio

For more details see the [install docs](../workflow/install.md).

The Helm chart `seldon-core-operator` will require clusterwide RBAC and should be installed by a cluster admin.

##### Dropping support for KSonnet

Ksonnet is now deprecated. You should convert to using Helm to install Seldon Core.
