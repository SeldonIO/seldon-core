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

function build_push_python {
    (cd wrappers/s2i/python/build_scripts \
	    && ./build_all_local.sh \
	    && ./push_all.sh)
    PYTHON_EXIT_VALUE=$?
}

function build_push_operator {
    make \
	-C operator \
	docker-build \
	docker-push \
	docker-build-redhat \
	docker-push-redhat	
    OPERATOR_EXIT_VALUE=$?
}

function build_push_executor {
    make \
	-C executor \
	docker-build \
	docker-push \
	docker-build-redhat \
	docker-push-redhat	
    EXECUTOR_EXIT_VALUE=$?
}

function build_push_engine {
    make \
	-C testing/scripts \
	build_protos
    make \
	-C engine \
	build_image \
	push_to_registry \
	docker-build-redhat \
	docker-push-redhat	
    ENGINE_EXIT_VALUE=$?
}

function build_push_mock {
    make \
	-C examples/models/mean_classifier \
	build \
	push
    MOCK_MODEL_EXIT_VALUE=$?
}

function build_push_alibi_detect {
    make \
	-C components/alibi-detect-server \
	docker-build \
	docker-push 
    ALIBI_DETECT_EXIT_VALUE=$?
}

function build_push_request_logger {
    make \
	-C components/seldon-request-logger \
        build_image \
	push_image 
    LOGGER_EXIT_VALUE=$?
}

function build_push_sklearnserver {
    make \
	-C servers/sklearnserver \
        build \
	push 
    SKLEARN_EXIT_VALUE=$?
}

function build_push_mlflowserver {
    make \
	-C servers/mlflowserver \
        build \
	push 
    MLFLOW_EXIT_VALUE=$?
}

function build_push_xgboostserver {
    make \
	-C servers/xgboostserver \
        build \
	push 
    XGBOOST_EXIT_VALUE=$?
}

function build_push_tfproxy {
    make \
	-C servers/tfserving_proxy \
        build \
	push 
    TFPROXY_EXIT_VALUE=$?
}

function build_push_alibi_explainer {
    make \
	-C components/alibi-explain-server \
        docker-build \
	docker-push 
    EXPLAIN_EXIT_VALUE=$?
}

function build_push_storage_initializer {
    make \
	-C components/storage-initializer \
        docker-build \
	docker-push 
    STORAGE_INITIALIZER_EXIT_VALUE=$?
}

function build_push_mab {
    make \
	-C components/routers/epsilon-greedy \
        build \
	push 
    MAB_EXIT_VALUE=$?
}


build_push_python
build_push_operator
build_push_executor
build_push_engine
build_push_mock
build_push_alibi_detect
build_push_request_logger
build_push_sklearnserver
build_push_mlflowserver
build_push_xgboostserver
build_push_tfproxy
build_push_alibi_explainer
build_push_storage_initializer
build_push_mab

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
echo "Mock model exit value: $MOCK_MODEL_EXIT_VALUE"
echo "Alibi Detect exit value: $ALIBI_DETECT_EXIT_VALUE"
echo "Request Logger exit value: $LOGGER_EXIT_VALUE"
echo "Tensorflow Proxy exit value: $TFPROXY_EXIT_VALUE"
echo "MAB exit value: $MAB_EXIT_VALUE"

exit $((${PYTHON_EXIT_VALUE} \
    + ${OPERATOR_EXIT_VALUE} \
    + ${ENGINE_EXIT_VALUE} \
    + ${EXECUTOR_EXIT_VALUE} \
    + ${MOCK_MODEL_EXIT_VALUE} \
    + ${ALIBI_DETECT_EXIT_VALUE} \
    + ${LOGGER_EXIT_VALUE} \
    + ${SKLEARN_EXIT_VALUE} \
    + ${MLFLOW_EXIT_VALUE} \
    + ${XGBOOST_EXIT_VALUE} \
    + ${TFPROXY_EXIT_VALUE} \
    + ${STORAGE_INITIALIZER_EXIT_VALUE} \
    + ${MAB_EXIT_VALUE} \
    + ${EXPLAIN_EXIT_VALUE}))


