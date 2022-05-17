CUSTOM_IMAGE_TAG ?= latest
DOCKERHUB_USERNAME ?= seldonio

CONTROLLER_IMG ?= ${DOCKERHUB_USERNAME}/seldonv2-controller:${CUSTOM_IMAGE_TAG}
SCHEDULER_IMG ?= ${DOCKERHUB_USERNAME}/seldon-scheduler:${CUSTOM_IMAGE_TAG}
AGENT_IMG ?= ${DOCKERHUB_USERNAME}/seldon-agent:${CUSTOM_IMAGE_TAG}
RCLONE_IMG ?= ${DOCKERHUB_USERNAME}/seldon-rclone:${CUSTOM_IMAGE_TAG}
MODELGATEWAY_IMG ?= ${DOCKERHUB_USERNAME}/seldon-modelgateway:${CUSTOM_IMAGE_TAG}
PIPELINEGATEWAY_IMG ?= ${DOCKERHUB_USERNAME}/seldon-pipelinegateway:${CUSTOM_IMAGE_TAG}
DATAFLOW_IMG ?= ${DOCKERHUB_USERNAME}/seldon-dataflow-engine:${CUSTOM_IMAGE_TAG}
ENVOY_IMG ?= ${DOCKERHUB_USERNAME}/seldon-envoy:${CUSTOM_IMAGE_TAG}

MLSERVER_IMG ?= seldonio/mlserver:1.0.1
TRITON_IMG ?= nvcr.io/nvidia/tritonserver:21.12-py3


.PHONY: create-yaml
create-yaml:
	cd ../scheduler/k8s/scheduler && kustomize edit set image scheduler=${SCHEDULER_IMG}
	cd ../scheduler/k8s/envoy && kustomize edit set image envoy=${ENVOY_IMG}
	cd ../scheduler/k8s/modelgateway && kustomize edit set image modelgateway=${MODELGATEWAY_IMG}
	cd ../scheduler/k8s/pipelinegateway && kustomize edit set image pipelinegateway=${PIPELINEGATEWAY_IMG}
	cd ../scheduler/k8s/dataflow-engine && kustomize edit set image dataflow-engine=${DATAFLOW_IMG}

	cd ../operator/config/manager && kustomize edit set image controller=${CONTROLLER_IMG}
	cd ../operator/config/serverconfigs && kustomize edit set image agent=${AGENT_IMG}
	cd ../operator/config/serverconfigs && kustomize edit set image rclone=${RCLONE_IMG}
	cd ../operator/config/serverconfigs && kustomize edit set image mlserver=${MLSERVER_IMG}
	cd ../operator/config/serverconfigs && kustomize edit set image triton=${TRITON_IMG}

	kustomize build kustomize/crds/ > yaml/seldon-v2-crds.yaml
	kustomize build kustomize/components/ > yaml/seldon-v2-components.yaml
	kustomize build kustomize/servers/ > yaml/seldon-v2-servers.yaml

.PHONY: create-helm-charts
create-helm-charts:
	kustomize build kustomize/helm-crds/ > helm-charts/seldon-core-v2-crds/templates/seldon-v2-crds.yaml
	kustomize build kustomize/helm-components/ > helm-charts/seldon-core-v2-setup/templates/seldon-v2-components.yaml
	kustomize build kustomize/helm-servers/ > helm-charts/seldon-core-v2-setup/templates/seldon-v2-servers.yaml
