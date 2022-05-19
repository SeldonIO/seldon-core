# CLI

Seldon provides a CLI to allow easy management and testing of Model, Experiment, Pipeline and Explainer resources.

At present this needs to be built by hand from the operator folder.

```
make build-seldon
```

Then place the `bin/seldon` executable in your path.

 * [cli docs](./docs/seldon.md)

```{toctree}
:maxdepth: 1
:hidden:

docs/seldon.md
docs/seldon_model.md
docs/seldon_experiment.md
docs/seldon_pipeline.md
docs/seldon_server.md
docs/seldon_model_infer.md
docs/seldon_model_load.md
docs/seldon_model_status.md
docs/seldon_model_unload.md
docs/seldon_experiment_start.md
docs/seldon_experiment_status.md
docs/seldon_experiment_stop.md
docs/seldon_pipeline_infer.md
docs/seldon_pipeline_load.md
docs/seldon_pipeline_status.md
docs/seldon_pipeline_unload.md
docs/seldon_pipeline_inspect.md
docs/seldon_server_status.md
```

