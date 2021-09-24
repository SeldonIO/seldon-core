#!/usr/bin/env bash
set -o errexit
set -o nounset
set -o pipefail

# Path to Envoy
ENVOY=${ENVOY:-/usr/local/bin/envoy}

## Start Envoy with sample bootstrap config.
${ENVOY} -c hack/bootstrap.yaml # --drain-time-s 1  # -l debug
