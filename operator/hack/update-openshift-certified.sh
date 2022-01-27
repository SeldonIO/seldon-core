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
SKLEARNSERVER_IMAGE=registry.connect.redhat.com/seldonio/sklearnserver
XGBOOSTSERVER_IMAGE=registry.connect.redhat.com/seldonio/xgboostserver
MLFLOWSERVER_IMAGE=registry.connect.redhat.com/seldonio/mlflowserver
TFPROXY_IMAGE=registry.connect.redhat.com/seldonio/tfproxy
TENSORFLOW_IMAGE=registry.connect.redhat.com/seldonio/tensorflow-serving
EXPLAINER_IMAGE=registry.connect.redhat.com/seldonio/alibiexplainer

mkdir -p certified


function update_images {
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_EXECUTOR$\)#\1\2\n\1  value: '${EXECUTOR_IMAGE}:${VERSION}'#' ${CSV_FOLDER}/seldon-operator-certified.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_ENGINE$\)#\1\2\n\1  value: '${ENGINE_IMAGE}:${VERSION}'#' ${CSV_FOLDER}/seldon-operator-certified.clusterserviceversion.yaml
    sed -i 's#\(^.*image: \)docker.io/seldonio/seldon-core-operator:.*$#\1'${OPERATOR_IMAGE}:${VERSION}'#' ${CSV_FOLDER}/seldon-operator-certified.clusterserviceversion.yaml
    sed -i 's#\(^.*containerImage: \)seldonio/seldon-core-operator:.*$#\1'${OPERATOR_IMAGE}:${VERSION}'#' ${CSV_FOLDER}/seldon-operator-certified.clusterserviceversion.yaml
    sed -i 's#seldonio/mock_classifier_rest:1.3#'${MOCK_IMAGE_SHA}'#' ${CSV_FOLDER}/seldon-operator-certified.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_STORAGE_INITIALIZER$\)#\1\2\n\1  value: '${STORAGE_INITIALIZER_IMAGE}:${VERSION}'#' ${CSV_FOLDER}/seldon-operator-certified.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_SKLEARNSERVER$\)#\1\2\n\1  value: '${SKLEARNSERVER_IMAGE}:${VERSION}'#' ${CSV_FOLDER}/seldon-operator-certified.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_XGBOOSTSERVER$\)#\1\2\n\1  value: '${XGBOOSTSERVER_IMAGE}:${VERSION}'#' ${CSV_FOLDER}/seldon-operator-certified.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_MLFLOWSERVER$\)#\1\2\n\1  value: '${MLFLOWSERVER_IMAGE}:${VERSION}'#' ${CSV_FOLDER}/seldon-operator-certified.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_TFPROXY$\)#\1\2\n\1  value: '${TFPROXY_IMAGE}:${VERSION}'#' ${CSV_FOLDER}/seldon-operator-certified.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_TENSORFLOW$\)#\1\2\n\1  value: '${TENSORFLOW_IMAGE}:2.1.0'#' ${CSV_FOLDER}/seldon-operator-certified.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_EXPLAINER$\)#\1\2\n\1  value: '${EXPLAINER_IMAGE}:${VERSION}'#' ${CSV_FOLDER}/seldon-operator-certified.clusterserviceversion.yaml
    sed -i 's#\(^.*\)\(- name: RELATED_IMAGE_MOCK_CLASSIFIER$\)#\1\2\n\1  value: '${MOCK_IMAGE}:${VERSION}'#' ${CSV_FOLDER}/seldon-operator-certified.clusterserviceversion.yaml
}


VERSION=$1
CSV_FOLDER=bundle-certified/manifests
update_images
