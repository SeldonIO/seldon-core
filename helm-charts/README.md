## Seldon Core Setup

Helm charts are published to our official repo. An example install:

The core charts for installing Seldon Core are shown below:

 * [seldon-core-controller](https://github.com/SeldonIO/seldon-core/tree/master/helm-charts/seldon-core-controller)
   * Main helm chart for installing Seldon Core CRD and Controller
 * [seldon-core-oauth-gateway](https://github.com/SeldonIO/seldon-core/tree/master/helm-charts/seldon-core-oauth-gateway)
   * Seldon OAuth Gateway
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
