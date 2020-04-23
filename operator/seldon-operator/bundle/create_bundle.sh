#!/bin/bash

set -o nounset
set -o errexit

STARTUP_DIR="$( cd "$( dirname "$0" )" && pwd )"
OPERATOR_IMAGE=registry.connect.redhat.com/seldonio/seldon-core-operator
EXECUTOR_IMAGE=registry.connect.redhat.com/seldonio/seldon-core-executor
ENGINE_IMAGE=registry.connect.redhat.com/seldonio/seldon-engine
MOCK_IMAGE=registry.connect.redhat.com/seldonio/mock-classifier

mkdir -p certified

function copy_files {
    mkdir -p certified/${VERSION}
    cp ../deploy/olm-catalog/seldon-operator/${VERSION}/seldon-operator.v${VERSION}.clusterserviceversion.yaml certified/${VERSION}
    cp ../deploy/olm-catalog/seldon-operator/${VERSION}/machinelearning.seldon.io_seldondeployment_crd.yaml certified/${VERSION}
    cp ../deploy/olm-catalog/seldon-operator/seldon-operator.package.yaml certified
}

function update_images {
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_EXECUTOR$\)#\1\2\n\1  value: '${EXECUTOR_IMAGE}:${VERSION}'#' certified/${VERSION}/seldon-operator.v${VERSION}.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_ENGINE$\)#\1\2\n\1  value: '${ENGINE_IMAGE}:${VERSION}'#' certified/${VERSION}/seldon-operator.v${VERSION}.clusterserviceversion.yaml
    sed -i 's#\(^.*image: \)seldonio/seldon-core-operator:.*$#\1'${OPERATOR_IMAGE}:${VERSION}'#' certified/${VERSION}/seldon-operator.v${VERSION}.clusterserviceversion.yaml
    sed -i 's#\(^.*containerImage: \)seldonio/seldon-core-operator:.*$#\1'${OPERATOR_IMAGE}:${VERSION}'#' certified/${VERSION}/seldon-operator.v${VERSION}.clusterserviceversion.yaml
    sed -i 's#seldonio/mock_classifier_rest:1.3#'${MOCK_IMAGE}:${VERSION}'#' certified/${VERSION}/seldon-operator.v${VERSION}.clusterserviceversion.yaml    
}

function update_package {
    sed -i 's/packageName: seldon-operator/packageName: seldon-operator-certified/' certified/seldon-operator.package.yaml
}

VERSION=$1
copy_files
update_images
update_package
