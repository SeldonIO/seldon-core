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


(cd wrappers/s2i/python/build_scripts \
    && ./build_all_local.sh \
    && ./push_all.sh)
PYTHON_EXIT_VALUE=$?

make \
    -C operator \
    docker-build \
    docker-push
OPERATOR_EXIT_VALUE=$?

make \
    -C executor \
    docker-build \
    docker-push
EXECUTOR_EXIT_VALUE=$?

make \
    -C testing/scripts \
    build_protos
make \
    -C engine \
    build_image \
    push_to_registry
ENGINE_EXIT_VALUE=$?


#######################################
# EXIT STOPS COMMANDS FROM HERE ONWARDS
set -o errexit

# CLEANING DOCKER
docker ps -aq | xargs -r docker rm -f || true
service docker stop || true

# NOW THAT WE'VE CLEANED WE CAN EXIT ON TEST EXIT VALUE
echo "Python exit value: $PYTHON_EXIT_VALUE"
echo "Operator exit value: $OPERATOR_EXIT_VALUE"
echo "Engine exit value: $ENGINE_EXIT_VALUE"
echo "Executor exit value: $EXECUTOR_EXIT_VALUE"

exit $((${PYTHON_EXIT_VALUE} \
    + ${OPERATOR_EXIT_VALUE} \
    + ${ENGINE_EXIT_VALUE} \
    + ${EXECUTOR_EXIT_VALUE}))


