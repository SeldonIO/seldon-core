import pytest
import json
import numpy as np
import pickle
import tensorflow as tf
import base64
from seldon_core.proto import prediction_pb2
from seldon_core.flask_utils import SeldonMicroserviceException
import seldon_core.utils as scu
from google.protobuf.struct_pb2 import Value


class UserObject(object):
    def __init__(self, metrics_ok=True, ret_nparray=False, ret_meta=False):
        self.metrics_ok = metrics_ok
        self.ret_nparray = ret_nparray
        self.nparray = np.array([1, 2, 3])
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
        else:
            print("Predict called - will run identity function")
            print(X)
            return X

    def feedback(self, features, feature_names, reward, truth):
        print("Feedback called")

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


def test_create_reponse_nparray():
    user_model = UserObject()
    request = prediction_pb2.SeldonMessage()
    raw_response = np.array([[1, 2, 3]])
    sm = scu.construct_response(user_model, True, request, raw_response)
    assert sm.data.WhichOneof("data_oneof") == "tensor"
    assert sm.data.tensor.values == [1, 2, 3]


def test_create_reponse_ndarray():
    user_model = UserObject()
    request_data = np.array([[5, 6, 7]])
    datadef = scu.array_to_grpc_datadef("ndarray", request_data)
    request = prediction_pb2.SeldonMessage(data=datadef)
    raw_response = np.array([[1, 2, 3]])
    sm = scu.construct_response(user_model, True, request, raw_response)
    assert sm.data.WhichOneof("data_oneof") == "ndarray"


def test_create_reponse_tensor():
    user_model = UserObject()
    request_data = np.array([[5, 6, 7]])
    datadef = scu.array_to_grpc_datadef("tensor", request_data)
    request = prediction_pb2.SeldonMessage(data=datadef)
    raw_response = np.array([[1, 2, 3]])
    sm = scu.construct_response(user_model, True, request, raw_response)
    assert sm.data.WhichOneof("data_oneof") == "tensor"


def test_create_response_strdata():
    user_model = UserObject()
    request_data = np.array([[5, 6, 7]])
    datadef = scu.array_to_grpc_datadef("ndarray", request_data)
    request = prediction_pb2.SeldonMessage(data=datadef)
    raw_response = "hello world"
    sm = scu.construct_response(user_model, True, request, raw_response)
    assert sm.data.WhichOneof("data_oneof") == None
    assert len(sm.strData) > 0


def test_create_response_jsondata():
    user_model = UserObject()
    request_data = np.array([[5, 6, 7]])
    datadef = scu.array_to_grpc_datadef("ndarray", request_data)
    request = prediction_pb2.SeldonMessage(data=datadef)
    raw_response = {"output": "data"}
    sm = scu.construct_response(user_model, True, request, raw_response)
    assert sm.data.WhichOneof("data_oneof") == None
    emptyValue = Value()
    assert sm.jsonData != emptyValue


def test_create_reponse_list():
    user_model = UserObject()
    request_data = np.array([[5, 6, 7]])
    datadef = scu.array_to_grpc_datadef("tensor", request_data)
    request = prediction_pb2.SeldonMessage(data=datadef)
    raw_response = ["one", "two", "three"]
    sm = scu.construct_response(user_model, True, request, raw_response)
    assert sm.data.WhichOneof("data_oneof") == "ndarray"


def test_create_reponse_binary():
    user_model = UserObject()
    request_data = np.array([[5, 6, 7]])
    datadef = scu.array_to_grpc_datadef("tensor", request_data)
    request = prediction_pb2.SeldonMessage(data=datadef)
    raw_response = b"binary"
    sm = scu.construct_response(user_model, True, request, raw_response)
    assert sm.data.WhichOneof("data_oneof") == None
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


def test_json_to_seldon_message_ndarry():
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
    bdata_base64 = base64.b64encode(serialized).decode('utf-8')
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
    data = {"jsonData": {"some": "value"}}
    requestProto = scu.json_to_seldon_message(data)
    assert len(requestProto.data.tensor.values) == 0
    assert requestProto.WhichOneof("data_oneof") == "jsonData"
    (json_data, meta, datadef, _) = scu.extract_request_parts(requestProto)
    assert not isinstance(json_data, np.ndarray)
    assert json_data == {"some": "value"}


def test_json_to_seldon_message_bad_data():
    with pytest.raises(SeldonMicroserviceException):
        data = {"foo": "bar"}
        requestProto = scu.json_to_seldon_message(data)


def test_json_to_feedback():
    data = {"request": {"data": {"tensor": {"shape": [1, 1], "values": [1]}}},
            "response": {"data": {"tensor": {"shape": [1, 1], "values": [2]}}}, "reward": 1.0}
    requestProto = scu.json_to_feedback(data)
    assert requestProto.request.data.tensor.values == [1.0]
    assert requestProto.response.data.tensor.values == [2.0]


def test_json_to_feedback_bad_data():
    with pytest.raises(SeldonMicroserviceException):
        data = {"requestBAD": {"data": {"tensor": {"shape": [1, 1], "values": [1]}}},
                "response": {"data": {"tensor": {"shape": [1, 1], "values": [2]}}}, "reward": 1.0}
        requestProto = scu.json_to_feedback(data)


def test_json_to_seldon_messages():
    data = {"seldonMessages": [{"data": {"tensor": {"shape": [1, 1], "values": [1]}}},
                               {"data": {"tensor": {"shape": [1, 1], "values": [2]}}}]}
    requestProto = scu.json_to_seldon_messages(data)
    assert requestProto.seldonMessages[0].data.tensor.values == [1]
    assert requestProto.seldonMessages[1].data.tensor.values == [2]
    assert len(requestProto.seldonMessages) == 2


def test_seldon_message_to_json():
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(
            shape=(2, 1),
            values=arr
        )
    )
    request = prediction_pb2.SeldonMessage(data=datadef)
    dict = scu.seldon_message_to_json(request)
    assert dict["data"]["tensor"]["values"] == [1, 2]


def test_get_data_from_proto_tensor():
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(
            shape=(2, 1),
            values=arr
        )
    )
    request = prediction_pb2.SeldonMessage(data=datadef)
    arr: np.ndarray = scu.get_data_from_proto(request)
    assert arr.shape == (2, 1)
    assert arr[0][0] == 1
    assert arr[1][0] == 2


def test_get_data_from_proto_ndarray():
    arr = np.array([[1], [2]])
    lv = scu.array_to_list_value(arr)
    datadef = prediction_pb2.DefaultData(
        ndarray=lv
    )
    request = prediction_pb2.SeldonMessage(data=datadef)
    arr: np.ndarray = scu.get_data_from_proto(request)
    assert arr.shape == (2, 1)
    assert arr[0][0] == 1
    assert arr[1][0] == 2


def test_get_data_from_proto_tftensor():
    arr = np.array([[1], [2]])
    datadef = prediction_pb2.DefaultData(
        tftensor=tf.make_tensor_proto(arr)
    )
    request = prediction_pb2.SeldonMessage(data=datadef)
    arr: np.ndarray = scu.get_data_from_proto(request)
    assert arr.shape == (2, 1)
    assert arr[0][0] == 1
    assert arr[1][0] == 2


def test_proto_array_to_tftensor():
    arr = np.array([[1, 2, 3], [4, 5, 6]])
    datadef = scu.array_to_grpc_datadef("tftensor", arr, [])
    print(datadef)
    assert datadef.tftensor.tensor_shape.dim[0].size == 2
    assert datadef.tftensor.tensor_shape.dim[1].size == 3
    assert datadef.tftensor.dtype == 9


def test_proto_tftensor_to_array():
    names = ["a", "b"]
    array = np.array([[1, 2], [3, 4]])
    datadef = prediction_pb2.DefaultData(
        names=names,
        tftensor=tf.make_tensor_proto(array)
    )
    array2 = scu.grpc_datadef_to_array(datadef)
    assert array.shape == array2.shape
    assert np.array_equal(array, array2)
