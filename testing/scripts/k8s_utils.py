from subprocess import run, Popen
import signal
import subprocess
import os
import time

from retrying import retry

API_AMBASSADOR = "localhost:8003"


def wait_for_shutdown(deploymentName):
    ret = run("kubectl get deploy/" + deploymentName, shell=True)
    while ret.returncode == 0:
        time.sleep(1)
        ret = run("kubectl get deploy/" + deploymentName, shell=True)


def get_seldon_version():
    completedProcess = Popen(
        "cat ../../version.txt", shell=True, stdout=subprocess.PIPE
    )
    output = completedProcess.stdout.readline()
    version = output.decode("utf-8").strip()
    return version
