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
    make s2i_build_base_images
    S2I_EXIT_VALUE=$?

    # CREATE PROTOS
    make build_protos
    PROTOS_EXIT_VALUE=$?

    # KIND CLUSTER SETUP
    make kind_setup
    SETUP_EXIT_VALUE=$?

    ## INSTALL ALL REQUIRED DEPENDENCIES
    #make -C ../../python install-dev
    #INSTALL_EXIT_VALUE=$?

    ## RUNNING TESTS AND CAPTURING ERROR
    #make test
    #TEST_EXIT_VALUE=$?
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

