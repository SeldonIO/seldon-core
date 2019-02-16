import pytest
import json
import numpy as np
import pickle
import tensorflow as tf
from google.protobuf import json_format
from tensorflow.core.framework.tensor_pb2 import TensorProto
import base64
from seldon_core.proto import prediction_pb2
from seldon_core.microservice import SeldonMicroserviceException
import seldon_core.utils as scu


def test_normal_data():
    data = {"data": {"tensor": {"shape": [1, 1], "values": [1]}}}
    requestProto = scu.json_to_seldon_message(data)
    (arr, meta, datadef, _) = scu.extract_request_parts(requestProto)
    assert isinstance(arr, np.ndarray)
    assert arr.shape[0] == 1
    assert arr.shape[1] == 1
    assert arr[0][0] == 1


def test_bin_data():
    a = np.array([1, 2, 3])
    serialized = pickle.dumps(a)
    bdata_base64 = base64.b64encode(serialized).decode('utf-8')
    data = {"binData": bdata_base64}
    requestProto = scu.json_to_seldon_message(data)
    (arr, meta, datadef, _) = scu.extract_request_parts(requestProto)
    assert not isinstance(arr, np.ndarray)
    assert arr == serialized


def test_str_data():
    data = {"strData": "my string data"}
    requestProto = scu.json_to_seldon_message(data)
    (arr, meta, datadef, _) = scu.extract_request_parts(requestProto)
    assert not isinstance(arr, np.ndarray)
    assert arr == "my string data"


def test_bad_data():
    with pytest.raises(SeldonMicroserviceException):
        data = {"foo": "bar"}
        requestProto = scu.json_to_seldon_message(data)
        (arr, meta, datadef, _) = scu.extract_request_parts(requestProto)


def test_proto_array_to_tftensor():
    arr = np.array([[1, 2, 3], [4, 5, 6]])
    datadef = scu.array_to_grpc_datadef(arr, [], "tftensor")
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
