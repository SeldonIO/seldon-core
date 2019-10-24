import pytest
from seldon_utils import *
from k8s_utils import *


def wait_for_shutdown(deploymentName):
    ret = run("kubectl get -n test1 deploy/" + deploymentName, shell=True)
    while ret.returncode == 0:
        time.sleep(1)
        ret = run("kubectl get -n test1 deploy/" + deploymentName, shell=True)


class TestClusterWide(object):

    # Test singe model helm script with 4 API methods
    def test_single_model(self):
        run("helm delete mymodel --purge", shell=True)
        run(
            "helm install ../../helm-charts/seldon-single-model --name mymodel --set oauth.key=oauth-key --set oauth.secret=oauth-secret --namespace test1",
            shell=True,
            check=True,
        )
        wait_for_rollout("mymodel-mymodel-7cd068f")
        initial_rest_request("mymodel", "test1")
        print("Test Ambassador REST gateway")
        r = rest_request_ambassador("mymodel", "test1", API_AMBASSADOR)
        print(r.json())
        assert r.status_code == 200
        assert len(r.json()["data"]["tensor"]["values"]) == 1
        print("Test Ambassador gRPC gateway")
        r = grpc_request_ambassador2("mymodel", "test1", API_AMBASSADOR)
        print(r)
        run("helm delete mymodel --purge", shell=True)

    # Test AB Test model helm script with 4 API methods
    def test_abtest_model(self):
        run("helm delete myabtest --purge", shell=True)
        run(
            "helm install ../../helm-charts/seldon-abtest --name myabtest --set oauth.key=oauth-key --set oauth.secret=oauth-secret --namespace test1",
            shell=True,
            check=True,
        )
        wait_for_rollout("myabtest-myabtest-41de5b8")
        wait_for_rollout("myabtest-myabtest-df66c5c")
        initial_rest_request("myabtest", "test1")
        print("Test Ambassador REST gateway")
        r = rest_request_ambassador("myabtest", "test1", API_AMBASSADOR)
        print(r.json())
        assert r.status_code == 200
        assert len(r.json()["data"]["tensor"]["values"]) == 1
        print("Test Ambassador gRPC gateway")
        print(
            "WARNING SKIPPING FLAKY AMBASSADOR TEST UNTIL AMBASSADOR GRPC ISSUE FIXED.."
        )
        # r = grpc_request_ambassador2("myabtest", "test1", API_AMBASSADOR)
        # print(r)
        run("helm delete myabtest --purge", shell=True)

    # Test MAB Test model helm script with 4 API methods
    def test_mab_model(self):
        run("helm delete mymab --purge", shell=True)
        run(
            "helm install ../../helm-charts/seldon-mab --name mymab --set oauth.key=oauth-key --set oauth.secret=oauth-secret --namespace test1",
            shell=True,
            check=True,
        )
        wait_for_rollout("mymab-mymab-41de5b8")
        wait_for_rollout("mymab-mymab-b8038b2")
        wait_for_rollout("mymab-mymab-df66c5c")
        initial_rest_request("mymab", "test1")
        print("Test Ambassador REST gateway")
        r = rest_request_ambassador("mymab", "test1", API_AMBASSADOR)
        print(r.json())
        assert r.status_code == 200
        assert len(r.json()["data"]["tensor"]["values"]) == 1
        print("Test Ambassador gRPC gateway")
        print(
            "WARNING SKIPPING FLAKY AMBASSADOR TEST UNTIL AMBASSADOR GRPC ISSUE FIXED.."
        )
        # r = grpc_request_ambassador2("mymab", "test1", API_AMBASSADOR)
        # print(r)
        run("helm delete mymab --purge", shell=True)
