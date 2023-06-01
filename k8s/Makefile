CUSTOM_IMAGE_TAG ?= latest
NEW_VERSION ?= 0.0.0

HELM_CRD_BASE := helm-charts/seldon-core-v2-crds/templates
HELM_COMPONENTS_BASE := helm-charts/seldon-core-v2-setup/templates
HELM_SERVERS_BASE := helm-charts/seldon-core-v2-runtime/templates

.PHONY: create
create: create-helm-charts create-yaml

.PHONY: create-helm-charts
create-helm-charts:
	sed "s/#TAG_VERSION_PLACEHOLDER#/${CUSTOM_IMAGE_TAG}/g" helm-charts/seldon-core-v2-setup/values.yaml.template > helm-charts/seldon-core-v2-setup/values.yaml
	kustomize build kustomize/helm-crds/ > ${HELM_CRD_BASE}/seldon-v2-crds.yaml
	kustomize build kustomize/helm-components-ns/ > ${HELM_COMPONENTS_BASE}/seldon-v2-components.yaml
	kustomize build kustomize/helm-servers/ > ${HELM_SERVERS_BASE}/seldon-v2-servers.yaml
	sed -i 's/\(.*\)\(imagePullSecrets:\).*/{{- with .Values.imagePullSecrets }}\n\1\2\n\1{{- toYaml . | nindent 8 }}\n{{- end }}/g' ${HELM_COMPONENTS_BASE}/seldon-v2-components.yaml
	sed -n --file=add-helm-toggle-hodometer.sed ${HELM_COMPONENTS_BASE}/seldon-v2-components.yaml \
		> ${HELM_COMPONENTS_BASE}/.components.yaml
	sed 's/HACK_REMOVE_ME//' ${HELM_COMPONENTS_BASE}/.components.yaml \
		> ${HELM_COMPONENTS_BASE}/seldon-v2-components.yaml
	rm ${HELM_COMPONENTS_BASE}/.components.yaml
	sed 's/HACK_REMOVE_ME//' ${HELM_SERVERS_BASE}/seldon-v2-servers.yaml \
		> ${HELM_SERVERS_BASE}/.seldon-v2-servers.yaml
	mv ${HELM_SERVERS_BASE}/.seldon-v2-servers.yaml ${HELM_SERVERS_BASE}/seldon-v2-servers.yaml

.PHONY: create-yaml
create-yaml:
	helm template seldon-core-v2-certs ./helm-charts/seldon-core-v2-certs > yaml/certs.yaml
	helm template seldon-core-v2-crds ./helm-charts/seldon-core-v2-crds > yaml/crds.yaml
	helm template seldon-core-v2-components ./helm-charts/seldon-core-v2-setup > yaml/components.yaml
	helm template seldon-core-v2-servers ./helm-charts/seldon-core-v2-runtime > yaml/servers.yaml

.PHONY: set-chart-version
set-chart-version:
	sed -i -r 's/(version|appVersion): .*/\1: ${NEW_VERSION}/' helm-charts/seldon-core-v2-certs/Chart.yaml
	sed -i -r 's/(version|appVersion): .*/\1: ${NEW_VERSION}/' helm-charts/seldon-core-v2-crds/Chart.yaml
	sed -i -r 's/(version|appVersion): .*/\1: ${NEW_VERSION}/' helm-charts/seldon-core-v2-setup/Chart.yaml
	sed -i -r 's/(version|appVersion): .*/\1: ${NEW_VERSION}/' helm-charts/seldon-core-v2-servers/Chart.yaml
