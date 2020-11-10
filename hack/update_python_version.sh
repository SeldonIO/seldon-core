#!/usr/bin/env bash

set -o nounset
set -o errexit
set -o pipefail

STARTUP_DIR="$( cd "$( dirname "$0" )" && pwd )"

if [ "$#" -ne 1 ]; then
    echo "You must enter <new-python-version>"
    exit 1
fi

NEW_VERSION=$1

cd ${STARTUP_DIR}/.. \
  && find . \
      -type f \( -path './examples/*.ipynb' -or -path './doc/*.md'  -or -path './examples/*Makefile' \) \
      -exec grep -El 'seldon-core-s2i-python3[67]?:[^\$\n ]+' \{\} \; | xargs -n1 -r sed -Ei "s/(seldon-core-s2i-python3)([67]?:)([^\$\n ]+)/\1\2${NEW_VERSION}/g"



