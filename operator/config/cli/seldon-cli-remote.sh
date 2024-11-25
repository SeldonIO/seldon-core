#!/bin/sh
N="${NAMESPACE:-seldon-mesh}"

echo "Running command: $*, in namespace: $N" 
kubectl exec seldon-cli -n ${N} -- $*