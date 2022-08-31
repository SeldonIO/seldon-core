#!/bin/bash

set -o nounset -o errexit -o pipefail

# Retries a command a with backoff.
# Based on https://stackoverflow.com/a/8351489/376366
# and https://gist.github.com/fernandoacorreia/b4fa9ae88c67fa6759d271b743e96063
#
# The retry count is given by ATTEMPTS (default 7), the
# initial backoff timeout is given by TIMEOUT in seconds
# (default 1). With default settings, it will try for about 1 minute.
#
# Successive backoffs double the timeout.
function with_backoff {
  local max_attempts=${ATTEMPTS-7}
  local timeout=${TIMEOUT-1}
  local attempt=1
  local exitCode=0

  while (( $attempt < $max_attempts ))
  do
    set +e
    "$@"
    exitCode=$?
    set -e

    if [[ $exitCode == 0 ]]
    then
      break
    fi

    echo "Attempt $attempt/$max_attempts failed ($@). Retrying in ${timeout}s..." 1>&2
    sleep $timeout
    attempt=$(( attempt + 1 ))
    timeout=$(( timeout * 2 ))
  done

  if [[ $exitCode != 0 ]]
  then
    echo "Attempt $attempt/$max_attempts failed ($@). Maximum retries exceeded." 1>&2
  fi

  return $exitCode
}
