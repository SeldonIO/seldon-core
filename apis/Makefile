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
		./mlops/scheduler/storage.proto \
		./mlops/chainer/chainer.proto \
		./mlops/v2_dataplane/v2_dataplane.proto


PROTOC_PYTHON_OUT := --python_out=./python
PROTOC_PYTHON_OPTIONS = $(PROTOC_PYTHON_OUT)

.PHONY: build-v2-python
build-v2-python:
	protoc \
		$(PROTOC_IMPORT_PATH) \
		$(PROTOC_PYTHON_OPTIONS) \
		./mlops/v2_dataplane/v2_dataplane.proto

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
