VERSION := $(shell cat ../version.txt)
# Image URL to use all building/pushing image targets
IMG ?= seldonio/seldon-core-executor:${VERSION}

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...


# Build manager binary
executor: fmt vet
	go build -o executor main.go


.PHONY: copy_protos
copy_protos:
	cp -r ../proto/tensorflow/tensorflow/** proto/tensorflow


.PHONY: compile_seldon_proto
compile_seldon_proto:
	cp ../proto/prediction.proto api/grpc
	cd api/grpc && protoc -I. -I${GOPATH}/src/github.com/tensorflow/tensorflow --go_out=paths=source_relative,plugins=grpc:. prediction.proto
	rm api/grpc/prediction.proto

# https://github.com/tensorflow/serving/issues/1365#issuecomment-525351995
.PHONY: compile_tensorflow_proto
compile_tensorflow_proto:
	git clone -b r1.15 https://github.com/tensorflow/tensorflow.git
	git clone -b r1.14 https://github.com/tensorflow/serving.git
	go run protoc.go
	go mod edit -replace=github.com/tensorflow/tensorflow/tensorflow/go/core=./proto/tensorflow/core
	cd proto/tensorflow/core && go mod init github.com/tensorflow/tensorflow/tensorflow/go/core && cd -
	go build ./proto/tensorflow/serving

# Run tests
test: fmt vet
	go test ./api/... ./predictor/...  -coverprofile cover.out

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}


kind-image-install: docker-build
	docker save ${IMG} > executor.tar
	kind load image-archive executor.tar --loglevel trace


.PHONY: clean
clean:
	rm -rf vendor
	rm -rf tensorflow
	rm -rf serving

start_test_grpc_server:
	./executor --sdep seldon-model --namespace default --predictor example --file samples/model_test_grpc.yaml --http_port 8000 --grpc_port 5000 --transport grpc

start_dummy_grpc_model:
	cd ../examples/models/mean_classifier && make run_grpc_local 

test_grpc_local:
	cd proto && grpcurl -d '{"data":{"ndarray":[[1.0,2.0]]}}' -plaintext -proto ./prediction.proto  0.0.0.0:5000 seldon.protos.Seldon/Predict


start_test_rest_server:
	./executor --sdep seldon-model --namespace default --predictor example --file samples/model_test_rest.yaml --http_port 8000 

start_test_rest_server_logger:
	./executor --sdep seldon-model --namespace default --predictor example --file samples/model_test_rest_logger.yaml --http_port 8000 

start_dummy_rest_model:
	cd ../examples/models/mean_classifier && make run_rest_local 

start_dummy_logsink:
	docker run -it -p 2222:80 --rm -t mendhak/http-https-echo

test_rest_local:
	curl -v localhost:8000/api/v0.1/predictions -d '{"data":{"ndarray":[[1.0,2.0]]}}' 
