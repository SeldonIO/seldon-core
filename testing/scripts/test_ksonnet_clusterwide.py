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

@pytest.mark.usefixtures("seldon_java_images")
@pytest.mark.usefixtures("clusterwide_seldon_ksonnet")
class TestSingleNamespace(object):
    
    # Test singe model helm script with 4 API methods
    def test_single_model(self):
        run('cd my-model && ks delete default && ks component rm mymodel', shell=True)
        run('kubectl delete sdep --all', shell=True)
        run('cd my-model && ks generate seldon-serve-simple-v1alpha2 mymodel --image seldonio/mock_classifier:1.0 --oauthKey=oauth-key --oauthSecret=oauth-secret && ks apply default -c mymodel', shell=True, check=True)       
        wait_for_rollout("mymodel-mymodel-025d03d")
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
        r = grpc_request_ambassador("mymodel","test1",API_AMBASSADOR)        
        print(r)
        r = grpc_request_api_gateway("oauth-key","oauth-secret","test1",rest_endpoint=API_GATEWAY_REST,grpc_endpoint=API_GATEWAY_GRPC)
        print(r)
        run('cd my-model && ks delete default -c mymodel && ks component rm mymodel', shell=True)        

    # Test AB Test model helm script with 4 API methods
    def test_abtest_model(self):
        run('cd my-model && ks delete default && ks component rm mymodel', shell=True)
        run('kubectl delete sdep --all', shell=True)
        run('cd my-model && ks generate seldon-abtest-v1alpha2 myabtest --imageA seldonio/mock_classifier:1.0 --imageB seldonio/mock_classifier:1.0 --oauthKey=oauth-key --oauthSecret=oauth-secret &&     ks apply default -c myabtest', shell=True)
        wait_for_rollout("myabtest-myabtest-41de5b8")
        wait_for_rollout("myabtest-myabtest-df66c5c")        
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
        r = grpc_request_ambassador("myabtest","test1",API_AMBASSADOR)        
        print(r)
        r = grpc_request_api_gateway("oauth-key","oauth-secret","test1",rest_endpoint=API_GATEWAY_REST,grpc_endpoint=API_GATEWAY_GRPC)
        print(r)
        run('cd my-model &&     ks delete default -c myabtest &&  ks component rm myabtest', shell=True)

    # Test MAB Test model helm script with 4 API methods
    def test_mab_model(self):
        run('cd my-model && ks delete default && ks component rm mymab', shell=True)
        run('kubectl delete sdep --all', shell=True)
        run('cd my-model &&     ks generate seldon-mab-v1alpha2 mymab --imageA seldonio/mock_classifier:1.0 --imageB seldonio/mock_classifier:1.0 --oauthKey=oauth-key --oauthSecret=oauth-secret &&     ks apply default -c mymab', shell=True) 
        wait_for_rollout("mymab-mymab-41de5b8")
        wait_for_rollout("mymab-mymab-b8038b2")
        wait_for_rollout("mymab-mymab-df66c5c")                
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
        r = grpc_request_ambassador("mymab","test1",API_AMBASSADOR)        
        print(r)
        r = grpc_request_api_gateway("oauth-key","oauth-secret","test1",rest_endpoint=API_GATEWAY_REST,grpc_endpoint=API_GATEWAY_GRPC)
        print(r)
        run('cd my-model && ks delete default && ks component rm mymab', shell=True)        
        
