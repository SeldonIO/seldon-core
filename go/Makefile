SELDON_CORE_DIR=..

.PHONY: compile_proto
compile_proto:
	rm -rf pkg/api && mkdir -p pkg/api

	cp ${SELDON_CORE_DIR}/proto/prediction.proto pkg/api

	$(MAKE) -C ${SELDON_CORE_DIR}/proto/tensorflow/ create_protos
	cp -r $(SELDON_CORE_DIR)/proto/tensorflow/tensorflow pkg/api

	cd pkg/api && protoc -I. -I./tensorflow --go_out=paths=source_relative,plugins=grpc:. *.proto
	cd pkg/api && protoc -I. -I./tensorflow --go_out=paths=source_relative,plugins=grpc:. ./tensorflow/core/framework/*.proto

	find ./pkg -name "*.pb.go" -type f -exec perl -e "s/github\.com\/tensorflow\/tensorflow\/tensorflow\/go/github\.com\/seldonio\/seldon-core\/go\/pkg\/api\/tensorflow/g" -pi {} +
	find ./pkg -name "*.proto" -type f -delete