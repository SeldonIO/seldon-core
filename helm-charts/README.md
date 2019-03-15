## Seldon Core Setup

Helm charts are published to our official repo. An example install:

```bash
helm install seldon-core-crd --name seldon-core-crd --repo https://storage.googleapis.com/seldon-charts \
     --set usage_metrics.enabled=true
```

The core charts for installing Seldon Core.

 * [seldon-core-crd](https://github.com/SeldonIO/seldon-core/tree/master/helm-charts/seldon-core-crd)
   * SeldonDeployment Custom Resource Definition and Spartakus usage metrics
 * [seldon-core](https://github.com/SeldonIO/seldon-core/tree/master/helm-charts/seldon-core)
   * Main helm chart for installing Seldon Core
 * [seldon-core-analytics](https://github.com/SeldonIO/seldon-core/tree/master/helm-charts/seldon-core-analytics)
   * Example Prometheus and Grafana setup with demonstration Grafana dashboard for Seldon Core


## Seldon Core Inference Graph Templates

A set of charts to provide example templates for creating particular inference graphs using Seldon Core

 * [seldon-single-model](https://github.com/SeldonIO/seldon-core/tree/master/helm-charts/seldon-single-model)
   * Serve a single model with attached Persistent Volume.
 * [seldon-abtest](https://github.com/SeldonIO/seldon-core/tree/master/helm-charts/seldon-abtest)
   * Serve an AB test between two models.
 * [seldon-mab](https://github.com/SeldonIO/seldon-core/tree/master/helm-charts/seldon-mab)
   * Serve a multi-armed bandit between two models.
 * [seldon-openvino](https://github.com/SeldonIO/seldon-core/tree/master/helm-charts/seldon-openvino)
   * Deploy a single model with Intel OpenVINO model server.
 * [seldon-od-model](https://github.com/SeldonIO/seldon-core/tree/master/helm-charts/seldon-od-model) and [seldon-od-transformer](https://github.com/SeldonIO/seldon-core/tree/master/helm-charts/seldon-od-transformer)
   * Serve one of the following Outlier Detector components as either models or transformers:
     * [Isolation Forest](https://github.com/SeldonIO/seldon-core/tree/master/components/outlier-detection/isolation-forest)
     * [Variational Auto-Encoder](https://github.com/SeldonIO/seldon-core/tree/master/components/outlier-detection/vae)
     * [Sequence-to-Sequence-LSTM](https://github.com/SeldonIO/seldon-core/tree/master/components/outlier-detection/seq2seq-lstm)
     * [Mahalanobis Distance](https://github.com/SeldonIO/seldon-core/tree/master/components/outlier-detection/mahalanobis)

For examples of using some of the above charts see [here](https://github.com/SeldonIO/seldon-core/tree/master/notebooks/helm_examples.ipynb).

## Misc

 * [seldon-core-loadtesting](https://github.com/SeldonIO/seldon-core/tree/master/helm-charts/seldon-core-loadtesting)
   * Utility to load test
