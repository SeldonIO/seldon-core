import pytest
from google.protobuf import json_format
import json

from seldon_core.flask_utils import SeldonMicroserviceException
from seldon_core.proto import prediction_pb2, prediction_pb2_grpc
from seldon_core.metrics import *
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
