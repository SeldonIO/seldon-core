CUSTOM_IMAGE_TAG ?= latest
NEW_VERSION ?= 0.0.0
SELDON_MESH_NAMESPACE ?= seldon-mesh

HELM_CRD_BASE := helm-charts/seldon-core-v2-crds/templates
HELM_COMPONENTS_BASE := helm-charts/seldon-core-v2-setup/templates
HELM_RUNTIME_BASE := helm-charts/seldon-core-v2-runtime/templates
HELM_SERVERS_BASE := helm-charts/seldon-core-v2-servers/templates

.PHONY: create
create: create-helm-charts create-yaml

.PHONY: create-helm-charts
create-helm-charts:
	sed "s/#TAG_VERSION_PLACEHOLDER#/${CUSTOM_IMAGE_TAG}/g" helm-charts/seldon-core-v2-setup/values.yaml.template > helm-charts/seldon-core-v2-setup/values.yaml
	kustomize build kustomize/helm-crds/ > ${HELM_CRD_BASE}/seldon-v2-crds.yaml
	kustomize build kustomize/helm-components-ns/ > ${HELM_COMPONENTS_BASE}/seldon-v2-components.yaml
	kustomize build kustomize/helm-servers/ > ${HELM_SERVERS_BASE}/seldon-v2-servers.yaml
	sed -i 's/\(.*\)\(imagePullSecrets:\).*/{{- with .Values.imagePullSecrets }}\n\1\2\n\1{{- toYaml . | nindent 8 }}\n{{- end }}/g' ${HELM_COMPONENTS_BASE}/seldon-v2-components.yaml
	sed -i 's/HACK_REMOVE_ME//' ${HELM_COMPONENTS_BASE}/seldon-v2-components.yaml
	sed 's/labelz:/labels:\n\ \ \ /' ${HELM_SERVERS_BASE}/seldon-v2-servers.yaml > ${HELM_SERVERS_BASE}/.seldon-v2-servers.yaml
	sed -i 's/annotationz:/annotations:\n\ \ \ /' ${HELM_SERVERS_BASE}/.seldon-v2-servers.yaml
	sed -i 's/HACK_REMOVE_ME//' ${HELM_SERVERS_BASE}/.seldon-v2-servers.yaml
	sed -zi 's#\(apiVersion: rbac.authorization.k8s.io/v1\nkind: ClusterRole.*\- get\n\-\-\-\)#{{- if .Values.controller.clusterwide -}}\n\1\n{{- end }}#' ${HELM_COMPONENTS_BASE}/seldon-v2-components.yaml
	sed -zi 's#\(apiVersion: rbac.authorization.k8s.io/v1\nkind: ClusterRoleBinding.*ServiceAccount\n  name: seldon-v2-controller-manager\n  namespace:[^-]\{27\}\n\-\-\-\)#{{- if .Values.controller.clusterwide -}}\n\1\n{{- end }}#' ${HELM_COMPONENTS_BASE}/seldon-v2-components.yaml
	mv ${HELM_SERVERS_BASE}/.seldon-v2-servers.yaml ${HELM_SERVERS_BASE}/seldon-v2-servers.yaml

.PHONY: create-yaml
create-yaml:
	helm template -n ${SELDON_MESH_NAMESPACE} seldon-core-v2-certs ./helm-charts/seldon-core-v2-certs | grep -v "namespace:" > yaml/certs.yaml
	helm template seldon-core-v2-crds ./helm-charts/seldon-core-v2-crds > yaml/crds.yaml
	helm template seldon-core-v2-components ./helm-charts/seldon-core-v2-setup  | grep -v "namespace:" > yaml/components.yaml
	helm template seldon-core-v2-runtime ./helm-charts/seldon-core-v2-runtime  | grep -v "namespace:" > yaml/runtime.yaml
	helm template seldon-core-v2-servers ./helm-charts/seldon-core-v2-servers  | grep -v "namespace:" > yaml/servers.yaml

.PHONY: set-chart-version
set-chart-version:
	sed -i -r 's/(version|appVersion): .*/\1: ${NEW_VERSION}/' helm-charts/seldon-core-v2-certs/Chart.yaml
	sed -i -r 's/(version|appVersion): .*/\1: ${NEW_VERSION}/' helm-charts/seldon-core-v2-crds/Chart.yaml
	sed -i -r 's/(version|appVersion): .*/\1: ${NEW_VERSION}/' helm-charts/seldon-core-v2-setup/Chart.yaml
	sed -i -r 's/(version|appVersion): .*/\1: ${NEW_VERSION}/' helm-charts/seldon-core-v2-runtime/Chart.yaml
	sed -i -r 's/(version|appVersion): .*/\1: ${NEW_VERSION}/' helm-charts/seldon-core-v2-servers/Chart.yaml
