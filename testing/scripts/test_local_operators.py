import os
import time
import logging
import pytest
from subprocess import run
from seldon_e2e_utils import (
    wait_for_rollout,
    rest_request_ambassador,
    initial_rest_request,
    retry_run,
    API_AMBASSADOR,
    API_ISTIO_GATEWAY,
)


class TestLocalOperators(object):
    def test_namespace_operator(self):
        namespace = "test-namespaced-operator"
        retry_run(f"kubectl create namespace {namespace}")
        retry_run(
            f"helm install seldon ../../helm-charts/seldon-core-operator --namespace {namespace} --set istio.enabled=true --set istio.gateway=seldon-gateway --set certManager.enabled=false --set crd.create=false --set singleNamespace=true"
        )
        retry_run(f"kubectl apply -f ../resources/graph1.json -n {namespace}")
        wait_for_rollout("mymodel-mymodel-e2eb561", namespace)
        logging.warning("Initial request")
        r = initial_rest_request("mymodel", namespace, endpoint=API_AMBASSADOR)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
        logging.warning("Success for test_namespace_operator")
        run(f"kubectl delete -f ../resources/graph1.json -n {namespace}", shell=True)
        run(f"kubectl delete namespace {namespace}", shell=True)

    def test_labelled_operator(self):
        namespace = "test-labelled-operator"
        retry_run(f"kubectl create namespace {namespace}")
        retry_run(
            f"helm install seldon ../../helm-charts/seldon-core-operator --namespace {namespace} --set istio.enabled=true --set istio.gateway=seldon-gateway --set certManager.enabled=false --set crd.create=false --set controllerId=seldon-id1"
        )
        retry_run(f"kubectl apply -f ../resources/model_controller_id.yaml -n default")
        wait_for_rollout("test-c1-example-cf749e0", "default")
        logging.warning("Initial request")
        r = initial_rest_request("test-c1", "default", endpoint=API_AMBASSADOR)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
        logging.warning("Success for test_labelled_operator")
        run(
            f"kubectl delete -f ../resources/model_controller_id.yaml -n default",
            shell=True,
        )
        run(f"kubectl delete namespace {namespace}", shell=True)
