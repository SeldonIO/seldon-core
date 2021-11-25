#!/bin/bash

set -o errexit
set -o pipefail
set -o nounset

# Collect Environment Variables
TESTS_TO_RUN="${SELDON_E2E_TESTS_TO_RUN:-all}"
echo "Test run type is: [$TESTS_TO_RUN]"

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

    run_end_to_end_tests() {

        echo "Files changed in python or wrapper folder:"
        git --no-pager diff --exit-code --name-only origin/master ../../python ../../wrappers/s2i/python
        PYTHON_MODIFIED=$?
        if [[ $PYTHON_MODIFIED -gt 0 ]]; then
            make s2i_build_base_images
            PYTHON_EXIT_VALUE=$?
            if [[ $PYTHON_EXIT_VALUE -gt 0 ]]; then
                echo "Python build returned errors"
                return 1
            fi
        else
            echo "SKIPPING PYTHON IMAGE BUILD..."
        fi

        echo "Files changed in operator folder:"
        git --no-pager diff --exit-code --name-only origin/master ../../operator
        OPERATOR_MODIFIED=$?
        if [[ $OPERATOR_MODIFIED -gt 0 ]]; then
            make kind_build_operator
            OPERATOR_EXIT_VALUE=$?
            if [[ $OPERATOR_EXIT_VALUE -gt 0 ]]; then
                echo "Operator build returned errors"
                return 1
            fi
        else
            echo "SKIPPING OPERATOR IMAGE BUILD..."
        fi

        echo "Files changed in engine folder:"
        git --no-pager diff --exit-code --name-only origin/master ../../engine
        ENGINE_MODIFIED=$?
        if [[ $ENGINE_MODIFIED -gt 0 ]]; then
            make build_protos
            PROTO_EXIT_VALUE=$?
            if [[ $PROTO_EXIT_VALUE -gt 0 ]]; then
                return 1
            fi
            make kind_build_engine
            ENGINE_EXIT_VALUE=$?
            if [[ $ENGINE_EXIT_VALUE -gt 0 ]]; then
                echo "Engine build returned errors"
                return 1
            fi
        else
            echo "SKIPPING ENGINE IMAGE BUILD..."
        fi

        echo "Files changed in executor folder:"
        git --no-pager diff --exit-code --name-only origin/master ../../executor
        EXECUTOR_MODIFIED=$?
        if [[ $EXECUTOR_MODIFIED -gt 0 ]]; then
            make kind_build_executor
            EXECUTOR_EXIT_VALUE=$?
            if [[ $EXECUTOR_EXIT_VALUE -gt 0 ]]; then
                echo "Executor build returned errors"
                return 1
            fi
        else
            echo "SKIPPING EXECUTOR IMAGE BUILD..."
        fi

        echo "Files changed in test models folder:"
        git --no-pager diff --exit-code --name-only origin/master ../../examples/models/mean_classifier/ ../../examples/models/testing/ ../../testing/docker/echo-model/ ../../testing/docker/fixed-model
        TEST_MODELS_MODIFIED=$?
        if [[ $TEST_MODELS_MODIFIED -gt 0 ]]; then
            make kind_build_test_models
            KIND_BUILD_EXIT_VALUE=$?
            if [[ $KIND_BUILD_EXIT_VALUE -gt 0 ]]; then
                echo "Kind build has errors"
                return 1
            fi
        else
            echo "SKIPPING TEST MODEL BUILD..."
        fi

        echo "Files changed in prepackaged, python, or wrapper folder:"
        git --no-pager diff --exit-code --name-only origin/master ../../servers ../../integrations
        PREPACKAGED_MODIFIED=$?
        if [[ $PREPACKAGED_MODIFIED -gt 0 ]] && [[ $PYTHON_MODIFIED -gt 0 ]]; then
            make kind_build_prepackaged
            PREPACKAGED_EXIT_VALUE=$?
            if [[ $PREPACKAGED_EXIT_VALUE -gt 0 ]]; then
                echo "Prepackaged server build returned errors"
                return 1
            fi
        else
            echo "SKIPPING PREPACKAGED IMAGE BUILD..."
        fi

        echo "Files changed in alibi explain folder:"
        git --no-pager diff --exit-code --name-only origin/master ../../components/alibi-explain-server/
        ALIBI_EXPLAIN_MODIFIED=$?
        if [[ $ALIBI_EXPLAIN_MODIFIED -gt 0 ]]; then
            make kind_build_alibi_explain
            ALIBI_EXPLAIN_EXIT_VALUE=$?
            if [[ $ALIBI_EXPLAIN_EXIT_VALUE -gt 0 ]]; then
                echo "Alibi Explain server build returned errors"
                return 1
            fi
        else
            echo "SKIPPING ALIBI EXPLAIN IMAGE BUILD..."
        fi

        echo "Files changed in alibi detect folder:"
        git --no-pager diff --exit-code --name-only origin/master ../../components/alibi-detect-server
        ALIBI_DETECT_MODIFIED=$?
        if [[ $ALIBI_DETECT_MODIFIED -gt 0 ]]; then
            make kind_build_alibi_detect
            ALIBI_DETECT_EXIT_VALUE=$?
            if [[ $ALIBI_DETECT_EXIT_VALUE -gt 0 ]]; then
                echo "Alibi Detect server build returned errors"
                return 1
            fi
        else
            echo "SKIPPING ALIBI DETECT IMAGE BUILD..."
        fi

        echo "Files changed in rclone storage initializer folder:"
        git --no-pager diff --exit-code --name-only origin/master ../../components/rclone-storage-initializer/
        RCLONE_STRORAGE_INITIALIZER_MODIFIED=$?
        if [[ $RCLONE_STRORAGE_INITIALIZER_MODIFIED -gt 0 ]]; then
            make kind_build_rclone_storage_initializer
            RCLONE_STRORAGE_INITIALIZER_EXIT_VALUE=$?
            if [[ $RCLONE_STRORAGE_INITIALIZER_EXIT_VALUE -gt 0 ]]; then
                echo "rclone storage initializer build returned errors"
                return 1
            fi
        else
            echo "SKIPPING RCLONE STORAGE INITIALIZER IMAGE BUILD..."
        fi

        echo "Files changed in misc folders:"
        git --no-pager diff --exit-code --name-only origin/master ../../components/seldon-request-logger ../../components/storage-initializer ../../components/routers/epsilon-greedy
        MISC_MODIFIED=$?
        if [[ $MISC_MODIFIED -gt 0 ]]; then
            make kind_build_misc
            MISC_EXIT_VALUE=$?
            if [[ $MISC_EXIT_VALUE -gt 0 ]]; then
                echo "Misc server build returned errors"
                return 1
            fi
        else
            echo "SKIPPING MISC IMAGE BUILD..."
        fi

        # KIND CLUSTER SETUP
        make kind_setup
        SETUP_EXIT_VALUE=$?
        if [[ $SETUP_EXIT_VALUE -gt 0 ]]; then
            echo "Kind setup returned errors"
            return 1
        fi

        ## INSTALL ALL REQUIRED DEPENDENCIES
        make -C ../../python install_dev
        INSTALL_EXIT_VALUE=$?
        if [[ $INSTALL_EXIT_VALUE -gt 0 ]]; then
            echo "Dependency installation returned errors"
            return 1
        fi

        ## RUNNING TESTS AND CAPTURING ERROR
        if [ "$TESTS_TO_RUN" == "all" ]; then
            make test_parallel
            make test_sequential
            make test_notebooks
        elif [ "$TESTS_TO_RUN" == "notebooks" ]; then
            make test_notebooks
        elif [ "$TESTS_TO_RUN" == "base" ]; then
            make test_parallel test_sequential
        elif [ "$TESTS_TO_RUN" == "parallel" ]; then
            make test_parallel
        elif [ "$TESTS_TO_RUN" == "benchmark" ]; then
            make test_benchmark
        fi
        TEST_EXIT_VALUE=$?
        if [[ $TEST_EXIT_VALUE -gt 0 ]]; then
            echo "Test returned errors"
            return 1
        fi

        # If we reach this point return success
        return 0
    }
    # We run the piece above
    run_end_to_end_tests
    RUN_EXIT_VALUE=$?

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

# NOW THAT WE'VE CLEANED WE CAN EXIT ON RUN EXIT VALUE
exit ${RUN_EXIT_VALUE}
