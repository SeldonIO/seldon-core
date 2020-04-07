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
    validate_metrics,
    COUNTER,
    BINS,
)
from seldon_core.user_model import client_custom_metrics

import seldon_core


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
    assert client_custom_metrics(c, SeldonMetrics()) == c.metrics()


def test_component_bad():
    with pytest.raises(SeldonMicroserviceException):
        c = Component(False)
        client_custom_metrics(c, SeldonMetrics())


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

    def aggregate(self, X, features_names):
        logging.info("Aggregate called")
        return X[0]

    def transform_input(self, X, feature_names):
        logging.info("Transform input called")
        return X

    def transform_output(self, X, feature_names):
        logging.info("Transform output called")
        return X

    def route(self, X, feature_names):
        logging.info("Route called")
        return 22

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


class UserObjectLowLevel:
    _metrics = [
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

    def predict_raw(self, msg):
        logging.info("Predict raw called")
        return {
            "meta": {"metrics": self._metrics},
            "data": {"names": ["input"], "data": ["output"]},
        }

    def aggregate_raw(self, msg):
        logging.info("Aggregate raw called")
        return {
            "meta": {"metrics": self._metrics},
            "data": {"names": ["input"], "data": ["output"]},
        }

    def transform_input_raw(self, msg):
        logging.info("Transform input raw called")
        return {
            "meta": {"metrics": self._metrics},
            "data": {"names": ["input"], "data": ["output"]},
        }

    def transform_output_raw(self, msg):
        logging.info("Transform output raw called")
        return {
            "meta": {"metrics": self._metrics},
            "data": {"names": ["input"], "data": ["output"]},
        }

    def route_raw(self, msg):
        logging.info("Route raw called")
        return {
            "meta": {"metrics": self._metrics},
            "data": {"names": ["input"], "data": [22]},
        }


class UserObjectLowLevelGrpc:
    _metrics = [
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

    def predict_raw(self, msg):
        logging.info("Predict raw called")

        meta = prediction_pb2.Meta()
        json_format.ParseDict({"metrics": self._metrics}, meta)

        arr = np.array([9, 9])
        datadef = prediction_pb2.DefaultData(
            tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
        )
        request = prediction_pb2.SeldonMessage(data=datadef, meta=meta)
        return request

    def aggregate_raw(self, msg):
        logging.info("Aggregate raw called")

        meta = prediction_pb2.Meta()
        json_format.ParseDict({"metrics": self._metrics}, meta)

        arr = np.array([9, 9])
        datadef = prediction_pb2.DefaultData(
            tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
        )
        request = prediction_pb2.SeldonMessage(data=datadef, meta=meta)
        return request

    def transform_input_raw(self, msg):
        logging.info("Transform input raw called")

        meta = prediction_pb2.Meta()
        json_format.ParseDict({"metrics": self._metrics}, meta)

        arr = np.array([9, 9])
        datadef = prediction_pb2.DefaultData(
            tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
        )
        request = prediction_pb2.SeldonMessage(data=datadef, meta=meta)
        return request

    def transform_output_raw(self, msg):
        logging.info("Transform output raw called")

        meta = prediction_pb2.Meta()
        json_format.ParseDict({"metrics": self._metrics}, meta)

        arr = np.array([9, 9])
        datadef = prediction_pb2.DefaultData(
            tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
        )
        request = prediction_pb2.SeldonMessage(data=datadef, meta=meta)
        return request

    def route_raw(self, msg):
        logging.info("Route raw called")

        meta = prediction_pb2.Meta()
        json_format.ParseDict({"metrics": self._metrics}, meta)

        arr = np.array([22])
        datadef = prediction_pb2.DefaultData(
            tensor=prediction_pb2.Tensor(shape=(1, 1), values=arr)
        )
        request = prediction_pb2.SeldonMessage(data=datadef, meta=meta)
        return request


def verify_seldon_metrics(data, mycounter_value, histogram_entries):
    assert data["GAUGE", "mygauge"]["value"] == 100
    assert data["GAUGE", "customtag"]["value"] == 200
    assert data["GAUGE", "customtag"]["tags"] == {"mytag": "mytagvalue"}
    assert data["COUNTER", "mycounter"]["value"] == mycounter_value
    assert np.allclose(
        np.histogram(histogram_entries, BINS)[0], data["TIMER", "mytimer"]["value"][0]
    )
    assert np.allclose(data["TIMER", "mytimer"]["value"][1], np.sum(histogram_entries))


@pytest.mark.parametrize("cls", [UserObject, UserObjectLowLevel])
def test_seldon_metrics_predict(cls, client_gets_metrics):
    user_object = cls()
    seldon_metrics = SeldonMetrics()

    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()

    rv = client.get('/predict?json={"data": {"names": ["input"], "ndarray": ["data"]}}')
    assert rv.status_code == 200
    assert ("metrics" in json.loads(rv.data)["meta"]) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 1, [0.0202])

    rv = client.get('/predict?json={"data": {"names": ["input"], "ndarray": ["data"]}}')
    assert rv.status_code == 200
    assert ("metrics" in json.loads(rv.data)["meta"]) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 2, [0.0202, 0.0202])


@pytest.mark.parametrize("cls", [UserObject, UserObjectLowLevel])
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
    verify_seldon_metrics(data, 1, [0.0202])

    rv = client.get(
        '/aggregate?json={"seldonMessages": [{"data": {"names": ["input"], "ndarray": ["data"]}}]}'
    )
    assert rv.status_code == 200
    assert ("metrics" in json.loads(rv.data)["meta"]) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 2, [0.0202, 0.0202])


@pytest.mark.parametrize("cls", [UserObject, UserObjectLowLevel])
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
    verify_seldon_metrics(data, 1, [0.0202])

    rv = client.get(
        '/transform-input?json={"data": {"names": ["input"], "ndarray": ["data"]}}'
    )
    assert rv.status_code == 200
    assert ("metrics" in json.loads(rv.data)["meta"]) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 2, [0.0202, 0.0202])


@pytest.mark.parametrize("cls", [UserObject, UserObjectLowLevel])
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
    verify_seldon_metrics(data, 1, [0.0202])

    rv = client.get(
        '/transform-output?json={"data": {"names": ["input"], "ndarray": ["data"]}}'
    )
    assert rv.status_code == 200
    assert ("metrics" in json.loads(rv.data)["meta"]) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 2, [0.0202, 0.0202])


@pytest.mark.parametrize("cls", [UserObject, UserObjectLowLevel])
def test_seldon_metrics_route(cls, client_gets_metrics):
    user_object = cls()
    seldon_metrics = SeldonMetrics()

    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()

    rv = client.get('/route?json={"data": {"names": ["input"], "ndarray": ["data"]}}')
    assert rv.status_code == 200
    assert ("metrics" in json.loads(rv.data)["meta"]) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 1, [0.0202])

    rv = client.get('/route?json={"data": {"names": ["input"], "ndarray": ["data"]}}')
    assert rv.status_code == 200
    assert ("metrics" in json.loads(rv.data)["meta"]) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 2, [0.0202, 0.0202])


@pytest.mark.parametrize("cls", [UserObject, UserObjectLowLevelGrpc])
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
    verify_seldon_metrics(data, 1, [0.0202])
    resp = app.Predict(request, None)
    assert (
        "metrics" in json.loads(json_format.MessageToJson(resp))["meta"]
    ) == client_gets_metrics
    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 2, [0.0202, 0.0202])


@pytest.mark.parametrize("cls", [UserObject, UserObjectLowLevelGrpc])
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
    verify_seldon_metrics(data, 1, [0.0202])

    resp = app.Aggregate(request, None)
    assert (
        "metrics" in json.loads(json_format.MessageToJson(resp))["meta"]
    ) == client_gets_metrics
    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 2, [0.0202, 0.0202])


@pytest.mark.parametrize("cls", [UserObject, UserObjectLowLevelGrpc])
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
    verify_seldon_metrics(data, 1, [0.0202])

    resp = app.TransformInput(request, None)
    assert (
        "metrics" in json.loads(json_format.MessageToJson(resp))["meta"]
    ) == client_gets_metrics
    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 2, [0.0202, 0.0202])


@pytest.mark.parametrize("cls", [UserObject, UserObjectLowLevelGrpc])
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
    verify_seldon_metrics(data, 1, [0.0202])

    resp = app.TransformOutput(request, None)
    assert (
        "metrics" in json.loads(json_format.MessageToJson(resp))["meta"]
    ) == client_gets_metrics
    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 2, [0.0202, 0.0202])


@pytest.mark.parametrize("cls", [UserObject, UserObjectLowLevelGrpc])
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
    verify_seldon_metrics(data, 1, [0.0202])
    resp = app.Route(request, None)
    assert (
        "metrics" in json.loads(json_format.MessageToJson(resp))["meta"]
    ) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 2, [0.0202, 0.0202])


@pytest.mark.parametrize("cls", [UserObject, UserObjectLowLevel])
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


@pytest.mark.parametrize("cls", [UserObject, UserObjectLowLevelGrpc])
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
