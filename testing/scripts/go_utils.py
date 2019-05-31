import pytest
from subprocess import run,Popen
import signal
import subprocess
import os
import time


def build_go_images(version):
    #run("rm -rf ${PWD}/go && export GOPATH=${PWD}/go && mkdir -p $GOPATH/src/github.com/seldonio/ && cd ./go/src/github.com/seldonio && git clone https://github.com/SeldonIO/seldon-operator.git && cd seldon-operator && make docker-build docker-push-local-private VERSION="+version, shell=True, check=True)
    run("rm -rf ${PWD}/go && export GOPATH=${PWD}/go && mkdir -p $GOPATH/src/github.com/seldonio/ && cd ./go/src/github.com/seldonio && git clone --single-branch --branch ambassador_update https://github.com/cliveseldon/seldon-operator.git && cd seldon-operator && make docker-build docker-push-local-private VERSION="+version, shell=True, check=True)    

