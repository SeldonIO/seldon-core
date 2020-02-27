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
        tail /var/log/docker.log
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

    echo "Files changed in python folder:"
    git --no-pager diff --exit-code --name-only origin/master ../../python
    PYTHON_MODIFIED=$?
    if [[ $PYTHON_MODIFIED -gt 0 ]]; then 
        make s2i_build_base_images
    else
        echo "SKIPPING PYTHON IMAGE BUILD..."
    fi

    echo "Files changed in operator folder:"
    git --no-pager diff --exit-code --name-only origin/master ../../operator
    OPERATOR_MODIFIED=$?
    if [[ $OPERATOR_MODIFIED -gt 0 ]]; then
        make kind_build_operator
        OPERATOR_EXIT_VALUE=$?
    else
        echo "SKIPPING OPERATOR IMAGE BUILD..."
    fi

    echo "Files changed in engine folder:"
    git --no-pager diff --exit-code --name-only origin/master ../../engine
    ENGINE_MODIFIED=$?
    if [[ $ENGINE_MODIFIED -gt 0 ]]; then
        make build_protos
        PROTO_EXIT_VALUE=$?
        make kind_build_engine
        ENGINE_EXIT_VALUE=$?
    else
        echo "SKIPPING ENGINE IMAGE BUILD..."
    fi

    echo "Files changed in executor folder:"
    git --no-pager diff --exit-code --name-only origin/master ../../executor
    EXECUTOR_MODIFIED=$?
    if [[ $EXECUTOR_MODIFIED -gt 0 ]]; then
        make kind_build_executor
        EXECUTOR_EXIT_VALUE=$?
    else
        echo "SKIPPING EXECUTOR IMAGE BUILD..."
    fi

    echo "Build fixed models"
    make kind_build_fixed_models

    # KIND CLUSTER SETUP
    make kind_setup
    SETUP_EXIT_VALUE=$?

    ## INSTALL ALL REQUIRED DEPENDENCIES
    make -C ../../python install_dev
    INSTALL_EXIT_VALUE=$?

    ## RUNNING TESTS AND CAPTURING ERROR
    make test_notebooks
    TEST_EXIT_VALUE=$?
else
    echo "Existing kind cluster or failure starting - ${KIND_EXIT_VALUE}"
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

