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


function update_images {
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_EXECUTOR$\)#\1\2\n\1  value: '${EXECUTOR_IMAGE}:${VERSION}'#' ${CSV_FOLDER}/seldon-operator.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_ENGINE$\)#\1\2\n\1  value: '${ENGINE_IMAGE}:${VERSION}'#' ${CSV_FOLDER}/seldon-operator.clusterserviceversion.yaml
    sed -i 's#\(^.*image: \)seldonio/seldon-core-operator:.*$#\1'${OPERATOR_IMAGE}:${VERSION}'#' ${CSV_FOLDER}/seldon-operator.clusterserviceversion.yaml
    sed -i 's#\(^.*containerImage: \)seldonio/seldon-core-operator:.*$#\1'${OPERATOR_IMAGE}:${VERSION}'#' ${CSV_FOLDER}/seldon-operator.clusterserviceversion.yaml
    sed -i 's#seldonio/mock_classifier_rest:1.3#'${MOCK_IMAGE_SHA}'#' ${CSV_FOLDER}/seldon-operator.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_STORAGE_INITIALIZER$\)#\1\2\n\1  value: '${STORAGE_INITIALIZER_IMAGE}:${VERSION}'#' ${CSV_FOLDER}/seldon-operator.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_SKLEARNSERVER_REST$\)#\1\2\n\1  value: '${SKLEARNSERVER_REST_IMAGE}:${VERSION}'#' ${CSV_FOLDER}/seldon-operator.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_SKLEARNSERVER_GRPC$\)#\1\2\n\1  value: '${SKLEARNSERVER_GRPC_IMAGE}:${VERSION}'#' ${CSV_FOLDER}/seldon-operator.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_XGBOOSTSERVER_REST$\)#\1\2\n\1  value: '${XGBOOSTSERVER_REST_IMAGE}:${VERSION}'#' ${CSV_FOLDER}/seldon-operator.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_XGBOOSTSERVER_GRPC$\)#\1\2\n\1  value: '${XGBOOSTSERVER_GRPC_IMAGE}:${VERSION}'#' ${CSV_FOLDER}/seldon-operator.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_MLFLOWSERVER_REST$\)#\1\2\n\1  value: '${MLFLOWSERVER_REST_IMAGE}:${VERSION}'#' ${CSV_FOLDER}/seldon-operator.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_MLFLOWSERVER_GRPC$\)#\1\2\n\1  value: '${MLFLOWSERVER_GRPC_IMAGE}:${VERSION}'#' ${CSV_FOLDER}/seldon-operator.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_TFPROXY_REST$\)#\1\2\n\1  value: '${TFPROXY_REST_IMAGE}:${VERSION}'#' ${CSV_FOLDER}/seldon-operator.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_TFPROXY_GRPC$\)#\1\2\n\1  value: '${TFPROXY_GRPC_IMAGE}:${VERSION}'#' ${CSV_FOLDER}/seldon-operator.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_TENSORFLOW$\)#\1\2\n\1  value: '${TENSORFLOW_IMAGE}:2.1.0'#' ${CSV_FOLDER}/seldon-operator.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_EXPLAINER$\)#\1\2\n\1  value: '${EXPLAINER_IMAGE}:${VERSION}'#' ${CSV_FOLDER}/seldon-operator.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_MOCK_CLASSIFIER$\)#\1\2\n\1  value: '${MOCK_IMAGE}:${VERSION}'#' ${CSV_FOLDER}/seldon-operator.clusterserviceversion.yaml
}


VERSION=$1
CSV_FOLDER=packagemanifests-certified/${VERSION}
update_images
