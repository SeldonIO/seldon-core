#!/bin/bash -e
#
# S2I run script for the 'seldon-core-s2i-java-jni' image.
# The run script executes the server that runs your application.
#
# For more information see the documentation:
#	https://github.com/openshift/source-to-image/blob/master/docs/builder_image.md
#

# Check environment vars
if [[ -z "$JAVA_IMPORT_PATH" || -z "$SERVICE_TYPE" || -z "$PERSISTENCE" ]]; then

    echo "Failed to find required env vars JAVA_IMPORT_PATH, SERVICE_TYPE, PERSISTENCE"
    exit 1
fi

PAYLOAD_PASSTHROUGH=${PAYLOAD_PASSTHROUGH:-true}
MODEL_NAME=${MODEL_NAME:-"JavaJNIServer"}
JAVA_JAR_PATH=${JAVA_JAR_PATH:-"./model.jar"}

echo "Starting microservice"
exec \
  env PAYLOAD_PASSTHROUGH="$PAYLOAD_PASSTHROUGH" \
  env JAVA_JAR_PATH="$JAVA_JAR_PATH" \
  env JAVA_IMPORT_PATH="$JAVA_IMPORT_PATH" \
  seldon-core-microservice \
    $MODEL_NAME \
    --service-type $SERVICE_TYPE \
    --persistence $PERSISTENCE

