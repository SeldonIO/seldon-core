from seldon_utils import *
from k8s_utils import *


def wait_for_shutdown(deploymentName):
    ret = run("kubectl get deploy/" + deploymentName, shell=True)
    while ret.returncode == 0:
        time.sleep(1)
        ret = run("kubectl get deploy/" + deploymentName, shell=True)


def wait_for_rollout(deploymentName):
    ret = run("kubectl rollout status deploy/" + deploymentName, shell=True)
    while ret.returncode > 0:
        time.sleep(1)
        ret = run("kubectl rollout status deploy/" + deploymentName, shell=True)


class TestRollingHttp(object):

    # Test updating a model with a new image version as the only change
    def test_rolling_update1(self):
        run("kubectl delete sdep --all", shell=True)
        wait_for_shutdown("mymodel-mymodel-e2eb561")
        run("kubectl apply -f ../resources/graph1.json", shell=True, check=True)
        wait_for_rollout("mymodel-mymodel-e2eb561")
        print("Initial request")
        r = initial_rest_request("mymodel", "seldon")
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
        run("kubectl apply -f ../resources/graph2.json", shell=True, check=True)
        i = 0
        for i in range(100):
            r = rest_request_ambassador("mymodel", "seldon", API_AMBASSADOR)
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

    # test changing the image version and the name of its container
    def test_rolling_update2(self):
        run("kubectl delete sdep --all", shell=True)
        wait_for_shutdown("mymodel-mymodel-e2eb561")
        run("kubectl apply -f ../resources/graph1.json", shell=True, check=True)
        wait_for_rollout("mymodel-mymodel-e2eb561")
        print("Initial request")
        r = initial_rest_request("mymodel", "seldon")
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
        run("kubectl apply -f ../resources/graph3.json", shell=True, check=True)
        i = 0
        for i in range(100):
            r = rest_request_ambassador("mymodel", "seldon", API_AMBASSADOR)
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

    # Test updating a model with a new resource request but same image
    def test_rolling_update3(self):
        run("kubectl delete sdep --all", shell=True)
        wait_for_shutdown("mymodel-mymodel-e2eb561")
        run("kubectl apply -f ../resources/graph1.json", shell=True, check=True)
        wait_for_rollout("mymodel-mymodel-e2eb561")
        print("Initial request")
        r = initial_rest_request("mymodel", "seldon")
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
        run("kubectl apply -f ../resources/graph4.json", shell=True, check=True)
        i = 0
        for i in range(50):
            r = rest_request_ambassador("mymodel", "seldon", API_AMBASSADOR)
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

    # Test updating a model with a multi deployment new model
    def test_rolling_update4(self):
        run("kubectl delete sdep --all", shell=True)
        wait_for_shutdown("mymodel-mymodel-e2eb561")
        run("kubectl apply -f ../resources/graph1.json", shell=True, check=True)
        wait_for_rollout("mymodel-mymodel-e2eb561")
        print("Initial request")
        r = initial_rest_request("mymodel", "seldon")
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
        run("kubectl apply -f ../resources/graph5.json", shell=True, check=True)
        i = 0
        for i in range(50):
            r = rest_request_ambassador("mymodel", "seldon", API_AMBASSADOR)
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

    # Test updating a model to a multi predictor model
    def test_rolling_update5(self):
        run("kubectl delete sdep --all", shell=True)
        wait_for_shutdown("mymodel-mymodel-e2eb561")
        run("kubectl apply -f ../resources/graph1.json", shell=True, check=True)
        wait_for_rollout("mymodel-mymodel-e2eb561")
        print("Initial request")
        r = initial_rest_request("mymodel", "seldon")
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
        run("kubectl apply -f ../resources/graph6.json", shell=True, check=True)
        i = 0
        for i in range(50):
            r = rest_request_ambassador("mymodel", "seldon", API_AMBASSADOR)
            print("Status code", r.status_code)
            assert r.status_code == 200
            res = r.json()
            print(res)
            assert (
                "complex-model" in res["meta"]["requestPath"]
                and res["meta"]["requestPath"]["complex-model"]
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
        print("Success for test_rolling_update5")

    # Test updating a model with a new image version as the only change
    def test_rolling_update6(self):
        run("kubectl delete sdep --all", shell=True)
        wait_for_shutdown("mymodel-mymodel-e2eb561")
        wait_for_shutdown("mymodel-mymodel-svc-orch-8e2a24b")
        run("kubectl apply -f ../resources/graph1svc.json", shell=True, check=True)
        wait_for_rollout("mymodel-mymodel-svc-orch-8e2a24b")
        wait_for_rollout("mymodel-mymodel-e2eb561")
        print("Initial request")
        r = initial_rest_request("mymodel", "seldon")
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
        run("kubectl apply -f ../resources/graph2svc.json", shell=True, check=True)
        i = 0
        for i in range(100):
            r = rest_request_ambassador("mymodel", "seldon", API_AMBASSADOR)
            print("Status code", r.status_code)
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
        print("Success for test_rolling_update6")

    # test changing the image version and the name of its container
    def test_rolling_update7(self):
        run("kubectl delete sdep --all", shell=True)
        wait_for_shutdown("mymodel-mymodel-e2eb561")
        wait_for_shutdown("mymodel-mymodel-svc-orch-8e2a24b")
        run("kubectl apply -f ../resources/graph1svc.json", shell=True, check=True)
        wait_for_rollout("mymodel-mymodel-svc-orch-8e2a24b")
        wait_for_rollout("mymodel-mymodel-e2eb561")
        print("Initial request")
        r = initial_rest_request("mymodel", "seldon")
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
        run("kubectl apply -f ../resources/graph3svc.json", shell=True, check=True)
        i = 0
        for i in range(100):
            r = rest_request_ambassador("mymodel", "seldon", API_AMBASSADOR)
            print("Status code", r.status_code)
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
        print("Success for test_rolling_update7")

    # Test updating a model with a new resource request but same image
    def test_rolling_update8(self):
        run("kubectl delete sdep --all", shell=True)
        wait_for_shutdown("mymodel-mymodel-e2eb561")
        wait_for_shutdown("mymodel-mymodel-svc-orch-8e2a24b")
        run("kubectl apply -f ../resources/graph1svc.json", shell=True, check=True)
        wait_for_rollout("mymodel-mymodel-svc-orch-8e2a24b")
        wait_for_rollout("mymodel-mymodel-e2eb561")
        r = initial_rest_request("mymodel", "seldon")
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
        run("kubectl apply -f ../resources/graph4svc.json", shell=True, check=True)
        i = 0
        for i in range(50):
            r = rest_request_ambassador("mymodel", "seldon", API_AMBASSADOR)
            print("Status code", r.status_code)
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
        print("Success for test_rolling_update8")

    # Test updating a model with a multi deployment new model
    def test_rolling_update9(self):
        run("kubectl delete sdep --all", shell=True)
        wait_for_shutdown("mymodel-mymodel-e2eb561")
        wait_for_shutdown("mymodel-mymodel-svc-orch-8e2a24b")
        run("kubectl apply -f ../resources/graph1svc.json", shell=True, check=True)
        wait_for_rollout("mymodel-mymodel-svc-orch-8e2a24b")
        wait_for_rollout("mymodel-mymodel-e2eb561")
        r = initial_rest_request("mymodel", "seldon")
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
        run("kubectl apply -f ../resources/graph5svc.json", shell=True, check=True)
        i = 0
        for i in range(50):
            r = rest_request_ambassador("mymodel", "seldon", API_AMBASSADOR)
            print("Status code", r.status_code)
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
        print("Success for test_rolling_update9")

    # Test updating a model to a multi predictor model
    def test_rolling_update10(self):
        run("kubectl delete sdep --all", shell=True)
        wait_for_shutdown("mymodel-mymodel-e2eb561")
        wait_for_shutdown("mymodel-mymodel-svc-orch-8e2a24b")
        run("kubectl apply -f ../resources/graph1svc.json", shell=True, check=True)
        wait_for_rollout("mymodel-mymodel-svc-orch-8e2a24b")
        wait_for_rollout("mymodel-mymodel-e2eb561")
        r = initial_rest_request("mymodel", "seldon")
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["values"] == [1.0, 2.0, 3.0, 4.0]
        run("kubectl apply -f ../resources/graph6svc.json", shell=True, check=True)
        i = 0
        for i in range(50):
            r = rest_request_ambassador("mymodel", "seldon", API_AMBASSADOR)
            print("Status code", r.status_code)
            assert r.status_code == 200
            res = r.json()
            print(res)
            assert (
                "complex-model" in res["meta"]["requestPath"]
                and res["meta"]["requestPath"]["complex-model"]
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
        print("Success for test_rolling_update10")
