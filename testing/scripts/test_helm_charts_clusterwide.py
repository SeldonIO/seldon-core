import pytest
from seldon_utils import *
from k8s_utils import *


def wait_for_shutdown(deploymentName, namespace):
    ret = run(f"kubectl get -n {namespace} deploy/" + deploymentName, shell=True)
    while ret.returncode == 0:
        time.sleep(1)
        ret = run(f"kubectl get -n {namespace} deploy/" + deploymentName, shell=True)


class TestClusterWide(object):

    # Test singe model helm script with 4 API methods
    def test_single_model(self):
        namespace = "test-single-model"
        run(f"kubectl create namespace {namespace}", shell=true, check=true)
        run(
            f"helm install ../../helm-charts/seldon-single-model --name mymodel-{namespace} --set oauth.key=oauth-key --set oauth.secret=oauth-secret --namespace {namespace}",
            shell=True,
            check=True,
        )
        wait_for_rollout("mymodel-mymodel-7cd068f", namespace)
        initial_rest_request("mymodel", namespace)
        print("Test Ambassador REST gateway")
        r = rest_request_ambassador("mymodel", namespace, API_AMBASSADOR)
        print(r.json())
        assert r.status_code == 200
        assert len(r.json()["data"]["tensor"]["values"]) == 1
        print("Test Ambassador gRPC gateway")
        r = grpc_request_ambassador2("mymodel", namespace, API_AMBASSADOR)
        print(r)
        run(f"helm delete mymodel-{namespace} --purge", shell=True)
        run(f"kubectl delete namespace {namespace}", shell=true, check=true)

    # Test AB Test model helm script with 4 API methods
    def test_abtest_model(self):
        namespace = "test-abtest-model"
        run(f"kubectl create namespace {namespace}", shell=true, check=true)
        run(
            f"helm install ../../helm-charts/seldon-abtest --name myabtest-{namespace} --set oauth.key=oauth-key --set oauth.secret=oauth-secret --namespace {namespace}",
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
        # r = grpc_request_ambassador2("myabtest", "test1", API_AMBASSADOR)
        # print(r)
        run(f"helm delete myabtest-{namespace} --purge", shell=True)
        run(f"kubectl delete namespace {namespace}", shell=true, check=true)

    # Test MAB Test model helm script with 4 API methods
    def test_mab_model(self):
        namespace = "test-mab-model"
        run(f"kubectl create namespace {namespace}", shell=true, check=true)
        run(
            f"helm install ../../helm-charts/seldon-mab --name mymab-{namespace} --set oauth.key=oauth-key --set oauth.secret=oauth-secret --namespace {namespace}",
            shell=True,
            check=True,
        )
        wait_for_rollout("mymab-mymab-41de5b8", namespace)
        wait_for_rollout("mymab-mymab-b8038b2", namespace)
        wait_for_rollout("mymab-mymab-df66c5c", namespace)
        initial_rest_request("mymab", "test1")
        print("Test Ambassador REST gateway")
        r = rest_request_ambassador("mymab", namespace, API_AMBASSADOR)
        print(r.json())
        assert r.status_code == 200
        assert len(r.json()["data"]["tensor"]["values"]) == 1
        print("Test Ambassador gRPC gateway")
        print(
            "WARNING SKIPPING FLAKY AMBASSADOR TEST UNTIL AMBASSADOR GRPC ISSUE FIXED.."
        )
        # r = grpc_request_ambassador2("mymab", "test1", API_AMBASSADOR)
        # print(r)
        run(f"helm delete mymab-{namespace} --purge", shell=True)
        run(f"kubectl delete namespace {namespace}", shell=true, check=true)
