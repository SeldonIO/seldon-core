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

S2I_CREATE = "cd ../s2i/python/#TYPE# && s2i build -E environment#SUFFIX# . seldonio/seldon-core-s2i-python3:#VERSION# seldonio/test#TYPE##SUFFIX#:0.1"
IMAGE_NAME = "seldonio/test#TYPE##SUFFIX#:0.1"


def create_s2I_image(s2i_python_version, component_type, suffix):
    cmd = (
        S2I_CREATE.replace("#TYPE#", component_type)
        .replace("#SUFFIX#", suffix)
        .replace("#VERSION#", s2i_python_version)
    )
    logging.warning(cmd)
    run(cmd, shell=True, check=True)


def kind_push_s2i_image(component_type, suffix):
    img = get_image_name(component_type, suffix)
    cmd = "kind load docker-image " + img
    logging.warning(cmd)
    run(cmd, shell=True, check=True)


def get_image_name(component_type, suffix):
    return IMAGE_NAME.replace("#TYPE#", component_type).replace("#SUFFIX#", suffix)


def create_push_s2i_image(s2i_python_version, component_type, suffix):
    create_s2I_image(s2i_python_version, component_type, suffix)
    kind_push_s2i_image(component_type, suffix)


@pytest.mark.sequential
@pytest.mark.usefixtures("s2i_python_version")
class TestPythonS2i(object):
    def test_build_router(self, s2i_python_version):
        create_s2I_image(s2i_python_version, "router", "")
        img = get_image_name("router", "")
        run("docker run -d --rm --name 'router' " + img, shell=True, check=True)
        time.sleep(2)
        run("docker rm -f router", shell=True, check=True)

    def test_build_model(self, s2i_python_version):
        create_s2I_image(s2i_python_version, "model", "")
        img = get_image_name("model", "")
        run("docker run -d --rm --name 'model' " + img, shell=True, check=True)
        time.sleep(2)
        run("docker rm -f model", shell=True, check=True)

    def test_build_transformer(self, s2i_python_version):
        create_s2I_image(s2i_python_version, "transformer", "")
        img = get_image_name("transformer", "")
        run(
            "docker run -d --rm --name 'transformer' " + img, shell=True, check=True,
        )
        time.sleep(2)
        run("docker rm -f transformer", shell=True, check=True)

    def test_build_combiner(self, s2i_python_version):
        create_s2I_image(s2i_python_version, "combiner", "")
        img = get_image_name("combiner", "")
        logging.warning(img)
        run("docker run -d --rm --name 'combiner' " + img, shell=True, check=True)
        time.sleep(2)
        run("docker rm -f combiner", shell=True, check=True)


@pytest.mark.sequential
@pytest.mark.usefixtures("namespace")
@pytest.mark.usefixtures("s2i_python_version")
class TestPythonS2iK8s(object):
    def test_model(self, namespace, s2i_python_version):
        create_push_s2i_image(s2i_python_version, "model", "")
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
        create_push_s2i_image(s2i_python_version, "model", "_non200")
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
        assert r.status_code == 400
        assert r.json()["status"]["code"] == 400
        assert r.json()["status"]["info"] == "exception caught"
        assert r.json()["status"]["reason"] == "exception message"
        assert r.json()["status"]["status"] == "FAILURE"
        run(
            f"kubectl delete -f ../resources/s2i_python_model_non200.json -n {namespace}",
            shell=True,
        )

    def test_input_transformer(self, namespace, s2i_python_version):
        create_push_s2i_image(s2i_python_version, "transformer", "")
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

    def test_output_transformer(self, namespace, s2i_python_version):
        create_push_s2i_image(s2i_python_version, "transformer", "")
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

    def test_router(self, namespace, s2i_python_version):
        create_push_s2i_image(s2i_python_version, "model", "")
        create_push_s2i_image(s2i_python_version, "router", "")
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

    def test_combiner(self, namespace, s2i_python_version):
        create_push_s2i_image(s2i_python_version, "model", "")
        create_push_s2i_image(s2i_python_version, "combiner", "")
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
