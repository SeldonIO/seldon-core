import base64
import json
import logging
from typing import Dict, List, Union

import numpy as np
import pytest
from google.protobuf import json_format

from seldon_core.metrics import SeldonMetrics
from seldon_core.proto import prediction_pb2
from seldon_core.user_model import SeldonComponent
from seldon_core.utils import seldon_message_to_json
from seldon_core.wrapper import SeldonModelGRPC, get_grpc_server, get_rest_microservice


class UserObject:
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
            logging.info("Transform input called - will run identity function")
            logging.info(X)
            return X

    def transform_output(self, X, features_names, **kwargs):
        if self.ret_meta:
            self.inc_meta = kwargs.get("meta")
        if self.ret_nparray:
            return self.nparray
        else:
            logging.info("Transform output called - will run identity function")
            logging.info(X)
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


class UserObjectLowLevel:
    def __init__(self, metrics_ok=True, ret_nparray=False):
        self.metrics_ok = metrics_ok
        self.ret_nparray = ret_nparray
        self.nparray = np.array([1, 2, 3])

    def transform_input_rest(self, X):
        return {"data": {"ndarray": [9, 9]}}

    def transform_output_rest(self, X):
        return {"data": {"ndarray": [9, 9]}}

    def transform_input_grpc(self, X):
        arr = np.array([9, 9])
        datadef = prediction_pb2.DefaultData(
            tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
        )
        request = prediction_pb2.SeldonMessage(data=datadef)
        return request

    def transform_output_grpc(self, X):
        arr = np.array([9, 9])
        datadef = prediction_pb2.DefaultData(
            tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
        )
        request = prediction_pb2.SeldonMessage(data=datadef)
        return request


class UserObjectLowLevelGrpc:
    def __init__(self, metrics_ok=True, ret_nparray=False):
        self.metrics_ok = metrics_ok
        self.ret_nparray = ret_nparray
        self.nparray = np.array([1, 2, 3])

    def transform_input_grpc(self, X):
        arr = np.array([9, 9])
        datadef = prediction_pb2.DefaultData(
            tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
        )
        request = prediction_pb2.SeldonMessage(data=datadef)
        return request

    def transform_output_grpc(self, X):
        arr = np.array([9, 9])
        datadef = prediction_pb2.DefaultData(
            tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
        )
        request = prediction_pb2.SeldonMessage(data=datadef)
        return request


class UserObjectLowLevelRaw:
    def __init__(self, metrics_ok=True, ret_nparray=False):
        self.metrics_ok = metrics_ok
        self.ret_nparray = ret_nparray
        self.nparray = np.array([1, 2, 3])

    def transform_input_raw(
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

    def transform_output_raw(
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


class UserObjectLowLevelRawInherited(SeldonComponent):
    def __init__(self, metrics_ok=True, ret_nparray=False):
        self.metrics_ok = metrics_ok
        self.ret_nparray = ret_nparray
        self.nparray = np.array([1, 2, 3])

    def transform_input_raw(
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

    def transform_output_raw(
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


def test_transformer_input_ok():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/transform-input?json={"data":{"ndarray":[1]}}')
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]
    assert j["data"]["ndarray"] == [1]


def test_transformer_input_lowlevel_ok():
    user_object = UserObjectLowLevel()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/transform-input?json={"data":{"ndarray":[1]}}')
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["data"]["ndarray"] == [9, 9]


def test_transformer_input_lowlevel_raw_ok():
    user_object = UserObjectLowLevelRaw()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/transform-input?json={"data":{"ndarray":[1]}}')
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["data"]["tensor"]["values"] == [9, 9]


def test_transformer_input_lowlevel_raw_ingerited_ok():
    user_object = UserObjectLowLevelRawInherited()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/transform-input?json={"data":{"ndarray":[1]}}')
    logging.info(rv.data)
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["data"]["tensor"]["values"] == [9, 9]


def test_transformer_input_bin_data():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    bdata = b"123"
    bdata_base64 = base64.b64encode(bdata).decode("utf-8")
    rv = client.get('/transform-input?json={"binData":"' + bdata_base64 + '"}')
    j = json.loads(rv.data)
    sm = prediction_pb2.SeldonMessage()
    # Check we can parse response
    assert sm == json_format.Parse(rv.data, sm, ignore_unknown_fields=False)
    logging.info(j)
    assert rv.status_code == 200
    assert "binData" in j
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]


def test_transformer_input_bin_data_nparray():
    user_object = UserObject(ret_nparray=True)
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    bdata = b"123"
    bdata_base64 = base64.b64encode(bdata).decode("utf-8")
    rv = client.get('/transform-input?json={"binData":"' + bdata_base64 + '"}')
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["data"]["tensor"]["values"] == [1, 2, 3]
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]


def test_transform_input_no_json():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    uo = UserObject()
    rv = client.get("/transform-input?")
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 400


def test_transform_input_bad_metrics():
    user_object = UserObject(metrics_ok=False)
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/transform-input?json={"data":{"ndarray":[]}}')
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 500


def test_transform_input_gets_meta():
    user_object = UserObject(ret_meta=True)
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get(
        '/transform-input?json={"meta":{"puid": "abc"},"data":{"ndarray":[]}}'
    )
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["meta"]["tags"] == {"inc_meta": {"puid": "abc"}}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]


def test_transform_output_gets_meta():
    user_object = UserObject(ret_meta=True)
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get(
        '/transform-output?json={"meta":{"puid": "abc"},"data":{"ndarray":[]}}'
    )
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["meta"]["tags"] == {"inc_meta": {"puid": "abc"}}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]


def test_transform_input_passes_through_tags():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get(
        '/transform-input?json={"meta":{"tags":{"foo":"bar"}},"data":{"ndarray":[]}}'
    )
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["meta"]["tags"] == {"foo": "bar", "mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]


def test_transform_output_passes_through_tags():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get(
        '/transform-output?json={"meta":{"tags":{"foo":"bar"}},"data":{"ndarray":[]}}'
    )
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["meta"]["tags"] == {"foo": "bar", "mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]


def test_transform_input_passes_through_metrics():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get(
        '/transform-input?json={"meta":{"metrics":[{"key":"request_gauge","type":"GAUGE","value":100}]},"data":{"ndarray":[]}}'
    )
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == "request_gauge"
    assert j["meta"]["metrics"][0]["value"] == 100

    assert j["meta"]["metrics"][1]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][1]["value"] == user_object.metrics()[0]["value"]


def test_transform_output_passes_through_metrics():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get(
        '/transform-output?json={"meta":{"metrics":[{"key":"request_gauge","type":"GAUGE","value":100}]},"data":{"ndarray":[]}}'
    )
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == "request_gauge"
    assert j["meta"]["metrics"][0]["value"] == 100

    assert j["meta"]["metrics"][1]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][1]["value"] == user_object.metrics()[0]["value"]


def test_transformer_output_ok():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/transform-output?json={"data":{"ndarray":[1]}}')
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]
    assert j["data"]["ndarray"] == [1]


def test_transformer_output_lowlevel_ok():
    user_object = UserObjectLowLevel()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/transform-output?json={"data":{"ndarray":[1]}}')
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["data"]["ndarray"] == [9, 9]


def test_transformer_output_lowlevel_raw_ok():
    user_object = UserObjectLowLevelRaw()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/transform-output?json={"data":{"ndarray":[1]}}')
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["data"]["tensor"]["values"] == [9, 9]


def test_transformer_output_lowlevel_raw_inherited_ok():
    user_object = UserObjectLowLevelRawInherited()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/transform-output?json={"data":{"ndarray":[1]}}')
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["data"]["tensor"]["values"] == [9, 9]


def test_transformer_output_bin_data():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    bdata = b"123"
    bdata_base64 = base64.b64encode(bdata).decode("utf-8")
    rv = client.get('/transform-output?json={"binData":"' + bdata_base64 + '"}')
    j = json.loads(rv.data)
    sm = prediction_pb2.SeldonMessage()
    # Check we can parse response
    assert sm == json_format.Parse(rv.data, sm, ignore_unknown_fields=False)
    logging.info(j)
    assert rv.status_code == 200
    assert "binData" in j
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]


def test_transformer_output_bin_data_nparray():
    user_object = UserObject(ret_nparray=True)
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    bdata = b"123"
    bdata_base64 = base64.b64encode(bdata).decode("utf-8")
    rv = client.get('/transform-output?json={"binData":"' + bdata_base64 + '"}')
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["data"]["tensor"]["values"] == [1, 2, 3]
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]


def test_transform_output_no_json():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    uo = UserObject()
    rv = client.get("/transform-output?")
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 400


def test_transform_output_bad_metrics():
    user_object = UserObject(metrics_ok=False)
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/transform-output?json={"data":{"ndarray":[]}}')
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 500


def test_transform_input_proto_ok():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = SeldonModelGRPC(user_object, seldon_metrics)
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
    )
    request = prediction_pb2.SeldonMessage(data=datadef)
    resp = app.TransformInput(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    logging.info(j)
    assert j["meta"]["tags"] == {"mytag": 1}
    # add default type
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]
    assert j["data"]["tensor"]["shape"] == [2, 1]
    assert j["data"]["tensor"]["values"] == [1, 2]


def test_transform_input_proto_lowlevel_ok():
    user_object = UserObjectLowLevelGrpc()
    seldon_metrics = SeldonMetrics()
    app = SeldonModelGRPC(user_object, seldon_metrics)
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
    )
    request = prediction_pb2.SeldonMessage(data=datadef)
    resp = app.TransformInput(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    logging.info(j)
    assert j["data"]["tensor"]["shape"] == [2, 1]
    assert j["data"]["tensor"]["values"] == [9, 9]


def test_transform_input_proto_bin_data():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = SeldonModelGRPC(user_object, seldon_metrics)
    binData = b"\0\1"
    request = prediction_pb2.SeldonMessage(binData=binData)
    resp = app.TransformInput(request, None)
    assert resp.binData == binData


def test_transform_input_proto_bin_data_nparray():
    user_object = UserObject(ret_nparray=True)
    seldon_metrics = SeldonMetrics()
    app = SeldonModelGRPC(user_object, seldon_metrics)
    binData = b"\0\1"
    request = prediction_pb2.SeldonMessage(binData=binData)
    resp = app.TransformInput(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    logging.info(j)
    assert j["data"]["tensor"]["values"] == list(user_object.nparray.flatten())


def test_transform_output_proto_ok():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = SeldonModelGRPC(user_object, seldon_metrics)
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
    )
    request = prediction_pb2.SeldonMessage(data=datadef)
    resp = app.TransformOutput(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    logging.info(j)
    assert j["meta"]["tags"] == {"mytag": 1}
    # add default type
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]
    assert j["data"]["tensor"]["shape"] == [2, 1]
    assert j["data"]["tensor"]["values"] == [1, 2]


def test_transform_output_proto_lowlevel_ok():
    user_object = UserObjectLowLevelGrpc()
    seldon_metrics = SeldonMetrics()
    app = SeldonModelGRPC(user_object, seldon_metrics)
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
    )
    request = prediction_pb2.SeldonMessage(data=datadef)
    resp = app.TransformOutput(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    logging.info(j)
    assert j["data"]["tensor"]["shape"] == [2, 1]
    assert j["data"]["tensor"]["values"] == [9, 9]


def test_transform_output_proto_bin_data():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = SeldonModelGRPC(user_object, seldon_metrics)
    binData = b"\0\1"
    request = prediction_pb2.SeldonMessage(binData=binData)
    resp = app.TransformOutput(request, None)
    assert resp.binData == binData


def test_transform_output_proto_bin_data_nparray():
    user_object = UserObject(ret_nparray=True)
    seldon_metrics = SeldonMetrics()
    app = SeldonModelGRPC(user_object, seldon_metrics)
    binData = b"\0\1"
    request = prediction_pb2.SeldonMessage(binData=binData)
    resp = app.TransformOutput(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    logging.info(j)
    assert j["data"]["tensor"]["values"] == list(user_object.nparray.flatten())


def test_get_grpc_server():
    user_object = UserObject(ret_nparray=True)
    seldon_metrics = SeldonMetrics()
    server = get_grpc_server(user_object, seldon_metrics)


def test_transform_input_proto_gets_meta():
    user_object = UserObject(ret_meta=True)
    seldon_metrics = SeldonMetrics()
    app = SeldonModelGRPC(user_object, seldon_metrics)
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
    )
    meta = prediction_pb2.Meta()
    metaJson = {"puid": "abc"}
    json_format.ParseDict(metaJson, meta)
    request = prediction_pb2.SeldonMessage(data=datadef, meta=meta)
    resp = app.TransformInput(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    logging.info(j)
    assert j["meta"]["tags"] == {"inc_meta": {"puid": "abc"}}
    # add default type
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]
    assert j["data"]["tensor"]["shape"] == [2, 1]
    assert j["data"]["tensor"]["values"] == [1, 2]


def test_transform_output_proto_gets_meta():
    user_object = UserObject(ret_meta=True)
    seldon_metrics = SeldonMetrics()
    app = SeldonModelGRPC(user_object, seldon_metrics)
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
    )
    meta = prediction_pb2.Meta()
    metaJson = {"puid": "abc"}
    json_format.ParseDict(metaJson, meta)
    request = prediction_pb2.SeldonMessage(data=datadef, meta=meta)
    resp = app.TransformOutput(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    logging.info(j)
    assert j["meta"]["tags"] == {"inc_meta": {"puid": "abc"}}
    # add default type
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]
    assert j["data"]["tensor"]["shape"] == [2, 1]
    assert j["data"]["tensor"]["values"] == [1, 2]


def test_transform_proto_input_passes_through_tags():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = SeldonModelGRPC(user_object, seldon_metrics)
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
    )
    meta = prediction_pb2.Meta()
    json_format.ParseDict({"tags": {"foo": "bar"}}, meta)
    request = prediction_pb2.SeldonMessage(data=datadef, meta=meta)
    resp = app.TransformInput(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    logging.info(j)
    assert j["meta"]["tags"] == {"foo": "bar", "mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]
    assert j["data"]["tensor"]["shape"] == [2, 1]
    assert j["data"]["tensor"]["values"] == [1, 2]


def test_transform_proto_output_passes_through_tags():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = SeldonModelGRPC(user_object, seldon_metrics)
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
    )
    meta = prediction_pb2.Meta()
    json_format.ParseDict({"tags": {"foo": "bar"}}, meta)
    request = prediction_pb2.SeldonMessage(data=datadef, meta=meta)
    resp = app.TransformOutput(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    logging.info(j)
    assert j["meta"]["tags"] == {"foo": "bar", "mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]
    assert j["data"]["tensor"]["shape"] == [2, 1]
    assert j["data"]["tensor"]["values"] == [1, 2]


def test_transform_proto_input_passes_through_metrics():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = SeldonModelGRPC(user_object, seldon_metrics)
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
    )
    meta = prediction_pb2.Meta()
    json_format.ParseDict(
        {"metrics": [{"key": "request_gauge", "type": "GAUGE", "value": 100}]}, meta
    )
    request = prediction_pb2.SeldonMessage(data=datadef, meta=meta)
    resp = app.TransformInput(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    logging.info(j)
    assert j["meta"]["metrics"][0]["key"] == "request_gauge"
    assert j["meta"]["metrics"][0]["value"] == 100

    assert j["meta"]["metrics"][1]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][1]["value"] == user_object.metrics()[0]["value"]

    assert j["data"]["tensor"]["shape"] == [2, 1]
    assert j["data"]["tensor"]["values"] == [1, 2]


def test_transform_proto_output_passes_through_metrics():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = SeldonModelGRPC(user_object, seldon_metrics)
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
    )
    meta = prediction_pb2.Meta()
    json_format.ParseDict(
        {"metrics": [{"key": "request_gauge", "type": "GAUGE", "value": 100}]}, meta
    )
    request = prediction_pb2.SeldonMessage(data=datadef, meta=meta)
    resp = app.TransformOutput(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    logging.info(j)
    assert j["meta"]["metrics"][0]["key"] == "request_gauge"
    assert j["meta"]["metrics"][0]["value"] == 100

    assert j["meta"]["metrics"][1]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][1]["value"] == user_object.metrics()[0]["value"]

    assert j["data"]["tensor"]["shape"] == [2, 1]
    assert j["data"]["tensor"]["values"] == [1, 2]


def test_unimplemented_transform_input_raw_on_seldon_component():
    class CustomSeldonComponent(SeldonComponent):
        def transform_input(self, X, features_names, **kwargs):
            return X * 2

    user_object = CustomSeldonComponent()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/transform-input?json={"data":{"ndarray":[1]}}')
    j = json.loads(rv.data)

    logging.info(j)
    assert rv.status_code == 200
    assert j["data"]["ndarray"] == [2.0]


def test_unimplemented_transform_input_raw():
    class CustomObject:
        def transform_input(self, X, features_names, **kwargs):
            return X * 2

    user_object = CustomObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/transform-input?json={"data":{"ndarray":[1]}}')
    j = json.loads(rv.data)

    logging.info(j)
    assert rv.status_code == 200
    assert j["data"]["ndarray"] == [2.0]


def test_unimplemented_transform_output_raw_on_seldon_component():
    class CustomSeldonComponent(SeldonComponent):
        def transform_output(self, X, features_names, **kwargs):
            return X * 2

    user_object = CustomSeldonComponent()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/transform-output?json={"data":{"ndarray":[1]}}')
    j = json.loads(rv.data)

    logging.info(j)
    assert rv.status_code == 200
    assert j["data"]["ndarray"] == [2.0]


def test_unimplemented_transform_output_raw():
    class CustomObject:
        def transform_output(self, X, features_names, **kwargs):
            return X * 2

    user_object = CustomObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/transform-output?json={"data":{"ndarray":[1]}}')
    j = json.loads(rv.data)

    logging.info(j)
    assert rv.status_code == 200
    assert j["data"]["ndarray"] == [2.0]
