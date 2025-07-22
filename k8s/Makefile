CUSTOM_IMAGE_TAG ?= latest
NEW_VERSION ?= 0.0.0
SELDON_MESH_NAMESPACE ?= seldon-mesh

HELM_CRD_BASE := helm-charts/seldon-core-v2-crds/templates
HELM_COMPONENTS_BASE := helm-charts/seldon-core-v2-setup/templates
HELM_RUNTIME_BASE := helm-charts/seldon-core-v2-runtime/templates
HELM_SERVERS_BASE := helm-charts/seldon-core-v2-servers/templates

KUSTOMIZE_VERSION ?= v5.2.1

.PHONY: create
create: create-helm-charts create-helm-charts-experimental merge-configs create-yaml

.PHONY: create-helm-charts
create-helm-charts: kustomize
	sed "s/#TAG_VERSION_PLACEHOLDER#/${CUSTOM_IMAGE_TAG}/g" helm-charts/seldon-core-v2-setup/values.yaml.template > helm-charts/seldon-core-v2-setup/values.yaml
	$(KUSTOMIZE) build kustomize/helm-crds/ > ${HELM_CRD_BASE}/seldon-v2-crds.yaml
	$(KUSTOMIZE) build kustomize/helm-components-ns/ > ${HELM_COMPONENTS_BASE}/_components-statefulsets.tpl
	$(KUSTOMIZE) build kustomize/helm-servers/ > ${HELM_SERVERS_BASE}/_servers-statefulsets.tpl
	sed -i 's/\(.*\)\(imagePullSecrets:\).*/{{- with .Values.imagePullSecrets }}\n\1\2\n\1{{- toYaml . | nindent 8 }}\n{{- end }}/g' ${HELM_COMPONENTS_BASE}/_components-statefulsets.tpl
	sed -i 's/HACK_REMOVE_ME//' ${HELM_COMPONENTS_BASE}/_components-statefulsets.tpl
	# kustomizse has builtin helper fuctions for meta.labels and metadata.annotations, so the field names are changed to escape kustomize's extra processing
	sed 's/labelz:/labels:\n\ \ \ /' ${HELM_SERVERS_BASE}/_servers-statefulsets.tpl > ${HELM_SERVERS_BASE}/._servers-statefulsets.tpl
	sed -i 's/annotationz:/annotations:\n\ \ \ /' ${HELM_SERVERS_BASE}/._servers-statefulsets.tpl
	sed -i 's/HACK_REMOVE_ME//' ${HELM_SERVERS_BASE}/._servers-statefulsets.tpl
	sed -zi 's#\(apiVersion: rbac.authorization.k8s.io/v1\nkind: ClusterRole.*\- get\n\-\-\-\)#{{- if and (not .Values.controller.skipClusterRoleCreation) (or .Values.controller.clusterwide .Values.controller.watchNamespaces) -}}\n\1\n{{- end }}#' ${HELM_COMPONENTS_BASE}/_components-statefulsets.tpl
	sed -zi 's#\(apiVersion: rbac.authorization.k8s.io/v1\nkind: ClusterRoleBinding.*ServiceAccount\n  name: seldon-v2-controller-manager\n  namespace:[^-]\{27\}\n\-\-\-\)#{{- if or .Values.controller.clusterwide .Values.controller.watchNamespaces -}}\n\1\n{{- end }}#' ${HELM_COMPONENTS_BASE}/_components-statefulsets.tpl
	mv ${HELM_SERVERS_BASE}/._servers-statefulsets.tpl ${HELM_SERVERS_BASE}/_servers-statefulsets.tpl

.PHONY: create-helm-charts-experimental
create-helm-charts-experimental: kustomize
	$(KUSTOMIZE) build --load-restrictor LoadRestrictionsNone kustomize-experimental/helm-components-ns/ > ${HELM_COMPONENTS_BASE}/_components-deployments.tpl
	$(KUSTOMIZE) build --load-restrictor LoadRestrictionsNone kustomize-experimental/helm-servers/ > ${HELM_SERVERS_BASE}/_servers-deployments.tpl
	sed -i 's/\(.*\)\(imagePullSecrets:\).*/{{- with .Values.imagePullSecrets }}\n\1\2\n\1{{- toYaml . | nindent 8 }}\n{{- end }}/g' ${HELM_COMPONENTS_BASE}/_components-deployments.tpl
	sed -i 's/HACK_REMOVE_ME//' ${HELM_COMPONENTS_BASE}/_components-deployments.tpl
	sed 's/labelz:/labels:\n\ \ \ /' ${HELM_SERVERS_BASE}/_servers-deployments.tpl > ${HELM_SERVERS_BASE}/._servers-deployments.tpl
	sed -i 's/annotationz:/annotations:\n\ \ \ /' ${HELM_SERVERS_BASE}/._servers-deployments.tpl
	sed -i 's/HACK_REMOVE_ME//' ${HELM_SERVERS_BASE}/._servers-deployments.tpl
	sed -zi 's#\(apiVersion: rbac.authorization.k8s.io/v1\nkind: ClusterRole.*\- get\n\-\-\-\)#{{- if and (not .Values.controller.skipClusterRoleCreation) (or .Values.controller.clusterwide .Values.controller.watchNamespaces) -}}\n\1\n{{- end }}#' ${HELM_COMPONENTS_BASE}/_components-deployments.tpl
	sed -zi 's#\(apiVersion: rbac.authorization.k8s.io/v1\nkind: ClusterRoleBinding.*ServiceAccount\n  name: seldon-v2-controller-manager\n  namespace:[^-]\{27\}\n\-\-\-\)#{{- if or .Values.controller.clusterwide .Values.controller.watchNamespaces -}}\n\1\n{{- end }}#' ${HELM_COMPONENTS_BASE}/_components-deployments.tpl
	mv ${HELM_SERVERS_BASE}/._servers-deployments.tpl ${HELM_SERVERS_BASE}/_servers-deployments.tpl

.PHONY: merge-configs
merge-configs:
	sed -i '1s;^;{{- define "setup.statefulsets" }}\n;' ${HELM_COMPONENTS_BASE}/_components-statefulsets.tpl && echo '{{- end }}' >> ${HELM_COMPONENTS_BASE}/_components-statefulsets.tpl
	sed -i '1s;^;{{- define "setup.deployments" }}\n;' ${HELM_COMPONENTS_BASE}/_components-deployments.tpl && echo '{{- end }}' >> ${HELM_COMPONENTS_BASE}/_components-deployments.tpl
	@{ \
	echo '{{- if .Values.useDeploymentsForServers }}'; \
	echo '{{ include "setup.deployments" .  }}'; \
	echo '{{- else }}'; \
	echo '{{ include "setup.statefulsets" .  }}'; \
	echo '{{- end }}'; \
	} > ${HELM_COMPONENTS_BASE}/seldon-v2-components.yaml
	sed -i '1s;^;{{- define "servers.statefulsets" }}\n;' ${HELM_SERVERS_BASE}/_servers-statefulsets.tpl && echo '{{- end }}' >> ${HELM_SERVERS_BASE}/_servers-statefulsets.tpl
	sed -i '1s;^;{{- define "servers.deployments" }}\n;' ${HELM_SERVERS_BASE}/_servers-deployments.tpl && echo '{{- end }}' >> ${HELM_SERVERS_BASE}/_servers-deployments.tpl
	@{ \
	echo '{{- if .Values.useDeploymentsForServers }}'; \
	echo '{{ include "servers.deployments" .  }}'; \
	echo '{{- else }}'; \
	echo '{{ include "servers.statefulsets" .  }}'; \
	echo '{{- end }}'; \
	} > ${HELM_SERVERS_BASE}/seldon-v2-servers.yaml

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

KUSTOMIZE = $(shell pwd)/bin/kustomize
.PHONY: kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v5@$(KUSTOMIZE_VERSION))

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef
