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


# ONLY RUN IF PYTHON HAS BE MODIFIED
PYTHON_MODIFIED=`git diff --exit-code --quiet master python/`
if [[ $PYTHON_MODIFIED -gt 0 ]]; then 
    (cd wrappers/s2i/python/build_scripts \
        && ./build_all_local.sh \
        && ./push_all.sh)
    PYTHON_EXIT_VALUET=$?
else
    echo "SKIPPING PYTHON IMAGE BUILD..."
fi

# MORE EFFICIENT CLUSTER SETUP
OPERATOR_MODIFIED=`git diff --exit-code --quiet master operator/`
if [[ $OPERATOR_MODIFIED -gt 0 ]]; then
    make \
        -C operator \
        docker-build \
        docker-push
    OPERATOR_EXIT_VALUE=$?
else
    echo "SKIPPING OPERATOR IMAGE BUILD..."
fi

ENGINE_MODIFIED=`git diff --exit-code --quiet master engine/`
if [[ $ENGINE_MODIFIED -gt 0 ]]; then
    make \
        -C engine \
        build_image \
        push_to_registry
    ENGINE_EXIT_VALUE=$?
else
    echo "SKIPPING ENGINE IMAGE BUILD..."
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

