import pytest
import time
import subprocess
from subprocess import run,Popen
from seldon_utils import *
from k8s_utils import *

def wait_for_shutdown(deploymentName):
    ret = run("kubectl get -n test1 deploy/"+deploymentName, shell=True)
    while ret.returncode == 0:
        time.sleep(1)
        ret = run("kubectl get -n test1 deploy/"+deploymentName, shell=True)

def wait_for_rollout(deploymentName):
    ret = run("kubectl rollout status -n test1 deploy/"+deploymentName, shell=True)
    while ret.returncode > 0:
        time.sleep(1)
        ret = run("kubectl rollout status -n test1 deploy/"+deploymentName, shell=True)    

def initial_rest_request():
    r = rest_request_api_gateway("oauth-key","oauth-secret","test1",API_GATEWAY_REST)        
    if not r.status_code == 200:
        time.sleep(1)
        r = rest_request_api_gateway("oauth-key","oauth-secret","test1",API_GATEWAY_REST)
        if not r.status_code == 200:
            time.sleep(5)
            r = rest_request_api_gateway("oauth-key","oauth-secret","test1",API_GATEWAY_REST)
    return r

@pytest.mark.usefixtures("seldon_images")
@pytest.mark.usefixtures("clusterwide_seldon_helm")
class TestClusterWide(object):
    
    # Test singe model helm script with 4 API methods
    def test_single_model(self):
        run("helm delete mymodel --purge", shell=True)
        run("helm install ../../helm-charts/seldon-single-model --name mymodel --set oauth.key=oauth-key --set oauth.secret=oauth-secret --namespace test1", shell=True, check=True)
        wait_for_rollout("mymodel-mymodel-7cd068f")
        r = initial_rest_request()
        r = rest_request_api_gateway("oauth-key","oauth-secret","test1",API_GATEWAY_REST)
        res = r.json()
        print(res)
        assert r.status_code == 200
        assert len(r.json()["data"]["tensor"]["values"]) == 1
        r = rest_request_ambassador("mymodel","test1",API_AMBASSADOR)
        res = r.json()
        print(res)
        assert r.status_code == 200
        assert len(r.json()["data"]["tensor"]["values"]) == 1
        r = grpc_request_api_gateway2("oauth-key","oauth-secret","test1",rest_endpoint=API_GATEWAY_REST,grpc_endpoint=API_GATEWAY_GRPC)
        print(r)
        r = grpc_request_ambassador2("mymodel","test1",API_AMBASSADOR)
        print(r)
        run("helm delete mymodel --purge", shell=True)        

    # Test AB Test model helm script with 4 API methods
    def test_abtest_model(self):
        run("helm delete myabtest --purge", shell=True)
        run("helm install ../../helm-charts/seldon-abtest --name myabtest --set oauth.key=oauth-key --set oauth.secret=oauth-secret --namespace test1", shell=True, check=True)
        wait_for_rollout("myabtest-abtest-41de5b8")
        wait_for_rollout("myabtest-abtest-df66c5c")        
        r = initial_rest_request()
        r = rest_request_api_gateway("oauth-key","oauth-secret","test1",API_GATEWAY_REST)
        res = r.json()
        print(res)
        assert r.status_code == 200
        assert len(r.json()["data"]["tensor"]["values"]) == 1
        r = rest_request_ambassador("myabtest","test1",API_AMBASSADOR)
        res = r.json()
        print(res)
        assert r.status_code == 200
        assert len(r.json()["data"]["tensor"]["values"]) == 1
        r = grpc_request_api_gateway2("oauth-key","oauth-secret","test1",rest_endpoint=API_GATEWAY_REST,grpc_endpoint=API_GATEWAY_GRPC)
        print(r)
        r = grpc_request_ambassador2("myabtest","test1",API_AMBASSADOR)
        print(r)
        run("helm delete myabtest --purge", shell=True)        

    # Test MAB Test model helm script with 4 API methods
    def test_mab_model(self):
        run("helm delete mymab --purge", shell=True)
        run("helm install ../../helm-charts/seldon-mab --name mymab --set oauth.key=oauth-key --set oauth.secret=oauth-secret --namespace test1", shell=True, check=True)
        wait_for_rollout("mymab-abtest-41de5b8")
        wait_for_rollout("mymab-abtest-b8038b2")
        wait_for_rollout("mymab-abtest-df66c5c")                
        r = initial_rest_request()
        r = rest_request_api_gateway("oauth-key","oauth-secret","test1",API_GATEWAY_REST)
        res = r.json()
        print(res)
        assert r.status_code == 200
        assert len(r.json()["data"]["tensor"]["values"]) == 1
        r = rest_request_ambassador("mymab","test1",API_AMBASSADOR)
        res = r.json()
        print(res)
        assert r.status_code == 200
        assert len(r.json()["data"]["tensor"]["values"]) == 1
        r = grpc_request_api_gateway2("oauth-key","oauth-secret","test1",rest_endpoint=API_GATEWAY_REST,grpc_endpoint=API_GATEWAY_GRPC)
        print(r)
        r = grpc_request_ambassador2("mymab","test1",API_AMBASSADOR)
        print(r)
        run("helm delete mymab --purge", shell=True)        
        
