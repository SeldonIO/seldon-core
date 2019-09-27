SHELL=/bin/bash

create_protos:
	./create-k8s-protos.sh
	sed -i '' 's|import "k8s.io/apiextensions-apiserver/|//import "k8s.io/apiextensions-apiserver/|' k8s.io/api/core/v1/generated.proto

clean:
	@rm -rfv k8s.io


