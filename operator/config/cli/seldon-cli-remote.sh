#!/bin/sh

# This script is used to run commands on the seldon-cli pod in the seldon-mesh namespace
# Usage: ./seldon-cli-remote.sh <command>
# command: the command to run on the seldon-cli pod in the format: `seldon` <args> 

N="${NAMESPACE:-seldon-mesh}"

kubectl exec seldon-cli -n ${N} -- $*