import pytest
import time
import subprocess
from subprocess import run
import json
from seldon_utils import *
from k8s_utils import *

def wait_for_status(name):
    for attempts in range(5):
        completedProcess = run("kubectl get sdep "+name+" -o json", shell=True, check=True, stdout=subprocess.PIPE)
        jStr = completedProcess.stdout
        j = json.loads(jStr)
        if "status" in j:
            return
        else:
            print("Failed to find status - sleeping")
            time.sleep(5)

@pytest.mark.usefixtures("seldon_java_images")
@pytest.mark.usefixtures("single_namespace_seldon_helm")    
class TestBadGraphs(object):
    
    def test_duplicate_predictor_name(self):
        run("kubectl apply -f ../resources/bad_duplicate_predictor_name.json", shell=True, check=True)
        wait_for_status("dupname")
        completedProcess = run("kubectl get sdep dupname -o json", shell=True, check=True, stdout=subprocess.PIPE)
        jStr = completedProcess.stdout
        j = json.loads(jStr)
        assert j["status"]["state"] == "Failed"
        assert j["status"]["description"] == "Duplicate predictor name: mymodel"
        run("kubectl delete sdep --all", shell=True)    

    # Name in graph and that in PodTemplateSpec don't match
    def test_model_name_mismatch(self):
        run("kubectl apply -f ../resources/bad_name_mismatch.json", shell=True, check=True)
        wait_for_status("namemismatch")
        completedProcess = run("kubectl get sdep namemismatch -o json", shell=True, check=True, stdout=subprocess.PIPE)
        jStr = completedProcess.stdout
        j = json.loads(jStr)
        assert j["status"]["state"] == "Failed"
        assert j["status"]["description"] == "Can't find container for predictive unit with name complex-model_bad_name"
        run("kubectl delete sdep --all", shell=True)    
    
