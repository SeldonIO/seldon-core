#!/bin/bash

PYTHON_VERSION=""
TAG=""

s2i build . seldonio/seldon-core-s2i-python${PYTHON_VERSION}:${TAG} clean_text_transformer:0.1

