import subprocess
import json
from seldon_utils import *
from seldon_core.seldon_client import SeldonClient


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
        if "status" in j and j["status"] == "Available":
            return j
        else:
            print("Failed to find status - sleeping")
            time.sleep(5)


def wait_for_rollout(deploymentName):
    ret = run("kubectl rollout status deploy/" + deploymentName, shell=True)
    while ret.returncode > 0:
        time.sleep(1)
        ret = run("kubectl rollout status deploy/" + deploymentName, shell=True)


class TestPrepack(object):

    # Test prepackaged server for sklearn
    def test_sklearn(self):
        run("kubectl delete sdep --all", shell=True)
        run(
            "kubectl apply -f ../../servers/sklearnserver/samples/iris.yaml",
            shell=True,
            check=True,
        )
        wait_for_rollout("iris-default-4903e3c")
        wait_for_status("sklearn")
        print("Initial request")
        sc = SeldonClient(deployment_name="sklearn", namespace="seldon")
        r = sc.predict(gateway="ambassador", transport="rest", shape=(1, 4))
        assert r.success
        print("Success for test_prepack_sklearn")

    # Test prepackaged server for tfserving
    def test_tfserving(self):
        run("kubectl delete sdep --all", shell=True)
        run(
            "kubectl apply -f ../../servers/tfserving/samples/mnist_rest.yaml",
            shell=True,
            check=True,
        )
        wait_for_rollout("mnist-default-725903e")
        wait_for_status("tfserving")
        print("Initial request")
        sc = SeldonClient(deployment_name="tfserving", namespace="seldon")
        r = sc.predict(gateway="ambassador", transport="rest", shape=(1, 784))
        assert r.success
        print("Success for test_prepack_tfserving")

    # Test prepackaged server for xgboost
    def test_xgboost(self):
        run("kubectl delete sdep --all", shell=True)
        run(
            "kubectl apply -f ../../servers/xgboostserver/samples/iris.yaml",
            shell=True,
            check=True,
        )
        wait_for_rollout("iris-default-af1783b")
        wait_for_status("xgboost")
        print("Initial request")
        sc = SeldonClient(deployment_name="xgboost", namespace="seldon")
        r = sc.predict(gateway="ambassador", transport="rest", shape=(1, 4))
        assert r.success
        print("Success for test_prepack_xgboost")
