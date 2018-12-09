from subprocess import run,Popen
import signal
import subprocess
import os
API_AMBASSADOR="localhost:8003"
API_GATEWAY_REST="localhost:8002"
API_GATEWAY_GRPC="localhost:8004"

def create_seldon_single_namespace(request):
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
    def fin():
        run("helm delete seldon-core --purge", shell=True)
        run("helm delete seldon-core-crd --purge", shell=True)
        
    request.addfinalizer(fin)


def port_forward(request):
    print("Setup: Port forward")
    p1 = Popen("kubectl port-forward $(kubectl get pods -n seldon -l app=seldon-apiserver-container-app -o jsonpath='{.items[0].metadata.name}') -n seldon 8002:8080",stdout=subprocess.PIPE,shell=True, preexec_fn=os.setsid) 
    p2 = Popen("kubectl port-forward $(kubectl get pods -n seldon -l service=ambassador -o jsonpath='{.items[0].metadata.name}') -n seldon 8003:8080",stdout=subprocess.PIPE,shell=True, preexec_fn=os.setsid) 
    p3 = Popen("kubectl port-forward $(kubectl get pods -n seldon -l app=seldon-apiserver-container-app -o jsonpath='{.items[0].metadata.name}') -n seldon 8004:5000",stdout=subprocess.PIPE,shell=True, preexec_fn=os.setsid) 
    #, stdout=subprocess.PIPE
    def fin():        
        print("teardown port forward")
        os.killpg(os.getpgid(p1.pid), signal.SIGTERM)
        os.killpg(os.getpgid(p2.pid), signal.SIGTERM)
        os.killpg(os.getpgid(p3.pid), signal.SIGTERM)        
        
    request.addfinalizer(fin)

