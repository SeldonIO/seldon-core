import subprocess
import json
from seldon_utils import *


def wait_for_status(name):
    for attempts in range(7):
        completedProcess = run(
            "kubectl get sdep " + name + " -o json -n seldon",
            shell=True,
            check=True,
            stdout=subprocess.PIPE,
        )
        jStr = completedProcess.stdout
        j = json.loads(jStr)
        if "status" in j:
            return j
        else:
            print("Failed to find status - sleeping")
            time.sleep(5)


class TestBadGraphs(object):
    def test_duplicate_predictor_name(self):
        ret = run(
            "kubectl apply -f ../resources/bad_duplicate_predictor_name.json -n seldon",
            shell=True,
            check=False,
        )
        assert ret.returncode == 1

    # Name in graph and that in PodTemplateSpec don't match
    def test_model_name_mismatch(self):
        ret = run(
            "kubectl apply -f ../resources/bad_name_mismatch.json -n seldon",
            shell=True,
            check=False,
        )
        assert ret.returncode == 1
