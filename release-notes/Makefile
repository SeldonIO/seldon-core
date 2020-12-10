VERSION := $(shell cat ../version.txt)

install-dev:
	GO111MODULE=on go install k8s.io/release/cmd/release-notes

generate:
	release-notes \
		--github-org SeldonIO \
		--github-repo seldon-core \
		--required-author '' \
		--start-rev v1.2.1 \
		--end-rev master \
		--go-template go-template:release-notes.gotmpl \
		--output release-notes-$(VERSION).md \
		--release-version $(VERSION) \
		--dependencies false

