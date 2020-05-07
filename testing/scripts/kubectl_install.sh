#!/usr/bin/env bash
#
# kubectl install
#
# kubectl Release info:
# version: v1.17.0
#
set -o nounset
set -o errexit
set -o pipefail
set -o noclobber
set -o noglob
#set -o xtrace

FILE=tempresources/kubectl

# If it doesn't already exist, install kubectl binary into tempresources directory
[ ! -f "$FILE" ] && \
	mkdir -p tempresources && \
	curl -o tempresources/kubectl -LO https://storage.googleapis.com/kubernetes-release/release/v1.17.0/bin/linux/amd64/kubectl && \
	chmod +x tempresources/kubectl
exit 0
