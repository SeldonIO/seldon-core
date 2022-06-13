# Graph Deployment Options

In Seldon core there is the capability to have different mode of scopes in containerizing models and Seldon core components in the inference graph.
Each node of the inference graph will be a container in the Kubernetes cluster. Inference graph nodes containers could be encapsulated in a 
single or multiple kubernetes pods. The outer component of Seldon core are predictors which could contain one or more componentes that are referred 
by their name in constructing the inference graph in `spec.componentSpecs.graph`.

## Mode One: Single pod deployment

The following is an example of a Seldon core inference graph with a 
single predictor.
```bash
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: linear-pipeline-single-pod
spec:
  name: linear-pipeline
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier:1.0
          name: node-one
        - image: seldonio/mock_classifier:1.0
          name: node-two
        - image: seldonio/mock_classifier:1.0
          name: node-three
    graph:
      name: node-one
      type: MODEL
      children:
      - name: node-two
        type: MODEL
        children:
        - name: node-three
          type: MODEL
          children: []
    name: example
```

This will result in deploying all the graph nodes in a single pod:

```bash
kubectl get pods

NAME                                                       READY   STATUS    RESTARTS   AGE
seldon-c71cc2d950d44db1bc6afbeb0194c1da-5d8dddb8cb-xx4gv   5/5     Running   0          6m59s
```

## Mode Two: Separate pod deployment

Another way of deployment is to implement the each node of inference graph in a seperate predictor which will result in having separate pods for 
each inference graph node.

```bash
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: linear-pipeline-separate-pods
spec:
  name: linear-pipeline
  annotations:
    seldon.io/engine-separate-pod: "true"
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier:1.0 
          name: node-one
          imagePullPolicy: Always
    - spec:
        containers:
        - image: seldonio/mock_classifier:1.0
          name: node-two
          imagePullPolicy: Always
    - spec:
        containers:
          - image: seldonio/mock_classifier:1.0
            name: node-three
            imagePullPolicy: Always
    graph:
      name: node-one
      type: MODEL
      children:
      - name: node-two
        type: MODEL
        children:
        - name: node-three
          type: MODEL
          children: []
    name: example
```

This time it will result in having separate pods for each container.

```bash
kubectl get pods
NAME                                                              READY   STATUS    RESTARTS   AGE
linear-pipeline-separate-pods-example-0-node-one-6954fbbd5m7pcp   1/1     Running   0          4m33s
linear-pipeline-separate-pods-example-1-node-two-c4f55f689gxkkr   1/1     Running   0          4m33s
linear-pipeline-separate-pods-example-2-node-three-99667dcmg9kg   1/1     Running   0          4m33s
linear-pipeline-separate-pods-example-svc-orch-656c6bdf59-6m6nc   1/1     Running   0          4m33s
```
The most basic unit in Kubernetes are pods. This model will enable [scaling](scaling.md) at model level. In other words, you can 
scale each model separately while on the other hand having them in a single pod will change the granulity of scaling to the entire graph. However, 
on the other hand single pod deployment will need only a single [sidecar istio container](../ingress/istio.md)
that needs less resource request from the sidecar containers. Another potential difference is the less communication overhead in the single pod mode as
they will always be schduled on the same Kubernetes node.
