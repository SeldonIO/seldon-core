# Seldon Core Helm Charts

Helm charts are published to our official repo.

## Core Charts

The core charts for installing Seldon Core are shown below.

 * [seldon-core-operator](https://docs.seldon.io/projects/seldon-core/en/latest/charts/seldon-core-operator.html)
   * Main helm chart for installing Seldon Core CRD and Controller
 * [seldon-core-analytics](https://docs.seldon.io/projects/seldon-core/en/latest/charts/seldon-core-analytics.html)
   * Example Prometheus and Grafana setup with demonstration Grafana dashboard for Seldon Core


## Seldon Core Inference Graph Templates

A set of charts to provide example templates for creating particular inference graphs using Seldon Core

 * [seldon-single-model](https://docs.seldon.io/projects/seldon-core/en/latest/charts/seldon-single-model.html)
   * Serve a single model with attached Persistent Volume.
 * [seldon-abtest](https://docs.seldon.io/projects/seldon-core/en/latest/charts/seldon-abtest.html)
   * Serve an AB test between two models.
 * [seldon-mab](https://docs.seldon.io/projects/seldon-core/en/latest/charts/seldon-mab.html)
   * Serve a multi-armed bandit between two models.
 * [seldon-od-model](https://docs.seldon.io/projects/seldon-core/en/latest/charts/seldon-od-model.html) and [seldon-od-transformer](https://docs.seldon.io/projects/seldon-core/en/latest/charts/seldon-od-transformer.html)
   * Serve one of the following Outlier Detector components as either models or transformers:
     * [Isolation Forest](https://github.com/SeldonIO/seldon-core/tree/master/components/outlier-detection/isolation-forest)
     * [Variational Auto-Encoder](https://github.com/SeldonIO/seldon-core/tree/master/components/outlier-detection/vae)
     * [Sequence-to-Sequence-LSTM](https://github.com/SeldonIO/seldon-core/tree/master/components/outlier-detection/seq2seq-lstm)
     * [Mahalanobis Distance](https://github.com/SeldonIO/seldon-core/tree/master/components/outlier-detection/mahalanobis)

For examples of using some of the above charts see [here](https://github.com/SeldonIO/seldon-core/tree/master/notebooks/helm_examples.ipynb).

## Misc

 * [seldon-core-loadtesting](https://docs.seldon.io/projects/seldon-core/en/latest/charts/seldon-core-loadtesting.html)
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
