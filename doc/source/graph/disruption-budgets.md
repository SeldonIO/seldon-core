# Budgeting Disruptions

High availability is an important aspect in running production systems.
To this end, you can add Pod Disruption Budget Specifications to the Pod Template Specifications you create.
Depending on how you want your application to handle disruptions, you can define your disruption budget accordingly.

An example Seldon Deployment with disruption budgets defined can be seen below:

```yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: seldon-model
spec:
  name: test-deployment
  replicas: 2
  predictors:
  - componentSpecs:
    - pdbSpec:
        minAvailable: 90%
      spec:
        containers:
        - image: seldonio/mock_classifier_rest:1.3
          imagePullPolicy: IfNotPresent
          name: classifier
          resources:
            requests:
              cpu: '0.5'
        terminationGracePeriodSeconds: 1
    graph:
      children: []
      endpoint:
        type: REST
      name: classifier
      type: MODEL
    name: example
```

This example ensures that our serving capacity does not decrease by more than 10%.
