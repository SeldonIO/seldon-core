#!/usr/bin/env bash

set -o nounset
set -o errexit
set -o pipefail

STARTUP_DIR="$( cd "$( dirname "$0" )" && pwd )"

if [ "$#" -ne 2 ]; then
    echo "You must enter <existing-python-version> <new-python-version>"
    exit 1
fi

OLD_VERSION=$1
NEW_VERSION=$2

declare -a paths=('./examples/*.ipynb' './doc/*.md' './docs/*.md' './example/*Makefile' './integrations/*Makefile')
declare -a versions=('2' '3' '36' '37')

cd ../../..

for i in "${paths[@]}"
do
    for PYTHON_VERSION in "${versions[@]}"
    do
	echo "Updating python version ${PYTHON_VERSION} in $i from ${OLD_VERSION} to ${NEW_VERSION}"
	find . -type f -path "$i" -exec grep -l seldon-core-s2i-python${PYTHON_VERSION}:${OLD_VERSION} \{\} \; | xargs -n1 -r sed -i "s/seldon-core-s2i-python${PYTHON_VERSION}:${OLD_VERSION}/seldon-core-s2i-python${PYTHON_VERSION}:${NEW_VERSION}/g"
    done
done



