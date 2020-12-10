{{ template "chart.header" . }}
{{ template "chart.deprecationWarning" . }}

![Version: {{ .Version }}](https://img.shields.io/static/v1?label=Version&message={{ .Version | replace "-" "--" }}&color=informational&style=flat-square) 

{{ template "chart.description" . }}

## Usage

To use this chart, you will first need to add the `seldonio` Helm repo:

```bash
helm repo add seldonio https://storage.googleapis.com/seldon-charts
helm repo update
```

{{- /* We need to differentiate between "app charts" and "inference graph charts" */ -}}
{{- $appCharts := list "seldon-core-operator" "seldon-core-analytics" "seldon-core-loadtesting" -}}
{{ if has .Name $appCharts }}

Onca that's done, you should then be able to deploy the chart as:

```bash
kubectl create namespace seldon-system
helm install {{ .Name }} seldonio/{{ .Name }} --namespace seldon-system
```

{{ else }}

Once that's done, you should then be able to use the inference graph template as:

```bash
helm template $MY_MODEL_NAME seldonio/{{ .Name }} --namespace $MODELS_NAMESPACE
```

Note that you can also deploy the inference graph directly to your cluster
using:

```bash
helm install $MY_MODEL_NAME seldonio/{{ .Name }} --namespace $MODELS_NAMESPACE
```

{{ end -}}

{{ template "chart.homepageLine" . }}

{{ template "chart.maintainersSection" . }}

{{ template "chart.sourcesSection" . }}

{{ template "chart.requirementsSection" . }}

{{ template "chart.valuesSection" . }}
