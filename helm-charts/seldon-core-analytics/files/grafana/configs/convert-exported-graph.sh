#!/usr/bin/env bash

set -o nounset
set -o errexit
set -o pipefail
set -o xtrace

STARTUP_DIR="$( cd "$( dirname "$0" )" && pwd )"

cd "${STARTUP_DIR}"


cat predictions-analytics-dashboard.json | sed 's/\${DS_PROMETHEUS}/prometheus/' | sed 's/DS_PROMETHEUS/DS_PROM/' > tt
mv tt predictions-analytics-dashboard.json
