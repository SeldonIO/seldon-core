import subprocess
import json
from seldon_e2e_utils import (
    wait_for_rollout,
    initial_rest_request,
    retry_run,
    create_random_data,
    wait_for_status,
)
from subprocess import run
import time
import logging


class TestPrepack(object):

    # Test prepackaged server for sklearn
    def test_sklearn(self):
        namespace = "test-sklearn"
        retry_run(f"kubectl create namespace {namespace}")
        retry_run(
            f"kubectl apply -f ../../servers/sklearnserver/samples/iris.yaml -n {namespace}"
        )
        wait_for_rollout("iris-default-4903e3c", namespace)
        wait_for_status("sklearn", namespace)
        time.sleep(1)
        logging.warning("Initial request")
        r = initial_rest_request(
            "sklearn", namespace, rows=1, data_size=4, dtype="ndarray"
        )
        assert r.status_code == 200
        logging.warning("Success for test_prepack_sklearn")
        run(
            f"kubectl delete -f ../../servers/sklearnserver/samples/iris.yaml -n {namespace}",
            shell=True,
        )
        run(f"kubectl delete namespace {namespace}", shell=True)

    # Test prepackaged server for tfserving
    def test_tfserving(self):
        namespace = "test-tfserving"
        retry_run(f"kubectl create namespace {namespace}")
        retry_run(
            f"kubectl apply -f ../../servers/tfserving/samples/mnist_rest.yaml -n {namespace}"
        )
        wait_for_rollout("mnist-default-725903e", namespace)
        wait_for_status("tfserving", namespace)
        time.sleep(1)
        logging.warning("Initial request")
        r = initial_rest_request(
            "tfserving",
            namespace,
            data=[create_random_data(784)[1].tolist()],
            dtype="ndarray",
        )
        assert r.status_code == 200
        logging.warning("Success for test_prepack_tfserving")
        run(
            f"kubectl delete -f ../../servers/tfserving/samples/mnist_rest.yaml -n {namespace}",
            shell=True,
        )
        run(f"kubectl delete namespace {namespace}", shell=True)

    # Test prepackaged server for xgboost
    def test_xgboost(self):
        namespace = "test-xgboost"
        retry_run(f"kubectl create namespace {namespace}")
        retry_run(
            f"kubectl apply -f ../../servers/xgboostserver/samples/iris.yaml -n {namespace}"
        )
        wait_for_rollout("iris-default-af1783b", namespace)
        wait_for_status("xgboost", namespace)
        time.sleep(1)
        logging.warning("Initial request")
        r = initial_rest_request(
            "xgboost", namespace, data=[[0.1, 0.2, 0.3, 0.4]], dtype="ndarray"
        )
        assert r.status_code == 200
        logging.warning("Success for test_prepack_xgboost")
        run(
            f"kubectl delete -f ../../servers/xgboostserver/samples/iris.yaml -n {namespace}",
            shell=True,
        )
        run(f"kubectl delete namespace {namespace}", shell=True)
