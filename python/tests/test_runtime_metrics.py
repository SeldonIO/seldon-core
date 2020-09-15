import os
import logging
import pytest
import numpy as np
from google.protobuf import json_format
import json

from seldon_core.flask_utils import SeldonMicroserviceException
from seldon_core.proto import prediction_pb2
from seldon_core.wrapper import (
    get_rest_microservice,
    get_metrics_microservice,
    SeldonModelGRPC,
)
from seldon_core.metrics import (
    SeldonMetrics,
    create_counter,
    create_gauge,
    create_timer,
    split_image_tag,
    validate_metrics,
    COUNTER,
    BINS,
    FEEDBACK_METRIC_METHOD_TAG,
    PREDICT_METRIC_METHOD_TAG,
    INPUT_TRANSFORM_METRIC_METHOD_TAG,
    OUTPUT_TRANSFORM_METRIC_METHOD_TAG,
    ROUTER_METRIC_METHOD_TAG,
    AGGREGATE_METRIC_METHOD_TAG,
    HEALTH_METRIC_METHOD_TAG,
)
from seldon_core.user_model import client_custom_metrics, SeldonPrediction


RUNTIME_METRICS = [
    {"type": "GAUGE", "key": "runtime_gauge", "value": 42},
]


class UserObject:
    def predict(self, X, features_names):
        logging.info("Predict called")
        return SeldonPrediction(data=X, metrics=RUNTIME_METRICS)

    def aggregate(self, X, features_names):
        logging.info("Aggregate called")
        return SeldonPrediction(data=X[0], metrics=RUNTIME_METRICS)

    def transform_input(self, X, feature_names):
        logging.info("Transform input called")
        return SeldonPrediction(data=X, metrics=RUNTIME_METRICS)

    def transform_output(self, X, feature_names):
        logging.info("Transform output called")
        return SeldonPrediction(data=X, metrics=RUNTIME_METRICS)

    def route(self, X, feature_names):
        logging.info("Route called")
        return SeldonPrediction(data=22, metrics=RUNTIME_METRICS)

    def metrics(self):
        logging.info("Metrics called")
        return [
            {"type": "COUNTER", "key": "mycounter", "value": 1},
            {"type": "GAUGE", "key": "mygauge", "value": 100},
            {"type": "TIMER", "key": "mytimer", "value": 20.2},
            {
                "type": "GAUGE",
                "key": "customtag",
                "value": 200,
                "tags": {"mytag": "mytagvalue"},
            },
        ]


def verify_seldon_metrics(data, mycounter_value, histogram_entries, method):
    assert data["GAUGE", "runtime_gauge"]["value"] == 42
    assert data["GAUGE", "mygauge"]["value"] == 100
    assert data["GAUGE", "customtag"]["value"] == 200
    assert data["GAUGE", "customtag"]["tags"] == {
        "mytag": "mytagvalue",
        "method": method,
    }
    assert data["COUNTER", "mycounter"]["value"] == mycounter_value
    assert np.allclose(
        np.histogram(histogram_entries, BINS)[0], data["TIMER", "mytimer"]["value"][0]
    )
    assert np.allclose(data["TIMER", "mytimer"]["value"][1], np.sum(histogram_entries))


@pytest.mark.parametrize("cls", [UserObject])
def test_seldon_metrics_predict(cls, client_gets_metrics):
    user_object = cls()
    seldon_metrics = SeldonMetrics()

    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()

    rv = client.get('/predict?json={"data": {"names": ["input"], "ndarray": ["data"]}}')
    assert rv.status_code == 200
    assert ("metrics" in json.loads(rv.data)["meta"]) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 1, [0.0202], PREDICT_METRIC_METHOD_TAG)

    rv = client.get('/predict?json={"data": {"names": ["input"], "ndarray": ["data"]}}')
    assert rv.status_code == 200
    assert ("metrics" in json.loads(rv.data)["meta"]) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 2, [0.0202, 0.0202], PREDICT_METRIC_METHOD_TAG)


@pytest.mark.parametrize("cls", [UserObject])
def test_seldon_metrics_send_feedback(cls):
    user_object = cls()
    seldon_metrics = SeldonMetrics()

    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()

    rv = client.get('/send-feedback?json={"reward": 42}')
    assert rv.status_code == 200

    data = seldon_metrics.data[os.getpid()]

    assert data["COUNTER", "seldon_api_model_feedback_reward"] == {
        "value": 42.0,
        "tags": {"method": FEEDBACK_METRIC_METHOD_TAG},
    }

    rv = client.get('/send-feedback?json={"reward": 42}')
    assert rv.status_code == 200

    data = seldon_metrics.data[os.getpid()]

    assert data["COUNTER", "seldon_api_model_feedback_reward"] == {
        "value": 84.0,
        "tags": {"method": FEEDBACK_METRIC_METHOD_TAG},
    }


@pytest.mark.parametrize("cls", [UserObject])
def test_seldon_metrics_aggregate(cls, client_gets_metrics):
    user_object = cls()
    seldon_metrics = SeldonMetrics()

    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()

    rv = client.get(
        '/aggregate?json={"seldonMessages": [{"data": {"names": ["input"], "ndarray": ["data"]}}]}'
    )
    assert rv.status_code == 200
    assert ("metrics" in json.loads(rv.data)["meta"]) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 1, [0.0202], AGGREGATE_METRIC_METHOD_TAG)

    rv = client.get(
        '/aggregate?json={"seldonMessages": [{"data": {"names": ["input"], "ndarray": ["data"]}}]}'
    )
    assert rv.status_code == 200
    assert ("metrics" in json.loads(rv.data)["meta"]) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 2, [0.0202, 0.0202], AGGREGATE_METRIC_METHOD_TAG)


@pytest.mark.parametrize("cls", [UserObject])
def test_seldon_metrics_transform_input(cls, client_gets_metrics):
    user_object = cls()
    seldon_metrics = SeldonMetrics()

    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()

    rv = client.get(
        '/transform-input?json={"data": {"names": ["input"], "ndarray": ["data"]}}'
    )
    assert rv.status_code == 200
    assert ("metrics" in json.loads(rv.data)["meta"]) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 1, [0.0202], INPUT_TRANSFORM_METRIC_METHOD_TAG)

    rv = client.get(
        '/transform-input?json={"data": {"names": ["input"], "ndarray": ["data"]}}'
    )
    assert rv.status_code == 200
    assert ("metrics" in json.loads(rv.data)["meta"]) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 2, [0.0202, 0.0202], INPUT_TRANSFORM_METRIC_METHOD_TAG)


@pytest.mark.parametrize("cls", [UserObject])
def test_seldon_metrics_transform_output(cls, client_gets_metrics):
    user_object = cls()
    seldon_metrics = SeldonMetrics()

    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()

    rv = client.get(
        '/transform-output?json={"data": {"names": ["input"], "ndarray": ["data"]}}'
    )
    assert rv.status_code == 200
    assert ("metrics" in json.loads(rv.data)["meta"]) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 1, [0.0202], OUTPUT_TRANSFORM_METRIC_METHOD_TAG)

    rv = client.get(
        '/transform-output?json={"data": {"names": ["input"], "ndarray": ["data"]}}'
    )
    assert rv.status_code == 200
    assert ("metrics" in json.loads(rv.data)["meta"]) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 2, [0.0202, 0.0202], OUTPUT_TRANSFORM_METRIC_METHOD_TAG)


@pytest.mark.parametrize("cls", [UserObject])
def test_seldon_metrics_route(cls, client_gets_metrics):
    user_object = cls()
    seldon_metrics = SeldonMetrics()

    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()

    rv = client.get('/route?json={"data": {"names": ["input"], "ndarray": ["data"]}}')
    assert rv.status_code == 200
    assert ("metrics" in json.loads(rv.data)["meta"]) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 1, [0.0202], ROUTER_METRIC_METHOD_TAG)

    rv = client.get('/route?json={"data": {"names": ["input"], "ndarray": ["data"]}}')
    assert rv.status_code == 200
    assert ("metrics" in json.loads(rv.data)["meta"]) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 2, [0.0202, 0.0202], ROUTER_METRIC_METHOD_TAG)


@pytest.mark.parametrize("cls", [UserObject])
def test_proto_seldon_metrics_predict(cls, client_gets_metrics):
    user_object = cls()
    seldon_metrics = SeldonMetrics()

    app = SeldonModelGRPC(user_object, seldon_metrics)
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=(2, 1), values=np.array([1, 2]))
    )

    request = prediction_pb2.SeldonMessage(data=datadef)

    resp = app.Predict(request, None)
    assert (
        "metrics" in json.loads(json_format.MessageToJson(resp))["meta"]
    ) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 1, [0.0202], PREDICT_METRIC_METHOD_TAG)
    resp = app.Predict(request, None)
    assert (
        "metrics" in json.loads(json_format.MessageToJson(resp))["meta"]
    ) == client_gets_metrics
    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 2, [0.0202, 0.0202], PREDICT_METRIC_METHOD_TAG)


@pytest.mark.parametrize("cls", [UserObject])
def test_proto_seldon_metrics_aggregate(cls, client_gets_metrics):
    user_object = cls()
    seldon_metrics = SeldonMetrics()

    app = SeldonModelGRPC(user_object, seldon_metrics)

    arr1 = np.array([1, 2])
    datadef1 = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr1)
    )
    arr2 = np.array([3, 4])
    datadef2 = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr2)
    )
    msg1 = prediction_pb2.SeldonMessage(data=datadef1)
    msg2 = prediction_pb2.SeldonMessage(data=datadef2)

    request = prediction_pb2.SeldonMessageList(seldonMessages=[msg1, msg2])

    resp = app.Aggregate(request, None)
    assert (
        "metrics" in json.loads(json_format.MessageToJson(resp))["meta"]
    ) == client_gets_metrics
    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 1, [0.0202], AGGREGATE_METRIC_METHOD_TAG)

    resp = app.Aggregate(request, None)
    assert (
        "metrics" in json.loads(json_format.MessageToJson(resp))["meta"]
    ) == client_gets_metrics
    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 2, [0.0202, 0.0202], AGGREGATE_METRIC_METHOD_TAG)


@pytest.mark.parametrize("cls", [UserObject])
def test_proto_seldon_metrics_transform_input(cls, client_gets_metrics):
    user_object = cls()
    seldon_metrics = SeldonMetrics()

    app = SeldonModelGRPC(user_object, seldon_metrics)
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=(2, 1), values=np.array([1, 2]))
    )

    request = prediction_pb2.SeldonMessage(data=datadef)

    resp = app.TransformInput(request, None)
    assert (
        "metrics" in json.loads(json_format.MessageToJson(resp))["meta"]
    ) == client_gets_metrics
    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 1, [0.0202], INPUT_TRANSFORM_METRIC_METHOD_TAG)

    resp = app.TransformInput(request, None)
    assert (
        "metrics" in json.loads(json_format.MessageToJson(resp))["meta"]
    ) == client_gets_metrics
    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 2, [0.0202, 0.0202], INPUT_TRANSFORM_METRIC_METHOD_TAG)


@pytest.mark.parametrize("cls", [UserObject])
def test_proto_seldon_metrics_transform_output(cls, client_gets_metrics):
    user_object = cls()
    seldon_metrics = SeldonMetrics()

    app = SeldonModelGRPC(user_object, seldon_metrics)
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=(2, 1), values=np.array([1, 2]))
    )

    request = prediction_pb2.SeldonMessage(data=datadef)

    resp = app.TransformOutput(request, None)
    assert (
        "metrics" in json.loads(json_format.MessageToJson(resp))["meta"]
    ) == client_gets_metrics
    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 1, [0.0202], OUTPUT_TRANSFORM_METRIC_METHOD_TAG)

    resp = app.TransformOutput(request, None)
    assert (
        "metrics" in json.loads(json_format.MessageToJson(resp))["meta"]
    ) == client_gets_metrics
    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 2, [0.0202, 0.0202], OUTPUT_TRANSFORM_METRIC_METHOD_TAG)


@pytest.mark.parametrize("cls", [UserObject])
def test_proto_seldon_metrics_route(cls, client_gets_metrics):
    user_object = cls()
    seldon_metrics = SeldonMetrics()

    app = SeldonModelGRPC(user_object, seldon_metrics)
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=(2, 1), values=np.array([1, 2]))
    )

    request = prediction_pb2.SeldonMessage(data=datadef)
    resp = app.Route(request, None)
    assert (
        "metrics" in json.loads(json_format.MessageToJson(resp))["meta"]
    ) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 1, [0.0202], ROUTER_METRIC_METHOD_TAG)
    resp = app.Route(request, None)
    assert (
        "metrics" in json.loads(json_format.MessageToJson(resp))["meta"]
    ) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 2, [0.0202, 0.0202], ROUTER_METRIC_METHOD_TAG)


@pytest.mark.parametrize("cls", [UserObject])
def test_seldon_metrics_endpoint(cls, client_gets_metrics):
    def _match_label(line):
        _data, value = line.split()
        name, labels = _data.split()[0].split("{")
        labels = labels[:-1]
        return name, value, eval(f"dict({labels})")

    def _iterate_metrics(text):
        for line in text.split("\n"):
            if not line or line[0] == "#":
                continue
            yield _match_label(line)

    user_object = cls()
    seldon_metrics = SeldonMetrics()

    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()

    metrics_app = get_metrics_microservice(seldon_metrics)
    metrics_client = metrics_app.test_client()

    rv = metrics_client.get("/metrics")
    assert rv.status_code == 200
    assert rv.data.decode() == ""

    rv = client.get('/predict?json={"data": {"names": ["input"], "ndarray": ["data"]}}')
    assert rv.status_code == 200
    assert ("metrics" in json.loads(rv.data)["meta"]) == client_gets_metrics

    rv = metrics_client.get("/metrics")
    text = rv.data.decode()

    timer_present = False
    for name, value, labels in _iterate_metrics(text):
        if name == "mytimer_bucket":
            timer_present = True

        if name == "mycounter_total":
            assert value == "1.0"
            assert labels["worker_id"] == str(os.getpid())

        if name == "mygauge":
            assert value == "100.0"
            assert labels["worker_id"] == str(os.getpid())

        if name == "customtag":
            assert value == "200.0"
            assert labels["mytag"] == "mytagvalue"

    assert timer_present


@pytest.mark.parametrize("cls", [UserObject])
def test_proto_seldon_metrics_endpoint(cls, client_gets_metrics):
    def _match_label(line):
        _data, value = line.split()
        name, labels = _data.split()[0].split("{")
        labels = labels[:-1]
        return name, value, eval(f"dict({labels})")

    def _iterate_metrics(text):
        for line in text.split("\n"):
            if not line or line[0] == "#":
                continue
            yield _match_label(line)

    user_object = cls()
    seldon_metrics = SeldonMetrics()

    app = SeldonModelGRPC(user_object, seldon_metrics)
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=(2, 1), values=np.array([1, 2]))
    )

    request = prediction_pb2.SeldonMessage(data=datadef)

    metrics_app = get_metrics_microservice(seldon_metrics)
    metrics_client = metrics_app.test_client()

    rv = metrics_client.get("/metrics")
    assert rv.status_code == 200
    assert rv.data.decode() == ""

    resp = app.Predict(request, None)
    assert (
        "metrics" in json.loads(json_format.MessageToJson(resp))["meta"]
    ) == client_gets_metrics
    rv = metrics_client.get("/metrics")
    text = rv.data.decode()

    timer_present = False
    for name, value, labels in _iterate_metrics(text):
        if name == "mytimer_bucket":
            timer_present = True

        if name == "mycounter_total":
            assert value == "1.0"
            assert labels["worker_id"] == str(os.getpid())

        if name == "mygauge":
            assert value == "100.0"
            assert labels["worker_id"] == str(os.getpid())

        if name == "customtag":
            assert value == "200.0"
            assert labels["mytag"] == "mytagvalue"

    assert timer_present
