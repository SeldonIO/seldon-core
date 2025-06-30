# Scaling Replicas

## Replica Settings

Replicas settings can be provided at several levels with the most specific taking precedence, from most general to most specific as shown below:

  * `.spec.replicas`
  * `.spec.predictors[].replicas`
  * `.spec.predictors[].componentSpecs[].replicas`

If you use the annotation `seldon.io/engine-separate-pod` you can also set the number of replicas for the service orchestrator in:

 * `.spec.predictors[].svcOrchSpec.replicas`

As illustration, a contrived example showing various options is shown below:

```yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: test-replicas
spec:
  replicas: 1
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier_rest:1.3
          name: classifier
    - spec:
        containers:
        - image: seldonio/mock_classifier_rest:1.3
          name: classifier2
      replicas: 3
    graph:
      endpoint:
        type: REST
      name: classifier
      type: MODEL
      children:
      - name: classifier2
        type: MODEL
        endpoint:
          type: REST
    name: example
    replicas: 2
    traffic: 50
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier_rest:1.3
          name: classifier3
    graph:
      children: []
      endpoint:
        type: REST
      name: classifier3
      type: MODEL
    name: example2
    traffic: 50

```

 * classfier will have a deployment with 2 replicas as specified by the predictor it is defined within
 * classifier2 will have a deployment with 3 replicas as that is specified in its componentSpec
 * classifier3 will have 1 replica as it takes its value from `.spec.replicas`

For more details see [a worked example for the above replica settings](../examples/scale.html).

## Scale replicas

Its is possible to use the `kubectl scale` command to set the `replicas` value of the SeldonDeployment. For simple inference graphs this can be an easy way to scale them up and down. For example:

```yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: seldon-scale
spec:
  replicas: 1  
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
```

One can scale this Seldon Deployment up using the command:

```console
kubectl scale --replicas=2 sdep/seldon-scale
```

For more details you can follow [a worked example of scaling](../examples/scale.html).

## Autoscaling Seldon Deployments

To autoscale your Seldon Deployment resources you can add Horizontal Pod Template Specifications to the Pod Template Specifications you create. There are two steps:

  1. Ensure you have a resource request for the metric you want to scale on if it is a standard metric such as cpu or memory. This has to be done for every container in the seldondeployment, except for the seldon-container-image and the storage initializer. Some combinations of protocol and server type may spawn additional support containers; resource requests have to be added to those containers as well.
  2. Add a HPA Spec referring to this Deployment.

We presently support the `autoscaling/v2beta1` definition in the existing `metrics` field as well as the `autoscaling/v2` definition in the `metricsv2` field of the SeldonDeployment `hpaSpec`. In both cases they will create a K8s `autoscaling/v2` HPA which means you will need to be running a [Kubernetes cluster of >= 1.23](https://kubernetes.io/docs/reference/using-api/deprecation-guide/#horizontalpodautoscaler-v125).

To illustrate this we have an example Seldon Deployment below with the `v2` definition:

```yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: seldon-model
spec:
  name: test-deployment
  predictors:
  - componentSpecs:
    - hpaSpec:
        maxReplicas: 3
        metricsv2:
        - resource:
            name: cpu
            target:
              type: Utilization
              averageUtilization: 70
          type: Resource
        minReplicas: 1
      spec:
        containers:
        - image: seldonio/mock_classifier:1.5.0-dev
          imagePullPolicy: IfNotPresent
          name: classifier
          resources:
            requests:
              cpu: '0.5'
        terminationGracePeriodSeconds: 1
    graph:
      children: []
      name: classifier
      type: MODEL
    name: example
```

The key points here are:

 * We define a CPU request for our container. This is required to allow us to utilize cpu autoscaling in Kubernetes.
 * We define an HPA associated with our componentSpec which scales on CPU when the average CPU is above 70% up to a maximum of 3 replicas.

Once deployed, the HPA resource may take a few minutes to start up. To check status of the HPA resource, `kubectl describe hpa -n <podname>` may be used.

An example using the `v2beta1` definition is shown below:

```yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: seldon-model
spec:
  name: test-deployment
  predictors:
  - componentSpecs:
    - hpaSpec:
        maxReplicas: 3
        metrics:
        - resource:
            name: cpu
            targetAverageUtilization: 70
          type: Resource
        minReplicas: 1
      spec:
        containers:
        - image: seldonio/mock_classifier:1.5.0-dev
          imagePullPolicy: IfNotPresent
          name: classifier
          resources:
            requests:
              cpu: '0.5'
        terminationGracePeriodSeconds: 1
    graph:
      children: []
      name: classifier
      type: MODEL
    name: example

```

For worked examples see [this notebook](../examples/autoscaling_example.html).
