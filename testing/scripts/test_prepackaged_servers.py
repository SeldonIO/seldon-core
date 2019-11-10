import subprocess
import json
from seldon_utils import *
from seldon_core.seldon_client import SeldonClient


def wait_for_status(name, namespace):
    for attempts in range(7):
        completedProcess = run(
            f"kubectl get sdep " + name + " -o json -n {namespace}",
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


def wait_for_rollout(deploymentName, namespace):
    ret = run(
        f"kubectl rollout status deploy/{deploymentName} -n {namespace}", shell=True
    )
    while ret.returncode > 0:
        time.sleep(1)
        ret = run(
            f"kubectl rollout status deploy/{deploymentName} -n {namespace}", shell=True
        )


class TestPrepack(object):

    # Test prepackaged server for sklearn
    def test_sklearn(self):
        namespace = "test-sklearn"
        run(f"kubectl create namespace {namespace}", shell=True, check=True)
        run(
            f"kubectl apply -f ../../servers/sklearnserver/samples/iris.yaml -n {namespace}",
            shell=True,
            check=True,
        )
        wait_for_rollout("iris-default-4903e3c", namespace)
        wait_for_status("sklearn", namespace)
        print("Initial request")
        sc = SeldonClient(deployment_name="sklearn", namespace=namespace)
        r = sc.predict(gateway="ambassador", transport="rest", shape=(1, 4))
        assert r.success
        print("Success for test_prepack_sklearn")
        run(
            f"kubectl delete -f ../../servers/sklearnserver/samples/iris.yaml -n {namespace}",
            shell=True,
            check=True,
        )
        run(f"kubectl delete namespace {namespace}", shell=True, check=True)

    # Test prepackaged server for tfserving
    def test_tfserving(self):
        namespace = "test-tfserving"
        run(f"kubectl create namespace {namespace}", shell=True, check=True)
        run(
            f"kubectl apply -f ../../servers/tfserving/samples/mnist_rest.yaml -n {namespace}",
            shell=True,
            check=True,
        )
        wait_for_rollout("mnist-default-725903e", namespace)
        wait_for_status("tfserving", namespace)
        print("Initial request")
        sc = SeldonClient(deployment_name="tfserving", namespace=namespace)
        r = sc.predict(gateway="ambassador", transport="rest", shape=(1, 784))
        assert r.success
        print("Success for test_prepack_tfserving")
        run(
            f"kubectl delete -f ../../servers/tfserving/samples/mnist_rest.yaml -n {namespace}",
            shell=True,
            check=True,
        )
        run(f"kubectl delete namespace {namespace}", shell=True, check=True)

    # Test prepackaged server for xgboost
    def test_xgboost(self):
        namespace = "test-xgboost"
        run(f"kubectl create namespace {namespace}", shell=True, check=True)
        run(
            f"kubectl apply -f ../../servers/xgboostserver/samples/iris.yaml -n {namespace}",
            shell=True,
            check=True,
        )
        wait_for_rollout("iris-default-af1783b", namespace)
        wait_for_status("xgboost", namespace)
        print("Initial request")
        sc = SeldonClient(deployment_name="xgboost", namespace=namespace)
        r = sc.predict(gateway="ambassador", transport="rest", shape=(1, 4))
        assert r.success
        print("Success for test_prepack_xgboost")
        run(
            f"kubectl delete -f ../../servers/xgboostserver/samples/iris.yaml -n {namespace}",
            shell=True,
            check=True,
        )
        run(f"kubectl delete namespace {namespace}", shell=True, check=True)
