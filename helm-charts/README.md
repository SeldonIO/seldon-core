# Seldon Core Helm Charts

Helm charts are published to our official repo.

## Core Charts

The core chart for installing Seldon Core is shown below.

 * [seldon-core-operator](https://docs.seldon.io/projects/seldon-core/en/latest/charts/seldon-core-operator.html)
   * Main helm chart for installing Seldon Core CRD and Controller


## Seldon Core Inference Graph Templates

A set of charts to provide example templates for creating particular inference graphs using Seldon Core

 * [seldon-single-model](./seldon-single-model/README.md)
   * Serve a single model with attached Persistent Volume.
 * [seldon-abtest](./seldon-abtest/README.md)
   * Serve an AB test between two models.
 * [seldon-mab](./seldon-mab/README.md)
   * Serve a multi-armed bandit between two models.
 * [seldon-od-model](./seldon-od-model/README.md) and [seldon-od-transformer](./seldon-od-transformer/README.md)

For examples of using some of the above charts see [here](https://github.com/SeldonIO/seldon-core/tree/master/notebooks/helm_examples.ipynb).

## Misc

 * [seldon-core-loadtesting](./seldon-core-loadtesting/README.md)
   * Utility to load test

## Documentation

To generate the documentation of our Helm charts, we use
[`helm-docs`](https://github.com/norwoodj/helm-docs).
This tool will read the metadata included in the `Chart.yaml` and `values.yaml`
to generate a `README.md` page.

### Generating documentation locally

You can install the latest version of `helm-docs` using the `install` target of
the `Makefile`:

```shell
make install
```

Afterwards, you can use the `docs` target of the `Makefile`:

```shell
make docs
```
