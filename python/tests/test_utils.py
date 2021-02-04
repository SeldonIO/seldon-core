import base64
import logging
import pickle

import numpy as np
import pytest
from google.protobuf import any_pb2
from google.protobuf.struct_pb2 import Value

import seldon_core.utils as scu
from seldon_core.flask_utils import SeldonMicroserviceException
from seldon_core.imports_helper import _TF_PRESENT
from seldon_core.proto import prediction_pb2

from .utils import skipif_tf_missing

if _TF_PRESENT:
    import tensorflow as tf


class UserObject:
    def __init__(
        self, metrics_ok=True, ret_nparray=False, ret_meta=False, ret_dict=False
    ):
        self.metrics_ok = metrics_ok
        self.ret_nparray = ret_nparray
        self.nparray = np.array([1, 2, 3])
        self.dict = {"output": "data"}
        self.ret_meta = ret_meta

    def predict(self, X, features_names, **kwargs):
        """
        Return a prediction.

        Parameters
        ----------
        X : array-like
        feature_names : array of feature names (optional)
        """
        if self.ret_meta:
            self.inc_meta = kwargs.get("meta")
        if self.ret_nparray:
            return self.nparray
        elif self.ret_dict:
            return self.dict
        else:
            logging.info("Predict called - will run identity function")
            logging.info(X)
            return X

    def feedback(self, features, feature_names, reward, truth):
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


def test_create_rest_reponse_nparray():
    user_model = UserObject()
    request = {}
    raw_response = np.array([[1, 2, 3]])
    result = scu.construct_response_json(user_model, True, request, raw_response)
    assert "tensor" in result.get("data", {})
    assert result["data"]["tensor"]["values"] == [1, 2, 3]


def test_create_grpc_reponse_nparray():
    user_model = UserObject()
    request = prediction_pb2.SeldonMessage()
    raw_response = np.array([[1, 2, 3]])
    sm = scu.construct_response(user_model, True, request, raw_response)
    assert sm.data.WhichOneof("data_oneof") == "tensor"
    assert sm.data.tensor.values == [1, 2, 3]


def test_create_rest_reponse_text_ndarray():
    user_model = UserObject()
    request_data = np.array([["hello", "world"], ["hello", "another", "world"]])
    request = {"data": {"ndarray": request_data, "names": []}}
    (features, meta, datadef, data_type) = scu.extract_request_parts_json(request)
    raw_response = np.array([["hello", "world"], ["here", "another"]])
    result = scu.construct_response_json(user_model, True, request, raw_response)
    assert "ndarray" in result.get("data", {})
    assert np.array_equal(result["data"]["ndarray"], raw_response)
    assert datadef == request["data"]
    assert np.array_equal(features, request_data)
    assert data_type == "data"


def test_create_grpc_reponse_text_ndarray():
    user_model = UserObject()
    request_data = np.array([["hello", "world"], ["hello", "another", "world"]])
    datadef = scu.array_to_grpc_datadef("ndarray", request_data)
    request = prediction_pb2.SeldonMessage(data=datadef)
    (features, meta, datadef, data_type) = scu.extract_request_parts(request)
    raw_response = np.array([["hello", "world"], ["here", "another"]])
    sm = scu.construct_response(user_model, True, request, raw_response)
    assert sm.data.WhichOneof("data_oneof") == "ndarray"
    assert type(features[0]) == list
    assert np.array_equal(sm.data.ndarray, raw_response)
    assert datadef == request.data
    assert np.array_equal(features, request_data)
    assert data_type == "data"


def test_create_rest_reponse_ndarray():
    user_model = UserObject()
    request = {"data": {"ndarray": np.array([[5, 6, 7]]), "names": []}}
    raw_response = np.array([[1, 2, 3]])
    result = scu.construct_response_json(user_model, True, request, raw_response)
    assert "ndarray" in result.get("data", {})
    assert np.array_equal(result["data"]["ndarray"], raw_response)


def test_create_grpc_reponse_ndarray():
    user_model = UserObject()
    request_data = np.array([[5, 6, 7]])
    datadef = scu.array_to_grpc_datadef("ndarray", request_data)
    request = prediction_pb2.SeldonMessage(data=datadef)
    raw_response = np.array([[1, 2, 3]])
    sm = scu.construct_response(user_model, True, request, raw_response)
    assert sm.data.WhichOneof("data_oneof") == "ndarray"


def test_create_rest_reponse_tensor():
    user_model = UserObject()
    tensor = {"values": [1, 2, 3], "shape": (3,)}
    request = {"data": {"tensor": tensor, "names": []}}
    raw_response = np.array([1, 2, 3])
    result = scu.construct_response_json(user_model, True, request, raw_response)
    assert "tensor" in result.get("data", {})
    assert np.array_equal(result["data"]["tensor"], tensor)


def test_create_grpc_reponse_tensor():
    user_model = UserObject()
    request_data = np.array([[5, 6, 7]])
    datadef = scu.array_to_grpc_datadef("tensor", request_data)
    request = prediction_pb2.SeldonMessage(data=datadef)
    raw_response = np.array([[1, 2, 3]])
    sm = scu.construct_response(user_model, True, request, raw_response)
    assert sm.data.WhichOneof("data_oneof") == "tensor"


def test_create_rest_response_strdata():
    user_model = UserObject()
    request_data = "Request data"
    request = {"strData": request_data}
    raw_response = "hello world"
    sm = scu.construct_response_json(user_model, True, request, raw_response)
    assert "strData" in sm
    assert len(sm["strData"]) > 0
    assert sm["strData"] == raw_response


def test_create_grpc_response_strdata():
    user_model = UserObject()
    request_data = np.array([[5, 6, 7]])
    datadef = scu.array_to_grpc_datadef("ndarray", request_data)
    request = prediction_pb2.SeldonMessage(data=datadef)
    raw_response = "hello world"
    sm = scu.construct_response(user_model, True, request, raw_response)
    assert sm.data.WhichOneof("data_oneof") is None
    assert len(sm.strData) > 0


def test_create_grpc_response_jsondata():
    user_model = UserObject()
    request_data = np.array([[5, 6, 7]])
    datadef = scu.array_to_grpc_datadef("ndarray", request_data)
    request = prediction_pb2.SeldonMessage(data=datadef)
    raw_response = {"output": "data"}
    sm = scu.construct_response(user_model, True, request, raw_response)
    assert sm.data.WhichOneof("data_oneof") is None
    emptyValue = Value()
    assert sm.jsonData != emptyValue


def test_create_grpc_response_customdata():
    user_model = UserObject()
    request_data = np.array([[5, 6, 7]])
    datadef = scu.array_to_grpc_datadef("ndarray", request_data)
    request = prediction_pb2.SeldonMessage(data=datadef)
    raw_response = any_pb2.Any(value=b"testdata")
    sm = scu.construct_response(user_model, True, request, raw_response)
    assert sm.data.WhichOneof("data_oneof") is None
    emptyValue = Value()
    assert sm.customData != emptyValue


def test_create_rest_response_jsondata():
    user_model = UserObject()
    request_data = np.array([[5, 6, 7]])
    datadef = scu.array_to_rest_datadef("ndarray", request_data)
    json_request = {"jsonData": datadef}
    raw_response = {"output": "data"}
    json_response = scu.construct_response_json(
        user_model, True, json_request, raw_response
    )
    assert "data" not in json_response
    emptyValue = Value()
    assert json_response["jsonData"] != emptyValue


def test_create_rest_response_jsondata_with_array_input():
    user_model = UserObject(ret_dict=True)
    request_data = np.array([[5, 6, 7]])
    datadef = scu.array_to_rest_datadef("ndarray", request_data)
    json_request = {"data": datadef}
    raw_response = {"output": "data"}
    json_response = scu.construct_response_json(
        user_model, True, json_request, raw_response
    )
    assert "data" not in json_response
    assert json_response["jsonData"] == user_model.dict


def test_symmetric_json_conversion():
    request_data = np.array([[5, 6, 7]])
    datadef = scu.array_to_rest_datadef("ndarray", request_data)
    json_request = {"jsonData": datadef}
    seldon_message_request = scu.json_to_seldon_message(json_request)
    result_json_request = scu.seldon_message_to_json(seldon_message_request)
    assert json_request == result_json_request


def test_create_grpc_reponse_list():
    user_model = UserObject()
    request_data = np.array([[5, 6, 7]])
    datadef = scu.array_to_grpc_datadef("tensor", request_data)
    request = prediction_pb2.SeldonMessage(data=datadef)
    raw_response = ["one", "two", "three"]
    sm = scu.construct_response(user_model, True, request, raw_response)
    assert sm.data.WhichOneof("data_oneof") == "ndarray"


def test_create_rest_reponse_binary():
    user_model = UserObject()
    request_data = b"input"
    request = {"binData": request_data}
    raw_resp = b"binary"
    sm = scu.construct_response_json(user_model, True, request, raw_resp)
    resp_data = base64.b64encode(raw_resp).decode("utf-8")
    assert "strData" not in sm
    assert "binData" in sm
    assert sm["binData"] == resp_data


def test_create_grpc_reponse_binary():
    user_model = UserObject()
    request_data = np.array([[5, 6, 7]])
    datadef = scu.array_to_grpc_datadef("tensor", request_data)
    request = prediction_pb2.SeldonMessage(data=datadef)
    raw_response = b"binary"
    sm = scu.construct_response(user_model, True, request, raw_response)
    assert sm.data.WhichOneof("data_oneof") is None
    assert len(sm.strData) == 0
    assert len(sm.binData) > 0


def test_json_to_seldon_message_normal_data():
    data = {"data": {"tensor": {"shape": [1, 1], "values": [1]}}}
    requestProto = scu.json_to_seldon_message(data)
    assert requestProto.data.tensor.values == [1]
    assert requestProto.data.tensor.shape[0] == 1
    assert requestProto.data.tensor.shape[1] == 1
    assert len(requestProto.data.tensor.shape) == 2
    (arr, meta, datadef, _) = scu.extract_request_parts(requestProto)
    assert isinstance(arr, np.ndarray)
    assert arr.shape[0] == 1
    assert arr.shape[1] == 1
    assert arr[0][0] == 1


def test_json_to_seldon_message_ndarray():
    data = {"data": {"ndarray": [[1]]}}
    requestProto = scu.json_to_seldon_message(data)
    assert requestProto.data.ndarray[0][0] == 1
    (arr, meta, datadef, _) = scu.extract_request_parts(requestProto)
    assert isinstance(arr, np.ndarray)
    assert arr.shape[0] == 1
    assert arr.shape[1] == 1
    assert arr[0][0] == 1


def test_json_to_seldon_message_bin_data():
    a = np.array([1, 2, 3])
    serialized = pickle.dumps(a)
    bdata_base64 = base64.b64encode(serialized).decode("utf-8")
    data = {"binData": bdata_base64}
    requestProto = scu.json_to_seldon_message(data)
    assert len(requestProto.data.tensor.values) == 0
    assert requestProto.WhichOneof("data_oneof") == "binData"
    assert len(requestProto.binData) > 0
    (arr, meta, datadef, _) = scu.extract_request_parts(requestProto)
    assert not isinstance(arr, np.ndarray)
    assert arr == serialized


def test_json_to_seldon_message_str_data():
    data = {"strData": "my string data"}
    requestProto = scu.json_to_seldon_message(data)
    assert len(requestProto.data.tensor.values) == 0
    assert requestProto.WhichOneof("data_oneof") == "strData"
    (arr, meta, datadef, _) = scu.extract_request_parts(requestProto)
    assert not isinstance(arr, np.ndarray)
    assert arr == "my string data"


def test_json_to_seldon_message_json_data():
    json_data = {"jsonData": {"some": "value"}}
    (json_data, meta, datadef, _) = scu.extract_request_parts_json(json_data)
    assert not isinstance(json_data, np.ndarray)
    assert json_data == {"some": "value"}


def test_json_to_seldon_message_bad_data():
    with pytest.raises(SeldonMicroserviceException):
        data = {"foo": "bar"}
        scu.json_to_seldon_message(data)


def test_json_to_feedback():
    data = {
        "request": {"data": {"tensor": {"shape": [1, 1], "values": [1]}}},
        "response": {"data": {"tensor": {"shape": [1, 1], "values": [2]}}},
        "reward": 1.0,
    }
    requestProto = scu.json_to_feedback(data)
    assert requestProto.request.data.tensor.values == [1.0]
    assert requestProto.response.data.tensor.values == [2.0]


def test_json_to_feedback_bad_data():
    with pytest.raises(SeldonMicroserviceException):
        data = {
            "requestBAD": {"data": {"tensor": {"shape": [1, 1], "values": [1]}}},
            "response": {"data": {"tensor": {"shape": [1, 1], "values": [2]}}},
            "reward": 1.0,
        }
        scu.json_to_feedback(data)


def test_json_to_seldon_messages():
    data = {
        "seldonMessages": [
            {"data": {"tensor": {"shape": [1, 1], "values": [1]}}},
            {"data": {"tensor": {"shape": [1, 1], "values": [2]}}},
        ]
    }
    requestProto = scu.json_to_seldon_messages(data)
    assert requestProto.seldonMessages[0].data.tensor.values == [1]
    assert requestProto.seldonMessages[1].data.tensor.values == [2]
    assert len(requestProto.seldonMessages) == 2


def test_seldon_message_to_json():
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
    )
    request = prediction_pb2.SeldonMessage(data=datadef)
    dict = scu.seldon_message_to_json(request)
    assert dict["data"]["tensor"]["values"] == [1, 2]


def test_get_data_from_proto_tensor():
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
    )
    request = prediction_pb2.SeldonMessage(data=datadef)
    arr: np.ndarray = scu.get_data_from_proto(request)
    assert arr.shape == (2, 1)
    assert arr[0][0] == 1
    assert arr[1][0] == 2


def test_get_data_from_proto_ndarray():
    arr = np.array([[1], [2]])
    lv = scu.array_to_list_value(arr)
    datadef = prediction_pb2.DefaultData(ndarray=lv)
    request = prediction_pb2.SeldonMessage(data=datadef)
    arr: np.ndarray = scu.get_data_from_proto(request)
    assert arr.shape == (2, 1)
    assert arr[0][0] == 1
    assert arr[1][0] == 2


@skipif_tf_missing
def test_get_data_from_proto_tftensor():
    arr = np.array([[1], [2]])
    datadef = prediction_pb2.DefaultData(tftensor=tf.make_tensor_proto(arr))
    request = prediction_pb2.SeldonMessage(data=datadef)
    arr: np.ndarray = scu.get_data_from_proto(request)
    assert arr.shape == (2, 1)
    assert arr[0][0] == 1
    assert arr[1][0] == 2


@skipif_tf_missing
def test_proto_array_to_tftensor():
    arr = np.array([[1, 2, 3], [4, 5, 6]])
    datadef = scu.array_to_grpc_datadef("tftensor", arr, [])
    logging.info(datadef)
    assert datadef.tftensor.tensor_shape.dim[0].size == 2
    assert datadef.tftensor.tensor_shape.dim[1].size == 3
    assert datadef.tftensor.dtype == 9


@skipif_tf_missing
def test_proto_tftensor_to_array():
    names = ["a", "b"]
    array = np.array([[1, 2], [3, 4]])
    datadef = prediction_pb2.DefaultData(
        names=names, tftensor=tf.make_tensor_proto(array)
    )
    array2 = scu.grpc_datadef_to_array(datadef)
    assert array.shape == array2.shape
    assert np.array_equal(array, array2)


@pytest.mark.parametrize(
    "env,expected",
    [
        ({"FOO1": "BAR1"}, "BAR1"),
        ({"FOO2": "BAR2"}, "BAR2"),
        ({"FOO3": "BAR3"}, "BAR3"),
        ({"FOO1": "BAR1", "FOO2": "BAR2"}, "BAR1"),
        ({}, "DEF"),
    ],
)
def test_getenv(monkeypatch, env, expected):
    for env_var, env_value in env.items():
        monkeypatch.setenv(env_var, env_value)

    value = scu.getenv("FOO1", "FOO2", "FOO3", default="DEF")
    assert value == expected


@pytest.mark.parametrize(
    "env_val,expected",
    [
        ("TRUE", True),
        ("true", True),
        ("t", True),
        ("1", True),
        ("FALSE", False),
        ("false", False),
        ("f", False),
        ("0", False),
        (None, False),
    ],
)
def test_getenv_as_bool(monkeypatch, env_val, expected):
    env_var = "MY_BOOL_VAR"

    if env_val is not None:
        monkeypatch.setenv(env_var, env_val)

    value = scu.getenv_as_bool(env_var, default=False)
    assert value == expected
