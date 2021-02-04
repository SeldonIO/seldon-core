import json
import logging
from typing import Dict, List, Union

import numpy as np
from google.protobuf import json_format

from seldon_core.metrics import SeldonMetrics
from seldon_core.proto import prediction_pb2
from seldon_core.user_model import SeldonComponent
from seldon_core.utils import seldon_message_to_json
from seldon_core.wrapper import SeldonModelGRPC, get_grpc_server, get_rest_microservice


class MinimalUserObject:
    def route(self, X, features_names):
        return 22


class UserObject:
    def __init__(self, metrics_ok=True, ret_meta=False):
        self.metrics_ok = metrics_ok
        self.ret_meta = ret_meta

    def route(self, X, features_names, **kwargs):
        logging.info("Route called")
        if self.ret_meta:
            self.inc_meta = kwargs.get("meta")
        return 22

    def send_feedback(self, features, feature_names, reward, truth, routing=-1):
        logging.info("Feedback called")

    def tags(self):
        if self.ret_meta:
            return {"inc_meta": self.inc_meta}
        else:
            return {"mytag": 1}

    def metrics(self):
        if self.metrics_ok:
            return [{"type": "COUNTER", "key": "mycounter", "value": 1}]
        else:
            return [{"type": "BAD", "key": "mycounter", "value": 1}]


class UserObjectLowLevel:
    def __init__(self, metrics_ok=True, ret_nparray=False):
        self.metrics_ok = metrics_ok
        self.ret_nparray = ret_nparray
        self.nparray = np.array([1, 2, 3])

    def route_rest(self, request):
        return {"data": {"ndarray": [[1]]}}

    def route_grpc(self, request):
        arr = np.array([1])
        datadef = prediction_pb2.DefaultData(
            tensor=prediction_pb2.Tensor(shape=(1, 1), values=arr)
        )
        request = prediction_pb2.SeldonMessage(data=datadef)
        return request

    def send_feedback_rest(self, request):
        logging.info("Feedback called")

    def send_feedback_grpc(self, request):
        logging.info("Feedback called")


class UserObjectLowLevelGrpc:
    def __init__(self, metrics_ok=True, ret_nparray=False):
        self.metrics_ok = metrics_ok
        self.ret_nparray = ret_nparray
        self.nparray = np.array([1, 2, 3])

    def route_grpc(self, request):
        arr = np.array([1])
        datadef = prediction_pb2.DefaultData(
            tensor=prediction_pb2.Tensor(shape=(1, 1), values=arr)
        )
        request = prediction_pb2.SeldonMessage(data=datadef)
        return request

    def send_feedback_grpc(self, request):
        logging.info("Feedback called")


class UserObjectLowLevelRaw:
    def __init__(self, metrics_ok=True, ret_nparray=False):
        self.metrics_ok = metrics_ok
        self.ret_nparray = ret_nparray
        self.nparray = np.array([1, 2, 3])

    def route_raw(
        self, request: Union[prediction_pb2.SeldonMessage, List, Dict]
    ) -> Union[prediction_pb2.SeldonMessage, List, Dict]:

        is_proto = isinstance(request, prediction_pb2.SeldonMessage)

        arr = np.array([1])
        datadef = prediction_pb2.DefaultData(
            tensor=prediction_pb2.Tensor(shape=(1, 1), values=arr)
        )
        response = prediction_pb2.SeldonMessage(data=datadef)
        if is_proto:
            return response
        else:
            return seldon_message_to_json(response)

    def send_feedback_raw(self, request):
        logging.info("Feedback called")


class UserObjectBad:
    pass


def test_router_ok():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/route?json={"data":{"ndarray":[2]}}')
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]
    assert j["data"]["ndarray"] == [[22]]


def test_router_gets_meta():
    user_object = UserObject(ret_meta=True)
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/route?json={"meta":{"puid": "abc"}, "data":{"ndarray":[2]}}')
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["meta"]["tags"] == {"inc_meta": {"puid": "abc"}}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]


def test_router_meta_to_nonmeta_model():
    user_object = MinimalUserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/route?json={"meta":{"puid": "abc"}, "data":{"ndarray":[2]}}')
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["data"]["ndarray"] == [[22]]


def test_router_bad_user_object():
    user_object = UserObjectBad()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/route?json={"data":{"ndarray":[2]}}')
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 400
    assert j["status"]["info"] == "Route not defined"


def test_router_lowlevel_ok():
    user_object = UserObjectLowLevel()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/route?json={"data":{"ndarray":[2]}}')
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["data"]["ndarray"] == [[1]]


def test_router_lowlevel_raw_ok():
    user_object = UserObjectLowLevelRaw()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/route?json={"data":{"ndarray":[2]}}')
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["data"]["tensor"]["values"] == [1]


def test_router_no_json():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    uo = UserObject()
    rv = client.get("/route?")
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 400


def test_router_bad_metrics():
    user_object = UserObject(metrics_ok=False)
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/route?json={"data":{"ndarray":[]}}')
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 500


def test_router_feedback_ok():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get(
        '/send-feedback?json={"request":{"data":{"ndarray":[]}},"response":{"meta":{"routing":{"1":1}}},"reward":1.0}'
    )
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200


def test_router_feedback_lowlevel_ok():
    user_object = UserObjectLowLevel()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get(
        '/send-feedback?json={"request":{"data":{"ndarray":[]}},"reward":1.0}'
    )
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200


def test_router_proto_ok():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = SeldonModelGRPC(user_object, seldon_metrics)
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
    )
    request = prediction_pb2.SeldonMessage(data=datadef)
    resp = app.Route(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    logging.info(j)
    assert j["meta"]["tags"] == {"mytag": 1}
    # add default type
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]
    assert j["data"]["tensor"]["shape"] == [1, 1]
    assert j["data"]["tensor"]["values"] == [22]


def test_router_proto_lowlevel_ok():
    user_object = UserObjectLowLevelGrpc()
    seldon_metrics = SeldonMetrics()
    app = SeldonModelGRPC(user_object, seldon_metrics)
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
    )
    request = prediction_pb2.SeldonMessage(data=datadef)
    resp = app.Route(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    logging.info(j)
    assert j["data"]["tensor"]["shape"] == [1, 1]
    assert j["data"]["tensor"]["values"] == [1]


def test_router_proto_lowlevel_raw_ok():
    user_object = UserObjectLowLevelRaw()
    seldon_metrics = SeldonMetrics()
    app = SeldonModelGRPC(user_object, seldon_metrics)
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
    )
    request = prediction_pb2.SeldonMessage(data=datadef)
    resp = app.Route(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    logging.info(j)
    assert j["data"]["tensor"]["shape"] == [1, 1]
    assert j["data"]["tensor"]["values"] == [1]


def test_proto_feedback():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = SeldonModelGRPC(user_object, seldon_metrics)
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
    )
    meta = prediction_pb2.Meta()
    metaJson = {}
    routing = {"1": 1}
    metaJson["routing"] = routing
    json_format.ParseDict(metaJson, meta)

    request = prediction_pb2.SeldonMessage(data=datadef)
    response = prediction_pb2.SeldonMessage(meta=meta, data=datadef)
    feedback = prediction_pb2.Feedback(request=request, response=response, reward=1.0)
    resp = app.SendFeedback(feedback, None)


def test_get_grpc_server():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    server = get_grpc_server(user_object, seldon_metrics)


def test_unimplemented_route_raw_on_seldon_component():
    class CustomSeldonComponent(SeldonComponent):
        def route(self, X, features_names):
            return 53

    user_object = CustomSeldonComponent()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/route?json={"data":{"ndarray":[2]}}')
    j = json.loads(rv.data)

    logging.info(j)
    assert rv.status_code == 200
    assert j["data"]["ndarray"] == [[53]]


def test_unimplemented_route_raw():
    class CustomObject:
        def route(self, X, features_names):
            return 53

    user_object = CustomObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/route?json={"data":{"ndarray":[2]}}')
    j = json.loads(rv.data)

    logging.info(j)
    assert rv.status_code == 200
    assert j["data"]["ndarray"] == [[53]]
