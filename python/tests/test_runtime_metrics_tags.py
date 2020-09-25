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
from seldon_core.user_model import client_custom_metrics, SeldonResponse


RUNTIME_METRICS = [{"type": "GAUGE", "key": "runtime_gauge", "value": 42}]

RUNTIME_TAGS = {"runtime": "tag", "shared": "right one"}
EXPECTED_TAGS = {"static": "tag", **RUNTIME_TAGS}


class UserObject:
    def predict(self, X, features_names):
        logging.info("Predict called")
        return SeldonResponse(data=X, metrics=RUNTIME_METRICS, tags=RUNTIME_TAGS)

    def aggregate(self, X, features_names):
        logging.info("Aggregate called")
        return SeldonResponse(data=X[0], metrics=RUNTIME_METRICS, tags=RUNTIME_TAGS)

    def transform_input(self, X, feature_names):
        logging.info("Transform input called")
        return SeldonResponse(data=X, metrics=RUNTIME_METRICS, tags=RUNTIME_TAGS)

    def transform_output(self, X, feature_names):
        logging.info("Transform output called")
        return SeldonResponse(data=X, metrics=RUNTIME_METRICS, tags=RUNTIME_TAGS)

    def route(self, X, feature_names):
        logging.info("Route called")
        return SeldonResponse(data=22, metrics=RUNTIME_METRICS, tags=RUNTIME_TAGS)

    def send_feedback(self, X, feature_names, reward, truth, routing):
        return SeldonResponse(data=X, metrics=RUNTIME_METRICS, tags=RUNTIME_TAGS)

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

    def tags(self):
        return {"static": "tag", "shared": "not right one"}


def verify_seldon_metrics(data, mycounter_value, histogram_entries, method):
    expected_base_tags = {"method": method}
    base_tags_key = SeldonMetrics._generate_tags_key(expected_base_tags)
    expected_custom_tags = {"mytag": "mytagvalue", "method": method}
    custom_tags_key = SeldonMetrics._generate_tags_key(expected_custom_tags)
    assert data["GAUGE", "runtime_gauge", base_tags_key]["value"] == 42
    assert data["GAUGE", "mygauge", base_tags_key]["value"] == 100
    assert data["GAUGE", "customtag", custom_tags_key]["value"] == 200
    assert data["GAUGE", "customtag", custom_tags_key]["tags"] == expected_custom_tags
    assert data["COUNTER", "mycounter", base_tags_key]["value"] == mycounter_value
    assert np.allclose(
        np.histogram(histogram_entries, BINS)[0],
        data["TIMER", "mytimer", base_tags_key]["value"][0],
    )
    assert np.allclose(
        data["TIMER", "mytimer", base_tags_key]["value"][1], np.sum(histogram_entries)
    )


@pytest.mark.parametrize("cls", [UserObject])
def test_seldon_runtime_data_predict(cls, client_gets_metrics):
    user_object = cls()
    seldon_metrics = SeldonMetrics()

    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()

    rv = client.get('/predict?json={"data": {"names": ["input"], "ndarray": ["data"]}}')
    assert rv.status_code == 200
    j = json.loads(rv.data)
    assert j["data"]["ndarray"] == ["data"]
    assert j["meta"]["tags"] == EXPECTED_TAGS
    assert ("metrics" in j["meta"]) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 1, [0.0202], PREDICT_METRIC_METHOD_TAG)

    rv = client.get('/predict?json={"data": {"names": ["input"], "ndarray": ["data"]}}')
    assert rv.status_code == 200
    j = json.loads(rv.data)
    assert j["data"]["ndarray"] == ["data"]
    assert j["meta"]["tags"] == EXPECTED_TAGS
    assert ("metrics" in j["meta"]) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 2, [0.0202, 0.0202], PREDICT_METRIC_METHOD_TAG)


@pytest.mark.parametrize("cls", [UserObject])
def test_seldon_runtime_data_send_feedback(cls):
    user_object = cls()
    seldon_metrics = SeldonMetrics()

    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()

    rv = client.get('/send-feedback?json={"reward": 42}')
    assert rv.status_code == 200
    j = json.loads(rv.data)
    assert j["meta"]["tags"] == EXPECTED_TAGS

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 1, [0.0202], FEEDBACK_METRIC_METHOD_TAG)

    expected_base_tags = {"method": FEEDBACK_METRIC_METHOD_TAG}
    base_tags_key = SeldonMetrics._generate_tags_key(expected_base_tags)

    assert data["COUNTER", "seldon_api_model_feedback_reward", base_tags_key] == {
        "value": 42.0,
        "tags": expected_base_tags,
    }

    rv = client.get('/send-feedback?json={"reward": 42}')
    assert rv.status_code == 200

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 2, [0.0202, 0.0202], FEEDBACK_METRIC_METHOD_TAG)

    assert data["COUNTER", "seldon_api_model_feedback_reward", base_tags_key] == {
        "value": 84.0,
        "tags": expected_base_tags,
    }


@pytest.mark.parametrize("cls", [UserObject])
def test_seldon_runtime_data_aggregate(cls, client_gets_metrics):
    user_object = cls()
    seldon_metrics = SeldonMetrics()

    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()

    rv = client.get(
        '/aggregate?json={"seldonMessages": [{"data": {"names": ["input"], "ndarray": ["data"]}}]}'
    )
    assert rv.status_code == 200
    j = json.loads(rv.data)
    assert j["data"]["ndarray"] == ["data"]
    assert j["meta"]["tags"] == EXPECTED_TAGS
    assert ("metrics" in j["meta"]) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 1, [0.0202], AGGREGATE_METRIC_METHOD_TAG)

    rv = client.get(
        '/aggregate?json={"seldonMessages": [{"data": {"names": ["input"], "ndarray": ["data"]}}]}'
    )
    assert rv.status_code == 200
    j = json.loads(rv.data)
    assert j["data"]["ndarray"] == ["data"]
    assert j["meta"]["tags"] == EXPECTED_TAGS
    assert ("metrics" in j["meta"]) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 2, [0.0202, 0.0202], AGGREGATE_METRIC_METHOD_TAG)


@pytest.mark.parametrize("cls", [UserObject])
def test_seldon_runtime_data_transform_input(cls, client_gets_metrics):
    user_object = cls()
    seldon_metrics = SeldonMetrics()

    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()

    rv = client.get(
        '/transform-input?json={"data": {"names": ["input"], "ndarray": ["data"]}}'
    )
    assert rv.status_code == 200
    j = json.loads(rv.data)
    assert j["data"]["ndarray"] == ["data"]
    assert j["meta"]["tags"] == EXPECTED_TAGS
    assert ("metrics" in j["meta"]) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 1, [0.0202], INPUT_TRANSFORM_METRIC_METHOD_TAG)

    rv = client.get(
        '/transform-input?json={"data": {"names": ["input"], "ndarray": ["data"]}}'
    )
    assert rv.status_code == 200
    j = json.loads(rv.data)
    assert j["data"]["ndarray"] == ["data"]
    assert j["meta"]["tags"] == EXPECTED_TAGS
    assert ("metrics" in j["meta"]) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 2, [0.0202, 0.0202], INPUT_TRANSFORM_METRIC_METHOD_TAG)


@pytest.mark.parametrize("cls", [UserObject])
def test_seldon_runtime_data_transform_output(cls, client_gets_metrics):
    user_object = cls()
    seldon_metrics = SeldonMetrics()

    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()

    rv = client.get(
        '/transform-output?json={"data": {"names": ["input"], "ndarray": ["data"]}}'
    )
    assert rv.status_code == 200
    j = json.loads(rv.data)
    assert j["data"]["ndarray"] == ["data"]
    assert j["meta"]["tags"] == EXPECTED_TAGS
    assert ("metrics" in j["meta"]) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 1, [0.0202], OUTPUT_TRANSFORM_METRIC_METHOD_TAG)

    rv = client.get(
        '/transform-output?json={"data": {"names": ["input"], "ndarray": ["data"]}}'
    )
    assert rv.status_code == 200
    j = json.loads(rv.data)
    assert j["data"]["ndarray"] == ["data"]
    assert j["meta"]["tags"] == EXPECTED_TAGS
    assert ("metrics" in j["meta"]) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 2, [0.0202, 0.0202], OUTPUT_TRANSFORM_METRIC_METHOD_TAG)


@pytest.mark.parametrize("cls", [UserObject])
def test_seldon_runtime_data_route(cls, client_gets_metrics):
    user_object = cls()
    seldon_metrics = SeldonMetrics()

    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()

    rv = client.get('/route?json={"data": {"names": ["input"], "ndarray": ["data"]}}')
    assert rv.status_code == 200
    j = json.loads(rv.data)
    assert j["data"]["ndarray"] == [[22]]
    assert j["meta"]["tags"] == EXPECTED_TAGS
    assert ("metrics" in j["meta"]) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 1, [0.0202], ROUTER_METRIC_METHOD_TAG)

    rv = client.get('/route?json={"data": {"names": ["input"], "ndarray": ["data"]}}')
    assert rv.status_code == 200
    j = json.loads(rv.data)
    assert j["data"]["ndarray"] == [[22]]
    assert j["meta"]["tags"] == EXPECTED_TAGS
    assert ("metrics" in j["meta"]) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 2, [0.0202, 0.0202], ROUTER_METRIC_METHOD_TAG)


@pytest.mark.parametrize("cls", [UserObject])
def test_proto_seldon_runtime_data_predict(cls, client_gets_metrics):
    user_object = cls()
    seldon_metrics = SeldonMetrics()

    app = SeldonModelGRPC(user_object, seldon_metrics)
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=(2, 1), values=np.array([1, 2]))
    )

    request = prediction_pb2.SeldonMessage(data=datadef)

    resp = app.Predict(request, None)

    j = json.loads(json_format.MessageToJson(resp))
    assert j["data"] == {
        "names": ["t:0"],
        "tensor": {"shape": [2, 1], "values": [1.0, 2.0]},
    }
    assert j["meta"]["tags"] == EXPECTED_TAGS
    assert ("metrics" in j["meta"]) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 1, [0.0202], PREDICT_METRIC_METHOD_TAG)
    resp = app.Predict(request, None)
    j = json.loads(json_format.MessageToJson(resp))
    assert j["data"] == {
        "names": ["t:0"],
        "tensor": {"shape": [2, 1], "values": [1.0, 2.0]},
    }
    assert j["meta"]["tags"] == EXPECTED_TAGS
    assert ("metrics" in j["meta"]) == client_gets_metrics
    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 2, [0.0202, 0.0202], PREDICT_METRIC_METHOD_TAG)


@pytest.mark.parametrize("cls", [UserObject])
def test_proto_seldon_runtime_data_aggregate(cls, client_gets_metrics):
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
    j = json.loads(json_format.MessageToJson(resp))
    assert j["data"] == {
        "names": ["t:0"],
        "tensor": {"shape": [2, 1], "values": [1.0, 2.0]},
    }
    assert j["meta"]["tags"] == EXPECTED_TAGS
    assert ("metrics" in j["meta"]) == client_gets_metrics
    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 1, [0.0202], AGGREGATE_METRIC_METHOD_TAG)

    resp = app.Aggregate(request, None)
    j = json.loads(json_format.MessageToJson(resp))
    assert j["data"] == {
        "names": ["t:0"],
        "tensor": {"shape": [2, 1], "values": [1.0, 2.0]},
    }
    assert j["meta"]["tags"] == EXPECTED_TAGS
    assert ("metrics" in j["meta"]) == client_gets_metrics
    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 2, [0.0202, 0.0202], AGGREGATE_METRIC_METHOD_TAG)


@pytest.mark.parametrize("cls", [UserObject])
def test_proto_seldon_runtime_data_transform_input(cls, client_gets_metrics):
    user_object = cls()
    seldon_metrics = SeldonMetrics()

    app = SeldonModelGRPC(user_object, seldon_metrics)
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=(2, 1), values=np.array([1, 2]))
    )

    request = prediction_pb2.SeldonMessage(data=datadef)

    resp = app.TransformInput(request, None)
    j = json.loads(json_format.MessageToJson(resp))
    assert j["data"] == {
        "names": ["t:0"],
        "tensor": {"shape": [2, 1], "values": [1.0, 2.0]},
    }
    assert j["meta"]["tags"] == EXPECTED_TAGS
    assert ("metrics" in j["meta"]) == client_gets_metrics
    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 1, [0.0202], INPUT_TRANSFORM_METRIC_METHOD_TAG)

    resp = app.TransformInput(request, None)
    j = json.loads(json_format.MessageToJson(resp))
    assert j["data"] == {
        "names": ["t:0"],
        "tensor": {"shape": [2, 1], "values": [1.0, 2.0]},
    }
    assert j["meta"]["tags"] == EXPECTED_TAGS
    assert ("metrics" in j["meta"]) == client_gets_metrics
    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 2, [0.0202, 0.0202], INPUT_TRANSFORM_METRIC_METHOD_TAG)


@pytest.mark.parametrize("cls", [UserObject])
def test_proto_seldon_runtime_data_transform_output(cls, client_gets_metrics):
    user_object = cls()
    seldon_metrics = SeldonMetrics()

    app = SeldonModelGRPC(user_object, seldon_metrics)
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=(2, 1), values=np.array([1, 2]))
    )

    request = prediction_pb2.SeldonMessage(data=datadef)

    resp = app.TransformOutput(request, None)
    j = json.loads(json_format.MessageToJson(resp))
    assert j["data"] == {
        "names": ["t:0"],
        "tensor": {"shape": [2, 1], "values": [1.0, 2.0]},
    }
    assert j["meta"]["tags"] == EXPECTED_TAGS
    assert ("metrics" in j["meta"]) == client_gets_metrics
    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 1, [0.0202], OUTPUT_TRANSFORM_METRIC_METHOD_TAG)

    resp = app.TransformOutput(request, None)
    j = json.loads(json_format.MessageToJson(resp))
    assert j["data"] == {
        "names": ["t:0"],
        "tensor": {"shape": [2, 1], "values": [1.0, 2.0]},
    }
    assert j["meta"]["tags"] == EXPECTED_TAGS
    assert ("metrics" in j["meta"]) == client_gets_metrics
    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 2, [0.0202, 0.0202], OUTPUT_TRANSFORM_METRIC_METHOD_TAG)


@pytest.mark.parametrize("cls", [UserObject])
def test_proto_seldon_runtime_data_route(cls, client_gets_metrics):
    user_object = cls()
    seldon_metrics = SeldonMetrics()

    app = SeldonModelGRPC(user_object, seldon_metrics)
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=(2, 1), values=np.array([1, 2]))
    )

    request = prediction_pb2.SeldonMessage(data=datadef)
    resp = app.Route(request, None)
    j = json.loads(json_format.MessageToJson(resp))
    assert j["data"] == {
        "names": ["t:0"],
        "tensor": {"shape": [1, 1], "values": [22.0]},
    }
    assert j["meta"]["tags"] == EXPECTED_TAGS
    assert ("metrics" in j["meta"]) == client_gets_metrics

    data = seldon_metrics.data[os.getpid()]
    verify_seldon_metrics(data, 1, [0.0202], ROUTER_METRIC_METHOD_TAG)
    resp = app.Route(request, None)
    j = json.loads(json_format.MessageToJson(resp))
    assert j["data"] == {
        "names": ["t:0"],
        "tensor": {"shape": [1, 1], "values": [22.0]},
    }
    assert j["meta"]["tags"] == EXPECTED_TAGS
    assert ("metrics" in j["meta"]) == client_gets_metrics

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
    j = json.loads(json_format.MessageToJson(resp))
    assert j["data"] == {
        "names": ["t:0"],
        "tensor": {"shape": [2, 1], "values": [1.0, 2.0]},
    }
    assert j["meta"]["tags"] == EXPECTED_TAGS
    assert ("metrics" in j["meta"]) == client_gets_metrics
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
