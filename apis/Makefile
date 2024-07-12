# Description: Makefile for generating code from protobuf files for Seldon Core
#
#	By default, this Makefile will download the protoc compiler and any required
#	language-specific plugins locally, in the `tools` directory. The versions of
#	the downloaded tools are pinned to what is specified in the `VERSION SELECTION`
#	section below.
#
# Overwrite the PROTOC_GEN_*_VERSION variables in that section to use different
# releases.
#
# If you want to use versions of protoc and/or of the language-specific plugins
# already installed on your system, you can do so by overwriting the
# CMD_PROTOC* variables. This is not recommended because it offers significantly
# fewer guarantees about the compatibility of the generated code across the
# codebase
#

.PHONY: build
build: build-go build-kotlin build-v2-python update-copyright

### VERSION SELECTION ###

# Links in comments below point to the corresponding projects, for determining
# the latest version of protoc + language-specific plugins.

# https://github.com/protocolbuffers/protobuf/releases
PROTOC_VERSION ?= 21.10

# https://pkg.go.dev/google.golang.org/protobuf/cmd/protoc-gen-go
PROTOC_GEN_GO_VERSION ?= 1.28.1

# https://pkg.go.dev/google.golang.org/grpc/cmd/protoc-gen-go-grpc
PROTOC_GEN_GO_GRPC_VERSION ?= 1.2.0

# https://pypi.org/project/grpcio-tools/
PROTOC_GEN_PYTHON_GRPCIO_TOOLS_VERSION ?= 1.51.3

# https://github.com/grpc/grpc-java
# https://mvnrepository.com/artifact/io.grpc/protoc-gen-grpc-java
PROTOC_GEN_JAVA_GRPC_VERSION ?= 1.45.1

# https://github.com/grpc/grpc-kotlin
# https://mvnrepository.com/artifact/io.grpc/protoc-gen-grpc-kotlin
PROTOC_GEN_KOTLIN_GRPC_VERSION ?= 1.2.1
PROTOC_GEN_KOTLIN_GRPC_JDK ?= jdk7


################################################################################

MAKEFILE := $(lastword $(MAKEFILE_LIST))
BASE_DIR := $(realpath $(dir $(MAKEFILE)))
TOOLS_DIR := $(BASE_DIR)/tools

CMD_WGET   	?= wget
CMD_UNZIP 	?= unzip
CMD_GO			?= go
CMD_PYTHON	?= python3

PROTOC_IMPORT_PATH := -I.
PROTOBUF_RELEASE_URL ?= https://github.com/protocolbuffers/protobuf/releases
ARCH ?= $(shell uname -m)

# Check existence of required cmdline tools
$(TOOLS_DIR)/.check_%: \
	| $(TOOLS_DIR)/./
#
	@command -v $* > /dev/null; \
	if [ $$? -ne 0 ]; then \
		echo "[WARN] Missing required tool $(notdir $*); trying to install it"; \
		$(MAKE) --no-print-directory $*; \
		if [ $$? -ne 0 ]; then \
			echo "[ERROR] $(notdir $*) could not be installed"; \
			exit 1; \
		else \
			echo "[INFO] $(notdir $*) installed successfully"; \
			touch $(TOOLS_DIR)/.check_$(notdir $*); \
		fi \
	else \
		touch $(TOOLS_DIR)/.check_$(notdir $*); \
	fi

# Create tool directories when requested
$(TOOLS_DIR)/%/:
	@mkdir -p $@

################################################################################

### START PROTOC ###

CMD_PROTOC 	?= $(TOOLS_DIR)/bin/protoc-$(PROTOC_VERSION)
PROTOC_DOWNLOAD_URL ?= $(PROTOBUF_RELEASE_URL)/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-linux-$(ARCH).zip

# Download protoc
$(TOOLS_DIR)/bin/protoc-$(PROTOC_VERSION): \
	| $(TOOLS_DIR)/.check_$(CMD_WGET) \
		$(TOOLS_DIR)/.check_$(CMD_UNZIP)
#
	@$(CMD_WGET) -q $(PROTOC_DOWNLOAD_URL) -O protoc-$(PROTOC_VERSION)-linux-$(ARCH).zip
	@$(CMD_UNZIP) -o protoc-$(PROTOC_VERSION)-linux-$(ARCH).zip -d $(TOOLS_DIR)
	@mv $(TOOLS_DIR)/bin/protoc $(TOOLS_DIR)/bin/protoc-$(PROTOC_VERSION)
	@rm protoc-$(PROTOC_VERSION)-linux-$(ARCH).zip


################################################################################

### START GO ###

CMD_PROTOC_GEN_GO ?= $(TOOLS_DIR)/bin/protoc-gen-go-${PROTOC_GEN_GO_VERSION}
CMD_PROTOC_GEN_GO_GRPC ?= $(TOOLS_DIR)/bin/protoc-gen-go-grpc-${PROTOC_GEN_GO_GRPC_VERSION}

PROTOC_GEN_GO_NAME ?= google.golang.org/protobuf/cmd/protoc-gen-go
PROTOC_GEN_GO_GRPC_NAME ?= google.golang.org/grpc/cmd/protoc-gen-go-grpc
PROTOC_GO_PACKAGE_STYLE := import
PROTOC_GO_MODULE_PREFIX := github.com/seldonio/seldon-core/apis/go/v2
PROTOC_GO_OUTPUT_PREFIX := ./go

PROTOC_GO_PATH := --go_opt=paths=$(PROTOC_GO_PACKAGE_STYLE)
PROTOC_GO_MODULE := --go_opt=module=$(PROTOC_GO_MODULE_PREFIX)
PROTOC_GO_OUT := --go_out=$(PROTOC_GO_OUTPUT_PREFIX)
PROTOC_GO_GRPC_PATH := --go-grpc_opt=paths=$(PROTOC_GO_PACKAGE_STYLE)
PROTOC_GO_GRPC_MODULE := --go-grpc_opt=module=$(PROTOC_GO_MODULE_PREFIX)
PROTOC_GO_GRPC_OUT := --go-grpc_out=$(PROTOC_GO_OUTPUT_PREFIX)
PROTOC_GEN_GO_PLUGIN := --plugin=protoc-gen-go=$(TOOLS_DIR)/bin/protoc-gen-go-$(PROTOC_GEN_GO_VERSION)
PROTOC_GEN_GO_GRPC_PLUGIN := --plugin=protoc-gen-go-grpc=$(TOOLS_DIR)/bin/protoc-gen-go-grpc-$(PROTOC_GEN_GO_GRPC_VERSION)
PROTOC_GO_OPTIONS = $(PROTOC_GO_PATH) \
										$(PROTOC_GO_MODULE) \
										$(PROTOC_GO_OUT) \
										$(PROTOC_GO_GRPC_PATH) \
										$(PROTOC_GO_GRPC_MODULE) \
										$(PROTOC_GEN_GO_PLUGIN) \
										$(PROTOC_GEN_GO_GRPC_PLUGIN) \
										$(PROTOC_GO_GRPC_OUT)

# Install protoc-gen-go
$(TOOLS_DIR)/bin/protoc-gen-go-$(PROTOC_GEN_GO_VERSION): \
	| $(TOOLS_DIR)/.check_$(CMD_GO)
#
	@GOBIN=$(TOOLS_DIR)/bin $(CMD_GO) install $(PROTOC_GEN_GO_NAME)@v$(PROTOC_GEN_GO_VERSION)
	@mv $(TOOLS_DIR)/bin/protoc-gen-go $(TOOLS_DIR)/bin/protoc-gen-go-$(PROTOC_GEN_GO_VERSION)

# Install protoc-gen-go-grpc
$(TOOLS_DIR)/bin/protoc-gen-go-grpc-$(PROTOC_GEN_GO_GRPC_VERSION): \
	| $(TOOLS_DIR)/.check_$(CMD_GO)
#
	@GOBIN=$(TOOLS_DIR)/bin $(CMD_GO) install $(PROTOC_GEN_GO_GRPC_NAME)@v$(PROTOC_GEN_GO_GRPC_VERSION)
	@mv $(TOOLS_DIR)/bin/protoc-gen-go-grpc $(TOOLS_DIR)/bin/protoc-gen-go-grpc-$(PROTOC_GEN_GO_GRPC_VERSION)

# Generate go code from protobuf files
.PHONY: build-go
.ONESHELL: build-go
build-go: \
	| $(TOOLS_DIR)/.check_$(CMD_PROTOC) \
	  $(TOOLS_DIR)/.check_$(CMD_PROTOC_GEN_GO) \
		$(TOOLS_DIR)/.check_$(CMD_PROTOC_GEN_GO_GRPC)
#
	@echo "[INFO] Generating Go code from protobuf files..."
	@mkdir -p $(PROTOC_GO_OUTPUT_PREFIX)
	@$(CMD_PROTOC) \
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
PROTOC_PYTHON_VENV_DIR = $(TOOLS_DIR)/python/.venv.grpcio-tools-$(PROTOC_GEN_PYTHON_GRPCIO_TOOLS_VERSION)

# Create python virtual environment and install grpcio-tools
$(PROTOC_PYTHON_VENV_DIR):
	@$(CMD_PYTHON) -m venv $(PROTOC_PYTHON_VENV_DIR)
	@. $(PROTOC_PYTHON_VENV_DIR)/bin/activate; \
	 pip install pip -U; \
	 pip install grpcio-tools==$(PROTOC_GEN_PYTHON_GRPCIO_TOOLS_VERSION) -U;
	@touch $(TOOLS_DIR)/.check_python-grpcio-tools-$(PROTOC_GEN_PYTHON_GRPCIO_TOOLS_VERSION)

# Generate python code from protobuf files
.PHONY: build-v2-python
build-v2-python: \
	|	$(PROTOC_PYTHON_VENV_DIR)
#
	@echo "[INFO] Generating Python code from protobuf files..."
	@mkdir -p python
	@. $(PROTOC_PYTHON_VENV_DIR)/bin/activate; \
	 $(CMD_PYTHON) -m grpc_tools.protoc \
		$(PROTOC_IMPORT_PATH) \
		$(PROTOC_PYTHON_OPTIONS) \
		--pyi_out=./python \
		./mlops/v2_dataplane/v2_dataplane.proto

################################################################################

### START JVM ###

PROTOC_JAVA_GRPC_BIN ?= protoc-gen-grpc-java-$(PROTOC_GEN_JAVA_GRPC_VERSION)-linux-$(ARCH).exe
PROTOC_JAVA_GRPC_URL ?= https://repo1.maven.org/maven2/io/grpc/protoc-gen-grpc-java/$(PROTOC_GEN_JAVA_GRPC_VERSION)/$(PROTOC_JAVA_GRPC_BIN)

PROTOC_KOTLIN_GRPC_JAR ?= protoc-gen-grpc-kotlin-$(PROTOC_GEN_KOTLIN_GRPC_VERSION)-$(PROTOC_GEN_KOTLIN_GRPC_JDK).jar
PROTOC_KOTLIN_GRPC_URL ?= https://repo1.maven.org/maven2/io/grpc/protoc-gen-grpc-kotlin/$(PROTOC_GEN_KOTLIN_GRPC_VERSION)/$(PROTOC_KOTLIN_GRPC_JAR)
PROTOC_KOTLIN_GRPC_BIN_WRAPPER := protoc-gen-grpc-kotlin-$(PROTOC_GEN_KOTLIN_GRPC_VERSION).sh

CMD_PROTOC_GEN_JAVA_GRPC ?= $(TOOLS_DIR)/jvm/$(PROTOC_JAVA_GRPC_BIN)
CMD_PROTOC_GEN_KOTLIN_GRPC ?= $(TOOLS_DIR)/jvm/$(PROTOC_KOTLIN_GRPC_BIN_WRAPPER)

# Download protoc-gen-grpc-java
$(TOOLS_DIR)/jvm/$(PROTOC_JAVA_GRPC_BIN): \
	| $(TOOLS_DIR)/jvm/
#
	@$(CMD_WGET) -q $(PROTOC_JAVA_GRPC_URL) -O $@
	@chmod u+x $@

# Download protoc-gen-grpc-kotlin
$(TOOLS_DIR)/jvm/$(PROTOC_KOTLIN_GRPC_JAR): \
	| $(TOOLS_DIR)/jvm/
#
	@$(CMD_WGET) -q $(PROTOC_KOTLIN_GRPC_URL) -O $@

# Create executable wrapper for protoc-gen-grpc-kotlin
$(TOOLS_DIR)/jvm/$(PROTOC_KOTLIN_GRPC_BIN_WRAPPER): \
	  $(TOOLS_DIR)/jvm/$(PROTOC_KOTLIN_GRPC_JAR) \
	| $(TOOLS_DIR)/jvm/
#
	@echo '#!/bin/bash' > $@
	@echo 'java -jar $(TOOLS_DIR)/jvm/$(PROTOC_KOTLIN_GRPC_JAR) "$$@"' >> $@
	@chmod u+x $@

# Generate java/kotlin code from protobuf files
.PHONY: build-kotlin
build-kotlin: \
	| $(TOOLS_DIR)/.check_$(CMD_PROTOC) \
	  $(TOOLS_DIR)/.check_$(CMD_PROTOC_GEN_JAVA_GRPC) \
		$(TOOLS_DIR)/.check_$(CMD_PROTOC_GEN_KOTLIN_GRPC)
#
	@echo "[INFO] Generating Java/Kotlin code from protobuf files..."
	@cd $(BASE_DIR)/mlops/chainer && $(CMD_PROTOC) \
		--proto_path=. \
		--plugin=protoc-gen-grpc-java=$(CMD_PROTOC_GEN_JAVA_GRPC) \
		--java_out=./kotlin \
		--grpc-java_out=./kotlin \
		--plugin=protoc-gen-grpckt=$(CMD_PROTOC_GEN_KOTLIN_GRPC) \
		--kotlin_out=./kotlin \
		--grpckt_out=./kotlin \
		chainer.proto
	@cd $(BASE_DIR)/mlops/v2_dataplane && $(CMD_PROTOC) \
		--proto_path=. \
		--plugin=protoc-gen-grpc-java=$(CMD_PROTOC_GEN_JAVA_GRPC) \
		--java_out=./kotlin \
		--grpc-java_out=./kotlin \
		--plugin=protoc-gen-grpckt=$(CMD_PROTOC_GEN_KOTLIN_GRPC) \
		--kotlin_out=./kotlin \
		--grpckt_out=./kotlin \
		v2_dataplane.proto

################################################################################

.PHONY: update-copyright
update-copyright:
	@echo "[INFO] Updating copyright headers..."
	@$(MAKE) -C $(BASE_DIR)/.. --no-print-directory update-copyright
