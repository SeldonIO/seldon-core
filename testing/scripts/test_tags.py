import json
import logging
import time
from subprocess import run

import numpy as np
import pytest
from google.protobuf import json_format

from seldon_e2e_utils import (
    API_AMBASSADOR,
    grpc_request_ambassador,
    initial_grpc_request,
    initial_rest_request,
    rest_request_ambassador,
    retry_run,
    wait_for_rollout,
    wait_for_status,
)

S2I_CREATE = """cd ../s2i/python-features/tags && \
    s2i build -E environment_{model}_{api_type} . \
    seldonio/seldon-core-s2i-python3:{s2i_python_version} \
    seldonio/test_tags_{model}_{api_type}:0.1
"""
IMAGE_NAME = "seldonio/test_tags_{model}_{api_type}:0.1"


def create_s2i_image(s2i_python_version, model, api_type):
    cmd = S2I_CREATE.format(
        s2i_python_version=s2i_python_version, model=model, api_type=api_type
    )

    logging.info(cmd)
    run(cmd, shell=True, check=True)


def kind_push_s2i_image(model, api_type):
    img = get_image_name(model, api_type)
    cmd = "kind load docker-image " + img
    logging.info(cmd)
    run(cmd, shell=True, check=True)


def get_image_name(model, api_type):
    return IMAGE_NAME.format(model=model, api_type=api_type)


def create_push_s2i_image(s2i_python_version, model, api_type):
    create_s2i_image(s2i_python_version, model, api_type)
    kind_push_s2i_image(model, api_type)


@pytest.mark.sequential
@pytest.mark.usefixtures("s2i_python_version")
class TestTagsPythonS2i(object):
    def test_build_model_one_rest(self, s2i_python_version):
        create_s2i_image(s2i_python_version, "one", "rest")
        img = get_image_name("one", "rest")
        run("docker run -d --rm --name 'model-one-rest' " + img, shell=True, check=True)
        time.sleep(2)
        run("docker rm -f model-one-rest", shell=True, check=True)

    def test_build_model_two_rest(self, s2i_python_version):
        create_s2i_image(s2i_python_version, "two", "rest")
        img = get_image_name("two", "rest")
        run("docker run -d --rm --name 'model-two-rest' " + img, shell=True, check=True)
        time.sleep(2)
        run("docker rm -f model-two-rest", shell=True, check=True)

    def test_build_model_one_grpc(self, s2i_python_version):
        create_s2i_image(s2i_python_version, "one", "grpc")
        img = get_image_name("one", "grpc")
        run("docker run -d --rm --name 'model-one-grpc' " + img, shell=True, check=True)
        time.sleep(2)
        run("docker rm -f model-one-grpc", shell=True, check=True)

    def test_build_model_two_grpc(self, s2i_python_version):
        create_s2i_image(s2i_python_version, "two", "grpc")
        img = get_image_name("two", "grpc")
        run("docker run -d --rm --name 'model-two-grpc' " + img, shell=True, check=True)
        time.sleep(2)
        run("docker rm -f model-two-grpc", shell=True, check=True)

    def test_build_combiner_rest(self, s2i_python_version):
        create_s2i_image(s2i_python_version, "combiner", "rest")
        img = get_image_name("combiner", "rest")
        run("docker run -d --rm --name 'combiner-rest' " + img, shell=True, check=True)
        time.sleep(2)
        run("docker rm -f combiner-rest", shell=True, check=True)

    def test_build_combiner_grpc(self, s2i_python_version):
        create_s2i_image(s2i_python_version, "combiner", "grpc")
        img = get_image_name("combiner", "grpc")
        run("docker run -d --rm --name 'combiner-grpc' " + img, shell=True, check=True)
        time.sleep(2)
        run("docker rm -f combiner-grpc", shell=True, check=True)


@pytest.mark.sequential
@pytest.mark.usefixtures("namespace")
@pytest.mark.usefixtures("s2i_python_version")
class TestTagsPythonS2iK8s(object):
    def test_model_single_rest(self, namespace, s2i_python_version):
        create_push_s2i_image(s2i_python_version, "one", "rest")
        retry_run(f"kubectl apply -f ../resources/tags_single_rest.json -n {namespace}")
        wait_for_status("mymodel-tags-single", namespace)
        wait_for_rollout("mymodel-tags-single", namespace)
        r = initial_rest_request("mymodel-tags-single", namespace)
        arr = np.array([[1, 2, 3]])
        r = rest_request_ambassador(
            "mymodel-tags-single", namespace, API_AMBASSADOR, data=arr
        )
        res = r.json()
        logging.info(res)
        assert r.status_code == 200
        assert res["data"]["ndarray"] == ["model-1"]
        assert res["meta"]["tags"] == {"common": 1, "model-1": "yes"}
        run(
            f"kubectl delete -f ../resources/tags_single_rest.json -n {namespace}",
            shell=True,
        )

    def test_model_graph_rest(self, namespace, s2i_python_version):
        create_push_s2i_image(s2i_python_version, "one", "rest")
        create_push_s2i_image(s2i_python_version, "two", "rest")
        retry_run(f"kubectl apply -f ../resources/tags_graph_rest.json -n {namespace}")
        wait_for_status("mymodel-tags-graph", namespace)
        wait_for_rollout("mymodel-tags-graph", namespace)
        r = initial_rest_request("mymodel-tags-graph", namespace)
        arr = np.array([[1, 2, 3]])
        r = rest_request_ambassador(
            "mymodel-tags-graph", namespace, API_AMBASSADOR, data=arr
        )
        res = r.json()
        logging.info(res)
        assert r.status_code == 200
        assert res["data"]["ndarray"] == ["model-2"]
        assert res["meta"]["tags"] == {"common": 2, "model-1": "yes", "model-2": "yes"}
        run(
            f"kubectl delete -f ../resources/tags_graph_rest.json -n {namespace}",
            shell=True,
        )

    def test_model_combiner_rest(self, namespace, s2i_python_version):
        create_push_s2i_image(s2i_python_version, "one", "rest")
        create_push_s2i_image(s2i_python_version, "two", "rest")
        create_push_s2i_image(s2i_python_version, "combiner", "rest")
        retry_run(
            f"kubectl apply -f ../resources/tags_combiner_rest.json -n {namespace}"
        )
        wait_for_status("mymodel-tags-combiner", namespace)
        wait_for_rollout("mymodel-tags-combiner", namespace)
        r = initial_rest_request("mymodel-tags-combiner", namespace)
        arr = np.array([[1, 2, 3]])
        r = rest_request_ambassador(
            "mymodel-tags-combiner", namespace, API_AMBASSADOR, data=arr
        )
        res = r.json()
        logging.info(res)
        assert r.status_code == 200
        assert res["data"]["ndarray"] == [["model-1"], ["model-2"]]
        assert res["meta"]["tags"] == {
            "combiner": "yes",
            "common": 2,
            "model-1": "yes",
            "model-2": "yes",
        }
        run(
            f"kubectl delete -f ../resources/tags_combiner_rest.json -n {namespace}",
            shell=True,
        )

    def test_model_single_grpc(self, namespace, s2i_python_version):
        create_push_s2i_image(s2i_python_version, "one", "grpc")
        retry_run(f"kubectl apply -f ../resources/tags_single_grpc.json -n {namespace}")
        wait_for_status("mymodel-tags-single", namespace)
        wait_for_rollout("mymodel-tags-single", namespace)
        r = initial_grpc_request("mymodel-tags-single", namespace)
        arr = np.array([[1, 2, 3]])
        r = grpc_request_ambassador(
            "mymodel-tags-single", namespace, API_AMBASSADOR, data=arr
        )
        res = json.loads(json_format.MessageToJson(r))
        logging.info(res)
        # assert r.status_code == 200
        assert res["data"]["ndarray"] == ["model-1"]
        assert res["meta"]["tags"] == {"common": 1, "model-1": "yes"}
        run(
            f"kubectl delete -f ../resources/tags_single_grpc.json -n {namespace}",
            shell=True,
        )

    def test_model_graph_grpc(self, namespace, s2i_python_version):
        create_push_s2i_image(s2i_python_version, "one", "grpc")
        create_push_s2i_image(s2i_python_version, "two", "grpc")
        retry_run(f"kubectl apply -f ../resources/tags_graph_grpc.json -n {namespace}")
        wait_for_status("mymodel-tags-graph", namespace)
        wait_for_rollout("mymodel-tags-graph", namespace)
        r = initial_grpc_request("mymodel-tags-graph", namespace)
        arr = np.array([[1, 2, 3]])
        r = grpc_request_ambassador(
            "mymodel-tags-graph", namespace, API_AMBASSADOR, data=arr
        )
        res = json.loads(json_format.MessageToJson(r))
        logging.info(res)
        # assert r.status_code == 200
        assert res["data"]["ndarray"] == ["model-2"]
        assert res["meta"]["tags"] == {"common": 2, "model-1": "yes", "model-2": "yes"}
        run(
            f"kubectl delete -f ../resources/tags_graph_grpc.json -n {namespace}",
            shell=True,
        )

    def test_model_combiner_grpc(self, namespace, s2i_python_version):
        create_push_s2i_image(s2i_python_version, "one", "grpc")
        create_push_s2i_image(s2i_python_version, "two", "grpc")
        create_push_s2i_image(s2i_python_version, "combiner", "grpc")
        retry_run(
            f"kubectl apply -f ../resources/tags_combiner_grpc.json -n {namespace}"
        )
        wait_for_status("mymodel-tags-combiner", namespace)
        wait_for_rollout("mymodel-tags-combiner", namespace)
        r = initial_grpc_request("mymodel-tags-combiner", namespace)
        arr = np.array([[1, 2, 3]])
        r = grpc_request_ambassador(
            "mymodel-tags-combiner", namespace, API_AMBASSADOR, data=arr
        )
        res = json.loads(json_format.MessageToJson(r))
        logging.info(res)
        # assert r.status_code == 200
        assert res["data"]["ndarray"] == [["model-1"], ["model-2"]]
        assert res["meta"]["tags"] == {
            "combiner": "yes",
            "common": 2,
            "model-1": "yes",
            "model-2": "yes",
        }
        run(
            f"kubectl delete -f ../resources/tags_combiner_grpc.json -n {namespace}",
            shell=True,
        )
