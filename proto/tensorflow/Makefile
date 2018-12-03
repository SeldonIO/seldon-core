SHELL=/bin/bash

tensorflow/core/framework/tensor.proto:
	./create-protos.sh

.PHONY: create_protos
create_protos: tensorflow/core/framework/tensor.proto

.PHONY: clean
clean:
	@rm -rfv tensorflow
