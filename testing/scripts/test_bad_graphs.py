import pytest
import time
import subprocess
from subprocess import run
import json

# Test updating a model with a new image version as the only change
def test_duplicate_predictor_name():
    run("kubectl apply -f ../resources/bad_duplicate_predictor_name.json", shell=True, check=True)
    time.sleep(10)
    completedProcess = run("kubectl get sdep dupname -o json", shell=True, check=True, stdout=subprocess.PIPE)
    jStr = completedProcess.stdout
    j = json.loads(jStr)
    assert j["status"]["state"] == "Failed"
    run("kubectl delete sdep --all", shell=True)    
    
