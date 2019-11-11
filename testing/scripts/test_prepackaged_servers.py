import subprocess
import json
from seldon_core.seldon_client import SeldonClient
from seldon_e2e_utils import wait_for_rollout, retry_run, wait_for_status
from subprocess import run
import time


class TestPrepack(object):

    # Test prepackaged server for sklearn
    def test_sklearn(self):
        namespace = "test-sklearn"
        run(f"kubectl create namespace {namespace}", shell=True, check=True)
        retry_run(
            f"kubectl apply -f ../../servers/sklearnserver/samples/iris.yaml -n {namespace}"
        )
        wait_for_rollout("iris-default-4903e3c", namespace)
        wait_for_status("sklearn", namespace)
        time.sleep(1)
        print("Initial request")
        sc = SeldonClient(deployment_name="sklearn", namespace=namespace)
        r = sc.predict(gateway="ambassador", transport="rest", shape=(1, 4))
        assert r.success
        print("Success for test_prepack_sklearn")
        run(
            f"kubectl delete -f ../../servers/sklearnserver/samples/iris.yaml -n {namespace}",
            shell=True,
        )
        run(f"kubectl delete namespace {namespace}", shell=True)

    # Test prepackaged server for tfserving
    def test_tfserving(self):
        namespace = "test-tfserving"
        run(f"kubectl create namespace {namespace}", shell=True, check=True)
        retry_run(
            f"kubectl apply -f ../../servers/tfserving/samples/mnist_rest.yaml -n {namespace}"
        )
        wait_for_rollout("mnist-default-725903e", namespace)
        wait_for_status("tfserving", namespace)
        time.sleep(1)
        print("Initial request")
        sc = SeldonClient(deployment_name="tfserving", namespace=namespace)
        r = sc.predict(gateway="ambassador", transport="rest", shape=(1, 784))
        assert r.success
        print("Success for test_prepack_tfserving")
        run(
            f"kubectl delete -f ../../servers/tfserving/samples/mnist_rest.yaml -n {namespace}",
            shell=True,
        )
        run(f"kubectl delete namespace {namespace}", shell=True)

    # Test prepackaged server for xgboost
    def test_xgboost(self):
        namespace = "test-xgboost"
        run(f"kubectl create namespace {namespace}", shell=True, check=True)
        retry_run(
            f"kubectl apply -f ../../servers/xgboostserver/samples/iris.yaml -n {namespace}"
        )
        wait_for_rollout("iris-default-af1783b", namespace)
        wait_for_status("xgboost", namespace)
        time.sleep(1)
        print("Initial request")
        sc = SeldonClient(deployment_name="xgboost", namespace=namespace)
        r = sc.predict(gateway="ambassador", transport="rest", shape=(1, 4))
        assert r.success
        print("Success for test_prepack_xgboost")
        run(
            f"kubectl delete -f ../../servers/xgboostserver/samples/iris.yaml -n {namespace}",
            shell=True,
        )
        run(f"kubectl delete namespace {namespace}", shell=True)
