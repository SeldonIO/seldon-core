#!/bin/bash

set -o nounset
set -o errexit

STARTUP_DIR="$( cd "$( dirname "$0" )" && pwd )"
OPERATOR_IMAGE=registry.connect.redhat.com/seldonio/seldon-core-operator
EXECUTOR_IMAGE=registry.connect.redhat.com/seldonio/seldon-core-executor
ENGINE_IMAGE=registry.connect.redhat.com/seldonio/seldon-engine
MOCK_IMAGE_SHA=registry.connect.redhat.com/seldonio/mock-classifier@sha256:482ee477c344badcaa80e850f4339db41957f9c2396ae24f9e398b67bd5c184e
MOCK_IMAGE=registry.connect.redhat.com/seldonio/mock-classifier
STORAGE_INITIALIZER_IMAGE=registry.connect.redhat.com/seldonio/storage-initializer
SKLEARNSERVER_REST_IMAGE=registry.connect.redhat.com/seldonio/sklearnserver-rest
SKLEARNSERVER_GRPC_IMAGE=registry.connect.redhat.com/seldonio/sklearnserver-grpc
XGBOOSTSERVER_REST_IMAGE=registry.connect.redhat.com/seldonio/xgboostserver-rest
XGBOOSTSERVER_GRPC_IMAGE=registry.connect.redhat.com/seldonio/xgboostserver-grpc
MLFLOWSERVER_REST_IMAGE=registry.connect.redhat.com/seldonio/mlflowserver-rest
MLFLOWSERVER_GRPC_IMAGE=registry.connect.redhat.com/seldonio/mlflowserver-grpc
TFPROXY_REST_IMAGE=registry.connect.redhat.com/seldonio/tfproxy-rest
TFPROXY_GRPC_IMAGE=registry.connect.redhat.com/seldonio/tfproxy-grpc
TENSORFLOW_IMAGE=registry.connect.redhat.com/seldonio/tensorflow-serving
EXPLAINER_IMAGE=registry.connect.redhat.com/seldonio/alibiexplainer

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
    sed -i 's#seldonio/mock_classifier_rest:1.3#'${MOCK_IMAGE_SHA}'#' certified/${VERSION}/seldon-operator.v${VERSION}.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_STORAGE_INITIALIZER$\)#\1\2\n\1  value: '${STORAGE_INITIALIZER_IMAGE}:${VERSION}'#' certified/${VERSION}/seldon-operator.v${VERSION}.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_SKLEARNSERVER_REST$\)#\1\2\n\1  value: '${SKLEARNSERVER_REST_IMAGE}:${VERSION}'#' certified/${VERSION}/seldon-operator.v${VERSION}.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_SKLEARNSERVER_GRPC$\)#\1\2\n\1  value: '${SKLEARNSERVER_GRPC_IMAGE}:${VERSION}'#' certified/${VERSION}/seldon-operator.v${VERSION}.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_XGBOOSTSERVER_REST$\)#\1\2\n\1  value: '${XGBOOSTSERVER_REST_IMAGE}:${VERSION}'#' certified/${VERSION}/seldon-operator.v${VERSION}.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_XGBOOSTSERVER_GRPC$\)#\1\2\n\1  value: '${XGBOOSTSERVER_GRPC_IMAGE}:${VERSION}'#' certified/${VERSION}/seldon-operator.v${VERSION}.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_MLFLOWSERVER_REST$\)#\1\2\n\1  value: '${MLFLOWSERVER_REST_IMAGE}:${VERSION}'#' certified/${VERSION}/seldon-operator.v${VERSION}.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_MLFLOWSERVER_GRPC$\)#\1\2\n\1  value: '${MLFLOWSERVER_GRPC_IMAGE}:${VERSION}'#' certified/${VERSION}/seldon-operator.v${VERSION}.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_TFPROXY_REST$\)#\1\2\n\1  value: '${TFPROXY_REST_IMAGE}:${VERSION}'#' certified/${VERSION}/seldon-operator.v${VERSION}.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_TFPROXY_GRPC$\)#\1\2\n\1  value: '${TFPROXY_GRPC_IMAGE}:${VERSION}'#' certified/${VERSION}/seldon-operator.v${VERSION}.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_TENSORFLOW$\)#\1\2\n\1  value: '${TENSORFLOW_IMAGE}:2.1.0'#' certified/${VERSION}/seldon-operator.v${VERSION}.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_EXPLAINER$\)#\1\2\n\1  value: '${EXPLAINER_IMAGE}:${VERSION}'#' certified/${VERSION}/seldon-operator.v${VERSION}.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_MOCK_CLASSIFIER$\)#\1\2\n\1  value: '${MOCK_IMAGE}:${VERSION}'#' certified/${VERSION}/seldon-operator.v${VERSION}.clusterserviceversion.yaml    
}

function update_package {
    sed -i 's/packageName: seldon-operator/packageName: seldon-operator-certified/' certified/seldon-operator.package.yaml
}

VERSION=$1
copy_files
update_images
update_package
