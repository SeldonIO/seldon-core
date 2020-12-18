#!/bin/ash

if [ $# -eq 0 ]
  then
    echo "USAGE: deploy.sh NAMESPACE MODEL_URI MODEL_VERSION"
fi

echo "namespace=${1} model.uri=${2} model.version=${3}"

if helm --namespace=${1} status rsi ; then
    helm --namespace=${1} upgrade rsi /charts/seldon-deployment --set model.uri=${2} --set model.version=${3}
else
    helm --namespace=${1} install rsi /charts/seldon-deployment --set model.uri=${2} --set model.version=${3}
fi
