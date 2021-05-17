import logging
from subprocess import run

from seldon_e2e_utils import (
    API_AMBASSADOR,
    assert_model,
    initial_rest_request,
    rest_request_ambassador,
    wait_for_rollout,
    wait_for_status,
)


class TestClusterWide(object):

    # Test singe model helm script with 4 API methods
    def test_single_model(self, namespace):
        command = (
            "helm install mymodel ../../helm-charts/seldon-single-model "
            f"--set model.image=seldonio/fixed-model:0.1 "
            f"--namespace {namespace}"
        )
        run(command, shell=True, check=True)

        wait_for_status("mymodel", namespace)
        wait_for_rollout("mymodel", namespace)

        assert_model("mymodel", namespace, initial=True)

        run("helm delete mymodel", shell=True)

    # Test AB Test model helm script with 4 API methods
    def test_abtest_model(self, namespace):
        command = (
            "helm install myabtest ../../helm-charts/seldon-abtest "
            f"--namespace {namespace}"
        )
        run(command, shell=True, check=True)
        wait_for_status("myabtest", namespace)
        wait_for_rollout("myabtest", namespace, expected_deployments=2)
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
        run("helm delete myabtest", shell=True)

    # Test MAB Test model helm script with 4 API methods
    def test_mab_model(self, namespace):
        run(
            f"helm install mymab ../../helm-charts/seldon-mab --namespace {namespace}",
            shell=True,
            check=True,
        )
        wait_for_status("mymab", namespace)
        wait_for_rollout("mymab", namespace, expected_deployments=3)
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
        run("helm delete mymab", shell=True)
