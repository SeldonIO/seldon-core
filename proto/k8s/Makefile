SHELL=/bin/bash

create_protos:
	./create-k8s-protos.sh
	sed -i.bak 's|import "k8s.io/apiextensions-apiserver/|//import "k8s.io/apiextensions-apiserver/|' k8s.io/api/core/v1/generated.proto
	rm k8s.io/api/core/v1/generated.proto.bak

clean:
	@rm -rfv k8s.io


