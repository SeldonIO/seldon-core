{{ template "chart.header" . }}
{{ template "chart.deprecationWarning" . }}

{{ template "chart.versionBadge" . }}{{ template "chart.typeBadge" . }}{{ template "chart.appVersionBadge" . }}

{{ template "chart.description" . }}

## Usage

To use this chart, you will first need to add the `seldonio` Helm repo:

```shell
helm repo add seldonio https://storage.googleapis.com/seldon-charts
helm repo update
```

{{- /* We need to differentiate between "app charts" and "inference graph charts" */ -}}
{{- $appCharts := list "seldon-core-operator" "seldon-core-analytics" -}}
{{ if has .Name $appCharts }}

You can now deploy the chart as:

```shell
kubectl create namespace seldon-system
helm install {{ .Name }} seldonio/{{ .Name }} --namespace seldon-system
```

{{ else }}

Once that's done, you should be able to use the inference graph template as:

```shell
helm template $MY_MODEL_NAME seldonio/{{ .Name }} --namespace $MODELS_NAMESPACE
```

Note that you can also deploy the inference graph directly to your cluster
using:

```shell
helm install $MY_MODEL_NAME seldonio/{{ .Name }} --namespace $MODELS_NAMESPACE
```

{{ end -}}

{{ template "chart.homepageLine" . }}

{{ template "chart.maintainersSection" . }}

{{ template "chart.sourcesSection" . }}

{{ template "chart.requirementsSection" . }}

{{ template "chart.valuesSection" . }}
