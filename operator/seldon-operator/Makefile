
.PHONY: recreate-generated
generate-resources:
	rm -rf generated
	mkdir generated
	kustomize build ../config/lite -o generated

deploy/operator.yaml: generated
	python hack/create_resources.py --input generated --output deploy

# DO NOT RUN - used during development and can be used for next version
deploy/olm-catalog/seldon-operator/1.0.2/seldon-operator.v1.0.2.clusterserviceversion.yaml: 
	operator-sdk generate csv --csv-version 1.0.2 --from-version 0.1.5

# DO NOT RUN - used during development and can be used for next version
.PHONY: clean_1.1.0
clean_1.1.0:
	rm -f deploy/olm-catalog/seldon-operator/1.1.0/seldon-operator.v1.1.0.clusterserviceversion.yaml

# DO NOT RUN - used during development and can be used for next version
deploy/olm-catalog/seldon-operator/1.1.0/seldon-operator.v1.1.0.clusterserviceversion.yaml: 
	operator-sdk generate csv --csv-version 1.1.0 --from-version 1.0.0 --update-crds

#
# TESTS
#

scorecard:
	operator-sdk scorecard -o text --bundle deploy/olm-catalog/seldon-operator --kubeconfig ~/.kube/config

operator-courier_ui_validate:
	operator-courier verify --ui_validate_io deploy/olm-catalog/seldon-operator

# See https://github.com/operator-framework/community-operators/blob/master/docs/testing-operators.md
# Used to push bundle to quay.io for testing
operator-courier_push:
	operator-courier push deploy/olm-catalog/seldon-operator seldon seldon-operator 1.1.0 "$$QUAY_TOKEN"
