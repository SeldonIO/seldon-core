from subprocess import run, Popen
import signal
import subprocess
import os
import time

from retrying import retry

API_AMBASSADOR = "localhost:8003"
API_GATEWAY_REST = "localhost:8002"
API_GATEWAY_GRPC = "localhost:8004"


def setup_k8s():
    run("kubectl create namespace seldon", shell=True)
    run("kubectl create namespace test1", shell=True)
    run("kubectl config set-context $(kubectl config current-context) --namespace=seldon", shell=True)
    run(
        "kubectl create clusterrolebinding kube-system-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:default",
        shell=True)


def setup_helm():
    run("kubectl -n kube-system create sa tiller", shell=True)
    run("kubectl create clusterrolebinding tiller --clusterrole cluster-admin --serviceaccount=kube-system:tiller",
        shell=True)
    run("helm init --service-account tiller", shell=True)
    run("kubectl rollout status deploy/tiller-deploy -n kube-system", shell=True)


def wait_for_shutdown(deploymentName):
    ret = run("kubectl get deploy/" + deploymentName, shell=True)
    while ret.returncode == 0:
        time.sleep(1)
        ret = run("kubectl get deploy/" + deploymentName, shell=True)


def setup_finalizer_helm(request):
    def fin():
        run("helm delete seldon-core --purge", shell=True)
        run("helm delete ambassador --purge", shell=True)
        run("helm delete seldon-gateway --purge", shell=True)
        run("kubectl delete namespace seldon-system", shell=True)
        run("kubectl delete namespace seldon", shell=True)
        run("kubectl delete namespace test1", shell=True)

    if not request is None:
        request.addfinalizer(fin)


def get_seldon_version():
    completedProcess = Popen("grep version: ../../helm-charts/seldon-core-operator/Chart.yaml | cut -d: -f2", shell=True, stdout=subprocess.PIPE)
    output = completedProcess.stdout.readline()
    version = output.decode('utf-8').strip()
    return version


# Not using latest Ambassador due to gRPC issue. https://github.com/datawire/ambassador/issues/504
def create_ambassador():
    run("helm install stable/ambassador --name ambassador --set crds.keep=false --namespace seldon", shell=True)
    run("kubectl rollout status deployment.apps/ambassador --namespace seldon", shell=True)


def create_seldon_gateway(version):
    cmd="helm install ../../helm-charts/seldon-core-oauth-gateway --name seldon-gateway --namespace seldon --set singleNamespace=false --set image.repository=127.0.0.1:5000/seldonio/apife --set image.tag=" + version + " --set image.pullPolicy=Always"
    print(cmd)
    run(cmd, shell=True)
    run("kubectl rollout status deployment.apps/seldon-gateway-seldon-apiserver  --namespace seldon", shell=True)


def create_seldon_clusterwide_helm(request, version):
    setup_k8s()
    setup_helm()
    run(
        "helm install ../../helm-charts/seldon-core-operator --name seldon-core --namespace seldon-system --set image.repository=127.0.0.1:5000/seldonio/seldon-core-operator --set image.tag=" + version + " --set image.pullPolicy=Always --set engine.image.repository=127.0.0.1:5000/seldonio/engine --set engine.image.tag=" + version + " --set engine.image.pullPolicy=Always",
        shell=True)
    run('kubectl rollout status statefulset.apps/seldon-operator-controller-manager -n seldon-system', shell=True)
    create_ambassador()
    create_seldon_gateway(version)
    if not request is None:
        setup_finalizer_helm(request)


@retry(wait_exponential_multiplier=1000, wait_exponential_max=10000, stop_max_attempt_number=5)
def port_forward(request):
    print("Setup: Port forward")
    p1 = Popen("kubectl port-forward $(kubectl get pods -n seldon -l app=seldon-apiserver-container-app -o jsonpath='{.items[0].metadata.name}') -n seldon 8002:8080",stdout=subprocess.PIPE,shell=True, preexec_fn=os.setsid)
    p2 = Popen(
        "kubectl port-forward $(kubectl get pods -n seldon -l app.kubernetes.io/name=ambassador -o jsonpath='{.items[0].metadata.name}') -n seldon 8003:8080",
        stdout=subprocess.PIPE, shell=True, preexec_fn=os.setsid)

    p3 = Popen("kubectl port-forward $(kubectl get pods -n seldon -l app=seldon-apiserver-container-app -o jsonpath='{.items[0].metadata.name}') -n seldon 8004:5000",stdout=subprocess.PIPE,shell=True, preexec_fn=os.setsid)
    # , stdout=subprocess.PIPE
    def fin():
        print("teardown port forward")
        os.killpg(os.getpgid(p1.pid), signal.SIGTERM)
        os.killpg(os.getpgid(p2.pid), signal.SIGTERM)
        os.killpg(os.getpgid(p3.pid), signal.SIGTERM)

    if not request is None:
        request.addfinalizer(fin)


def create_docker_repo(request):
    run('kubectl apply -f ../resources/docker-private-registry.json -n default', shell=True)
    run('kubectl rollout status deploy/docker-private-registry-deployment -n default', shell=True)
    run('kubectl apply -f ../resources/docker-private-registry-proxy.json -n default', shell=True)

    def fin():
        return
        run('kubectl delete -f ../resources/docker-private-registry.json --ignore-not-found=true -n default',
            shell=True)
        run('kubectl delete -f ../resources/docker-private-registry-proxy.json --ignore-not-found=true -n default',
            shell=True)

    if not request is None:
        request.addfinalizer(fin)


@retry(wait_exponential_multiplier=1000, wait_exponential_max=10000, stop_max_attempt_number=5)
def port_forward_docker_repo(request):
    print("port-forward docker")
    p1 = Popen(
        "POD_NAME=$(kubectl get pods -l app=docker-private-registry -n default |sed -e '1d'|awk '{print $1}') && kubectl port-forward ${POD_NAME} 5000:5000 -n default",
        stdout=subprocess.PIPE, shell=True, preexec_fn=os.setsid)

    def fin():
        print("teardown port-foward docker")
        os.killpg(os.getpgid(p1.pid), signal.SIGTERM)

    if not request is None:
        request.addfinalizer(fin)
