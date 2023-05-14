CUSTOM_IMAGE_TAG ?= latest
DOCKERHUB_USERNAME ?= seldonio
NEW_VERSION ?= 0.0.0

AGENT_IMG ?= ${DOCKERHUB_USERNAME}/seldon-agent:${CUSTOM_IMAGE_TAG}
CONTROLLER_IMG ?= ${DOCKERHUB_USERNAME}/seldonv2-controller:${CUSTOM_IMAGE_TAG}
DATAFLOW_IMG ?= ${DOCKERHUB_USERNAME}/seldon-dataflow-engine:${CUSTOM_IMAGE_TAG}
ENVOY_IMG ?= ${DOCKERHUB_USERNAME}/seldon-envoy:${CUSTOM_IMAGE_TAG}
HODOMETER_IMG ?= ${DOCKERHUB_USERNAME}/seldon-hodometer:${CUSTOM_IMAGE_TAG}
MODELGATEWAY_IMG ?= ${DOCKERHUB_USERNAME}/seldon-modelgateway:${CUSTOM_IMAGE_TAG}
PIPELINEGATEWAY_IMG ?= ${DOCKERHUB_USERNAME}/seldon-pipelinegateway:${CUSTOM_IMAGE_TAG}
RCLONE_IMG ?= ${DOCKERHUB_USERNAME}/seldon-rclone:${CUSTOM_IMAGE_TAG}
SCHEDULER_IMG ?= ${DOCKERHUB_USERNAME}/seldon-scheduler:${CUSTOM_IMAGE_TAG}

MLSERVER_IMG ?= seldonio/mlserver:1.3.2
TRITON_IMG ?= nvcr.io/nvidia/tritonserver:23.03-py3

.PHONY: create
create: create-helm-charts create-yaml

.PHONY: create-yaml
create-yaml:
	helm template seldon-core-v2-certs ./helm-charts/seldon-core-v2-certs > yaml/certs.yaml
	helm template seldon-core-v2-crds ./helm-charts/seldon-core-v2-crds > yaml/crds.yaml
	helm template seldon-core-v2-components ./helm-charts/seldon-core-v2-setup > yaml/components.yaml
	helm template seldon-core-v2-servers ./helm-charts/seldon-core-v2-servers > yaml/servers.yaml

HELM_CRD_BASE := helm-charts/seldon-core-v2-crds/templates
HELM_COMPONENTS_BASE := helm-charts/seldon-core-v2-setup/templates
HELM_SERVERS_BASE := helm-charts/seldon-core-v2-servers/templates

.PHONY: create-helm-charts
create-helm-charts:
	sed "s/#TAG_VERSION_PLACEHOLDER#/${CUSTOM_IMAGE_TAG}/g" helm-charts/seldon-core-v2-setup/values.yaml.template > helm-charts/seldon-core-v2-setup/values.yaml
	kustomize build kustomize/helm-crds/ > ${HELM_CRD_BASE}/seldon-v2-crds.yaml
	kustomize build kustomize/helm-components-ns/ > ${HELM_COMPONENTS_BASE}/seldon-v2-components.yaml
	kustomize build kustomize/helm-servers/ > ${HELM_SERVERS_BASE}/seldon-v2-servers.yaml
	sed -n --file=add-helm-toggle-hodometer.sed ${HELM_COMPONENTS_BASE}/seldon-v2-components.yaml \
		> ${HELM_COMPONENTS_BASE}/.components.yaml
	sed 's/HACK_REMOVE_ME//' ${HELM_COMPONENTS_BASE}/.components.yaml \
		> ${HELM_COMPONENTS_BASE}/seldon-v2-components.yaml
	rm ${HELM_COMPONENTS_BASE}/.components.yaml
	sed 's/HACK_REMOVE_ME//' ${HELM_SERVERS_BASE}/seldon-v2-servers.yaml \
		> ${HELM_SERVERS_BASE}/.seldon-v2-servers.yaml
	mv ${HELM_SERVERS_BASE}/.seldon-v2-servers.yaml ${HELM_SERVERS_BASE}/seldon-v2-servers.yaml

.PHONY: set-chart-version
set-chart-version:
	sed -i "s/version: .*/version: ${NEW_VERSION}/" helm-charts/seldon-core-v2-certs/Chart.yaml
	sed -i "s/appVersion: .*/appVersion: ${NEW_VERSION}/" helm-charts/seldon-core-v2-certs/Chart.yaml
	sed -i "s/version: .*/version: ${NEW_VERSION}/" helm-charts/seldon-core-v2-crds/Chart.yaml
	sed -i "s/appVersion: .*/appVersion: ${NEW_VERSION}/" helm-charts/seldon-core-v2-crds/Chart.yaml
	sed -i "s/version: .*/version: ${NEW_VERSION}/" helm-charts/seldon-core-v2-setup/Chart.yaml
	sed -i "s/appVersion: .*/appVersion: ${NEW_VERSION}/" helm-charts/seldon-core-v2-setup/Chart.yaml
	sed -i "s/version: .*/version: ${NEW_VERSION}/" helm-charts/seldon-core-v2-servers/Chart.yaml
	sed -i "s/appVersion: .*/appVersion: ${NEW_VERSION}/" helm-charts/seldon-core-v2-servers/Chart.yaml
