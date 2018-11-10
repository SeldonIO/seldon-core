#!/bin/bash -e
#
# S2I run script for the 'seldon-core-s2i-nodejs' image.
# The run script executes the server that runs your application.
#
# For more information see the documentation:
#	https://github.com/openshift/source-to-image/blob/master/docs/builder_image.md
#

#check environment vars
if [[ -z "$MODEL_NAME" || -z "$API_TYPE" || -z "$SERVICE_TYPE" || -z "$PERSISTENCE" ]]; then

    echo "Failed to find required env vars MODEL_NAME, API_TYPE, SERVICE_TYPE, PERSISTENCE"
    exit 1

else
    cd /microservice
    echo "starting microservice"
    exec node microservice.js --model $MODEL_NAME --api $API_TYPE --service $SERVICE_TYPE --persistence $PERSISTENCE

fi

