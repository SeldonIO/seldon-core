import json
import logging
import time
from subprocess import run

import numpy as np
import pytest
from google.protobuf import json_format

from seldon_core.proto import prediction_pb2
from seldon_e2e_utils import (
    API_AMBASSADOR,
    grpc_request_ambassador,
    grpc_request_ambassador_metadata,
    initial_grpc_request,
    initial_rest_request,
    rest_request_ambassador,
    retry_run,
    wait_for_rollout,
    wait_for_status,
)

S2I_CREATE = """cd ../s2i/python-features/metadata && \
    s2i build -E environment_{model}_{api_type} . \
    seldonio/seldon-core-s2i-python37-ubi8:{s2i_python_version} \
    seldonio/test_metadata_{model}_{api_type}:0.1
"""
IMAGE_NAME = "seldonio/test_metadata_{model}_{api_type}:0.1"


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
    def test_build_modelmetadata_rest(self, s2i_python_version):
        create_s2i_image(s2i_python_version, "modelmetadata", "rest")
        img = get_image_name("modelmetadata", "rest")
        run(
            "docker run -d --rm --name 'model-modelmetadata-rest' " + img,
            shell=True,
            check=True,
        )
        time.sleep(2)
        run("docker rm -f model-modelmetadata-rest", shell=True, check=True)

    def test_build_modelmetadata_grpc(self, s2i_python_version):
        create_s2i_image(s2i_python_version, "modelmetadata", "grpc")
        img = get_image_name("modelmetadata", "grpc")
        run(
            "docker run -d --rm --name 'model-modelmetadata-grpc' " + img,
            shell=True,
            check=True,
        )
        time.sleep(2)
        run("docker rm -f model-modelmetadata-grpc", shell=True, check=True)


model_metadata = {
    "name": "my-model-name",
    "versions": ["my-model-version-01"],
    "platform": "seldon",
    "inputs": [
        {
            "messagetype": "tensor",
            "schema": {"names": ["a", "b", "c", "d"], "shape": [4]},
        }
    ],
    "outputs": [{"messagetype": "tensor", "schema": {"shape": [1]}}],
    "custom": {"tag-key": "tag-value"},
}

graph_metadata = {
    "name": "example",
    "models": {"my-model": model_metadata},
    "graphinputs": model_metadata["inputs"],
    "graphoutputs": model_metadata["outputs"],
}

graph_metadata_grpc = {
    "name": "example",
    "models": {"my-model": model_metadata},
    "inputs": model_metadata["inputs"],
    "outputs": model_metadata["outputs"],
    "custom": {"tag-key": "tag-value"},
}


@pytest.mark.sequential
@pytest.mark.usefixtures("namespace")
@pytest.mark.usefixtures("s2i_python_version")
class TestTagsPythonS2iK8s(object):
    def test_modelmetadata_rest(self, namespace, s2i_python_version):
        create_push_s2i_image(s2i_python_version, "modelmetadata", "rest")
        retry_run(
            f"kubectl apply -f ../resources/metadata_modelmetadata_rest.yaml -n {namespace}"
        )
        wait_for_status("mymodel-modelmetadata", namespace)
        wait_for_rollout("mymodel-modelmetadata", namespace)
        r = initial_rest_request("mymodel-modelmetadata", namespace)
        arr = np.array([[1, 2, 3]])
        r = rest_request_ambassador(
            "mymodel-modelmetadata", namespace, API_AMBASSADOR, data=arr
        )
        res = r.json()
        logging.info(res)
        assert r.status_code == 200

        r = rest_request_ambassador(
            "mymodel-modelmetadata", namespace, method="metadata", model_name="my-model"
        )

        assert r.status_code == 200

        res = r.json()
        logging.warning(res)

        assert res == model_metadata

        r = rest_request_ambassador(
            "mymodel-modelmetadata", namespace, method="graph-metadata"
        )

        assert r.status_code == 200

        res = r.json()
        logging.warning(res)

        assert res == graph_metadata

    def test_manifestmodelmetadata_rest(self, namespace, s2i_python_version):
        create_push_s2i_image(s2i_python_version, "manifestmetadata", "rest")
        retry_run(
            f"kubectl apply -f ../resources/metadata_manifestmetadata_rest.yaml -n {namespace}"
        )
        wait_for_status("mymodel-manifestmetadata", namespace)
        wait_for_rollout("mymodel-manifestmetadata", namespace)
        r = initial_rest_request("mymodel-manifestmetadata", namespace)
        arr = np.array([[1, 2, 3]])
        r = rest_request_ambassador(
            "mymodel-manifestmetadata", namespace, API_AMBASSADOR, data=arr
        )
        res = r.json()
        logging.info(res)
        assert r.status_code == 200

        r = rest_request_ambassador(
            "mymodel-manifestmetadata",
            namespace,
            method="metadata",
            model_name="my-model",
        )

        assert r.status_code == 200

        res = r.json()
        logging.warning(res)

        assert res == model_metadata

        r = rest_request_ambassador(
            "mymodel-manifestmetadata", namespace, method="graph-metadata"
        )

        assert r.status_code == 200

        res = r.json()
        logging.warning(res)

        assert res == graph_metadata

    def test_modelmetadata_grpc(self, namespace, s2i_python_version):
        create_push_s2i_image(s2i_python_version, "modelmetadata", "grpc")
        retry_run(
            f"kubectl apply -f ../resources/metadata_modelmetadata_grpc.yaml -n {namespace}"
        )
        wait_for_status("mymodel-modelmetadata", namespace)
        wait_for_rollout("mymodel-modelmetadata", namespace)
        r = initial_grpc_request("mymodel-modelmetadata", namespace)

        r = grpc_request_ambassador_metadata(
            "mymodel-modelmetadata", namespace, model_name="my-model"
        )

        res = json.loads(json_format.MessageToJson(r))
        logging.info(res)

        # Cast reference model metadata to proto and back in order to have int->float
        # infamous casting in google.protobuf.Value
        metadata_proto = prediction_pb2.SeldonModelMetadata()
        json_format.ParseDict(
            model_metadata, metadata_proto, ignore_unknown_fields=True
        )
        assert res == json.loads(json_format.MessageToJson(metadata_proto))

        r = grpc_request_ambassador_metadata("mymodel-modelmetadata", namespace)

        res = json.loads(json_format.MessageToJson(r))
        logging.info(res)

        graph_metadata_proto = prediction_pb2.SeldonGraphMetadata()
        json_format.ParseDict(
            graph_metadata_grpc, graph_metadata_proto, ignore_unknown_fields=True
        )
        assert res == json.loads(json_format.MessageToJson(graph_metadata_proto))
