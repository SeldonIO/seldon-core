import pytest
import time
import subprocess
from subprocess import run, Popen
from seldon_utils import *
from k8s_utils import *
import numpy as np

S2I_CREATE = "cd ../s2i/python/#TYPE# && s2i build -E environment_#API# . seldonio/seldon-core-s2i-python3:#VERSION# seldonio/test#TYPE#_#API#:0.1"
IMAGE_NAME = "seldonio/test#TYPE#_#API#:0.1"


def create_s2I_image(s2i_python_version, component_type, api_type):
    cmd = (
        S2I_CREATE.replace("#TYPE#", component_type)
        .replace("#API#", api_type)
        .replace("#VERSION#", s2i_python_version)
    )
    print(cmd)
    run(cmd, shell=True, check=True)


def kind_push_s2i_image(component_type, api_type):
    img = get_image_name(component_type, api_type)
    cmd = "kind load docker-image " + img + " --loglevel trace"
    print(cmd)
    run(cmd, shell=True, check=True)


def get_image_name(component_type, api_type):
    return IMAGE_NAME.replace("#TYPE#", component_type).replace("#API#", api_type)


def create_push_s2i_image(s2i_python_version, component_type, api_type):
    create_s2I_image(s2i_python_version, component_type, api_type)
    kind_push_s2i_image(component_type, api_type)


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


@pytest.mark.usefixtures("s2i_python_version")
class TestPythonS2iK8s(object):
    def test_model_rest(self, s2i_python_version):
        tester = S2IK8S()
        tester.test_model_rest(s2i_python_version)

    def test_input_transformer_rest(self, s2i_python_version):
        tester = S2IK8S()
        tester.test_input_transformer_rest(s2i_python_version)

    def test_output_transformer_rest(self, s2i_python_version):
        tester = S2IK8S()
        tester.test_output_transformer_rest(s2i_python_version)

    def test_router_rest(self, s2i_python_version):
        tester = S2IK8S()
        tester.test_router_rest(s2i_python_version)

    def test_combiner_rest(self, s2i_python_version):
        tester = S2IK8S()
        tester.test_combiner_rest(s2i_python_version)


class S2IK8S(object):
    def test_model_rest(self, s2i_python_version):
        namespace = "s2i-test-model-rest"
        run(f"kubectl create namespace {namespace}", shell=True, check=True)
        create_push_s2i_image(s2i_python_version, "model", "rest")
        run(
            f"kubectl apply -f ../resources/s2i_python_model.json -n {namespace}",
            shell=True,
            check=True,
        )
        wait_for_rollout("mymodel-mymodel-8715075", namespace)
        r = initial_rest_request("mymodel", namespace)
        arr = np.array([[1, 2, 3]])
        r = rest_request_ambassador("mymodel", namespace, API_AMBASSADOR, data=arr)
        res = r.json()
        print(res)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["shape"] == [1, 3]
        assert r.json()["data"]["tensor"]["values"] == [2, 3, 4]
        run(
            f"kubectl delete -f ../resources/s2i_python_model.json -n {namespace}",
            shell=True,
            check=True,
        )
        run(f"kubectl delete namespace {namespace}", shell=True, check=True)

    def test_input_transformer_rest(self, s2i_python_version):
        namespace = "s2i-test-input-transformer-rest"
        run(f"kubectl create namespace {namespace}", shell=True, check=True)
        create_push_s2i_image(s2i_python_version, "transformer", "rest")
        run(
            f"kubectl apply -f ../resources/s2i_python_transformer.json -n {namespace}",
            shell=True,
            check=True,
        )
        wait_for_rollout("mytrans-mytrans-1f278ae", namespace)
        r = initial_rest_request("mytrans", namespace)
        arr = np.array([[1, 2, 3]])
        r = rest_request_ambassador("mytrans", namespace, API_AMBASSADOR, data=arr)
        res = r.json()
        print(res)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["shape"] == [1, 3]
        assert r.json()["data"]["tensor"]["values"] == [2, 3, 4]
        run(
            f"kubectl delete -f ../resources/s2i_python_transformer.json -n {namespace}",
            shell=True,
            check=True,
        )
        run(f"kubectl create namespace {namespace}", shell=True, check=True)

    def test_output_transformer_rest(self, s2i_python_version):
        namespace = "s2i-test-output-transformer-rest"
        run(f"kubectl create namespace {namespace}", shell=True, check=True)
        create_push_s2i_image(s2i_python_version, "transformer", "rest")
        run(
            f"kubectl apply -f ../resources/s2i_python_output_transformer.json -n {namespace}",
            shell=True,
            check=True,
        )
        wait_for_rollout("mytrans-mytrans-52996cb", namespace)
        r = initial_rest_request("mytrans", namespace)
        arr = np.array([[1, 2, 3]])
        r = rest_request_ambassador("mytrans", namespace, API_AMBASSADOR, data=arr)
        res = r.json()
        print(res)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["shape"] == [1, 3]
        assert r.json()["data"]["tensor"]["values"] == [3, 4, 5]
        run(
            f"kubectl delete -f ../resources/s2i_python_output_transformer.json -n {namespace}",
            shell=True,
            check=True,
        )
        run(f"kubectl create namespace {namespace}", shell=True, check=True)

    def test_router_rest(self, s2i_python_version):
        namespace = "s2i-test-router-rest"
        run(f"kubectl create namespace {namespace}", shell=True, check=True)
        create_push_s2i_image(s2i_python_version, "model", "rest")
        create_push_s2i_image(s2i_python_version, "router", "rest")
        run(
            f"kubectl apply -f ../resources/s2i_python_router.json -n {namespace}",
            shell=True,
            check=True,
        )
        wait_for_rollout("myrouter-myrouter-340ed69", namespace)
        r = initial_rest_request("myrouter", namespace)
        arr = np.array([[1, 2, 3]])
        r = rest_request_ambassador("myrouter", namespace, API_AMBASSADOR, data=arr)
        res = r.json()
        print(res)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["shape"] == [1, 3]
        assert r.json()["data"]["tensor"]["values"] == [2, 3, 4]
        run(
            f"kubectl delete -f ../resources/s2i_python_router.json -n {namespace}",
            shell=True,
            check=True,
        )
        run(f"kubectl delete namespace {namespace}", shell=True, check=True)

    def test_combiner_rest(self, s2i_python_version):
        namespace = "s2i-test-combiner-rest"
        run(f"kubectl create namespace {namespace}", shell=True, check=True)
        create_push_s2i_image(s2i_python_version, "model", "rest")
        create_push_s2i_image(s2i_python_version, "combiner", "rest")
        run(
            f"kubectl apply -f ../resources/s2i_python_combiner.json -n {namespace}",
            shell=True,
            check=True,
        )
        wait_for_rollout("mycombiner-mycombiner-acc7c4d", namespace)
        r = initial_rest_request("mycombiner", namespace)
        arr = np.array([[1, 2, 3]])
        r = rest_request_ambassador("mycombiner", namespace, API_AMBASSADOR, data=arr)
        res = r.json()
        print(res)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["shape"] == [1, 3]
        assert r.json()["data"]["tensor"]["values"] == [3, 4, 5]
        run(
            f"kubectl delete -f ../resources/s2i_python_combiner.json -n {namespace}",
            shell=True,
            check=True,
        )
        run(f"kubectl delete namespace {namespace}", shell=True, check=True)
