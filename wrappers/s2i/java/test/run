#!/bin/bash
#
# The 'run' performs a simple test that verifies the S2I image.
# The main focus here is to exercise the S2I scripts.
#
# For more information see the documentation:
# https://github.com/openshift/source-to-image/blob/master/docs/builder_image.md
#
# IMAGE_NAME_BUILD specifies a name of the candidate build image used for testing.
# The image has to be available before this script is executed.
# IMAGE_NAME_RUNTIME specifies a name of the candidate runtime image used for testing.
# The image has to be available before this script is executed.
#
IMAGE_NAME_BUILD=${IMAGE_NAME_BUILD-seldon-core-s2i-java-build-candidate}
IMAGE_NAME_RUNTIME=${IMAGE_NAME_RUNTIME-seldon-core-s2i-java-runtime-candidate}

# Determining system utility executables (darwin compatibility check)
READLINK_EXEC="readlink"
MKTEMP_EXEC="mktemp"
if [[ "$OSTYPE" =~ 'darwin' ]]; then
  ! type -a "greadlink" &>"/dev/null" || READLINK_EXEC="greadlink"
  ! type -a "gmktemp" &>"/dev/null" || MKTEMP_EXEC="gmktemp"
fi

test_dir="$($READLINK_EXEC -zf $(dirname "${BASH_SOURCE[0]}"))"
image_dir=$($READLINK_EXEC -zf ${test_dir}/..)


# Since we built the candidate image locally, we don't want S2I to attempt to pull
# it from Docker hub
s2i_args="--pull-policy=never --loglevel=2"

# Port the image exposes service to be tested
test_port=5000

image_exists() {
  docker inspect $1 &>/dev/null
}

container_exists() {
  image_exists $(cat $cid_file)
}

container_ip() {
  if [ ! -z "$DOCKER_HOST" ] && [[ "$OSTYPE" =~ 'darwin' ]]; then
    docker-machine ip
  else
    docker inspect --format="{{ .NetworkSettings.IPAddress }}" $(cat $cid_file)
  fi
}

container_port() {
  if [ ! -z "$DOCKER_HOST" ] && [[ "$OSTYPE" =~ 'darwin' ]]; then
    docker inspect --format="{{(index .NetworkSettings.Ports \"$test_port/tcp\" 0).HostPort}}" "$(cat "${cid_file}")"
  else
    echo $test_port
  fi
}

run_s2i_build() {
    prefix=$1
    s2i build ${s2i_args} file://${test_dir}/${prefix}-template-app  ${IMAGE_NAME_BUILD} --runtime-image ${IMAGE_NAME_RUNTIME} ${IMAGE_NAME_RUNTIME}-testapp 
}

prepare() {
  prefix=$1
  if ! image_exists ${IMAGE_NAME_BUILD}; then
    echo "ERROR: The image ${IMAGE_NAME_BUILD} must exist before this script is executed."
    exit 1
  fi
  if ! image_exists ${IMAGE_NAME_RUNTIME}; then
    echo "ERROR: The image ${IMAGE_NAME_RUNTIME} must exist before this script is executed."
    exit 1
  fi
  # s2i build requires the application is a valid 'Git' repository
  pushd ${test_dir}/${prefix}-template-app >/dev/null
  git init
  git config user.email "build@localhost" && git config user.name "builder"
  git add -A && git commit -m "Sample commit"
  popd >/dev/null
  run_s2i_build ${prefix}
}

run_test_application() {
  docker run --rm --cidfile=${cid_file} -p ${test_port} ${IMAGE_NAME_RUNTIME}-testapp
}

cleanup() {
  if [ -f $cid_file ]; then
    if container_exists; then
      docker stop $(cat $cid_file)
    fi
  fi
  if image_exists ${IMAGE_NAME_RUNTIME}-testapp; then
    docker rmi ${IMAGE_NAME_RUNTIME}-testapp
  fi
}

check_result() {
  local result="$1"
  if [[ "$result" != "0" ]]; then
    echo "S2I image '${IMAGE_NAME_RUNTIME}' test FAILED (exit code: ${result})"
    cleanup
    exit $result
  fi
}

wait_for_cid() {
  local max_attempts=10
  local sleep_time=1
  local attempt=1
  local result=1
  while [ $attempt -le $max_attempts ]; do
    [ -f $cid_file ] && break
    echo "Waiting for container to start..."
    attempt=$(( $attempt + 1 ))
    sleep $sleep_time
  done
}

test_usage() {
  echo "Testing 's2i usage'..."
  s2i usage ${s2i_args} ${IMAGE_NAME_BUILD} &>/dev/null
}

test_seldonMessage() {
  local endpoint=$1
  echo "Testing $type HTTP connection (http://$(container_ip):$(container_port)${endpoint})"
  local max_attempts=10
  local sleep_time=1
  local attempt=1
  local result=1
  while [ $attempt -le $max_attempts ]; do
    data='{"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}}'
    echo "Sending GET request to http://$(container_ip):$(container_port)${endpoint}"
    response_code=$(curl -s -w %{http_code} -o /dev/null -d "json=${data}" http://$(container_ip):$(container_port)${endpoint})
    status=$?
    if [ $status -eq 0 ]; then
      if [ $response_code -eq 200 ]; then
        result=0
      fi
      break
    fi
    attempt=$(( $attempt + 1 ))
    sleep $sleep_time
  done
  return $result
}

test_feedback() {
  local endpoint=$1
  echo "Testing $type HTTP connection (http://$(container_ip):$(container_port)${endpoint})"
  local max_attempts=10
  local sleep_time=1
  local attempt=1
  local result=1
  while [ $attempt -le $max_attempts ]; do
    data='{"request":{"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}},"response":{"meta":{"routing":{"router":0}},"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}},"reward":1}'
    echo "Sending GET request to http://$(container_ip):$(container_port)${endpoint}"
    response_code=$(curl -s -w %{http_code} -o /dev/null -d "json=${data}" http://$(container_ip):$(container_port)${endpoint})
    status=$?
    if [ $status -eq 0 ]; then
      if [ $response_code -eq 200 ]; then
        result=0
      fi
      break
    fi
    attempt=$(( $attempt + 1 ))
    sleep $sleep_time
  done
  return $result
}

# Build the application image twice to ensure the 'save-artifacts' and
# 'restore-artifacts' scripts are working properly
array=( 'model' )
for i in "${array[@]}"
do
    cid_file=$($MKTEMP_EXEC -u --suffix=.cid)
    echo $i

    prepare ${i}
    run_s2i_build ${i}
    check_result $?

    # Verify the 'usage' script is working properly
    test_usage
    check_result $?

    # Verify that the HTTP connection can be established to test application container
    run_test_application &

    # Wait for the container to write its CID file
    wait_for_cid

    if [ "$i" = "model" ]; then
	test_seldonMessage "/predict"
	check_result $?
    elif [ "$i" = "router" ]; then
	test_seldonMessage "/route"
	check_result $?
	test_feedback "/send-feedback"
	check_result $?
    elif [ "$i" = "transformer" ]; then
	test_seldonMessage "/transform-input"
	check_result $?
	test_seldonMessage "/transform-output"
	check_result $?
    fi



    cleanup
done
