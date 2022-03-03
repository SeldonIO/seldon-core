PROTOC_IMPORT_PATH := -I.

PROTOC_GO_PATH := --go_opt=paths=source_relative
PROTOC_GO_OUT := --go_out=.
PROTOC_GO_GRPC_PATH := --go-grpc_opt=paths=source_relative
PROTOC_GO_GRPC_OUT := --go-grpc_out=.
PROTOC_GO_OPTIONS = $(PROTOC_GO_PATH) $(PROTOC_GO_OUT) $(PROTOC_GO_GRPC_PATH) $(PROTOC_GO_GRPC_OUT)

.PHONY: build
build:
	protoc \
		$(PROTOC_IMPORT_PATH) \
		$(PROTOC_GO_OPTIONS) \
		./mlops/agent/agent.proto \
		./mlops/agent_debug/agent_debug.proto \
		./mlops/proxy/proxy.proto \
		./mlops/scheduler/scheduler.proto \
		./mlops/v2_dataplane/v2_dataplane.proto


PROTOC_PYTHON_OUT := --python_out=./python
PROTOC_PYTHON_OPTIONS = $(PROTOC_PYTHON_OUT)

.PHONY: build-v2-python
build-v2-python:
	protoc \
		$(PROTOC_IMPORT_PATH) \
		$(PROTOC_PYTHON_OPTIONS) \
		./mlops/v2_dataplane/v2_dataplane.proto
