# Pandas Query Model

This model allows a [Pandas](https://pandas.pydata.org/)  query to be run in the input to select rows. An example is shown below:

```{literalinclude} ../../../../../samples/models/choice1.yaml 
:language: yaml
```

This invocation check filters for tensor A having value 1.

  * The model also returns a tensor called `status` which indicates the operation run and whether it was a success. If no rows satisfy the query then just a `status` tensor output will be returned.
  * Further details on Pandas query can be found [here](https://pandas.pydata.org/docs/reference/api/pandas.DataFrame.query.html)


This model can be useful for conditional Pipelines. For example, you could have two invocations of this model:

```{literalinclude} ../../../../../samples/models/choice1.yaml 
:language: yaml
```

and

```{literalinclude} ../../../../../samples/models/choice2.yaml 
:language: yaml
```

By including these in a Pipeline as follows we can define conditional routes:

```{literalinclude} ../../../../../samples/pipelines/choice.yaml 
:language: yaml
```

Here the mul10 model will be called if the choice-is-one model succeeds and the add10 model will be called if the choice-is-two model succeeds.

The full notebook can be found [here](../../examples/pandasquery.md)


