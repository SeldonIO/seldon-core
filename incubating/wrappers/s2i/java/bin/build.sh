#!/usr/bin/env bash
set -x

#
# Print the usage and exit.
#
function usage () {
  echo "Usage: $0 [OPTIONS] VERSION"
  echo "\r\n"
  echo "Options: "
  echo " --build-source-image   (OPTIONAL) override the build image source image"
  echo " --build-runtime-image  (OPTIONAL) override the runtime image source image"
  exit 1
}

# Parse the arguments.
# loop until last - 1, we assume the last arg is the model name
while [[ $# -ge 2 ]]
do
  key="$1"
  case ${key} in
  --build-source-image)
    BUILD_SOURCE_IMAGE="$2"
    shift # past key
    shift # past value
  ;;
  --runtime-source-image)
    RUNTIME_SOURCE_IMAGE="$2"
    shift # past key
    shift # past value
  ;;
  *)
	usage
  ;;
esac
done

# assign last parameter to model
VERSION=${@:${#@}}
IMAGE_BUILD_BUILD_ARGS=""
IMAGE_RUNTIME_BUILD_ARGS=""

# Validate the arguments.
if [[ -z "$VERSION" ]]; then
  echo "Missing argument: VERSION."
  usage
fi

# Prepare Docker build args ( if supplied )
if ! [[ -z "$BUILD_SOURCE_IMAGE" ]]; then
  IMAGE_BUILD_BUILD_ARGS=$(echo "$IMAGE_BUILD_BUILD_ARGS --build-arg IMAGE_SOURCE=$BUILD_SOURCE_IMAGE" | awk '{$1=$1};1')
fi
if ! [[ -z "$RUNTIME_SOURCE_IMAGE" ]]; then
  IMAGE_RUNTIME_BUILD_ARGS=$(echo "$IMAGE_RUNTIME_BUILD_ARGS --build-arg IMAGE_SOURCE=$RUNTIME_SOURCE_IMAGE" | awk '{$1=$1};1')
fi

IMAGE_NAME_BUILD=docker.io/seldonio/seldon-core-s2i-java-build:${VERSION}
IMAGE_NAME_RUNTIME=docker.io/seldonio/seldon-core-s2i-java-runtime:${VERSION}

echo "S2I build image: building..."
docker build \
    -f Dockerfile.build \
    -t ${IMAGE_NAME_BUILD} \
    ${IMAGE_BUILD_BUILD_ARGS} \
    .
echo "S2I build image: check Java version..."
docker run --entrypoint "/bin/bash" ${IMAGE_NAME_BUILD} "-c" "java --version"

echo "S2I runtime image: building..."
docker build \
    -f Dockerfile.runtime \
    -t ${IMAGE_NAME_RUNTIME} \
    ${IMAGE_RUNTIME_BUILD_ARGS} \
    .
echo "S2I runtime image: check Java version..."
docker run --entrypoint "/bin/bash" ${IMAGE_NAME_RUNTIME} "-c" "java --version"
