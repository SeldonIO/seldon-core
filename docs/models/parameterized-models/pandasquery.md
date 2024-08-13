# Pandas Query Model

This model allows a [Pandas](https://pandas.pydata.org/)  query to be run in the input to select rows. An example is shown below:

```yaml
# samples/models/choice1.yaml

```

This invocation check filters for tensor A having value 1.

* The model also returns a tensor called `status` which indicates the operation run and whether it
was a success. If no rows satisfy the query then just a `status` tensor output will be returned.
* Further details on Pandas query can be found [here](https://pandas.pydata.org/docs/reference/api/pandas.DataFrame.query.html)


This model can be useful for conditional Pipelines. For example, you could have two invocations of this model:

```yaml
# samples/models/choice1.yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: choice-is-one
spec:
  storageUri: "gs://seldon-models/scv2/examples/pandasquery"
  requirements:
  - mlserver
  - python
  parameters:
  - name: query
    value: "choice == 1"
```

and

```yaml
# samples/models/choice2.yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: choice-is-two
spec:
  storageUri: "gs://seldon-models/scv2/examples/pandasquery"
  requirements:
  - mlserver
  - python
  parameters:
  - name: query
    value: "choice == 2"
```

By including these in a Pipeline as follows we can define conditional routes:

```yaml
# samples/pipelines/choice.yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: choice
spec:
  steps:
  - name: choice-is-one
  - name: mul10
    inputs:
    - choice.inputs.INPUT
    triggers:
    - choice-is-one.outputs.choice
  - name: choice-is-two
  - name: add10
    inputs:
    - choice.inputs.INPUT
    triggers:
    - choice-is-two.outputs.choice
  output:
    steps:
    - mul10
    - add10
    stepsJoin: any
```

Here the mul10 model will be called if the choice-is-one model succeeds and the add10 model will
be called if the choice-is-two model succeeds.

The full notebook can be found [here](../examples/pandasquery.md)
