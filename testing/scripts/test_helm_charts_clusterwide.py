import pytest
from seldon_e2e_utils import (
    wait_for_rollout,
    initial_rest_request,
    initial_grpc_request,
    rest_request_ambassador,
    grpc_request_ambassador2,
    retry_run,
    API_AMBASSADOR,
)
from subprocess import run
import logging
import time


class TestClusterWide(object):

    # Test singe model helm script with 4 API methods
    def test_rest_single_model(self):
        namespace = "test-single-rest-model"
        retry_run(f"kubectl create namespace {namespace}")
        run(
            f"helm install mymodel ../../helm-charts/seldon-single-model --namespace {namespace}",
            shell=True,
            check=True,
        )
        wait_for_rollout(f"mymodel-mymodel-de240ba", namespace)
        initial_rest_request("mymodel", namespace)
        logging.warning("Test Ambassador REST gateway")
        r = rest_request_ambassador("mymodel", namespace, API_AMBASSADOR)
        logging.warning(r.json())
        assert r.status_code == 200
        assert len(r.json()["data"]["tensor"]["values"]) == 1
        run(f"kubectl delete namespace {namespace}", shell=True)

    def test_grpc_single_model(self):
        namespace = "test-single-grpc-model"
        retry_run(f"kubectl create namespace {namespace}")
        run(
            f"helm install mymodel ../../helm-charts/seldon-single-model --namespace {namespace} --set protocol=GRPC",
            shell=True,
            check=True,
        )
        wait_for_rollout(f"mymodel-mymodel-2a00e84", namespace)
        time.sleep(
            5
        )  # Seems to be needed for consistent Ambassador grpc call. If called too early it seems will always fail.
        initial_grpc_request("mymodel", namespace)
        logging.warning("Test Ambassador gRPC gateway")
        r = grpc_request_ambassador2("mymodel", namespace, API_AMBASSADOR)
        logging.warning(r)
        run(f"kubectl delete namespace {namespace}", shell=True)

    # Test AB Test model helm script with 4 API methods
    def test_rest_abtest_model(self):
        namespace = "test-abtest-model"
        retry_run(f"kubectl create namespace {namespace}")
        run(
            f"helm install myabtest ../../helm-charts/seldon-abtest --namespace {namespace}",
            shell=True,
            check=True,
        )
        wait_for_rollout("myabtest-myabtest-0cce7b2", namespace)
        wait_for_rollout("myabtest-myabtest-ba661ba", namespace)
        initial_rest_request("myabtest", namespace)
        logging.warning("Test Ambassador REST gateway")
        r = rest_request_ambassador("myabtest", namespace, API_AMBASSADOR)
        logging.warning(r.json())
        assert r.status_code == 200
        assert len(r.json()["data"]["tensor"]["values"]) == 1
        logging.warning("Test Ambassador gRPC gateway")
        logging.warning(
            "WARNING SKIPPING FLAKY AMBASSADOR TEST UNTIL AMBASSADOR GRPC ISSUE FIXED.."
        )
        run(f"kubectl delete namespace {namespace}", shell=True)

    # Test MAB Test model helm script with 4 API methods
    def test_rest_mab_model(self):
        namespace = "test-mab-model"
        retry_run(f"kubectl create namespace {namespace}")
        run(
            f"helm install mymab ../../helm-charts/seldon-mab --namespace {namespace}",
            shell=True,
            check=True,
        )
        wait_for_rollout("mymab-mymab-0cce7b2", namespace)
        wait_for_rollout("mymab-mymab-4ae7b7c", namespace)
        wait_for_rollout("mymab-mymab-ba661ba", namespace)
        initial_rest_request("mymab", namespace)
        logging.warning("Test Ambassador REST gateway")
        r = rest_request_ambassador("mymab", namespace, API_AMBASSADOR)
        logging.warning(r.json())
        assert r.status_code == 200
        assert len(r.json()["data"]["tensor"]["values"]) == 1
        logging.warning("Test Ambassador gRPC gateway")
        logging.warning(
            "WARNING SKIPPING FLAKY AMBASSADOR TEST UNTIL AMBASSADOR GRPC ISSUE FIXED.."
        )
        run(f"kubectl delete namespace {namespace}", shell=True)
