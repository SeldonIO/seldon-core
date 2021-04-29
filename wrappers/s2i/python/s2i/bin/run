#!/bin/bash -e
#
# S2I run script for the 'seldon-core-s2i-python' image.
# The run script executes the server that runs your application.
#
# For more information see the documentation:
#	https://github.com/openshift/source-to-image/blob/master/docs/builder_image.md
#

function _get_conda_envs() {
    # Printed value will be list of Conda envs
    conda env list | tail -n +3 | head -n -1 | cut -d' ' -f1
}

function _is_env_present() {
    # $1 is the Conda env we are probing
    # Return value will be 0 if present (TRUE)

    local _conda_envs=$(_get_conda_envs)
    echo $_conda_envs | grep -qw $1
    return $?
}

#check environment vars
if [[ -z "$MODEL_NAME" || -z "$SERVICE_TYPE" ]]; then

    echo "Failed to find required env vars MODEL_NAME, SERVICE_TYPE"
    exit 1

else
    cd /microservice

    if [[ -x before-run ]]; then
        echo "Executing before-run script"
        ./before-run
    fi

    CONDA_ENV_NAME=${CONDA_ENV_NAME:-microservice}
    if _is_env_present $CONDA_ENV_NAME; then
        echo "Activating Conda environment '$CONDA_ENV_NAME'"
        # We need to initialise Conda so that it can run on /bin/sh
        source /etc/profile.d/conda.sh
        conda activate $CONDA_ENV_NAME
    fi

    echo "starting microservice"
    if [ -n "$PERSISTENCE" ]; then
      exec seldon-core-microservice $MODEL_NAME --service-type $SERVICE_TYPE --persistence $PERSISTENCE
    else
      exec seldon-core-microservice $MODEL_NAME --service-type $SERVICE_TYPE
    fi
fi
