import pytest
import json
import numpy as np
from google.protobuf import json_format
import base64

from seldon_core.wrapper import get_rest_microservice, SeldonModelGRPC, get_grpc_server
from seldon_core.proto import prediction_pb2
from seldon_core.utils import seldon_message_to_json
from seldon_core.user_model import SeldonComponent

from typing import List, Dict, Union


class UserObject(object):
    def __init__(self, metrics_ok=True, ret_nparray=False):
        self.metrics_ok = metrics_ok
        self.ret_nparray = ret_nparray
        self.nparray = np.array([1, 2, 3])

    def aggregate(self, Xs, features_names):
        if self.ret_nparray:
            return self.nparray
        else:
            print("Aggregate input called - will return first item")
            print(Xs)
            return Xs[0]

    def tags(self):
        return {"mytag": 1}

    def metrics(self):
        if self.metrics_ok:
            return [{"type": "COUNTER", "key": "mycounter", "value": 1}]
        else:
            return [{"type": "BAD", "key": "mycounter", "value": 1}]


class UserObjectLowLevel(object):
    def __init__(self, metrics_ok=True, ret_nparray=False):
        self.metrics_ok = metrics_ok
        self.ret_nparray = ret_nparray
        self.nparray = np.array([1, 2, 3])

    def aggregate_rest(self, Xs):
        return {"data": {"ndarray": [9, 9]}}

    def aggregate_grpc(
        self, request: Union[prediction_pb2.SeldonMessage, List, Dict]
    ) -> Union[prediction_pb2.SeldonMessage, List, Dict]:

        is_proto = isinstance(request, prediction_pb2.SeldonMessage)

        arr = np.array([9, 9])
        datadef = prediction_pb2.DefaultData(
            tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
        )
        response = prediction_pb2.SeldonMessage(data=datadef)
        if is_proto:
            return response
        else:
            return seldon_message_to_json(response)


class UserObjectLowLevelGrpc(object):
    def __init__(self, metrics_ok=True, ret_nparray=False):
        self.metrics_ok = metrics_ok
        self.ret_nparray = ret_nparray
        self.nparray = np.array([1, 2, 3])

    def aggregate_grpc(self, X):
        arr = np.array([9, 9])
        datadef = prediction_pb2.DefaultData(
            tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
        )
        request = prediction_pb2.SeldonMessage(data=datadef)
        return request


class UserObjectBad(object):
    pass


def test_aggreate_ok():
    user_object = UserObject()
    app = get_rest_microservice(user_object)
    client = app.test_client()
    rv = client.get('/aggregate?json={"seldonMessages":[{"data":{"ndarray":[1]}}]}')
    print(rv)
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 200
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]
    assert j["data"]["ndarray"] == [1]


def test_aggreate_bad_user_object():
    user_object = UserObjectBad()
    app = get_rest_microservice(user_object)
    client = app.test_client()
    rv = client.get('/aggregate?json={"seldonMessages":[{"data":{"ndarray":[1]}}]}')
    print(rv)
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 400
    assert j["status"]["info"] == "Aggregate not defined"


def test_aggreate_invalid_message():
    user_object = UserObject()
    app = get_rest_microservice(user_object)
    client = app.test_client()
    rv = client.get('/aggregate?json={"wrong":[{"data":{"ndarray":[1]}}]}')
    assert rv.status_code == 400
    j = json.loads(rv.data)
    print(j)
    assert j["status"]["reason"] == "MICROSERVICE_BAD_DATA"


def test_aggreate_no_list():
    user_object = UserObject()
    app = get_rest_microservice(user_object)
    client = app.test_client()
    rv = client.get('/aggregate?json={"seldonMessages":{"data":{"ndarray":[1]}}}')
    assert rv.status_code == 400
    j = json.loads(rv.data)
    print(j)
    assert j["status"]["reason"] == "MICROSERVICE_BAD_DATA"


def test_aggreate_bad_messages():
    user_object = UserObject()
    app = get_rest_microservice(user_object)
    client = app.test_client()
    rv = client.get('/aggregate?json={"seldonMessages":[{"data2":{"ndarray":[1]}}]}')
    assert rv.status_code == 400
    j = json.loads(rv.data)
    print(j)
    assert j["status"]["reason"] == "MICROSERVICE_BAD_DATA"


def test_aggreate_ok_2messages():
    user_object = UserObject()
    app = get_rest_microservice(user_object)
    client = app.test_client()
    rv = client.get(
        '/aggregate?json={"seldonMessages":[{"data":{"ndarray":[1]}},{"data":{"ndarray":[2]}}]}'
    )
    print(rv)
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 200
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]
    assert j["data"]["ndarray"] == [1]


def test_aggreate_ok_bindata():
    user_object = UserObject()
    app = get_rest_microservice(user_object)
    client = app.test_client()
    bdata = b"123"
    bdata_base64 = base64.b64encode(bdata).decode("utf-8")
    rv = client.get(
        '/aggregate?json={"seldonMessages":[{"binData":"'
        + bdata_base64
        + '"},{"binData":"'
        + bdata_base64
        + '"}]}'
    )
    print(rv)
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 200
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]
    assert j["binData"] == bdata_base64


def test_aggreate_ok_strdata():
    user_object = UserObject()
    app = get_rest_microservice(user_object)
    client = app.test_client()
    rv = client.get(
        '/aggregate?json={"seldonMessages":[{"strData":"123"},{"strData":"456"}]}'
    )
    print(rv)
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 200
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]
    assert j["strData"] == "123"


def test_aggregate_bad_metrics():
    user_object = UserObject(metrics_ok=False)
    app = get_rest_microservice(user_object)
    client = app.test_client()
    rv = client.get(
        '/aggregate?json={"seldonMessages":[{"data":{"ndarray":[1]}},{"data":{"ndarray":[2]}}]}'
    )
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 400


def test_aggreate_ok_lowlevel():
    user_object = UserObjectLowLevel()
    app = get_rest_microservice(user_object)
    client = app.test_client()
    rv = client.get(
        '/aggregate?json={"seldonMessages":[{"data":{"ndarray":[1]}},{"data":{"ndarray":[2]}}]}'
    )
    print(rv)
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 200
    assert j["data"]["ndarray"] == [9, 9]


def test_aggregate_proto_ok():
    user_object = UserObject()
    app = SeldonModelGRPC(user_object)
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
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    print(j)
    assert j["meta"]["tags"] == {"mytag": 1}
    # add default type
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]
    assert j["data"]["tensor"]["shape"] == [2, 1]
    assert j["data"]["tensor"]["values"] == [1, 2]


def test_aggregate_proto_bin_data():
    user_object = UserObject()
    app = SeldonModelGRPC(user_object)
    binData = b"\0\1"
    msg1 = prediction_pb2.SeldonMessage(binData=binData)
    request = prediction_pb2.SeldonMessageList(seldonMessages=[msg1])
    resp = app.Aggregate(request, None)
    assert resp.binData == binData


def test_aggregate_proto_lowlevel_ok():
    user_object = UserObjectLowLevelGrpc()
    app = SeldonModelGRPC(user_object)
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
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    print(j)
    assert j["data"]["tensor"]["shape"] == [2, 1]
    assert j["data"]["tensor"]["values"] == [9, 9]


def test_get_grpc_server():
    user_object = UserObject(ret_nparray=True)
    server = get_grpc_server(user_object)


def test_unimplemented_aggregate_raw_on_seldon_component():
    class CustomSeldonComponent(SeldonComponent):
        def aggregate(self, Xs, features_names):
            return sum(Xs) * 2

    user_object = CustomSeldonComponent()
    app = get_rest_microservice(user_object)
    client = app.test_client()
    rv = client.get(
        '/aggregate?json={"seldonMessages":[{"data":{"ndarray":[1]}},{"data":{"ndarray":[2]}}]}'
    )
    j = json.loads(rv.data)

    print(j)
    assert rv.status_code == 200
    assert j["data"]["ndarray"] == [6.0]


def test_unimplemented_aggregate_raw():
    class CustomObject(object):
        def aggregate(self, Xs, features_names):
            return sum(Xs) * 2

    user_object = CustomObject()
    app = get_rest_microservice(user_object)
    client = app.test_client()
    rv = client.get(
        '/aggregate?json={"seldonMessages":[{"data":{"ndarray":[1]}},{"data":{"ndarray":[2]}}]}'
    )
    j = json.loads(rv.data)

    print(j)
    assert rv.status_code == 200
    assert j["data"]["ndarray"] == [6.0]
