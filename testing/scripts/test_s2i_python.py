import pytest
import time
import subprocess
from subprocess import run, Popen
from seldon_utils import *
from k8s_utils import *
import numpy as np

S2I_CREATE = "cd ../s2i/python/#TYPE# && s2i build -E environment_#API# . seldonio/seldon-core-s2i-python3:#VERSION# 127.0.0.1:5000/seldonio/test#TYPE#_#API#:0.1"
IMAGE_NAME = "127.0.0.1:5000/seldonio/test#TYPE#_#API#:0.1"


def create_s2I_image(s2i_python_version, component_type, api_type):
    cmd = S2I_CREATE.replace("#TYPE#", component_type).replace("#API#", api_type).replace("#VERSION#",
                                                                                          s2i_python_version)
    print(cmd)
    run(cmd, shell=True, check=True)


def push_s2i_image(component_type, api_type):
    img = get_image_name(component_type, api_type)
    run("docker push " + img, shell=True, check=True)


def get_image_name(component_type, api_type):
    return IMAGE_NAME.replace("#TYPE#", component_type).replace("#API#", api_type)


def create_push_s2i_image(s2i_python_version, component_type, api_type):
    create_s2I_image(s2i_python_version, component_type, api_type)
    push_s2i_image(component_type, api_type)


@pytest.mark.usefixtures("seldon_images")
@pytest.mark.usefixtures("setup_python_s2i")
@pytest.mark.usefixtures("s2i_python_version")
class TestPythonS2i(object):

    def test_build_router_rest(self, s2i_python_version):
        create_s2I_image(s2i_python_version, "router", "rest")
        img = get_image_name("router", "rest")
        run("docker run -d --rm --name 'router' " + img, shell=True, check=True)
        time.sleep(2)
        run("docker rm -f router", shell=True, check=True)

    def test_build_router_grpc(self, s2i_python_version):
        create_s2I_image(s2i_python_version, "router", "grpc")
        img = get_image_name("router", "grpc")
        run("docker run -d --rm --name 'router' " + img, shell=True, check=True)
        time.sleep(2)
        run("docker rm -f router", shell=True, check=True)

    def test_build_model_rest(self, s2i_python_version):
        create_s2I_image(s2i_python_version, "model", "rest")
        img = get_image_name("model", "rest")
        run("docker run -d --rm --name 'model' " + img, shell=True, check=True)
        time.sleep(2)
        run("docker rm -f model", shell=True, check=True)

    def test_build_model_grpc(self, s2i_python_version):
        create_s2I_image(s2i_python_version, "model", "grpc")
        img = get_image_name("model", "grpc")
        run("docker run -d --rm --name 'model' " + img, shell=True, check=True)
        time.sleep(2)
        run("docker rm -f model", shell=True, check=True)

    def test_build_transformer_rest(self, s2i_python_version):
        create_s2I_image(s2i_python_version, "transformer", "rest")
        img = get_image_name("transformer", "rest")
        run("docker run -d --rm --name 'transformer' " + img, shell=True, check=True)
        time.sleep(2)
        run("docker rm -f transformer", shell=True, check=True)

    def test_build_transformer_grpc(self, s2i_python_version):
        create_s2I_image(s2i_python_version, "transformer", "grpc")
        img = get_image_name("transformer", "grpc")
        run("docker run -d --rm --name 'transformer' " + img, shell=True, check=True)
        time.sleep(2)
        run("docker rm -f transformer", shell=True, check=True)

    def test_build_combiner_rest(self, s2i_python_version):
        create_s2I_image(s2i_python_version, "combiner", "rest")
        img = get_image_name("combiner", "rest")
        print(img)
        run("docker run -d --rm --name 'combiner' " + img, shell=True, check=True)
        time.sleep(2)
        run("docker rm -f combiner", shell=True, check=True)

    def test_build_combiner_grpc(self, s2i_python_version):
        create_s2I_image(s2i_python_version, "combiner", "grpc")
        img = get_image_name("combiner", "grpc")
        run("docker run -d --rm --name 'combiner' " + img, shell=True, check=True)
        time.sleep(2)
        run("docker rm -f combiner", shell=True, check=True)


def wait_for_rollout(deploymentName):
    ret = run("kubectl rollout status deploy/" + deploymentName, shell=True)
    while ret.returncode > 0:
        time.sleep(1)
        ret = run("kubectl rollout status deploy/" + deploymentName, shell=True)


@pytest.mark.usefixtures("seldon_images")
@pytest.mark.usefixtures("setup_python_s2i")
@pytest.mark.usefixtures("s2i_python_version")
@pytest.mark.usefixtures("clusterwide_seldon_helm")
class TestPythonS2iK8s(object):

    def __init__(self):
        self.tester = S2IK8S()

    def test_model_rest(self, s2i_python_version):
        self.tester.test_model_rest(s2i_python_version)

    def test_input_transformer_rest(self, s2i_python_version):
        self.tester.test_input_transformer_rest(s2i_python_version)

    def test_output_transformer_rest(self, s2i_python_version):
        self.tester.test_output_transformer_rest(s2i_python_version)

    def test_router_rest(self, s2i_python_version):
        self.tester.test_router_rest(s2i_python_version)

    def test_combiner_rest(self, s2i_python_version):
        self.tester.test_combiner_rest(s2i_python_version)


class S2IK8S(object):

    def test_model_rest(self, s2i_python_version):
        run("kubectl delete sdep --all", shell=True)
        create_push_s2i_image(s2i_python_version, "model", "rest")
        run("kubectl apply -f ../resources/s2i_python_model.json", shell=True, check=True)
        wait_for_rollout("mymodel-mymodel-b55624a")
        r = initial_rest_request("mymodel", "seldon")
        arr = np.array([[1, 2, 3]])
        r = rest_request_ambassador("mymodel", "seldon", API_AMBASSADOR, data=arr)
        res = r.json()
        print(res)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["shape"] == [1, 3]
        assert r.json()["data"]["tensor"]["values"] == [2, 3, 4]
        run("kubectl delete sdep --all", shell=True)

    def test_input_transformer_rest(self, s2i_python_version):
        run("kubectl delete sdep --all", shell=True)
        create_push_s2i_image(s2i_python_version, "transformer", "rest")
        run("kubectl apply -f ../resources/s2i_python_transformer.json", shell=True, check=True)
        wait_for_rollout("mytrans-mytrans-01bb8ff")
        r = initial_rest_request("mytrans", "seldon")
        arr = np.array([[1, 2, 3]])
        r = rest_request_ambassador("mytrans", "seldon", API_AMBASSADOR, data=arr)
        res = r.json()
        print(res)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["shape"] == [1, 3]
        assert r.json()["data"]["tensor"]["values"] == [2, 3, 4]
        run("kubectl delete sdep --all", shell=True)

    def test_output_transformer_rest(self, s2i_python_version):
        run("kubectl delete sdep --all", shell=True)
        create_push_s2i_image(s2i_python_version, "transformer", "rest")
        run("kubectl apply -f ../resources/s2i_python_output_transformer.json", shell=True, check=True)
        wait_for_rollout("mytrans-mytrans-14c8191")
        r = initial_rest_request("mytrans", "seldon")
        arr = np.array([[1, 2, 3]])
        r = rest_request_ambassador("mytrans", "seldon", API_AMBASSADOR, data=arr)
        res = r.json()
        print(res)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["shape"] == [1, 3]
        assert r.json()["data"]["tensor"]["values"] == [3, 4, 5]
        run("kubectl delete sdep --all", shell=True)

    def test_router_rest(self, s2i_python_version):
        run("kubectl delete sdep --all", shell=True)
        create_push_s2i_image(s2i_python_version, "model", "rest")
        create_push_s2i_image(s2i_python_version, "router", "rest")
        run("kubectl apply -f ../resources/s2i_python_router.json", shell=True, check=True)
        wait_for_rollout("myrouter-myrouter-afd55a5")
        r = initial_rest_request("myrouter", "seldon")
        arr = np.array([[1, 2, 3]])
        r = rest_request_ambassador("myrouter", "seldon", API_AMBASSADOR, data=arr)
        res = r.json()
        print(res)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["shape"] == [1, 3]
        assert r.json()["data"]["tensor"]["values"] == [2, 3, 4]
        run("kubectl delete sdep --all", shell=True)

    def test_combiner_rest(self, s2i_python_version):
        run("kubectl delete sdep --all", shell=True)
        create_push_s2i_image(s2i_python_version, "model", "rest")
        create_push_s2i_image(s2i_python_version, "combiner", "rest")
        run("kubectl apply -f ../resources/s2i_python_combiner.json", shell=True, check=True)
        wait_for_rollout("mycombiner-mycombiner-e5d0d2c")
        r = initial_rest_request("mycombiner", "seldon")
        arr = np.array([[1, 2, 3]])
        r = rest_request_ambassador("mycombiner", "seldon", API_AMBASSADOR, data=arr)
        res = r.json()
        print(res)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["shape"] == [1, 3]
        assert r.json()["data"]["tensor"]["values"] == [3, 4, 5]
        run("kubectl delete sdep --all", shell=True)

