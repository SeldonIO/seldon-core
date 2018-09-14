#!/bin/bash -e
#
# S2I assemble script for the 'seldon-core-s2i-nodejs' image.
# The 'assemble' script builds your application source so that it is ready to run.
#
# For more information refer to the documentation:
#	https://github.com/openshift/source-to-image/blob/master/docs/builder_image.md
#

# If the 'seldon-core-s2i-r' assemble script is executed with the '-h' flag, print the usage.
if [[ "$1" == "-h" ]]; then
	exec /usr/libexec/s2i/usage
fi


if [[ -z "$MODEL_NAME" ]]; then

    echo "Failed to find required env var MODEL_NAME"
    exit 1
fi

if [[ -z "$API_TYPE" ]]; then

    echo "Failed to find required env var API_TYPE, should be either REST or GRPC."
    exit 1
fi

if [[ -z "$SERVICE_TYPE" ]]; then

    echo "Failed to find required env var SERVICE_TYPE, should be one of MODEL, ROUTER, TRANSFORMER, COMBINER."
    exit 1
fi

if [[ -z "$PERSISTENCE" ]]; then

    echo "Failed to find required env var PERSISTENCE, should be 0 or 1."
    exit 1
fi


cd /microservice
mkdir model

# Restore artifacts from the previous build (if they exist).
#
if [ "$(ls /tmp/artifacts/ 2>/dev/null)" ]; then
  echo "---> Restoring build artifacts..."
  mv /tmp/artifacts/. ./
fi

echo "---> Installing application source..."
cp -Rf /tmp/src/. ./model/

if [[ -f package.json ]]; then
  echo "---> Installing dependencies ..."
  cd model
  rm -rf node_modules
  rm -f package-lock.json
  npm install
fi

