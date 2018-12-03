#!/bin/bash

release=${1:-"master"}

echo Downloading proto files for ${release}

base=https://raw.githubusercontent.com/tensorflow
tensorflow_base=${base}/tensorflow/${release}

base_folder=tensorflow/core/framework/
mkdir -p ${base_folder}

curl -s ${tensorflow_base}/tensorflow/core/framework/types.proto > ${base_folder}/types.proto
curl -s ${tensorflow_base}/tensorflow/core/framework/resource_handle.proto > ${base_folder}/resource_handle.proto
curl -s ${tensorflow_base}/tensorflow/core/framework/tensor_shape.proto > ${base_folder}/tensor_shape.proto
curl -s ${tensorflow_base}/tensorflow/core/framework/tensor.proto > ${base_folder}/tensor.proto

