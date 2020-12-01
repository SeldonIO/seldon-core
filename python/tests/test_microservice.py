import logging
import requests
import pytest
import grpc
import numpy as np

from os.path import dirname, join
from unittest import mock
from tenacity import Retrying, stop_after_attempt, wait_exponential
from seldon_core import __version__
from seldon_core.proto import prediction_pb2
from seldon_core.proto import prediction_pb2_grpc
from seldon_core.utils import NONIMPLEMENTED_MSG
from seldon_core import microservice
from seldon_core.flask_utils import SeldonMicroserviceException
from google.protobuf import json_format

from .conftest import RESOURCES_PATH

DEFAULT_ROUTING = {"routing": {NONIMPLEMENTED_MSG: -1}}


def retry_method(method, args=(), kwargs={}, stop_after=5, max_sleep=10):
    for attempt in Retrying(
        wait=wait_exponential(max=max_sleep),
        stop=stop_after_attempt(stop_after),
        reraise=True,
    ):
        with attempt:
            logging.info(f"Calling method... try: {attempt.retry_state.attempt_number}")
            return method(*args, **kwargs)


def test_microservice_version():
    fname = join(dirname(__file__), "..", "..", "version.txt")
    with open(fname, "r") as f:
        version = f.readline().strip()
    assert version == __version__


def test_model_template_app_rest(microservice):
    data = '{"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}}'
    response = requests.get("http://127.0.0.1:9000/predict", params="json=%s" % data)
    response.raise_for_status()
    assert response.json() == {
        "data": {"names": ["t:0", "t:1"], "ndarray": [[1.0, 2.0]]},
        "meta": {**DEFAULT_ROUTING},
    }

    data = (
        '{"request":{"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}},'
        '"response":{"meta":{"routing":{"router":0}},"data":{"names":["a","b"],'
        '"ndarray":[[1.0,2.0]]}},"reward":1}'
    )
    response = requests.get(
        "http://127.0.0.1:9000/send-feedback", params="json=%s" % data
    )
    response.raise_for_status()
    assert response.json() == {"data": {"ndarray": []}, "meta": {**DEFAULT_ROUTING}}


def test_model_template_app_rest_tags(microservice):
    data = '{"meta":{"tags":{"foo":"bar"}},"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}}'
    response = requests.get("http://127.0.0.1:9000/predict", params="json=%s" % data)
    response.raise_for_status()
    assert response.json() == {
        "data": {"names": ["t:0", "t:1"], "ndarray": [[1.0, 2.0]]},
        "meta": {"tags": {"foo": "bar"}, **DEFAULT_ROUTING},
    }


def test_model_template_app_rest_metrics(microservice):
    data = '{"meta":{"metrics":[{"key":"mygauge","type":"GAUGE","value":100}]},"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}}'
    response = requests.get("http://127.0.0.1:9000/predict", params="json=%s" % data)
    response.raise_for_status()
    assert response.json() == {
        "data": {"names": ["t:0", "t:1"], "ndarray": [[1.0, 2.0]]},
        "meta": {
            "metrics": [{"key": "mygauge", "type": "GAUGE", "value": 100}],
            **DEFAULT_ROUTING,
        },
    }


def test_model_template_app_rest_metrics_endpoint(microservice):
    response = requests.get("http://127.0.0.1:6005/metrics-endpoint")
    # This just tests if endpoint exists and replies with 200
    assert response.status_code == 200


@pytest.mark.parametrize(
    "microservice", [{"app_name": "model-template-app2"}], indirect=True
)
def test_model_template_app_rest_submodule(microservice):
    data = '{"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}}'
    response = requests.get("http://127.0.0.1:9000/predict", params="json=%s" % data)
    response.raise_for_status()
    assert response.json() == {
        "data": {"names": ["t:0", "t:1"], "ndarray": [[1.0, 2.0]]},
        "meta": {**DEFAULT_ROUTING},
    }

    data = (
        '{"request":{"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}},'
        '"response":{"meta":{"routing":{"router":0}},"data":{"names":["a","b"],'
        '"ndarray":[[1.0,2.0]]}},"reward":1}'
    )
    response = requests.get(
        "http://127.0.0.1:9000/send-feedback", params="json=%s" % data
    )
    response.raise_for_status()
    assert response.json() == {"data": {"ndarray": []}, "meta": {**DEFAULT_ROUTING}}


def test_model_template_app_grpc(microservice):
    data = np.array([[1, 2]])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=data.shape, values=data.flatten())
    )
    request = prediction_pb2.SeldonMessage(data=datadef)
    channel = grpc.insecure_channel("0.0.0.0:5000")
    stub = prediction_pb2_grpc.ModelStub(channel)
    response = retry_method(stub.Predict, kwargs=dict(request=request))
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


def test_model_template_app_grpc_tags(microservice):
    data = np.array([[1, 2]])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=data.shape, values=data.flatten())
    )

    meta = prediction_pb2.Meta()
    json_format.ParseDict({"tags": {"foo": "bar"}}, meta)

    request = prediction_pb2.SeldonMessage(data=datadef, meta=meta)
    channel = grpc.insecure_channel("0.0.0.0:5000")
    stub = prediction_pb2_grpc.ModelStub(channel)
    response = retry_method(stub.Predict, kwargs=dict(request=request))
    assert response.data.tensor.shape[0] == 1
    assert response.data.tensor.shape[1] == 2
    assert response.data.tensor.values[0] == 1
    assert response.data.tensor.values[1] == 2

    assert response.meta.tags["foo"].string_value == "bar"


def test_model_template_app_grpc_metrics(microservice):
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
    response = retry_method(stub.Predict, kwargs=dict(request=request))
    assert response.data.tensor.shape[0] == 1
    assert response.data.tensor.shape[1] == 2
    assert response.data.tensor.values[0] == 1
    assert response.data.tensor.values[1] == 2

    assert response.meta.metrics[0].key == "mygauge"
    assert response.meta.metrics[0].value == 100


@pytest.mark.parametrize(
    "microservice",
    [
        {
            "tracing": True,
            "envs": {
                "JAEGER_CONFIG_PATH": join(
                    RESOURCES_PATH, "tracing_config/tracing.yaml"
                )
            },
        }
    ],
    indirect=True,
)
def test_model_template_app_tracing_config(microservice):
    data = '{"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}}'
    response = requests.get("http://127.0.0.1:9000/predict", params="json=%s" % data)
    response.raise_for_status()
    assert response.json() == {
        "data": {"names": ["t:0", "t:1"], "ndarray": [[1.0, 2.0]]},
        "meta": {**DEFAULT_ROUTING},
    }

    data = (
        '{"request":{"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}},'
        '"response":{"meta":{"routing":{"router":0}},"data":{"names":["a","b"],'
        '"ndarray":[[1.0,2.0]]}},"reward":1}'
    )
    response = requests.get(
        "http://127.0.0.1:9000/send-feedback", params="json=%s" % data
    )
    response.raise_for_status()
    assert response.json() == {"data": {"ndarray": []}, "meta": {**DEFAULT_ROUTING}}


def test_model_template_bad_params():
    params = [
        join(dirname(__file__), "model-template-app"),
        "seldon-core-microservice",
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
