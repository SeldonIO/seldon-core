.PHONY: dataplane.yaml

MLSERVER_RELEASE := master
MLSERVER_BASE := https://raw.githubusercontent.com/SeldonIO/MLServer/${MLSERVER_RELEASE}

dataplane.yaml:
	curl -s -L ${MLSERVER_BASE}/openapi/dataplane.yaml > ./dataplane.yaml
