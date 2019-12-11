.PHONY: compile_tensorflow_proto
compile_tensorflow_proto:
	./generate_tensorflow_proto.sh

.PHONY: compile_proto
compile_proto: compile_tensorflow_proto
	mkdir -p pkg/api
	cp -r vendor/github.com/tensorflow  pkg/api/
	cp ../proto/prediction.proto pkg/api
	cd pkg/api && protoc -I. -I${CURDIR}/tensorflow --go_out=paths=source_relative,plugins=grpc:. prediction.proto
	rm pkg/api/prediction.proto
	find ./pkg -name "*.pb.go" -type f -exec perl -e "s/github\.com\/tensorflow/github\.com\/seldonio\/seldon-core\/go\/pkg\/api\/tensorflow/g" -pi {} +