# seldon-od-transformer

![Version: 0.2.0](https://img.shields.io/static/v1?label=Version&message=0.2.0&color=informational&style=flat-square)

Chart to deploy an outlier detector as a transformer in an inference graph.

## Usage

To use this chart, you will first need to add the `seldonio` Helm repo:

```bash
helm repo add seldonio https://storage.googleapis.com/seldon-charts
helm repo update
```

Once that's done, you should then be able to use the inference graph template as:

```bash
helm template $MY_MODEL_NAME seldonio/seldon-od-transformer --namespace $MODELS_NAMESPACE
```

Note that you can also deploy the inference graph directly to your cluster
using:

```bash
helm install $MY_MODEL_NAME seldonio/seldon-od-transformer --namespace $MODELS_NAMESPACE
```

**Homepage:** <https://github.com/SeldonIO/seldon-core>

## Source Code

* <https://github.com/SeldonIO/seldon-core>
* <https://github.com/SeldonIO/seldon-core/tree/master/helm-charts/seldon-od-transformer>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| model.image.name | string | `"seldonio/mock_classifier:1.0"` |  |
| model.name | string | `"classifier"` |  |
| name | string | `"seldon-od-transformer"` |  |
| outlierDetection.enabled | bool | `true` |  |
| outlierDetection.isolationforest.image.name | string | `"seldonio/outlier-if-tranformer:0.1"` |  |
| outlierDetection.isolationforest.load_path | string | `"./models/"` |  |
| outlierDetection.isolationforest.model_name | string | `"if"` |  |
| outlierDetection.isolationforest.threshold | int | `0` |  |
| outlierDetection.mahalanobis.image.name | string | `"seldonio/outlier-mahalanobis-tranformer:0.1"` |  |
| outlierDetection.mahalanobis.max_n | int | `-1` |  |
| outlierDetection.mahalanobis.n_components | int | `3` |  |
| outlierDetection.mahalanobis.n_stdev | int | `3` |  |
| outlierDetection.mahalanobis.start_clip | int | `50` |  |
| outlierDetection.mahalanobis.threshold | int | `25` |  |
| outlierDetection.name | string | `"outlier-detector"` |  |
| outlierDetection.parameterTypes.load_path | string | `"STRING"` |  |
| outlierDetection.parameterTypes.max_n | string | `"INT"` |  |
| outlierDetection.parameterTypes.model_name | string | `"STRING"` |  |
| outlierDetection.parameterTypes.n_components | string | `"INT"` |  |
| outlierDetection.parameterTypes.n_stdev | string | `"FLOAT"` |  |
| outlierDetection.parameterTypes.reservoir_size | string | `"INT"` |  |
| outlierDetection.parameterTypes.start_clip | string | `"INT"` |  |
| outlierDetection.parameterTypes.threshold | string | `"FLOAT"` |  |
| outlierDetection.seq2seq.image.name | string | `"seldonio/outlier-s2s-lstm-tranformer:0.1"` |  |
| outlierDetection.seq2seq.load_path | string | `"./models/"` |  |
| outlierDetection.seq2seq.model_name | string | `"seq2seq"` |  |
| outlierDetection.seq2seq.reservoir_size | int | `50000` |  |
| outlierDetection.seq2seq.threshold | float | `0.003` |  |
| outlierDetection.type | string | `"vae"` | Type of outlier detector. Valid values are: `vae`, `mahalanobis`, `seq2seq` and `isolationforest`. |
| outlierDetection.vae.image.name | string | `"seldonio/outlier-vae-tranformer:0.1"` |  |
| outlierDetection.vae.load_path | string | `"./models/"` |  |
| outlierDetection.vae.model_name | string | `"vae"` |  |
| outlierDetection.vae.reservoir_size | int | `50000` |  |
| outlierDetection.vae.threshold | int | `10` |  |
| predictorLabels.fluentd | string | `"true"` |  |
| predictorLabels.version | string | `"v1"` |  |
| replicas | int | `1` |  |
| sdepLabels.app | string | `"seldon"` |  |
