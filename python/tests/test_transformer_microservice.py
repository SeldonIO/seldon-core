import pytest
import json
import numpy as np
from google.protobuf import json_format
import base64

from seldon_core.transformer_microservice import get_rest_microservice, SeldonTransformerGRPC, get_grpc_server
from seldon_core.proto import prediction_pb2


class UserObject(object):
    def __init__(self, metrics_ok=True, ret_nparray=False, ret_meta=False):
        self.metrics_ok = metrics_ok
        self.ret_nparray = ret_nparray
        self.nparray = np.array([1, 2, 3])
        self.ret_meta = ret_meta

    def transform_input(self, X, features_names, **kwargs):
        if self.ret_meta:
            self.inc_meta = kwargs.get("meta")
        if self.ret_nparray:
            return self.nparray
        else:
            print("Transform input called - will run identity function")
            print(X)
            return X

    def transform_output(self, X, features_names, **kwargs):
        if self.ret_meta:
            self.inc_meta = kwargs.get("meta")
        if self.ret_nparray:
            return self.nparray
        else:
            print("Transform output called - will run identity function")
            print(X)
        return X

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

class UserObjectLowLevel(object):
    def __init__(self, metrics_ok=True, ret_nparray=False):
        self.metrics_ok = metrics_ok
        self.ret_nparray = ret_nparray
        self.nparray = np.array([1, 2, 3])

    def transform_input_rest(self, X):
        return {"data":{"ndarray":[9,9]}}

    def transform_output_rest(self, X):
        return {"data":{"ndarray":[9,9]}}

    def transform_input_grpc(self, X):
        arr = np.array([9, 9])
        datadef = prediction_pb2.DefaultData(
            tensor=prediction_pb2.Tensor(
                shape=(2, 1),
                values=arr
            )
        )
        request = prediction_pb2.SeldonMessage(data=datadef)
        return request

    def transform_output_grpc(self, X):
        arr = np.array([9, 9])
        datadef = prediction_pb2.DefaultData(
            tensor=prediction_pb2.Tensor(
                shape=(2, 1),
                values=arr
            )
        )
        request = prediction_pb2.SeldonMessage(data=datadef)
        return request


def test_transformer_input_ok():
    user_object = UserObject()
    app = get_rest_microservice(user_object, debug=True)
    client = app.test_client()
    rv = client.get('/transform-input?json={"data":{"ndarray":[1]}}')
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 200
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"] == user_object.metrics()
    assert j["data"]["ndarray"] == [1]

def test_transformer_input_lowlevel_ok():
    user_object = UserObjectLowLevel()
    app = get_rest_microservice(user_object, debug=True)
    client = app.test_client()
    rv = client.get('/transform-input?json={"data":{"ndarray":[1]}}')
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 200
    assert j["data"]["ndarray"] == [9, 9]


def test_transformer_input_bin_data():
    user_object = UserObject()
    app = get_rest_microservice(user_object, debug=True)
    client = app.test_client()
    bdata = b"123"
    bdata_base64 = base64.b64encode(bdata).decode('utf-8')
    rv = client.get('/transform-input?json={"binData":"' + bdata_base64 + '"}')
    j = json.loads(rv.data)
    sm = prediction_pb2.SeldonMessage()
    # Check we can parse response
    assert sm == json_format.Parse(rv.data, sm, ignore_unknown_fields=False)
    print(j)
    assert rv.status_code == 200
    assert "binData" in j
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"] == user_object.metrics()


def test_transformer_input_bin_data_nparray():
    user_object = UserObject(ret_nparray=True)
    app = get_rest_microservice(user_object, debug=True)
    client = app.test_client()
    rv = client.get('/transform-input?json={"binData":"123"}')
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 200
    assert j["data"]["ndarray"] == [1, 2, 3]
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"] == user_object.metrics()


def test_tranform_input_no_json():
    user_object = UserObject()
    app = get_rest_microservice(user_object, debug=True)
    client = app.test_client()
    uo = UserObject()
    rv = client.get('/transform-input?')
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 400


def test_transform_input_bad_metrics():
    user_object = UserObject(metrics_ok=False)
    app = get_rest_microservice(user_object, debug=True)
    client = app.test_client()
    rv = client.get('/transform-input?json={"data":{"ndarray":[]}}')
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 400


def test_transform_input_gets_meta():
    user_object = UserObject(ret_meta=True)
    app = get_rest_microservice(user_object, debug=True)
    client = app.test_client()
    rv = client.get('/transform-input?json={"meta":{"puid": "abc"},"data":{"ndarray":[]}}')
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 200
    assert j["meta"]["tags"] == {"inc_meta":{"puid": "abc"}}
    assert j["meta"]["metrics"] == user_object.metrics()


def test_transform_output_gets_meta():
    user_object = UserObject(ret_meta=True)
    app = get_rest_microservice(user_object, debug=True)
    client = app.test_client()
    rv = client.get('/transform-output?json={"meta":{"puid": "abc"},"data":{"ndarray":[]}}')
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 200
    assert j["meta"]["tags"] == {"inc_meta":{"puid": "abc"}}
    assert j["meta"]["metrics"] == user_object.metrics()


def test_transformer_output_ok():
    user_object = UserObject()
    app = get_rest_microservice(user_object, debug=True)
    client = app.test_client()
    rv = client.get('/transform-output?json={"data":{"ndarray":[1]}}')
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 200
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"] == user_object.metrics()
    assert j["data"]["ndarray"] == [1]


def test_transformer_output_lowlevel_ok():
    user_object = UserObjectLowLevel()
    app = get_rest_microservice(user_object, debug=True)
    client = app.test_client()
    rv = client.get('/transform-output?json={"data":{"ndarray":[1]}}')
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 200
    assert j["data"]["ndarray"] == [9, 9]


def test_transformer_output_bin_data():
    user_object = UserObject()
    app = get_rest_microservice(user_object, debug=True)
    client = app.test_client()
    bdata = b"123"
    bdata_base64 = base64.b64encode(bdata).decode('utf-8')
    rv = client.get(
        '/transform-output?json={"binData":"' + bdata_base64 + '"}')
    j = json.loads(rv.data)
    sm = prediction_pb2.SeldonMessage()
    # Check we can parse response
    assert sm == json_format.Parse(rv.data, sm, ignore_unknown_fields=False)
    print(j)
    assert rv.status_code == 200
    assert "binData" in j
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"] == user_object.metrics()


def test_transformer_output_bin_data_nparray():
    user_object = UserObject(ret_nparray=True)
    app = get_rest_microservice(user_object, debug=True)
    client = app.test_client()
    rv = client.get('/transform-output?json={"binData":"123"}')
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 200
    assert j["data"]["ndarray"] == [1, 2, 3]
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"] == user_object.metrics()


def test_tranform_output_no_json():
    user_object = UserObject()
    app = get_rest_microservice(user_object, debug=True)
    client = app.test_client()
    uo = UserObject()
    rv = client.get('/transform-output?')
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 400


def test_transform_output_bad_metrics():
    user_object = UserObject(metrics_ok=False)
    app = get_rest_microservice(user_object, debug=True)
    client = app.test_client()
    rv = client.get('/transform-output?json={"data":{"ndarray":[]}}')
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 400


def test_transform_input_proto_ok():
    user_object = UserObject()
    app = SeldonTransformerGRPC(user_object)
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(
            shape=(2, 1),
            values=arr
        )
    )
    request = prediction_pb2.SeldonMessage(data=datadef)
    resp = app.TransformInput(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    print(j)
    assert j["meta"]["tags"] == {"mytag": 1}
    # add default type
    j["meta"]["metrics"][0]["type"] = "COUNTER"
    assert j["meta"]["metrics"] == user_object.metrics()
    assert j["data"]["tensor"]["shape"] == [2, 1]
    assert j["data"]["tensor"]["values"] == [1, 2]

def test_transform_input_proto_lowlevel_ok():
    user_object = UserObjectLowLevel()
    app = SeldonTransformerGRPC(user_object)
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(
            shape=(2, 1),
            values=arr
        )
    )
    request = prediction_pb2.SeldonMessage(data=datadef)
    resp = app.TransformInput(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    print(j)
    assert j["data"]["tensor"]["shape"] == [2, 1]
    assert j["data"]["tensor"]["values"] == [9, 9]



def test_transform_input_proto_bin_data():
    user_object = UserObject()
    app = SeldonTransformerGRPC(user_object)
    binData = b"\0\1"
    request = prediction_pb2.SeldonMessage(binData=binData)
    resp = app.TransformInput(request, None)
    assert resp.binData == binData


def test_transform_input_proto_bin_data_nparray():
    user_object = UserObject(ret_nparray=True)
    app = SeldonTransformerGRPC(user_object)
    binData = b"\0\1"
    request = prediction_pb2.SeldonMessage(binData=binData)
    resp = app.TransformInput(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    print(j)
    assert j["data"]["tensor"]["values"] == list(user_object.nparray.flatten())


def test_transform_output_proto_ok():
    user_object = UserObject()
    app = SeldonTransformerGRPC(user_object)
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(
            shape=(2, 1),
            values=arr
        )
    )
    request = prediction_pb2.SeldonMessage(data=datadef)
    resp = app.TransformOutput(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    print(j)
    assert j["meta"]["tags"] == {"mytag": 1}
    # add default type
    j["meta"]["metrics"][0]["type"] = "COUNTER"
    assert j["meta"]["metrics"] == user_object.metrics()
    assert j["data"]["tensor"]["shape"] == [2, 1]
    assert j["data"]["tensor"]["values"] == [1, 2]

def test_transform_output_proto_lowlevel_ok():
    user_object = UserObjectLowLevel()
    app = SeldonTransformerGRPC(user_object)
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(
            shape=(2, 1),
            values=arr
        )
    )
    request = prediction_pb2.SeldonMessage(data=datadef)
    resp = app.TransformOutput(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    print(j)
    assert j["data"]["tensor"]["shape"] == [2, 1]
    assert j["data"]["tensor"]["values"] == [9, 9]


def test_transform_output_proto_bin_data():
    user_object = UserObject()
    app = SeldonTransformerGRPC(user_object)
    binData = b"\0\1"
    request = prediction_pb2.SeldonMessage(binData=binData)
    resp = app.TransformOutput(request, None)
    assert resp.binData == binData


def test_transform_output_proto_bin_data_nparray():
    user_object = UserObject(ret_nparray=True)
    app = SeldonTransformerGRPC(user_object)
    binData = b"\0\1"
    request = prediction_pb2.SeldonMessage(binData=binData)
    resp = app.TransformOutput(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    print(j)
    assert j["data"]["tensor"]["values"] == list(user_object.nparray.flatten())


def test_get_grpc_server():
    user_object = UserObject(ret_nparray=True)
    server = get_grpc_server(user_object)


def test_transform_input_proto_gets_meta():
    user_object = UserObject(ret_meta=True)
    app = SeldonTransformerGRPC(user_object)
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(
            shape=(2, 1),
            values=arr
        )
    )
    meta = prediction_pb2.Meta()
    metaJson = {"puid":"abc"}
    json_format.ParseDict(metaJson, meta)
    request = prediction_pb2.SeldonMessage(data=datadef, meta=meta)
    resp = app.TransformInput(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    print(j)
    assert j["meta"]["tags"] == {"inc_meta":{"puid":"abc"}}
    # add default type
    j["meta"]["metrics"][0]["type"] = "COUNTER"
    assert j["meta"]["metrics"] == user_object.metrics()
    assert j["data"]["tensor"]["shape"] == [2, 1]
    assert j["data"]["tensor"]["values"] == [1, 2]


def test_transform_output_proto_gets_meta():
    user_object = UserObject(ret_meta=True)
    app = SeldonTransformerGRPC(user_object)
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(
            shape=(2, 1),
            values=arr
        )
    )
    meta = prediction_pb2.Meta()
    metaJson = {"puid":"abc"}
    json_format.ParseDict(metaJson, meta)
    request = prediction_pb2.SeldonMessage(data=datadef, meta=meta)
    resp = app.TransformOutput(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    print(j)
    assert j["meta"]["tags"] == {"inc_meta":{"puid":"abc"}}
    # add default type
    j["meta"]["metrics"][0]["type"] = "COUNTER"
    assert j["meta"]["metrics"] == user_object.metrics()
    assert j["data"]["tensor"]["shape"] == [2, 1]
    assert j["data"]["tensor"]["values"] == [1, 2]
