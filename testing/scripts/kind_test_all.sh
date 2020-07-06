#!/bin/bash

set -o errexit
set -o pipefail
set -o nounset

# Collect Environment Variables
TESTS_TO_RUN="${SELDON_E2E_TESTS_TO_RUN:-all}"
echo "Test run type is: [$TESTS_TO_RUN]"

#########################################################################
#                          STARTING DOCKER DAEMON                       #
#########################################################################

# This set up follows the best practices outlined by d2iq in their setup
# More at https://d2iq.com/blog/running-kind-inside-a-kubernetes-cluster-for-continuous-integration

## Determine cgroup parent for docker daemon.
## We need to make sure cgroups created by the docker daemon do not
## interfere with other cgroups on the host, and do not leak after this
## container is terminated.
#if [ -f /sys/fs/cgroup/systemd/release_agent ]; then
#  # This means the user has bind mounted host /sys/fs/cgroup to the
#  # same location in the container (e.g., using the following docker
#  # run flags: `-v /sys/fs/cgroup:/sys/fs/cgroup`). In this case, we
#  # need to make sure the docker daemon in the container does not
#  # pollute the host cgroups hierarchy.
#  # Note that `release_agent` file is only created at the root of a
#  # cgroup hierarchy.
#  CGROUP_PARENT="$(grep systemd /proc/self/cgroup | cut -d: -f3)/docker"
#else
#  CGROUP_PARENT="/docker"
#
#  # For each cgroup subsystem, Docker does a bind mount from the
#  # current cgroup to the root of the cgroup subsystem. For instance:
#  #   /sys/fs/cgroup/memory/docker/<cid> -> /sys/fs/cgroup/memory
#  #
#  # This will confuse some system software that manipulate cgroups
#  # (e.g., kubelet/cadvisor, etc.) sometimes because
#  # `/proc/<pid>/cgroup` is not affected by the bind mount. The
#  # following is a workaround to recreate the original cgroup
#  # environment by doing another bind mount for each subsystem.
#  CURRENT_CGROUP=$(grep systemd /proc/self/cgroup | cut -d: -f3)
#  CGROUP_SUBSYSTEMS=$(findmnt -lun -o source,target -t cgroup | grep "${CURRENT_CGROUP}" | awk '{print $2}')
#
#  echo "${CGROUP_SUBSYSTEMS}" |
#  while IFS= read -r SUBSYSTEM; do
#    mkdir -p "${SUBSYSTEM}${CURRENT_CGROUP}"
#    mount --bind "${SUBSYSTEM}" "${SUBSYSTEM}${CURRENT_CGROUP}"
#  done
#fi
#
#setsid dockerd \
#  --cgroup-parent="${CGROUP_PARENT}" \
#  --bip="${DOCKERD_BIP:-172.17.1.1/24}" \
#  --mtu="${DOCKERD_MTU:-1400}" \
#  --raw-logs \
#  ${DOCKER_ARGS:-} &
#
## Wait until dockerd is ready.
#until docker ps >/dev/null 2>&1
#do
#  echo "Waiting for dockerd..."
#  sleep 1
#done

sleep 3600 # Hold to debug

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

        echo "Files changed in python folder:"
        git --no-pager diff --exit-code --name-only origin/master ../../python
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

        echo "Build test models"
        make kind_build_test_models
        KIND_BUILD_EXIT_VALUE=$?
        if [[ $KIND_BUILD_EXIT_VALUE -gt 0 ]]; then
            echo "Kind build has errors"
            return 1
        fi

        echo "Files changed in prepackaged folder:"
        git --no-pager diff --exit-code --name-only origin/master ../../servers ../../integrations
        PREPACKAGED_MODIFIED=$?
        if [[ $PREPACKAGED_MODIFIED -gt 0 ]]; then
            make kind_build_prepackaged
            PREPACKAGED_EXIT_VALUE=$?
            if [[ $PREPACKAGED_EXIT_VALUE -gt 0 ]]; then
                echo "Prepackaged server build returned errors"
                return 1
            fi
        else
            echo "SKIPPING PREPACKAGED IMAGE BUILD..."
        fi

        echo "Files changed in alibi folder:"
        git --no-pager diff --exit-code --name-only origin/master ../../components/alibi-detect-server ../../components/alibi-explain-server/
        ALIBI_MODIFIED=$?
        if [[ $ALIBI_MODIFIED -gt 0 ]]; then
            make kind_build_alibi
            ALIBI_EXIT_VALUE=$?
            if [[ $ALIBI_EXIT_VALUE -gt 0 ]]; then
                echo "Alibi server build returned errors"
                return 1
            fi
        else
            echo "SKIPPING ALIBI IMAGE BUILD..."
        fi

        echo "Files changed in misc folders:"
        git --no-pager diff --exit-code --name-only origin/master ../../components/seldon-request-logger ../../components/storage-initializer
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
            make test_parallel test_sequential test_notebooks
        elif [ "$TESTS_TO_RUN" == "notebooks" ]; then
            make test_notebooks
        elif [ "$TESTS_TO_RUN" == "base" ]; then
            make test_parallel test_sequential
        elif [ "$TESTS_TO_RUN" == "parallel" ]; then
            make test_parallel
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
# DOCKER NO LONGER NEEDS TO BE CLEANED GIVEN THAT WE USE A DIFFERENT CGROUP
# Read more at https://github.com/kubernetes-sigs/kind/issues/303#issuecomment-654401156
# docker ps -aq | xargs -r docker rm -f || "Failed to stop remaining containers"
# service docker stop || echo "Failed to stop docker service"

RUN_EXIT_VALUE=0
echo "Finished tests exiting with value: $RUN_EXIT_VALUE"

# NOW THAT WE'VE CLEANED WE CAN EXIT ON RUN EXIT VALUE
exit ${RUN_EXIT_VALUE}

