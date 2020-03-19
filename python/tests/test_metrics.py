import os
import logging
import pytest
import numpy as np
from google.protobuf import json_format
import json

from seldon_core.flask_utils import SeldonMicroserviceException
from seldon_core.proto import prediction_pb2
from seldon_core.wrapper import get_rest_microservice, get_metrics_microservice
from seldon_core.metrics import (
    SeldonMetrics,
    create_counter,
    create_gauge,
    create_timer,
    validate_metrics,
    COUNTER,
    BINS,
)
from seldon_core.user_model import client_custom_metrics


def test_create_counter():
    v = create_counter("k", 1)
    assert v["type"] == "COUNTER"


def test_create_counter_invalid_value():
    with pytest.raises(TypeError):
        v = create_counter("k", "invalid")


def test_create_timer():
    v = create_timer("k", 1)
    assert v["type"] == "TIMER"


def test_create_timer_invalid_value():
    with pytest.raises(TypeError):
        v = create_timer("k", "invalid")


def test_create_gauge():
    v = create_gauge("k", 1)
    assert v["type"] == "GAUGE"


def test_create_gauge_invalid_value():
    with pytest.raises(TypeError):
        v = create_gauge("k", "invalid")


def test_validate_ok():
    assert validate_metrics([{"type": COUNTER, "key": "a", "value": 1}]) == True


def test_validate_bad_type():
    assert validate_metrics([{"type": "ABC", "key": "a", "value": 1}]) == False


def test_validate_no_type():
    assert validate_metrics([{"key": "a", "value": 1}]) == False


def test_validate_no_key():
    assert validate_metrics([{"type": COUNTER, "value": 1}]) == False


def test_validate_no_value():
    assert validate_metrics([{"type": COUNTER, "key": "a"}]) == False


def test_validate_bad_value():
    assert validate_metrics([{"type": COUNTER, "key": "a", "value": "1"}]) == False


def test_validate_no_list():
    assert validate_metrics({"type": COUNTER, "key": "a", "value": 1}) == False


class Component(object):
    def __init__(self, ok=True):
        self.ok = ok

    def metrics(self):
        if self.ok:
            return [{"type": COUNTER, "key": "a", "value": 1}]
        else:
            return [{"type": "bad", "key": "a", "value": 1}]


def test_component_ok():
    c = Component(True)
    assert client_custom_metrics(c) == c.metrics()


def test_component_bad():
    with pytest.raises(SeldonMicroserviceException):
        c = Component(False)
        client_custom_metrics(c)


def test_proto_metrics():
    metrics = [{"type": "COUNTER", "key": "a", "value": 1}]
    meta = prediction_pb2.Meta()
    for metric in metrics:
        mpb2 = meta.metrics.add()
        json_format.ParseDict(metric, mpb2)


def test_proto_tags():
    metric = {
        "tags": {"t1": "t2"},
        "metrics": [
            {"type": "COUNTER", "key": "mycounter", "value": 1.2},
            {"type": "GAUGE", "key": "mygauge", "value": 1.2},
            {"type": "TIMER", "key": "mytimer", "value": 1.2},
        ],
    }
    meta = prediction_pb2.Meta()
    json_format.ParseDict(metric, meta)
    jStr = json_format.MessageToJson(meta)
    j = json.loads(jStr)
    assert "mycounter" == j["metrics"][0]["key"]
    assert 1.2 == pytest.approx(j["metrics"][0]["value"], 0.01)
    assert "GAUGE" == j["metrics"][1]["type"]
    assert "mygauge" == j["metrics"][1]["key"]
    assert 1.2 == pytest.approx(j["metrics"][1]["value"], 0.01)
    assert "TIMER" == j["metrics"][2]["type"]
    assert "mytimer" == j["metrics"][2]["key"]
    assert 1.2 == pytest.approx(j["metrics"][2]["value"], 0.01)


class UserObject:
    def predict(self, X, features_names):
        logging.info("Predict called")
        return X

    def metrics(self):
        logging.info("metrics called")
        return [
            {"type": "COUNTER", "key": "mycounter", "value": 1},
            {"type": "GAUGE", "key": "mygauge", "value": 100},
            {"type": "TIMER", "key": "mytimer", "value": 20.2},
        ]


class UserObjectLowLevel:
    def predict_raw(self, msg):
        metrics = [
            {"type": "COUNTER", "key": "mycounter", "value": 1},
            {"type": "GAUGE", "key": "mygauge", "value": 100},
            {"type": "TIMER", "key": "mytimer", "value": 20.2},
        ]

        return {
            "meta": {"metrics": metrics},
            "data": msg["data"],
        }

@pytest.mark.parametrize("cls", [UserObject, UserObjectLowLevel])
def test_seldon_metrics_gauge(cls):
    user_object = cls()
    seldon_metrics = SeldonMetrics()

    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()

    rv = client.get('/predict?json={"data": {"names": ["input"], "ndarray": ["data"]}}')
    assert rv.status_code == 200

    data = seldon_metrics.data[os.getpid()]
    assert data["GAUGE", "mygauge"] == 100


@pytest.mark.parametrize("cls", [UserObject, UserObjectLowLevel])
def test_seldon_metrics_counter(cls):
    user_object = cls()
    seldon_metrics = SeldonMetrics()

    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()

    rv = client.get('/predict?json={"data": {"names": ["input"], "ndarray": ["data"]}}')
    assert rv.status_code == 200
    data = seldon_metrics.data[os.getpid()]
    assert data["COUNTER", "mycounter"] == 1

    rv = client.get('/predict?json={"data": {"names": ["input"], "ndarray": ["data"]}}')
    assert rv.status_code == 200
    data = seldon_metrics.data[os.getpid()]
    assert data["COUNTER", "mycounter"] == 2


@pytest.mark.parametrize("cls", [UserObject, UserObjectLowLevel])
def test_seldon_metrics_histogram(cls):
    user_object = cls()
    seldon_metrics = SeldonMetrics()

    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()

    rv = client.get('/predict?json={"data": {"names": ["input"], "ndarray": ["data"]}}')
    assert rv.status_code == 200
    data = seldon_metrics.data[os.getpid()]
    assert np.allclose(
        np.histogram([20.2 / 1000], BINS)[0], data["TIMER", "mytimer"][0]
    )
    assert np.allclose(data["TIMER", "mytimer"][1], 0.0202)

    rv = client.get('/predict?json={"data": {"names": ["input"], "ndarray": ["data"]}}')
    assert rv.status_code == 200
    data = seldon_metrics.data[os.getpid()]
    assert np.allclose(
        np.histogram([20.2 / 1000, 20.2 / 1000], BINS)[0], data["TIMER", "mytimer"][0]
    )
    assert np.allclose(data["TIMER", "mytimer"][1], 0.0404)


@pytest.mark.parametrize("cls", [UserObject, UserObjectLowLevel])
def test_seldon_metrics_endpoint(cls):
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

    assert timer_present
