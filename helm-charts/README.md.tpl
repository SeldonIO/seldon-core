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

Once that's done, you should be able to generate your `SeldonDeployment`
resources as:

```shell
helm template $MY_MODEL_NAME seldonio/{{ .Name }} --namespace $MODELS_NAMESPACE
```

Note that you can also install / deploy the chart directly to your cluster using:

```shell
helm install $MY_MODEL_NAME seldonio/{{ .Name }} --namespace $MODELS_NAMESPACE
```

{{ template "chart.homepageLine" . }}

{{ template "chart.maintainersSection" . }}

{{ template "chart.sourcesSection" . }}

{{ template "chart.requirementsSection" . }}

{{ template "chart.valuesSection" . }}
