#!/bin/bash

set -o errexit
set -o pipefail
set -o nounset

# FIRST WE START THE DOCKER DAEMON
service docker start
# the service can be started but the docker socket not ready, wait for ready
WAIT_N=0
while true; do
    # docker ps -q should only work if the daemon is ready
    docker ps -q > /dev/null 2>&1 && break
    if [[ ${WAIT_N} -lt 5 ]]; then
        WAIT_N=$((WAIT_N+1))
        echo "[SETUP] Waiting for Docker to be ready, sleeping for ${WAIT_N} seconds ..."
        sleep ${WAIT_N}
    else
        echo "[SETUP] Reached maximum attempts, not waiting any longer ..."
        break
    fi
done

#######################################
# AVOID EXIT ON ERROR FOR FOLLOWING CMDS
set +o errexit

# START CLUSTER 
make kind_create_cluster
KIND_EXIT_VALUE=$?

# Ensure we reach the kubeconfig path
export KUBECONFIG=$(kind get kubeconfig-path)

# ONLY RUN THE FOLLOWING IF SUCCESS
if [[ ${KIND_EXIT_VALUE} -eq 0 ]]; then
    # BUILD S2I BASE IMAGES
    PYTHON_MODIFIED=`git diff --exit-code --quiet master ../../python`
    if [[ $PYTHON_MODIFIED -gt 0 ]]; then 
        make s2i_build_base_images
    else
        echo "SKIPPING PYTHON IMAGE BUILD..."
    fi

    # MORE EFFICIENT CLUSTER SETUP
    OPERATOR_MODIFIED=`git diff --exit-code --quiet master ../../operator`
    if [[ $OPERATOR_MODIFIED -gt 0 ]]; then
        make kind_build_operator
        OPERATOR_EXIT_VALUE=$?
    else
        echo "SKIPPING OPERATOR IMAGE BUILD..."
    fi

    ENGINE_MODIFIED=`git diff --exit-code --quiet master ../../engine`
    if [[ $ENGINE_MODIFIED -gt 0 ]]; then
        make build_protos
        PROTO_EXIT_VALUE=$?
        make kind_build_engine
        ENGINE_EXIT_VALUE=$?
    else
        echo "SKIPPING ENGINE IMAGE BUILD..."
    fi

    # KIND CLUSTER SETUP
    make kind_setup
    SETUP_EXIT_VALUE=$?

    ## INSTALL ALL REQUIRED DEPENDENCIES
    make -C ../../python install_dev
    INSTALL_EXIT_VALUE=$?

    ## RUNNING TESTS AND CAPTURING ERROR
    make test
    TEST_EXIT_VALUE=$?
fi

# DELETE KIND CLUSTER
make kind_delete_cluster
DELETE_EXIT_VALUE=$?

#######################################
# EXIT STOPS COMMANDS FROM HERE ONWARDS
set -o errexit

# CLEANING DOCKER
docker ps -aq | xargs -r docker rm -f || true
service docker stop || true

# NOW THAT WE'VE CLEANED WE CAN EXIT ON TEST EXIT VALUE
exit ${TEST_EXIT_VALUE}

