import pytest
import time
from subprocess import run
from seldon_utils import *
API_AMBASSADOR="localhost:8003"

def wait_for_rollout(deploymentName):
    ret = run("kubectl rollout status deploy/"+deploymentName, shell=True)
    while ret.returncode > 0:
        time.sleep(1)
        ret = run("kubectl rollout status deploy/"+deploymentName, shell=True)

# Test updating a model with a new image version as the only change
def test_rolling_update1():
    run("kubectl delete sdep --all", shell=True)
    print("sleeping 30 secs...")
    time.sleep(30)
    run("kubectl apply -f ../resources/graph1.json", shell=True, check=True)
    wait_for_rollout("mymodel-mymodel-svc-orch")
    wait_for_rollout("mymodel-mymodel-e2eb561")
    r = rest_request_api_gateway("oauth-key","oauth-secret")        
    assert r.status_code == 200
    assert r.json()["data"]["tensor"]["values"] == [1.0,2.0,3.0,4.0]
    run("kubectl apply -f ../resources/graph2.json", shell=True, check=True)
    i = 0
    for i in range(100):
        r = rest_request_api_gateway("oauth-key","oauth-secret")
        res = r.json()
        print(res)
        assert r.status_code == 200
        assert ((res["meta"]["requestPath"]["complex-model"] == "seldonio/fixed-model:0.1" and res["data"]["tensor"]["values"] == [1.0,2.0,3.0,4.0]) or (res["meta"]["requestPath"]["complex-model"] == "seldonio/fixed-model:0.2" and res["data"]["tensor"]["values"] == [5.0,6.0,7.0,8.0]))
        if (not r.status_code == 200) or (res["data"]["tensor"]["values"] == [5.0,6.0,7.0,8.0]):
            break
        time.sleep(1)
    assert i < 100

# test changing the image version and the name of its container
def test_rolling_update2():
    run("kubectl delete sdep --all", shell=True)
    print("sleeping 30 secs...")
    time.sleep(30)
    run("kubectl apply -f ../resources/graph1.json", shell=True, check=True)
    wait_for_rollout("mymodel-mymodel-svc-orch")
    wait_for_rollout("mymodel-mymodel-e2eb561")
    r = rest_request_api_gateway("oauth-key","oauth-secret")        
    assert r.status_code == 200
    assert r.json()["data"]["tensor"]["values"] == [1.0,2.0,3.0,4.0]
    run("kubectl apply -f ../resources/graph3.json", shell=True, check=True)
    i = 0
    for i in range(100):
        r = rest_request_api_gateway("oauth-key","oauth-secret")
        res = r.json()
        print(res)
        assert r.status_code == 200
        assert (("complex-model" in res["meta"]["requestPath"] and res["meta"]["requestPath"]["complex-model"] == "seldonio/fixed-model:0.1" and res["data"]["tensor"]["values"] == [1.0,2.0,3.0,4.0]) or (res["meta"]["requestPath"]["complex-model2"] == "seldonio/fixed-model:0.2" and res["data"]["tensor"]["values"] == [5.0,6.0,7.0,8.0]))
        if (not r.status_code == 200) or (res["data"]["tensor"]["values"] == [5.0,6.0,7.0,8.0]):
            break
        time.sleep(1)
    assert i < 100

# Test updating a model with a new resource request but same image
def test_rolling_update3():
    run("kubectl delete sdep --all", shell=True)
    print("sleeping 30 secs...")
    time.sleep(30)
    run("kubectl apply -f ../resources/graph1.json", shell=True, check=True)
    wait_for_rollout("mymodel-mymodel-svc-orch")
    wait_for_rollout("mymodel-mymodel-e2eb561")
    r = rest_request_api_gateway("oauth-key","oauth-secret")        
    assert r.status_code == 200
    assert r.json()["data"]["tensor"]["values"] == [1.0,2.0,3.0,4.0]
    run("kubectl apply -f ../resources/graph4.json", shell=True, check=True)
    i = 0
    for i in range(50):
        r = rest_request_api_gateway("oauth-key","oauth-secret")
        res = r.json()
        print(res)
        assert r.status_code == 200
        assert ((res["meta"]["requestPath"]["complex-model"] == "seldonio/fixed-model:0.1" and res["data"]["tensor"]["values"] == [1.0,2.0,3.0,4.0]))
        time.sleep(1)
    assert i == 49


# Test updating a model with a multi deployment new model
def test_rolling_update4():
    run("kubectl delete sdep --all", shell=True)
    print("sleeping 30 secs...")
    time.sleep(30)
    run("kubectl apply -f ../resources/graph1.json", shell=True, check=True)
    wait_for_rollout("mymodel-mymodel-svc-orch")
    wait_for_rollout("mymodel-mymodel-e2eb561")
    r = rest_request_api_gateway("oauth-key","oauth-secret")        
    assert r.status_code == 200
    assert r.json()["data"]["tensor"]["values"] == [1.0,2.0,3.0,4.0]
    run("kubectl apply -f ../resources/graph5.json", shell=True, check=True)
    i = 0
    for i in range(50):
        r = rest_request_api_gateway("oauth-key","oauth-secret")
        res = r.json()
        print(res)
        assert r.status_code == 200
        assert (("complex-model" in res["meta"]["requestPath"] and res["meta"]["requestPath"]["complex-model"] == "seldonio/fixed-model:0.1" and res["data"]["tensor"]["values"] == [1.0,2.0,3.0,4.0]) or (res["meta"]["requestPath"]["model1"] == "seldonio/fixed-model:0.1" and res["data"]["tensor"]["values"] == [1.0,2.0,3.0,4.0] and res["meta"]["requestPath"]["model2"] == "seldonio/fixed-model:0.1"))
        if (not r.status_code == 200) or ("model1" in res["meta"]["requestPath"]):
            break
        time.sleep(1)
    assert i < 100

# Test updating a model to a multi predictor model
def test_rolling_update5():
    run("kubectl delete sdep --all", shell=True)
    print("sleeping 30 secs...")
    time.sleep(30)
    run("kubectl apply -f ../resources/graph1.json", shell=True, check=True)
    wait_for_rollout("mymodel-mymodel-svc-orch")
    wait_for_rollout("mymodel-mymodel-e2eb561")
    r = rest_request_api_gateway("oauth-key","oauth-secret")        
    assert r.status_code == 200
    assert r.json()["data"]["tensor"]["values"] == [1.0,2.0,3.0,4.0]
    run("kubectl apply -f ../resources/graph6.json", shell=True, check=True)
    i = 0
    for i in range(50):
        r = rest_request_api_gateway("oauth-key","oauth-secret")
        res = r.json()
        print(res)
        assert r.status_code == 200
        assert (("complex-model" in res["meta"]["requestPath"] and res["meta"]["requestPath"]["complex-model"] == "seldonio/fixed-model:0.1" and res["data"]["tensor"]["values"] == [1.0,2.0,3.0,4.0]) or (res["meta"]["requestPath"]["complex-model"] == "seldonio/fixed-model:0.2" and res["data"]["tensor"]["values"] == [5.0,6.0,7.0,8.0]))
        if (not r.status_code == 200) or (res["data"]["tensor"]["values"] == [5.0,6.0,7.0,8.0]):        
            break
        time.sleep(1)
    assert i < 100
    
