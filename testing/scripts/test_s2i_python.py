import pytest
import time
from subprocess import run
import numpy as np
from seldon_e2e_utils import (
    wait_for_status,
    wait_for_rollout,
    rest_request_ambassador,
    initial_rest_request,
    retry_run,
    API_AMBASSADOR,
)
import logging

S2I_CREATE = "cd ../s2i/python/#TYPE# && s2i build -E environment_#API# . seldonio/seldon-core-s2i-python3:#VERSION# seldonio/test#TYPE#_#API#:0.1"
IMAGE_NAME = "seldonio/test#TYPE#_#API#:0.1"


def create_s2I_image(s2i_python_version, component_type, api_type):
    cmd = (
        S2I_CREATE.replace("#TYPE#", component_type)
        .replace("#API#", api_type)
        .replace("#VERSION#", s2i_python_version)
    )
    logging.warning(cmd)
    run(cmd, shell=True, check=True)


def kind_push_s2i_image(component_type, api_type):
    img = get_image_name(component_type, api_type)
    cmd = "kind load docker-image " + img + " --loglevel trace"
    logging.warning(cmd)
    run(cmd, shell=True, check=True)


def get_image_name(component_type, api_type):
    return IMAGE_NAME.replace("#TYPE#", component_type).replace("#API#", api_type)


def create_push_s2i_image(s2i_python_version, component_type, api_type):
    create_s2I_image(s2i_python_version, component_type, api_type)
    kind_push_s2i_image(component_type, api_type)


@pytest.mark.sequential
@pytest.mark.usefixtures("s2i_python_version")
class TestPythonS2i(object):
    def test_build_router_rest(self, s2i_python_version):
        create_s2I_image(s2i_python_version, "router", "rest")
        img = get_image_name("router", "rest")
        run("docker run -d --rm --name 'router-rest' " + img, shell=True, check=True)
        time.sleep(2)
        run("docker rm -f router-rest", shell=True, check=True)

    def test_build_router_grpc(self, s2i_python_version):
        create_s2I_image(s2i_python_version, "router", "grpc")
        img = get_image_name("router", "grpc")
        run("docker run -d --rm --name 'router-grpc' " + img, shell=True, check=True)
        time.sleep(2)
        run("docker rm -f router-grpc", shell=True, check=True)

    def test_build_model_rest(self, s2i_python_version):
        create_s2I_image(s2i_python_version, "model", "rest")
        img = get_image_name("model", "rest")
        run("docker run -d --rm --name 'model-rest' " + img, shell=True, check=True)
        time.sleep(2)
        run("docker rm -f model-rest", shell=True, check=True)

    def test_build_model_grpc(self, s2i_python_version):
        create_s2I_image(s2i_python_version, "model", "grpc")
        img = get_image_name("model", "grpc")
        run("docker run -d --rm --name 'model-grpc' " + img, shell=True, check=True)
        time.sleep(2)
        run("docker rm -f model-grpc", shell=True, check=True)

    def test_build_transformer_rest(self, s2i_python_version):
        create_s2I_image(s2i_python_version, "transformer", "rest")
        img = get_image_name("transformer", "rest")
        run(
            "docker run -d --rm --name 'transformer-rest' " + img,
            shell=True,
            check=True,
        )
        time.sleep(2)
        run("docker rm -f transformer-rest", shell=True, check=True)

    def test_build_transformer_grpc(self, s2i_python_version):
        create_s2I_image(s2i_python_version, "transformer", "grpc")
        img = get_image_name("transformer", "grpc")
        run(
            "docker run -d --rm --name 'transformer-grpc' " + img,
            shell=True,
            check=True,
        )
        time.sleep(2)
        run("docker rm -f transformer-grpc", shell=True, check=True)

    def test_build_combiner_rest(self, s2i_python_version):
        create_s2I_image(s2i_python_version, "combiner", "rest")
        img = get_image_name("combiner", "rest")
        logging.warning(img)
        run("docker run -d --rm --name 'combiner-rest' " + img, shell=True, check=True)
        time.sleep(2)
        run("docker rm -f combiner-rest", shell=True, check=True)

    def test_build_combiner_grpc(self, s2i_python_version):
        create_s2I_image(s2i_python_version, "combiner", "grpc")
        img = get_image_name("combiner", "grpc")
        run("docker run -d --rm --name 'combiner-grpc' " + img, shell=True, check=True)
        time.sleep(2)
        run("docker rm -f combiner-grpc", shell=True, check=True)


@pytest.mark.sequential
@pytest.mark.usefixtures("s2i_python_version")
class TestPythonS2iK8s(object):
    def test_model_rest(self, s2i_python_version):
        tester = S2IK8S()
        tester.test_model_rest(s2i_python_version)

    def test_model_rest_non200(self, s2i_python_version):
        tester = S2IK8S()
        tester.test_model_rest_non200(s2i_python_version)

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
    def test_model_rest(self, namespace, s2i_python_version):
        create_push_s2i_image(s2i_python_version, "model", "rest")
        retry_run(f"kubectl apply -f ../resources/s2i_python_model.json -n {namespace}")
        wait_for_status("mymodel", namespace)
        wait_for_rollout("mymodel", namespace)
        r = initial_rest_request("mymodel", namespace)
        arr = np.array([[1, 2, 3]])
        r = rest_request_ambassador("mymodel", namespace, API_AMBASSADOR, data=arr)
        res = r.json()
        logging.warning(res)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["shape"] == [1, 3]
        assert r.json()["data"]["tensor"]["values"] == [2, 3, 4]
        run(
            f"kubectl delete -f ../resources/s2i_python_model.json -n {namespace}",
            shell=True,
        )

    def test_model_rest_non200(self, namespace, s2i_python_version):
        create_push_s2i_image(s2i_python_version, "model", "rest_non200")
        retry_run(
            f"kubectl apply -f ../resources/s2i_python_model_non200.json -n {namespace}"
        )
        wait_for_status("mymodel", namespace)
        wait_for_rollout("mymodel", namespace)
        r = initial_rest_request("mymodel", namespace)
        arr = np.array([[1, 2, 3]])
        r = rest_request_ambassador("mymodel", namespace, API_AMBASSADOR, data=arr)
        res = r.json()
        logging.warning(res)
        assert r.status_code == 200
        assert r.json()["status"]["code"] == 400
        assert r.json()["status"]["reason"] == "exception message"
        assert r.json()["status"]["info"] == "exception caught"
        assert r.json()["status"]["status"] == "FAILURE"
        run(
            f"kubectl delete -f ../resources/s2i_python_model_non200.json -n {namespace}",
            shell=True,
        )

    def test_input_transformer_rest(self, namespace, s2i_python_version):
        create_push_s2i_image(s2i_python_version, "transformer", "rest")
        retry_run(
            f"kubectl apply -f ../resources/s2i_python_transformer.json -n {namespace}"
        )
        wait_for_status("mytrans", namespace)
        wait_for_rollout("mytrans", namespace)
        r = initial_rest_request("mytrans", namespace)
        arr = np.array([[1, 2, 3]])
        r = rest_request_ambassador("mytrans", namespace, API_AMBASSADOR, data=arr)
        res = r.json()
        logging.warning(res)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["shape"] == [1, 3]
        assert r.json()["data"]["tensor"]["values"] == [2, 3, 4]
        run(
            f"kubectl delete -f ../resources/s2i_python_transformer.json -n {namespace}",
            shell=True,
        )

    def test_output_transformer_rest(self, namespace, s2i_python_version):
        create_push_s2i_image(s2i_python_version, "transformer", "rest")
        retry_run(
            f"kubectl apply -f ../resources/s2i_python_output_transformer.json -n {namespace}"
        )
        wait_for_status("mytrans", namespace)
        wait_for_rollout("mytrans", namespace)
        r = initial_rest_request("mytrans", namespace)
        arr = np.array([[1, 2, 3]])
        r = rest_request_ambassador("mytrans", namespace, API_AMBASSADOR, data=arr)
        res = r.json()
        logging.warning(res)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["shape"] == [1, 3]
        assert r.json()["data"]["tensor"]["values"] == [3, 4, 5]
        run(
            f"kubectl delete -f ../resources/s2i_python_output_transformer.json -n {namespace}",
            shell=True,
        )

    def test_router_rest(self, namespace, s2i_python_version):
        create_push_s2i_image(s2i_python_version, "model", "rest")
        create_push_s2i_image(s2i_python_version, "router", "rest")
        retry_run(
            f"kubectl apply -f ../resources/s2i_python_router.json -n {namespace}"
        )
        wait_for_status("myrouter", namespace)
        wait_for_rollout("myrouter", namespace)
        r = initial_rest_request("myrouter", namespace)
        arr = np.array([[1, 2, 3]])
        r = rest_request_ambassador("myrouter", namespace, API_AMBASSADOR, data=arr)
        res = r.json()
        logging.warning(res)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["shape"] == [1, 3]
        assert r.json()["data"]["tensor"]["values"] == [2, 3, 4]
        run(
            f"kubectl delete -f ../resources/s2i_python_router.json -n {namespace}",
            shell=True,
        )

    def test_combiner_rest(self, namespace, s2i_python_version):
        create_push_s2i_image(s2i_python_version, "model", "rest")
        create_push_s2i_image(s2i_python_version, "combiner", "rest")
        retry_run(
            f"kubectl apply -f ../resources/s2i_python_combiner.json -n {namespace}"
        )
        wait_for_status("mycombiner", namespace)
        wait_for_rollout("mycombiner", namespace)
        r = initial_rest_request("mycombiner", namespace)
        arr = np.array([[1, 2, 3]])
        r = rest_request_ambassador("mycombiner", namespace, API_AMBASSADOR, data=arr)
        res = r.json()
        logging.warning(res)
        assert r.status_code == 200
        assert r.json()["data"]["tensor"]["shape"] == [1, 3]
        assert r.json()["data"]["tensor"]["values"] == [3, 4, 5]
        run(
            f"kubectl delete -f ../resources/s2i_python_combiner.json -n {namespace}",
            shell=True,
        )
