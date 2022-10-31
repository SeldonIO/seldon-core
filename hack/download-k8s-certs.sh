#!/bin/sh
set -eo pipefail

# <namespace> <secret> [<folder>]
# If folder not provided the secret contents will be stored in a folder with same name as secret.
# If used with TLS certificate secrets can be used in the seldon CLI config settings to connect to control and data plane.

namespace=$1
secret_name=$2
directory_to_save=${3-$2}

mkdir -p $directory_to_save

values=$(kubectl --namespace $namespace get secrets $secret_name -o json \
    | jq -r '.data | keys[] as $k | "\($k):\(.[$k])"')
for file_content in $values
do
    file_name=$(echo $file_content | cut -d':' -f1)
    echo $file_content | cut -d':' -f2 | base64 --decode > $directory_to_save/$file_name
    echo "Created $directory_to_save/$file_name"
done
