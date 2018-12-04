import pytest
import json
import numpy as np
import pickle
import tensorflow as tf
from google.protobuf import json_format
from tensorflow.core.framework.tensor_pb2 import TensorProto

from seldon_core.proto import prediction_pb2
from seldon_core.microservice import get_data_from_json, array_to_grpc_datadef, grpc_datadef_to_array, rest_datadef_to_array, array_to_rest_datadef
from seldon_core.microservice import SeldonMicroserviceException


def test_normal_data():
    data = {"data": {"tensor": {"shape": [1, 1], "values": [1]}}}
    arr = get_data_from_json(data)
    assert isinstance(arr, np.ndarray)
    assert arr.shape[0] == 1
    assert arr.shape[1] == 1
    assert arr[0][0] == 1


def test_bin_data():
    a = np.array([1, 2, 3])
    serialized = pickle.dumps(a)
    data = {"binData": serialized}
    arr = get_data_from_json(data)
    assert not isinstance(arr, np.ndarray)
    assert arr == serialized


def test_str_data():
    data = {"strData": "my string data"}
    arr = get_data_from_json(data)
    assert not isinstance(arr, np.ndarray)
    assert arr == "my string data"


def test_bad_data():
    with pytest.raises(SeldonMicroserviceException):
        data = {"foo": "bar"}
        arr = get_data_from_json(data)


def test_proto_array_to_tftensor():
    arr = np.array([[1, 2, 3], [4, 5, 6]])
    datadef = array_to_grpc_datadef(arr, [], "tftensor")
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
    array2 = grpc_datadef_to_array(datadef)
    assert array.shape == array2.shape
    assert np.array_equal(array, array2)


def test_json_tftensor_to_array():
    names = ["a", "b"]
    array = np.array([[1, 2], [3, 4]])
    datadef = prediction_pb2.DefaultData(
        names=names,
        tftensor=tf.make_tensor_proto(array)
    )
    jStr = json_format.MessageToJson(datadef)
    j = json.loads(jStr)
    array2 = rest_datadef_to_array(j)
    assert np.array_equal(array, array2)


def test_json_array_to_tftensor():
    array = np.array([[1, 2], [3, 4]])
    original_datadef = {"tftensor": {}}
    datadef = array_to_rest_datadef(array, [], original_datadef)
    assert "tftensor" in datadef
    tfp = TensorProto()
    json_format.ParseDict(datadef.get("tftensor"), tfp,
                          ignore_unknown_fields=False)
    array2 = tf.make_ndarray(tfp)
    assert np.array_equal(array, array2)
