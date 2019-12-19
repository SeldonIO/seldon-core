
SELDON_CORE_DIR=../../../../

.PHONY: compile_proto
compile_proto:
	rm -rf pkg/api && mkdir -p pkg/api

	cp ${SELDON_CORE_DIR}/proto/prediction.proto pkg/api

	$(MAKE) -C ${SELDON_CORE_DIR}/proto/tensorflow/ create_protos
	cp -r $(SELDON_CORE_DIR)/proto/tensorflow/tensorflow pkg/api

	cd pkg/api && protoc -I. -I./tensorflow --go_out=paths=source_relative,plugins=grpc:. *.proto
	cd pkg/api && protoc -I. -I./tensorflow --go_out=paths=source_relative,plugins=grpc:. ./tensorflow/core/framework/*.proto

	find ./pkg -name "*.pb.go" -type f -exec perl -e "s/github\.com\/tensorflow\/tensorflow\/tensorflow\/go/github\.com\/hemantkashniyal\/seldon-core\/incubating\/wrappers\/s2i\/go\/pkg\/api\/tensorflow/g" -pi {} +
	find ./pkg -name "*.proto" -type f -delete

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
