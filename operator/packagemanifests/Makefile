SHELL=/bin/bash

.PHONY: create_bundle_image
create_bundle_image_%:
	docker build . -f bundle-version.Dockerfile --build-arg VERSION=$* -t quay.io/seldon/seldon-operator:v$*

.PHONY: push_bundle_image
push_bundle_image_%:
	docker push quay.io/seldon/seldon-operator:v$*


create_bundles: create_bundle_image_1.2.2 create_bundle_image_1.2.1 create_bundle_image_1.2.0 create_bundle_image_1.1.0 create_bundle_image_1.0.0 create_bundle_image_0.1.5 create_bundle_image_0.1.4 create_bundle_image_0.1.3 create_bundle_image_0.1.2

push_bundles: push_bundle_image_1.2.2 push_bundle_image_1.2.1 push_bundle_image_1.2.0 push_bundle_image_1.1.0 push_bundle_image_1.0.0 push_bundle_image_0.1.5 push_bundle_image_0.1.4 push_bundle_image_0.1.3 push_bundle_image_0.1.2


build_push: create_bundles push_bundles
