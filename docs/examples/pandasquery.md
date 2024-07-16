# Conditional Pipeline with Pandas Query Model

The model is defined as an MLServer custom runtime and allows the user to pass in a custom pandas query as a parameter defined at model creation to be used to filter the data passed to the model.

```{literalinclude} ../../../../samples/examples/pandasquery/pandasquery/model.py
:language: python
```


```{include} ../../../../samples/examples/pandasquery/infer.md
```
