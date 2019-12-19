#!/usr/bin/env bash

set -o nounset
set -o errexit
set -o pipefail

# Run s2i build to create base images
make s2i_build_base_images

make kind_create_cluster
make kind_build_images
make kind_setup
