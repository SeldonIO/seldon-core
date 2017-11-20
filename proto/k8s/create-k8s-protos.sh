#!/bin/bash

# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

release=${1:-"master"}

echo Downloading proto files for ${release}

mkdir -p k8s.io/apimachinery/pkg/api/resource
mkdir -p k8s.io/apimachinery/pkg/apis/meta/v1
mkdir -p k8s.io/apimachinery/pkg/util/intstr
mkdir -p k8s.io/apimachinery/pkg/runtime/schema
mkdir -p k8s.io/apis/meta/v1

base=https://raw.githubusercontent.com/kubernetes
machinery_base=${base}/apimachinery/${release}
curl -s ${machinery_base}/pkg/api/resource/generated.proto \
	> k8s.io/apimachinery/pkg/api/resource/generated.proto

curl -s ${machinery_base}/pkg/apis/meta/v1/generated.proto \
	> k8s.io/apimachinery/pkg/apis/meta/v1/generated.proto

curl -s ${machinery_base}/pkg/util/intstr/generated.proto \
	> k8s.io/apimachinery/pkg/util/intstr/generated.proto

curl -s ${machinery_base}/pkg/runtime/generated.proto \
	> k8s.io/apimachinery/pkg/runtime/generated.proto

curl -s ${machinery_base}/pkg/runtime/schema/generated.proto \
	> k8s.io/apimachinery/pkg/runtime/schema/generated.proto

# There are currently no release branches for this file.
curl -s ${base}/api/master/core/v1/generated.proto > v1.proto


# The format here is <file-name>;<generated-class-name>
files="v1.proto;V1 \
       k8s.io/apimachinery/pkg/api/resource/generated.proto;Resource \
       k8s.io/apimachinery/pkg/apis/meta/v1/generated.proto;Meta \
       k8s.io/apimachinery/pkg/runtime/generated.proto;Runtime \
       k8s.io/apimachinery/pkg/runtime/schema/generated.proto;RuntimeSchema \
       k8s.io/apimachinery/pkg/util/intstr/generated.proto;IntStr"

proto_files=""

echo 'Munging proto file packages'

# This is a little hacky, but we know the go_package directive is in the
# right place, so add a marker, and then append more package declarations.
# Sorry, I like perl.
for info in ${files}; do
  file=$(echo ${info} | cut -d ";" -f 1)
  class=$(echo ${info} | cut -d ";" -f 2)
  proto_files="${file} ${proto_files}"
  perl -pi -e \
    's/option go_package = "(.*)";/option go_package = "$1";\n\/\/ PKG/' \
    ${file} 
  perl -pi -e \
    's/\/\/ PKG/\/\/ PKG\noption java_package = "io.kubernetes.client.proto";/' \
    ${file}
  perl -pi -e \
    "s/\/\/ PKG/\/\/ PKG\noption java_outer_classname = \"${class}\";/" \
    ${file}
  
  # Other package declarations can go here.
done

#echo "Generating code for $1"
#${proto} -I${dir} ${proto_files} --${1}_out=${2}
