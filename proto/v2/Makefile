.PHONY: README.md dataplane.proto

MLSERVER_RELEASE := master
MLSERVER_BASE := https://raw.githubusercontent.com/SeldonIO/MLServer/${MLSERVER_RELEASE}

dataplane.proto:
	curl -s -L ${MLSERVER_BASE}/proto/dataplane.proto > ./dataplane.proto

# Must be run from the directory of this Makefile, due to the usage of $(PWD)
README.md: dataplane.proto
	docker run -it --rm \
		-v $(PWD):/out \
		-v $(PWD):/templates \
		-v $(PWD):/protos \
		pseudomuto/protoc-gen-doc \
		--doc_opt=/templates/readme.md.tmpl,README.md
