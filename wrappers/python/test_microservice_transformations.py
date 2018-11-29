import pytest
from microservice import get_data_from_json
from microservice import SeldonMicroserviceException
import json
import numpy as np
import pickle

def test_normal_data():
    data = {"data":{"tensor":{"shape":[1,1],"values":[1]}}}
    arr = get_data_from_json(data)
    assert isinstance(arr, np.ndarray)
    assert arr.shape[0] == 1
    assert arr.shape[1] == 1
    assert arr[0][0] == 1


def test_bin_data():
    a = np.array([1,2,3])
    serialized = pickle.dumps(a)
    data = {"binData" : serialized }
    arr = get_data_from_json(data)
    assert not isinstance(arr, np.ndarray)
    assert arr == serialized


def test_str_data():
    data = {"strData":"my string data"}
    arr = get_data_from_json(data)
    assert not isinstance(arr, np.ndarray)
    assert arr == "my string data"

def test_bad_data():
    with pytest.raises(SeldonMicroserviceException):
        data = {"foo":"bar"}
        arr = get_data_from_json(data)
    
