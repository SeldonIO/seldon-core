.PHONY: build
build: build-go build-kotlin build-v2-python

PROTOC_IMPORT_PATH := -I.

################################################################################

### START GO ###

PROTOC_GO_PACKAGE_STYLE := import
PROTOC_GO_MODULE_PREFIX := github.com/seldonio/seldon-core/apis/go/v2
PROTOC_GO_OUTPUT_PREFIX := ./go

PROTOC_GO_PATH := --go_opt=paths=$(PROTOC_GO_PACKAGE_STYLE)
PROTOC_GO_MODULE := --go_opt=module=$(PROTOC_GO_MODULE_PREFIX)
PROTOC_GO_OUT := --go_out=$(PROTOC_GO_OUTPUT_PREFIX)
PROTOC_GO_GRPC_PATH := --go-grpc_opt=paths=$(PROTOC_GO_PACKAGE_STYLE)
PROTOC_GO_GRPC_MODULE := --go-grpc_opt=module=$(PROTOC_GO_MODULE_PREFIX)
PROTOC_GO_GRPC_OUT := --go-grpc_out=$(PROTOC_GO_OUTPUT_PREFIX)
PROTOC_GO_OPTIONS = $(PROTOC_GO_PATH) $(PROTOC_GO_MODULE) $(PROTOC_GO_OUT) $(PROTOC_GO_GRPC_PATH) $(PROTOC_GO_GRPC_MODULE) $(PROTOC_GO_GRPC_OUT)

.PHONY: build-go
.ONESHELL: build-go
build-go:
	mkdir -p $(PROTOC_GO_OUTPUT_PREFIX)
	protoc \
		$(PROTOC_IMPORT_PATH) \
		$(PROTOC_GO_OPTIONS) \
		./mlops/agent/agent.proto \
		./mlops/agent_debug/agent_debug.proto \
		./mlops/proxy/proxy.proto \
		./mlops/scheduler/scheduler.proto \
		./mlops/scheduler/storage.proto \
		./mlops/chainer/chainer.proto \
		./mlops/v2_dataplane/v2_dataplane.proto

.PHONY: clean-go
clean-go:
	rm -r ./go/mlops

################################################################################

### START PYTHON ###

PROTOC_PYTHON_OUT := --python_out=./python --grpc_python_out=./python 
PROTOC_PYTHON_OPTIONS = $(PROTOC_PYTHON_OUT)

.PHONY: python-deps
python-deps:
	pip install grpcio-tools -U

.PHONY: build-v2-python
build-v2-python: python-deps
	mkdir -p python
	python -m grpc_tools.protoc \
		$(PROTOC_IMPORT_PATH) \
		$(PROTOC_PYTHON_OPTIONS) \
		--pyi_out=./python \
		./mlops/v2_dataplane/v2_dataplane.proto

################################################################################

### START JVM ###

.PHONY: download-java-protoc
download-java-protoc:
	wget https://repo1.maven.org/maven2/io/grpc/protoc-gen-grpc-java/1.45.1/protoc-gen-grpc-java-1.45.1-linux-x86_64.exe
	chmod u+x protoc-gen-grpc-java-1.45.1-linux-x86_64.exe
	mv protoc-gen-grpc-java-1.45.1-linux-x86_64.exe kotlin

.PHONY: download-kotlin-protoc
download-kotlin-protoc:
	wget https://repo1.maven.org/maven2/io/grpc/protoc-gen-grpc-kotlin/1.2.1/protoc-gen-grpc-kotlin-1.2.1-jdk7.jar
	mv protoc-gen-grpc-kotlin-1.2.1-jdk7.jar kotlin

.PHONY: download-jvm-protoc
download-jvm-protoc: download-kotlin-protoc download-java-protoc

.PHONY: build-kotlin
build-kotlin:
	cd ./mlops/chainer && protoc \
		--proto_path=. \
		--plugin=protoc-gen-grpc-java=../../kotlin/protoc-gen-grpc-java-1.45.1-linux-x86_64.exe \
		--java_out=./kotlin \
		--grpc-java_out=./kotlin \
		--plugin=protoc-gen-grpckt=../../kotlin/protoc-gen-grpc-kotlin.sh \
		--kotlin_out=./kotlin \
		--grpckt_out=./kotlin \
		chainer.proto
