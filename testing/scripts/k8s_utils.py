from subprocess import run,Popen
import signal
import subprocess
import os
import time
import random
from retrying import retry
API_AMBASSADOR="localhost:8003"
API_GATEWAY_REST="localhost:8002"
API_GATEWAY_GRPC="localhost:8004"

def setup_k8s():
    run("kubectl create namespace seldon", shell=True)
    run("kubectl create namespace test1", shell=True)    
    run("kubectl config set-context $(kubectl config current-context) --namespace=seldon", shell=True)    
    run("kubectl create clusterrolebinding kube-system-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:default", shell=True)    

def setup_helm():
    run("kubectl -n kube-system create sa tiller", shell=True)
    run("kubectl create clusterrolebinding tiller --clusterrole cluster-admin --serviceaccount=kube-system:tiller", shell=True)
    run("helm init --service-account tiller", shell=True)
    run("kubectl rollout status deploy/tiller-deploy -n kube-system", shell=True)        

def wait_for_shutdown(deploymentName):
    ret = run("kubectl get deploy/"+deploymentName, shell=True)
    while ret.returncode == 0:
        time.sleep(1)
        ret = run("kubectl get deploy/"+deploymentName, shell=True)
        
def wait_seldon_ready():
    run('kubectl rollout status deploy/seldon-core-seldon-cluster-manager', shell=True)        
    run('kubectl rollout status deploy/seldon-core-seldon-apiserver', shell=True)        
    run('kubectl rollout status deploy/seldon-core-ambassador', shell=True)        

def setup_finalizer_helm(request):
    def fin():
        run("helm delete seldon-core --purge", shell=True)
        run("helm delete seldon-core-crd --purge", shell=True)
        run("kubectl delete namespace seldon", shell=True)
        run("kubectl delete namespace test1", shell=True)    
    request.addfinalizer(fin)

def setup_finalizer_ksonnet(request):
    def fin():
        run('cd my-ml-deployment && ks delete default', shell=True)
        wait_for_shutdown("seldon-core-seldon-cluster-manager")
        wait_for_shutdown("seldon-core-seldon-apiserver")
        wait_for_shutdown("seldon-core-ambassador")        
        run('rm -rf my-ml-deployment', shell=True)
        run("kubectl delete namespace seldon", shell=True)
        run("kubectl delete namespace test1", shell=True)            
    request.addfinalizer(fin)

def get_seldon_version():
    completedProcess = Popen("cat ../../cluster-manager/target/version.txt", shell=True, stdout=subprocess.PIPE)
    output = completedProcess.stdout.readline()
    version = output.decode('utf-8').rstrip()
    return version

    
def create_seldon_single_namespace_helm(request,version):
    setup_k8s()
    setup_helm()
    run("helm install ../../helm-charts/seldon-core-crd --name seldon-core-crd --set usage_metrics.enabled=true", shell=True)
    run("helm install ../../helm-charts/seldon-core --name seldon-core --namespace seldon  --set ambassador.enabled=true --set apife.image.name=127.0.0.1:5000/seldonio/apife:"+version+" --set cluster_manager.image.name=127.0.0.1:5000/seldonio/cluster-manager:"+version+" --set engine.image.name=127.0.0.1:5000/seldonio/engine:"+version, shell=True)
    wait_seldon_ready()
    setup_finalizer_helm(request)

def create_seldon_clusterwide_helm(request,version):
    setup_k8s()
    setup_helm()    
    run("helm install ../../helm-charts/seldon-core-crd --name seldon-core-crd --set usage_metrics.enabled=true", shell=True)
    run("helm install ../../helm-charts/seldon-core --name seldon-core --namespace seldon  --set single_namespace=false --set ambassador.rbac.namespaced=false --set ambassador.enabled=true --set ambassador.env.AMBASSADOR_SINGLE_NAMESPACE=false --set apife.image.name=127.0.0.1:5000/seldonio/apife:"+version+" --set cluster_manager.image.name=127.0.0.1:5000/seldonio/cluster-manager:"+version+" --set engine.image.name=127.0.0.1:5000/seldonio/engine:"+version, shell=True)
    wait_seldon_ready()
    setup_finalizer_helm(request)

def create_seldon_single_namespace_ksonnet(request):
    setup_k8s()
    run('rm -rf my-ml-deployment && ks init my-ml-deployment ', shell=True)
    run('cd my-ml-deployment &&     ks registry add seldon-core ../../../seldon-core &&     ks pkg install seldon-core/seldon-core@master &&     ks generate seldon-core seldon-core --withApife=true --withAmbassador=true --singleNamespace=true --namespace=seldon --withRbac=true --registry=127.0.0.1:5000', shell=True)
    run('cd my-ml-deployment &&       ks apply default', shell=True)
    run('rm -rf my-model && ks init my-model --namespace seldon', shell=True)
    run('cd my-model && ks registry add seldon-core ../../../seldon-core && ks pkg install seldon-core/seldon-core@master', shell=True)
    wait_seldon_ready()
    setup_finalizer_ksonnet(request)

def create_seldon_clusterwide_ksonnet(request):
    setup_k8s()
    run('rm -rf my-ml-deployment && ks init my-ml-deployment ', shell=True)
    run('cd my-ml-deployment &&     ks registry add seldon-core ../../../seldon-core &&     ks pkg install seldon-core/seldon-core@master &&     ks generate seldon-core seldon-core --withApife=true --withAmbassador=true --singleNamespace=false --namespace=seldon --withRbac=true --registry=127.0.0.1:5000', shell=True)
    run('cd my-ml-deployment &&       ks apply default', shell=True)
    run('rm -rf my-model && ks init my-model --namespace test1', shell=True)
    run('cd my-model && ks registry add seldon-core ../../../seldon-core && ks pkg install seldon-core/seldon-core@master', shell=True)
    wait_seldon_ready()
    setup_finalizer_ksonnet(request)

@retry(wait_exponential_multiplier=1000, wait_exponential_max=10000, stop_max_attempt_number=7)
def port_forward(request):
    print("Setup: Port forward")
    p1 = Popen("kubectl port-forward $(kubectl get pods -n seldon -l app=seldon-apiserver-container-app -o jsonpath='{.items[0].metadata.name}') -n seldon 8002:8080",stdout=subprocess.PIPE,shell=True, preexec_fn=os.setsid) 
    p2 = Popen("kubectl port-forward $(kubectl get pods -n seldon -l app.kubernetes.io/name=ambassador -o jsonpath='{.items[0].metadata.name}') -n seldon 8003:8080",stdout=subprocess.PIPE,shell=True, preexec_fn=os.setsid)
    p3 = Popen("kubectl port-forward $(kubectl get pods -n seldon -l app=seldon-apiserver-container-app -o jsonpath='{.items[0].metadata.name}') -n seldon 8004:5000",stdout=subprocess.PIPE,shell=True, preexec_fn=os.setsid) 
    #, stdout=subprocess.PIPE
    def fin():        
        print("teardown port forward")
        os.killpg(os.getpgid(p1.pid), signal.SIGTERM)
        os.killpg(os.getpgid(p2.pid), signal.SIGTERM)
        os.killpg(os.getpgid(p3.pid), signal.SIGTERM)        
        
    request.addfinalizer(fin)

def create_docker_repo(request):
    run('kubectl apply -f ../resources/docker-private-registry.json -n default', shell=True)
    run('kubectl rollout status deploy/docker-private-registry-deployment -n default', shell=True)
    run('kubectl apply -f ../resources/docker-private-registry-proxy.json -n default', shell=True)

    def fin():
        run('kubectl delete -f ../resources/docker-private-registry.json --ignore-not-found=true -n default', shell=True)
        run('kubectl delete -f ../resources/docker-private-registry-proxy.json --ignore-not-found=true -n default', shell=True)
        
    request.addfinalizer(fin)        

@retry(wait_exponential_multiplier=1000, wait_exponential_max=10000, stop_max_attempt_number=7)
def port_forward_docker_repo(request):
    print("port-forward docker")
    p1 = Popen("POD_NAME=$(kubectl get pods -l app=docker-private-registry -n default |sed -e '1d'|awk '{print $1}') && kubectl port-forward ${POD_NAME} 5000:5000 -n default",stdout=subprocess.PIPE,shell=True, preexec_fn=os.setsid)

    def fin():
        print("teardown port-foward docker")
        os.killpg(os.getpgid(p1.pid), signal.SIGTERM)

    request.addfinalizer(fin)

    
