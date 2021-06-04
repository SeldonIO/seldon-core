import os
import json
from alibiexplainer.v2_http import KFServingV2RequestHandler
import numpy as np

TESTS_PATH = os.path.dirname(__file__)
TESTDATA_PATH = os.path.join(TESTS_PATH, "data")


def test_extract_request():
    truck_file = os.path.join(TESTDATA_PATH,"truck-v2.json")
    with open(truck_file,"r") as f:
        d = json.load(f)
        rh = KFServingV2RequestHandler()
        arr = np.array(rh.extract_request(d))
        assert arr.shape[0] == 1
        assert arr.shape[1] == 32
        assert arr.shape[2] == 32
        assert arr.shape[3] == 3


def test_create_request():
    truck_file = os.path.join(TESTDATA_PATH, "truck-v2.json")
    with open(truck_file, "r") as f:
        d = json.load(f)
        rh = KFServingV2RequestHandler()
        arr = np.array(rh.extract_request(d))
        name = rh.extract_name(d)
        ty = rh.extract_type(d)
        print(arr.dtype)
        request = rh.create_request(arr, name, ty)
        assert request["inputs"][0]["name"] == name
        assert request["inputs"][0]["datatype"] == ty


def test_extract_response():
    truck_file = os.path.join(TESTDATA_PATH, "truck-v2-response.json")
    with open(truck_file, "r") as f:
        d = json.load(f)
        rh = KFServingV2RequestHandler()
        arr = rh.extract_response(d)
        assert arr.shape[0] == 1
        assert arr.shape[1] == 10
