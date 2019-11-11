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


echo "Files changed in python folder:"
git --no-pager diff --exit-code --name-only origin/master python
PYTHON_MODIFIED=$?
if [[ $PYTHON_MODIFIED -gt 0 ]]; then 
    (cd wrappers/s2i/python/build_scripts \
        && ./build_all_local.sh \
        && ./push_all.sh)
    PYTHON_EXIT_VALUE=$?
else
    echo "SKIPPING PYTHON IMAGE BUILD..."
    PYTHON_EXIT_VALUE=0
fi

echo "Files changed in operator folder:"
git --no-pager diff --exit-code --name-only origin/master operator
OPERATOR_MODIFIED=$?
if [[ $OPERATOR_MODIFIED -gt 0 ]]; then
    make \
        -C operator \
        docker-build \
        docker-push
    OPERATOR_EXIT_VALUE=$?
else
    echo "SKIPPING OPERATOR IMAGE BUILD..."
    OPERATOR_EXIT_VALUE=0
fi

echo "Files changed in engine folder:"
git --no-pager diff --exit-code --name-only origin/master engine
ENGINE_MODIFIED=$?
if [[ $ENGINE_MODIFIED -gt 0 ]]; then
    make \
        -c testing/scripts
        build_protos
    make \
        -C engine \
        build_image \
        push_to_registry
    ENGINE_EXIT_VALUE=$?
else
    echo "SKIPPING ENGINE IMAGE BUILD..."
    ENGINE_EXIT_VALUE=0
fi



#######################################
# EXIT STOPS COMMANDS FROM HERE ONWARDS
set -o errexit

# CLEANING DOCKER
docker ps -aq | xargs -r docker rm -f || true
service docker stop || true

# NOW THAT WE'VE CLEANED WE CAN EXIT ON TEST EXIT VALUE
exit $((${PYTHON_EXIT_VALUE} \
    + ${OPERATOR_EXIT_VALUE} \
    + ${ENGINE_EXIT_VALUE}))

