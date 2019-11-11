import pytest
from seldon_e2e_utils import (
    wait_for_rollout,
    initial_rest_request,
    rest_request_ambassador,
    grpc_request_ambassador2,
    API_AMBASSADOR,
)
from subprocess import run


class TestClusterWide(object):

    # Test singe model helm script with 4 API methods
    def test_single_model(self):
        namespace = "test-single-model"
        run(f"kubectl create namespace {namespace}", shell=True, check=True)
        run(
            f"helm install ../../helm-charts/seldon-single-model --name mymodel --set oauth.key=oauth-key --set oauth.secret=oauth-secret --namespace {namespace}",
            shell=True,
            check=True,
        )
        wait_for_rollout(f"mymodel-mymodel-7cd068f", namespace)
        initial_rest_request("mymodel", namespace)
        print("Test Ambassador REST gateway")
        r = rest_request_ambassador("mymodel", namespace, API_AMBASSADOR)
        print(r.json())
        assert r.status_code == 200
        assert len(r.json()["data"]["tensor"]["values"]) == 1
        print("Test Ambassador gRPC gateway")
        r = grpc_request_ambassador2("mymodel", namespace, API_AMBASSADOR)
        print(r)
        run(f"helm delete mymodel --purge", shell=True)
        run(f"kubectl delete namespace {namespace}", shell=True, check=True)

    # Test AB Test model helm script with 4 API methods
    def test_abtest_model(self):
        namespace = "test-abtest-model"
        run(f"kubectl create namespace {namespace}", shell=True, check=True)
        run(
            f"helm install ../../helm-charts/seldon-abtest --name myabtest --set oauth.key=oauth-key --set oauth.secret=oauth-secret --namespace {namespace}",
            shell=True,
            check=True,
        )
        wait_for_rollout("myabtest-myabtest-41de5b8", namespace)
        wait_for_rollout("myabtest-myabtest-df66c5c", namespace)
        initial_rest_request("myabtest", namespace)
        print("Test Ambassador REST gateway")
        r = rest_request_ambassador("myabtest", namespace, API_AMBASSADOR)
        print(r.json())
        assert r.status_code == 200
        assert len(r.json()["data"]["tensor"]["values"]) == 1
        print("Test Ambassador gRPC gateway")
        print(
            "WARNING SKIPPING FLAKY AMBASSADOR TEST UNTIL AMBASSADOR GRPC ISSUE FIXED.."
        )
        run(f"helm delete myabtest --purge", shell=True)
        run(f"kubectl delete namespace {namespace}", shell=True, check=True)

    # Test MAB Test model helm script with 4 API methods
    def test_mab_model(self):
        namespace = "test-mab-model"
        run(f"kubectl create namespace {namespace}", shell=True, check=True)
        run(
            f"helm install ../../helm-charts/seldon-mab --name mymab --set oauth.key=oauth-key --set oauth.secret=oauth-secret --namespace {namespace}",
            shell=True,
            check=True,
        )
        wait_for_rollout("mymab-mymab-41de5b8", namespace)
        wait_for_rollout("mymab-mymab-b8038b2", namespace)
        wait_for_rollout("mymab-mymab-df66c5c", namespace)
        initial_rest_request("mymab", namespace)
        print("Test Ambassador REST gateway")
        r = rest_request_ambassador("mymab", namespace, API_AMBASSADOR)
        print(r.json())
        assert r.status_code == 200
        assert len(r.json()["data"]["tensor"]["values"]) == 1
        print("Test Ambassador gRPC gateway")
        print(
            "WARNING SKIPPING FLAKY AMBASSADOR TEST UNTIL AMBASSADOR GRPC ISSUE FIXED.."
        )
        run(f"helm delete mymab --purge", shell=True)
        run(f"kubectl delete namespace {namespace}", shell=True, check=True)
