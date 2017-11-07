SHELL=/bin/bash

create_protos:
	./create-k8s-protos.sh
	sed --in-place 's|import "k8s.io/apiextensions-apiserver/|//import "k8s.io/apiextensions-apiserver/|' v1.proto

clean:
	@rm -rfv k8s.io
	@rm v1.proto

