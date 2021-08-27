KIND_IMAGE ?= kindest/node:v1.18.15
KIND_NAME ?= ansible

AMBASSADOR_PORT ?= 8003
ISTIO_PORT ?= 8004
MINIO_PORT ?= 8090


# Kind Cluster
kind-cluster:
	kind create cluster --name ${KIND_NAME} --image ${KIND_IMAGE} || echo "already there"
	kind export kubeconfig --name ${KIND_NAME}
	kubectl create ns seldon || echo "already there"
	kubectl config set-context --current --namespace seldon

kind-delete:
	kind delete cluster --name ${KIND_NAME}

# Port Forwards
port-forward-ambassador:
	kubectl port-forward -n ambassador svc/ambassador ${AMBASSADOR_PORT}:80

port-forward-istio:
	kubectl port-forward  -n istio-system svc/istio-ingressgateway ${ISTIO_PORT}:80

port-forward-minio:
	kubectl port-forward -n minio-system svc/minio ${MINIO_PORT}:9000
