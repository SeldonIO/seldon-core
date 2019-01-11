

.PHONY: compile_tensorflow_proto
compile_tensorflow_proto:
	./generate_tensorflow_proto.sh

.PHONY: compile_proto
compile_proto:
	cp ../../../proto/prediction.proto pkg/api
	cd pkg/api && protoc -I. -I${GOPATH}/src/github.com/tensorflow/tensorflow --go_out=paths=source_relative,plugins=grpc:. prediction.proto
	rm pkg/api/prediction.proto


.PHONY: build_docker
build_docker:
	docker build -t seldonio/gomodel:0.1 .

.PHONY: test_docker_rest
test_docker_rest:
	docker run -d --name "gomodel" -p 10000:10000 -e SERVER_TYPE='rest' --rm seldonio/gomodel:0.1

.PHONY: test_docker_grpc
test_docker_grpc:
	docker run -d --name "gomodel" -p 10000:10000 --rm seldonio/gomodel:0.1

.PHONY: clean
clean:
	rm server
