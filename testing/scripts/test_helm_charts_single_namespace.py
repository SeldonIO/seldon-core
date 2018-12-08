import pytest
import time
import subprocess
from subprocess import run,Popen
from seldon_utils import *
API_AMBASSADOR="localhost:8003"
API_GATEWAY_REST="localhost:8002"
API_GATEWAY_GRPC="localhost:8004"

def create_seldon(request):
    run("kubectl create namespace seldon", shell=True)
    run("kubectl config set-context $(kubectl config current-context) --namespace=seldon", shell=True)    
    run("kubectl create clusterrolebinding kube-system-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:default", shell=True)    
    run("kubectl -n kube-system create sa tiller", shell=True)
    run("kubectl create clusterrolebinding tiller --clusterrole cluster-admin --serviceaccount=kube-system:tiller", shell=True)
    run("helm init --service-account tiller", shell=True)
    run("kubectl rollout status deploy/tiller-deploy -n kube-system", shell=True)        
    run("helm install ../../helm-charts/seldon-core-crd --name seldon-core-crd --set usage_metrics.enabled=true", shell=True)
    run("helm install ../../helm-charts/seldon-core --name seldon-core --namespace seldon  --set ambassador.enabled=true", shell=True)        
    run('kubectl rollout status deploy/seldon-core-seldon-cluster-manager', shell=True)        
    run('kubectl rollout status deploy/seldon-core-seldon-apiserver', shell=True)        
    run('kubectl rollout status deploy/seldon-core-ambassador', shell=True)        


def port_forward(request):
    print("Setup: Port forward")
    p1 = Popen("kubectl port-forward $(kubectl get pods -n seldon -l app=seldon-apiserver-container-app -o jsonpath='{.items[0].metadata.name}') -n seldon 8002:8080", shell=True)
    p2 = Popen("kubectl port-forward $(kubectl get pods -n seldon -l service=ambassador -o jsonpath='{.items[0].metadata.name}') -n seldon 8003:8080", shell=True)
    p3 = Popen("kubectl port-forward $(kubectl get pods -n seldon -l app=seldon-apiserver-container-app -o jsonpath='{.items[0].metadata.name}') -n seldon 8004:5000", shell=True)    
    #, stdout=subprocess.PIPE
    def fin():
        print("teardown port forward")
        p1.kill()
        time.sleep(1)
        p2.kill()
        time.sleep(2)
        p3.kill()
        time.sleep(2)
        
    request.addfinalizer(fin)

@pytest.fixture(scope="session",autouse=True)
def pre_test_setup(request):
    create_seldon(request)
    port_forward(request)

    
def wait_for_shutdown(deploymentName):
    ret = run("kubectl get deploy/"+deploymentName, shell=True)
    while ret.returncode == 0:
        time.sleep(1)
        ret = run("kubectl get deploy/"+deploymentName, shell=True)


def wait_for_rollout(deploymentName):
    ret = run("kubectl rollout status deploy/"+deploymentName, shell=True)
    while ret.returncode > 0:
        time.sleep(1)
        ret = run("kubectl rollout status deploy/"+deploymentName, shell=True)    

def initial_rest_request():
    r = rest_request_api_gateway("oauth-key","oauth-secret",None,API_GATEWAY_REST)        
    if not r.status_code == 200:
        time.sleep(1)
        r = rest_request_api_gateway("oauth-key","oauth-secret",None,API_GATEWAY_REST)
        if not r.status_code == 200:
            time.sleep(5)
            r = rest_request_api_gateway("oauth-key","oauth-secret",None,API_GATEWAY_REST)
    return r

class TestSingleNamespace(object):
    
    # Test singe model helm script with 4 API methods
    def test_single_model(self):
        run("helm delete mymodel --purge", shell=True)
        run("helm install ../../helm-charts/seldon-single-model --name mymodel --set oauth.key=oauth-key --set oauth.secret=oauth-secret", shell=True, check=True)
        wait_for_rollout("mymodel-mymodel-7cd068f")
        r = initial_rest_request()
        r = rest_request_api_gateway("oauth-key","oauth-secret",None,API_GATEWAY_REST)
        res = r.json()
        print(res)
        assert r.status_code == 200
        assert len(r.json()["data"]["tensor"]["values"]) == 1
        r = rest_request_ambassador("mymodel",None,API_AMBASSADOR)
        res = r.json()
        print(res)
        assert r.status_code == 200
        assert len(r.json()["data"]["tensor"]["values"]) == 1
        r = grpc_request_ambassador("mymodel",None,API_AMBASSADOR)        
        print(r)
        r = grpc_request_api_gateway("oauth-key","oauth-secret",None,rest_endpoint=API_GATEWAY_REST,grpc_endpoint=API_GATEWAY_GRPC)
        print(r)
        run("helm delete mymodel --purge", shell=True)        

    # Test AB Test model helm script with 4 API methods
    def test_abtest_model(self):
        run("helm delete myabtest --purge", shell=True)
        run("helm install ../../helm-charts/seldon-abtest --name myabtest --set oauth.key=oauth-key --set oauth.secret=oauth-secret", shell=True, check=True)
        wait_for_rollout("myabtest-abtest-41de5b8")
        wait_for_rollout("myabtest-abtest-df66c5c")        
        r = initial_rest_request()
        r = rest_request_api_gateway("oauth-key","oauth-secret",None,API_GATEWAY_REST)
        res = r.json()
        print(res)
        assert r.status_code == 200
        assert len(r.json()["data"]["tensor"]["values"]) == 1
        r = rest_request_ambassador("myabtest",None,API_AMBASSADOR)
        res = r.json()
        print(res)
        assert r.status_code == 200
        assert len(r.json()["data"]["tensor"]["values"]) == 1
        r = grpc_request_ambassador("myabtest",None,API_AMBASSADOR)        
        print(r)
        r = grpc_request_api_gateway("oauth-key","oauth-secret",None,rest_endpoint=API_GATEWAY_REST,grpc_endpoint=API_GATEWAY_GRPC)
        print(r)
        run("helm delete myabtest --purge", shell=True)        

    # Test MAB Test model helm script with 4 API methods
    def test_mab_model(self):
        run("helm delete mymab --purge", shell=True)
        run("helm install ../../helm-charts/seldon-mab --name mymab --set oauth.key=oauth-key --set oauth.secret=oauth-secret", shell=True, check=True)
        wait_for_rollout("mymab-abtest-41de5b8")
        wait_for_rollout("mymab-abtest-b8038b2")
        wait_for_rollout("mymab-abtest-df66c5c")                
        r = initial_rest_request()
        r = rest_request_api_gateway("oauth-key","oauth-secret",None,API_GATEWAY_REST)
        res = r.json()
        print(res)
        assert r.status_code == 200
        assert len(r.json()["data"]["tensor"]["values"]) == 1
        r = rest_request_ambassador("mymab",None,API_AMBASSADOR)
        res = r.json()
        print(res)
        assert r.status_code == 200
        assert len(r.json()["data"]["tensor"]["values"]) == 1
        r = grpc_request_ambassador("mymab",None,API_AMBASSADOR)        
        print(r)
        r = grpc_request_api_gateway("oauth-key","oauth-secret",None,rest_endpoint=API_GATEWAY_REST,grpc_endpoint=API_GATEWAY_GRPC)
        print(r)
        run("helm delete mymab --purge", shell=True)        
        
