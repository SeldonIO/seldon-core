.PHONY: README.md

README.md:
	docker run -it --rm \
		-v $(PWD):/out \
		-v $(PWD):/templates \
		-v $(PWD):/protos \
		pseudomuto/protoc-gen-doc \
		--doc_opt=/templates/template.md,README.md
