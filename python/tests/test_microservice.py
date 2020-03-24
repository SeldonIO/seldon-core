import logging
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
from google.protobuf import json_format


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
                "PREDICTIVE_UNIT_SERVICE_PORT": "5000",
                "PREDICTIVE_UNIT_METRICS_SERVICE_PORT": "6005",
                "PREDICTIVE_UNIT_METRICS_ENDPOINT": "/metrics-endpoint",
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
        logging.info("starting: %s", " ".join(cmd))
        logging.info("cwd: %s", app_location)
        # stdout=PIPE, stderr=PIPE,
        p = Popen(cmd, cwd=app_location, env=env_vars, preexec_fn=os.setsid)

        time.sleep(1)
        for q in range(10):
            s1 = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            r1 = s1.connect_ex(("127.0.0.1", 5000))
            s2 = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            r2 = s2.connect_ex(("127.0.0.1", 6005))
            if r1 == 0 and r2 == 0:
                break
            time.sleep(5)
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
def test_model_template_app_rest_tags(tracing):
    with start_microservice(
        join(dirname(__file__), "model-template-app"), tracing=tracing
    ):
        data = '{"meta":{"tags":{"foo":"bar"}},"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}}'
        response = requests.get(
            "http://127.0.0.1:5000/predict", params="json=%s" % data
        )
        response.raise_for_status()
        assert response.json() == {
            "data": {"names": ["t:0", "t:1"], "ndarray": [[1.0, 2.0]]},
            "meta": {"tags": {"foo": "bar"}},
        }


@pytest.mark.parametrize("tracing", [(False), (True)])
def test_model_template_app_rest_metrics(tracing):
    with start_microservice(
        join(dirname(__file__), "model-template-app"), tracing=tracing
    ):
        data = '{"meta":{"metrics":[{"key":"mygauge","type":"GAUGE","value":100}]},"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}}'
        response = requests.get(
            "http://127.0.0.1:5000/predict", params="json=%s" % data
        )
        response.raise_for_status()
        assert response.json() == {
            "data": {"names": ["t:0", "t:1"], "ndarray": [[1.0, 2.0]]},
            "meta": {"metrics": [{"key": "mygauge", "type": "GAUGE", "value": 100}]},
        }


@pytest.mark.parametrize("tracing", [(False), (True)])
def test_model_template_app_rest_metrics_endpoint(tracing):
    with start_microservice(
        join(dirname(__file__), "model-template-app"), tracing=tracing
    ):
        response = requests.get("http://127.0.0.1:6005/metrics-endpoint")
        # This just tests if endpoint exists and replies with 200
        assert response.status_code == 200


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


@pytest.mark.parametrize("tracing", [(False), (True)])
def test_model_template_app_grpc_tags(tracing):
    with start_microservice(
        join(dirname(__file__), "model-template-app"), tracing=tracing, grpc=True
    ):
        data = np.array([[1, 2]])
        datadef = prediction_pb2.DefaultData(
            tensor=prediction_pb2.Tensor(shape=data.shape, values=data.flatten())
        )

        meta = prediction_pb2.Meta()
        json_format.ParseDict({"tags": {"foo": "bar"}}, meta)

        request = prediction_pb2.SeldonMessage(data=datadef, meta=meta)
        channel = grpc.insecure_channel("0.0.0.0:5000")
        stub = prediction_pb2_grpc.ModelStub(channel)
        response = stub.Predict(request=request)
        assert response.data.tensor.shape[0] == 1
        assert response.data.tensor.shape[1] == 2
        assert response.data.tensor.values[0] == 1
        assert response.data.tensor.values[1] == 2

        assert response.meta.tags["foo"].string_value == "bar"


@pytest.mark.parametrize("tracing", [(False), (True)])
def test_model_template_app_grpc_metrics(tracing):
    with start_microservice(
        join(dirname(__file__), "model-template-app"), tracing=tracing, grpc=True
    ):
        data = np.array([[1, 2]])
        datadef = prediction_pb2.DefaultData(
            tensor=prediction_pb2.Tensor(shape=data.shape, values=data.flatten())
        )

        meta = prediction_pb2.Meta()
        json_format.ParseDict(
            {"metrics": [{"key": "mygauge", "type": "GAUGE", "value": 100}]}, meta
        )

        request = prediction_pb2.SeldonMessage(data=datadef, meta=meta)
        channel = grpc.insecure_channel("0.0.0.0:5000")
        stub = prediction_pb2_grpc.ModelStub(channel)
        response = stub.Predict(request=request)
        assert response.data.tensor.shape[0] == 1
        assert response.data.tensor.shape[1] == 2
        assert response.data.tensor.values[0] == 1
        assert response.data.tensor.values[1] == 2

        assert response.meta.metrics[0].key == "mygauge"
        assert response.meta.metrics[0].value == 100


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
        ('foo="bar"', {"foo": "bar"}),
        (' foo  =   "bar"  ', {"foo": "bar"}),
        ('key=  "assign==="', {"key": "assign==="}),
    ]
    for data, expected_annotation in read_data:
        with mock.patch("seldon_core.microservice.open", return_value=StringIO(data)):
            assert microservice.load_annotations() == expected_annotation
