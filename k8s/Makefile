RELEASE_VERSION ?= latest

CONTROLLER_IMG ?= seldonio/seldonv2-controller:${RELEASE_VERSION}
SCHEDULER_IMG ?= seldonio/seldon-scheduler:${RELEASE_VERSION}
AGENT_IMG ?= seldonio/seldon-agent:${RELEASE_VERSION}
RCLONE_IMG ?= seldonio/seldon-rclone:${RELEASE_VERSION}
MODELGATEWAY_IMG ?= seldonio/seldon-modelgateway:${RELEASE_VERSION}
PIPELINEGATEWAY_IMG ?= seldonio/seldon-pipelinegateway:${RELEASE_VERSION}
DATAFLOW_IMG ?= seldonio/seldon-dataflow-engine:${RELEASE_VERSION}
ENVOY_IMG ?= seldonio/seldon-envoy:${RELEASE_VERSION}

MLSERVER_IMG ?= seldonio/mlserver:1.0.1
TRITON_IMG ?= nvcr.io/nvidia/tritonserver:21.12-py3



create:
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

	kustomize build crds/ > seldon-v2-crds.yaml
	kustomize build components/ > seldon-v2-components.yaml
	kustomize build servers/ > seldon-v2-servers.yaml
