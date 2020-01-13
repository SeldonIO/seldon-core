from seldon_e2e_utils import (
    wait_for_rollout,
    wait_for_status,
    initial_rest_request,
    rest_request_ambassador,
    grpc_request_ambassador2,
    API_AMBASSADOR,
)
from subprocess import run
import logging


class TestClusterWide(object):

    # Test singe model helm script with 4 API methods
    def test_single_model(self, namespace):
        run(
            f"helm install mymodel ../../helm-charts/seldon-single-model --set oauth.key=oauth-key --set oauth.secret=oauth-secret --namespace {namespace}",
            shell=True,
            check=True,
        )
        wait_for_status("mymodel", namespace)
        wait_for_rollout("mymodel", namespace)
        initial_rest_request("mymodel", namespace)
        logging.warning("Test Ambassador REST gateway")
        r = rest_request_ambassador("mymodel", namespace, API_AMBASSADOR)
        logging.warning(r.json())
        assert r.status_code == 200
        assert len(r.json()["data"]["tensor"]["values"]) == 1
        logging.warning("Test Ambassador gRPC gateway")
        r = grpc_request_ambassador2("mymodel", namespace, API_AMBASSADOR)
        logging.warning(r)
        run(f"helm delete mymodel", shell=True)

    # Test AB Test model helm script with 4 API methods
    def test_abtest_model(self, namespace):
        run(
            f"helm install myabtest ../../helm-charts/seldon-abtest --set oauth.key=oauth-key --set oauth.secret=oauth-secret --namespace {namespace}",
            shell=True,
            check=True,
        )
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
        run(f"helm delete myabtest", shell=True)

    # Test MAB Test model helm script with 4 API methods
    def test_mab_model(self, namespace):
        run(
            f"helm install mymab ../../helm-charts/seldon-mab --set oauth.key=oauth-key --set oauth.secret=oauth-secret --namespace {namespace}",
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
        run(f"helm delete mymab", shell=True)
