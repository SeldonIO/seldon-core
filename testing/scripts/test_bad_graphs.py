import pytest
import time
import subprocess
from subprocess import run
import json

def test_duplicate_predictor_name():
    run("kubectl apply -f ../resources/bad_duplicate_predictor_name.json", shell=True, check=True)
    time.sleep(10)
    completedProcess = run("kubectl get sdep dupname -o json", shell=True, check=True, stdout=subprocess.PIPE)
    jStr = completedProcess.stdout
    j = json.loads(jStr)
    assert j["status"]["state"] == "Failed"
    assert j["status"]["description"] == "Duplicate predictor name: mymodel"
    run("kubectl delete sdep --all", shell=True)    

# Name in graph and that in PodTemplateSpec don't match
def test_model_name_mismatch():
    run("kubectl apply -f ../resources/bad_name_mismatch.json", shell=True, check=True)
    time.sleep(10)
    completedProcess = run("kubectl get sdep namemismatch -o json", shell=True, check=True, stdout=subprocess.PIPE)
    jStr = completedProcess.stdout
    j = json.loads(jStr)
    assert j["status"]["state"] == "Failed"
    assert j["status"]["description"] == "Can't find container for predictive unit with name complex-model_bad_name"
    run("kubectl delete sdep --all", shell=True)    
    
