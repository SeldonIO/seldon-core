from subprocess import run
from seldon_e2e_utils import (
    wait_for_rollout,
    rest_request_ambassador,
    initial_rest_request,
    retry_run,
    API_AMBASSADOR,
)
import time


class TestRollingHttp(object):

    # Test updating a model with a new image version as the only change
    def test_rolling_update1(self):
        namespace = "test-rolling-update-1"
        run(f"kubectl create namespace {namespace}", shell=True, check=True)
        retry_run(f"kubectl apply -f ../resources/graph1.json -n {namespace}")
        wait_for_rollout("mymodel-mymodel-e2eb561", namespace)
        print("Initial request")
        r = initial_rest_request("mymodel", namespace)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
        retry_run(f"kubectl apply -f ../resources/graph2.json -n {namespace}")
        i = 0
        for i in range(100):
            r = rest_request_ambassador("mymodel", namespace, API_AMBASSADOR)
            assert r.status_code == 200
            res = r.json()
            print(res)
            assert (
                res["meta"]["requestPath"]["complex-model"]
                == "seldonio/fixed-model:0.1"
                and res["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
            ) or (
                res["meta"]["requestPath"]["complex-model"]
                == "seldonio/fixed-model:0.2"
                and res["data"]["tensor"]["values"] == [5.0, 6.0, 7.0, 8.0]
            )
            if (not r.status_code == 200) or (
                res["data"]["tensor"]["values"] == [5.0, 6.0, 7.0, 8.0]
            ):
                break
            time.sleep(1)
        assert i < 100
        print("Success for test_rolling_update1")
        run(f"kubectl delete -f ../resources/graph1.json -n {namespace}", shell=True)
        run(f"kubectl delete -f ../resources/graph2.json -n {namespace}", shell=True)
        run(f"kubectl delete namespace {namespace}", shell=True)

    # test changing the image version and the name of its container
    def test_rolling_update2(self):
        namespace = "test-rolling-update-2"
        run(f"kubectl create namespace {namespace}", shell=True, check=True)
        retry_run(f"kubectl apply -f ../resources/graph1.json -n {namespace}")
        wait_for_rollout("mymodel-mymodel-e2eb561", namespace)
        print("Initial request")
        r = initial_rest_request("mymodel", namespace)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
        retry_run(f"kubectl apply -f ../resources/graph3.json -n {namespace}")
        i = 0
        for i in range(100):
            r = rest_request_ambassador("mymodel", namespace, API_AMBASSADOR)
            assert r.status_code == 200
            res = r.json()
            print(res)
            assert (
                "complex-model" in res["meta"]["requestPath"]
                and res["meta"]["requestPath"]["complex-model"]
                == "seldonio/fixed-model:0.1"
                and res["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
            ) or (
                res["meta"]["requestPath"]["complex-model2"]
                == "seldonio/fixed-model:0.2"
                and res["data"]["tensor"]["values"] == [5.0, 6.0, 7.0, 8.0]
            )
            if (not r.status_code == 200) or (
                res["data"]["tensor"]["values"] == [5.0, 6.0, 7.0, 8.0]
            ):
                break
            time.sleep(1)
        assert i < 100
        print("Success for test_rolling_update2")
        run(f"kubectl delete -f ../resources/graph1.json -n {namespace}", shell=True)
        run(f"kubectl delete -f ../resources/graph3.json -n {namespace}", shell=True)
        run(f"kubectl delete namespace {namespace}", shell=True)

    # Test updating a model with a new resource request but same image
    def test_rolling_update3(self):
        namespace = "test-rolling-updates-3"
        run(f"kubectl create namespace {namespace}", shell=True, check=True)
        retry_run(f"kubectl apply -f ../resources/graph1.json -n {namespace}")
        wait_for_rollout("mymodel-mymodel-e2eb561", namespace)
        print("Initial request")
        r = initial_rest_request("mymodel", namespace)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
        retry_run(f"kubectl apply -f ../resources/graph4.json -n {namespace}")
        i = 0
        for i in range(50):
            r = rest_request_ambassador("mymodel", namespace, API_AMBASSADOR)
            assert r.status_code == 200
            res = r.json()
            print(res)
            assert res["meta"]["requestPath"][
                "complex-model"
            ] == "seldonio/fixed-model:0.1" and res["data"]["tensor"]["values"] == [
                1.0,
                2.0,
                3.0,
                4.0,
            ]
            time.sleep(1)
        assert i == 49
        print("Success for test_rolling_update3")
        run(f"kubectl delete -f ../resources/graph1.json -n {namespace}", shell=True)
        run(f"kubectl delete -f ../resources/graph4.json -n {namespace}", shell=True)
        run(f"kubectl delete namespace {namespace}", shell=True)

    # Test updating a model with a multi deployment new model
    def test_rolling_update4(self):
        namespace = "test-rolling-update-4"
        run(f"kubectl create namespace {namespace}", shell=True, check=True)
        retry_run(f"kubectl apply -f ../resources/graph1.json -n {namespace}")
        wait_for_rollout("mymodel-mymodel-e2eb561", namespace)
        print("Initial request")
        r = initial_rest_request("mymodel", namespace)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
        retry_run(f"kubectl apply -f ../resources/graph5.json -n {namespace}")
        i = 0
        for i in range(50):
            r = rest_request_ambassador("mymodel", namespace, API_AMBASSADOR)
            assert r.status_code == 200
            res = r.json()
            print(res)
            assert (
                "complex-model" in res["meta"]["requestPath"]
                and res["meta"]["requestPath"]["complex-model"]
                == "seldonio/fixed-model:0.1"
                and res["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
            ) or (
                res["meta"]["requestPath"]["model1"] == "seldonio/fixed-model:0.1"
                and res["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
                and res["meta"]["requestPath"]["model2"] == "seldonio/fixed-model:0.1"
            )
            if (not r.status_code == 200) or ("model1" in res["meta"]["requestPath"]):
                break
            time.sleep(1)
        assert i < 100
        print("Success for test_rolling_update4")
        run(f"kubectl delete -f ../resources/graph1.json -n {namespace}", shell=True)
        run(f"kubectl delete -f ../resources/graph5.json -n {namespace}", shell=True)
        run(f"kubectl delete namespace {namespace}", shell=True)
