# Parameterized Models

The Model specification allows parameters to be passed to the loaded model to allow customization. For example:

```{literalinclude} ../../../../../samples/models/choice1.yaml 
:language: yaml
```

This capability is only available for MLServer custom model runtimes. The named keys and values will be added to the model-settings.json file for the provided model in the
`parameters.extra` Dict. MLServer models are able to read these values in their `load` method.

## Example Parameterized Models

 * [Pandas Query](./pandasquery.md)

```{toctree}
:maxdepth: 1
:hidden:

pandasquery.md
```
