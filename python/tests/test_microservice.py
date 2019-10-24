from contextlib import contextmanager
import os
from os.path import dirname, join
import socket
from subprocess import Popen
import time
import requests
import pytest
from seldon_core.proto import prediction_pb2
from seldon_core.proto import prediction_pb2_grpc
import seldon_core.microservice as microservice
from seldon_core.flask_utils import SeldonMicroserviceException
import grpc
import numpy as np
import signal
import unittest.mock as mock


@contextmanager
def start_microservice(app_location, tracing=False, grpc=False, envs={}):
    p = None
    try:
        # PYTHONUNBUFFERED=x
        # exec python -u microservice.py $MODEL_NAME $API_TYPE --service-type $SERVICE_TYPE --persistence $PERSISTENCE
        env_vars = dict(os.environ)
        env_vars.update(envs)
        env_vars.update(
            {
                "PYTHONUNBUFFERED": "x",
                "PYTHONPATH": app_location,
                "APP_HOST": "127.0.0.1",
                "SERVICE_PORT_ENV_NAME": "5000",
            }
        )
        with open(join(app_location, ".s2i", "environment")) as fh:
            for line in fh.readlines():
                line = line.strip()
                if line:
                    key, value = line.split("=", 1)
                    key, value = key.strip(), value.strip()
                    if key and value:
                        env_vars[key] = value
        if grpc:
            env_vars["API_TYPE"] = "GRPC"
        cmd = (
            "seldon-core-microservice",
            env_vars["MODEL_NAME"],
            env_vars["API_TYPE"],
            "--service-type",
            env_vars["SERVICE_TYPE"],
            "--persistence",
            env_vars["PERSISTENCE"],
        )
        if tracing:
            cmd = cmd + ("--tracing",)
        print("starting:", " ".join(cmd))
        print("cwd:", app_location)
        # stdout=PIPE, stderr=PIPE,
        p = Popen(cmd, cwd=app_location, env=env_vars, preexec_fn=os.setsid)

        for q in range(10):
            time.sleep(5)
            sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            result = sock.connect_ex(("127.0.0.1", 5000))
            if result == 0:
                break
        else:
            raise RuntimeError("Server did not bind to 127.0.0.1:5000")
        yield
    finally:
        if p:
            os.killpg(os.getpgid(p.pid), signal.SIGTERM)


@pytest.mark.parametrize("tracing", [(False), (True)])
def test_model_template_app_rest(tracing):
    with start_microservice(
        join(dirname(__file__), "model-template-app"), tracing=tracing
    ):
        data = '{"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}}'
        response = requests.get(
            "http://127.0.0.1:5000/predict", params="json=%s" % data
        )
        response.raise_for_status()
        assert response.json() == {
            "data": {"names": ["t:0", "t:1"], "ndarray": [[1.0, 2.0]]},
            "meta": {},
        }

        data = (
            '{"request":{"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}},'
            '"response":{"meta":{"routing":{"router":0}},"data":{"names":["a","b"],'
            '"ndarray":[[1.0,2.0]]}},"reward":1}'
        )
        response = requests.get(
            "http://127.0.0.1:5000/send-feedback", params="json=%s" % data
        )
        response.raise_for_status()
        assert response.json() == {"data": {"ndarray": []}, "meta": {}}


@pytest.mark.parametrize("tracing", [(False), (True)])
def test_model_template_app_rest_submodule(tracing):
    with start_microservice(
        join(dirname(__file__), "model-template-app2"), tracing=tracing
    ):
        data = '{"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}}'
        response = requests.get(
            "http://127.0.0.1:5000/predict", params="json=%s" % data
        )
        response.raise_for_status()
        assert response.json() == {
            "data": {"names": ["t:0", "t:1"], "ndarray": [[1.0, 2.0]]},
            "meta": {},
        }

        data = (
            '{"request":{"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}},'
            '"response":{"meta":{"routing":{"router":0}},"data":{"names":["a","b"],'
            '"ndarray":[[1.0,2.0]]}},"reward":1}'
        )
        response = requests.get(
            "http://127.0.0.1:5000/send-feedback", params="json=%s" % data
        )
        response.raise_for_status()
        assert response.json() == {"data": {"ndarray": []}, "meta": {}}


@pytest.mark.parametrize("tracing", [(False), (True)])
def test_model_template_app_grpc(tracing):
    with start_microservice(
        join(dirname(__file__), "model-template-app"), tracing=tracing, grpc=True
    ):
        data = np.array([[1, 2]])
        datadef = prediction_pb2.DefaultData(
            tensor=prediction_pb2.Tensor(shape=data.shape, values=data.flatten())
        )
        request = prediction_pb2.SeldonMessage(data=datadef)
        channel = grpc.insecure_channel("0.0.0.0:5000")
        stub = prediction_pb2_grpc.ModelStub(channel)
        response = stub.Predict(request=request)
        assert response.data.tensor.shape[0] == 1
        assert response.data.tensor.shape[1] == 2
        assert response.data.tensor.values[0] == 1
        assert response.data.tensor.values[1] == 2

        arr = np.array([1, 2])
        datadef = prediction_pb2.DefaultData(
            tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
        )
        request = prediction_pb2.SeldonMessage(data=datadef)
        feedback = prediction_pb2.Feedback(request=request, reward=1.0)
        response = stub.SendFeedback(request=request)


def test_model_template_app_tracing_config():
    envs = {
        "JAEGER_CONFIG_PATH": join(dirname(__file__), "tracing_config/tracing.yaml")
    }
    with start_microservice(
        join(dirname(__file__), "model-template-app"), tracing=True, envs=envs
    ):
        data = '{"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}}'
        response = requests.get(
            "http://127.0.0.1:5000/predict", params="json=%s" % data
        )
        response.raise_for_status()
        assert response.json() == {
            "data": {"names": ["t:0", "t:1"], "ndarray": [[1.0, 2.0]]},
            "meta": {},
        }

        data = (
            '{"request":{"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}},'
            '"response":{"meta":{"routing":{"router":0}},"data":{"names":["a","b"],'
            '"ndarray":[[1.0,2.0]]}},"reward":1}'
        )
        response = requests.get(
            "http://127.0.0.1:5000/send-feedback", params="json=%s" % data
        )
        response.raise_for_status()
        assert response.json() == {"data": {"ndarray": []}, "meta": {}}


def test_model_template_bad_params():
    params = [
        join(dirname(__file__), "model-template-app"),
        "seldon-core-microservice",
        "REST",
        "--parameters",
        '[{"type":"FLOAT","name":"foo","value":"abc"}]',
    ]
    with mock.patch("sys.argv", params):
        with pytest.raises(SeldonMicroserviceException):
            microservice.main()


def test_model_template_bad_params_type():
    params = [
        join(dirname(__file__), "model-template-app"),
        "seldon-core-microservice",
        "REST",
        "--parameters",
        '[{"type":"FOO","name":"foo","value":"abc"}]',
    ]
    with mock.patch("sys.argv", params):
        with pytest.raises(SeldonMicroserviceException):
            microservice.main()


@mock.patch("seldon_core.microservice.os.path.isfile", return_value=True)
def test_load_annotations(mock_isfile):
    from io import StringIO

    read_data = [
        ("", {}),
        ("\n\n", {}),
        ("foo=bar", {"foo": "bar"}),
        ("foo=bar\nx =y", {"foo": "bar", "x": "y"}),
        ("foo=bar\nfoo=baz\n", {"foo": "baz"}),
        (" foo  =   bar ", {"foo": "bar"}),
        ("key =  assign===", {"key": "assign==="}),
        ("foo=\nfoo", {"foo": ""}),
    ]
    for data, expected_annotation in read_data:
        with mock.patch("seldon_core.microservice.open", return_value=StringIO(data)):
            assert microservice.load_annotations() == expected_annotation
