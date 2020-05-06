SHELL := /bin/bash
VERSION ?= $(shell cat ../../version.txt)
PREV_VERSION=1.1.0
COMMUNITY_OPERATORS_BASE=/home/clive/work/seldon-core/redhat/community-operators

.PHONY: recreate-generated
generate-resources:
	rm -rf generated
	mkdir generated
	kustomize build ../config/lite -o generated

deploy/operator.yaml: generated
	python hack/create_resources.py --input generated --output deploy --version ${VERSION}

# DO NOT RUN - used during development and can be used for next version
deploy/olm-catalog/seldon-operator/1.0.2/seldon-operator.v1.0.2.clusterserviceversion.yaml: 
	operator-sdk generate csv --csv-version 1.0.2 --from-version 0.1.5

# DO NOT RUN - used during development and can be used for next version
.PHONY: clean_csv_${VERSION}
clean_csv_${VERSION}:
	rm -f deploy/olm-catalog/seldon-operator/${VERSION}/seldon-operator.v${VERSION}.clusterserviceversion.yaml

.PHONY: clean_csv
clean_csv: clean_csv_${VERSION}

# DO NOT RUN - used during development and can be used for next version
deploy/olm-catalog/seldon-operator/${VERSION}/seldon-operator.v${VERSION}.clusterserviceversion.yaml: 
	operator-sdk generate csv --csv-version ${VERSION} --from-version ${PREV_VERSION} --update-crds

update_csv: clean_csv_${VERSION} deploy/olm-catalog/seldon-operator/${VERSION}/seldon-operator.v${VERSION}.clusterserviceversion.yaml
	sed -i s/${PREV_VERSION}/${VERSION}/ deploy/olm-catalog/seldon-operator/seldon-operator.package.yaml
	sed -i 's/containerImage:.*/containerImage: seldonio\/seldon-core-operator:'${VERSION}'/' deploy/olm-catalog/seldon-operator/${VERSION}/seldon-operator.v${VERSION}.clusterserviceversion.yaml
#
# TESTS
#

scorecard:
	operator-sdk scorecard -o text --bundle deploy/olm-catalog/seldon-operator --kubeconfig ~/.kube/config

operator-courier_ui_validate:
	operator-courier verify --ui_validate_io deploy/olm-catalog/seldon-operator

# See https://github.com/operator-framework/community-operators/blob/master/docs/testing-operators.md
# Used to push bundle to quay.io for testing
community-quay-push:
	operator-courier push deploy/olm-catalog/seldon-operator seldon seldon-operator ${VERSION} "$$QUAY_TOKEN"


#
# RELEASE
#

# Set COMMUNITY_OPERATORS_BASE env

.PHONY: update_community
update_community:
	cp deploy/olm-catalog/seldon-operator/seldon-operator.package.yaml ${COMMUNITY_OPERATORS_BASE}/community-operators/seldon-operator
	cp -r deploy/olm-catalog/seldon-operator/${VERSION} ${COMMUNITY_OPERATORS_BASE}/community-operators/seldon-operator/${VERSION}

.PHONY: update_upstream
update_upstream:
	cp deploy/olm-catalog/seldon-operator/seldon-operator.package.yaml ${COMMUNITY_OPERATORS_BASE}/upstream-community-operators/seldon-operator
	cp -r deploy/olm-catalog/seldon-operator/${VERSION} ${COMMUNITY_OPERATORS_BASE}/upstream-community-operators/seldon-operator/${VERSION}

# See https://github.com/operator-framework/community-operators/blob/master/docs/using-scripts.md
.PHONY: upstream_lint
upstream_lint:
	cd ${COMMUNITY_OPERATORS_BASE} && make operator.verify OP_PATH=upstream-community-operators/seldon-operator VERBOSE=1

.PHONY: upstream_install
upstream_install:
	cd ${COMMUNITY_OPERATORS_BASE} && make operator.install OP_PATH=upstream-community-operators/seldon-operator

.PHONY: upstream_test
upstream_test:
	cd ${COMMUNITY_OPERATORS_BASE} && make operator.test OP_PATH=upstream-community-operators/seldon-operator



#
# Certified : https://redhat-connect.gitbook.io/certified-operator-guide/
#

certified_create_bundle:
	cd bundle && ./create_bundle.sh ${VERSION}

certified_validate:
	 operator-courier --verbose verify bundle/certified/

certified_validate_ui:
	operator-courier --verbose verify --ui_validate_io bundle/certified/


certified_push_quay:
	operator-courier push bundle/certified/ seldon seldon-operator-certified 1.1.0 "$$QUAY_TOKEN"


certified_zip:
	rm -f bundle/bundle.zip
	cd bundle/certified && zip -r ../bundle.zip .
