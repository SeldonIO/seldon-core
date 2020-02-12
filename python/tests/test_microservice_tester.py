from seldon_core.microservice_tester import (
    run_method,
    run_send_feedback,
    reconciliate_cont_type,
    SeldonTesterException,
)
from unittest import mock
from seldon_core.utils import array_to_grpc_datadef, seldon_message_to_json
from seldon_core.proto import prediction_pb2
import numpy as np
from os.path import dirname, join
import pytest
import json
import logging


class MockResponse:
    def __init__(self, json_data, status_code, reason="", text=""):
        self.json_data = json_data
        self.status_code = status_code
        self.reason = reason
        self.text = text

    def json(self):
        return self.json_data


def mocked_requests_post_success(url, *args, **kwargs):
    data = np.random.rand(1, 1)
    datadef = array_to_grpc_datadef("tensor", data)
    request = prediction_pb2.SeldonMessage(data=datadef)
    json = seldon_message_to_json(request)
    return MockResponse(json, 200, text="{}")


class Bunch(object):
    def __init__(self, adict):
        self.__dict__.update(adict)


@mock.patch("requests.post", side_effect=mocked_requests_post_success)
def test_predict_rest(mock_post):
    filename = join(dirname(__file__), "model-template-app", "contract.json")
    args_dict = {
        "contract": filename,
        "host": "a",
        "port": 1000,
        "n_requests": 1,
        "batch_size": 1,
        "endpoint": "predict",
        "prnt": True,
        "grpc": False,
        "tensor": True,
    }
    args = Bunch(args_dict)
    run_method(args, "predict")
    logging.info(mock_post.call_args[1])
    payload = json.loads(mock_post.call_args[1]["data"]["json"])
    assert payload["data"]["names"] == [
        "sepal_length",
        "sepal_width",
        "petal_length",
        "petal_width",
    ]


@mock.patch("requests.post", side_effect=mocked_requests_post_success)
def test_feedback_rest(mock_post):
    filename = join(dirname(__file__), "model-template-app", "contract.json")
    args_dict = {
        "contract": filename,
        "host": "a",
        "port": 1000,
        "n_requests": 1,
        "batch_size": 1,
        "endpoint": "feedback",
        "prnt": True,
        "grpc": False,
        "tensor": True,
    }
    args = Bunch(args_dict)
    run_send_feedback(args)


@mock.patch("requests.post", side_effect=mocked_requests_post_success)
def test_predict_rest_categorical(mock_post):
    filename = join(dirname(__file__), "resources", "contract.json")
    args_dict = {
        "contract": filename,
        "host": "a",
        "port": 1000,
        "n_requests": 1,
        "batch_size": 1,
        "endpoint": "predict",
        "prnt": True,
        "grpc": False,
        "tensor": False,
    }
    args = Bunch(args_dict)
    run_method(args, "predict")


def test_reconciliate_exception():
    arr = np.array([1, 2])
    with pytest.raises(SeldonTesterException):
        reconciliate_cont_type(arr, "FOO")


def test_bad_contract():
    with pytest.raises(SeldonTesterException):
        filename = join(dirname(__file__), "resources", "bad_contract.json")
        args_dict = {
            "contract": filename,
            "host": "a",
            "port": 1000,
            "n_requests": 1,
            "batch_size": 1,
            "endpoint": "feedback",
            "prnt": True,
            "grpc": False,
            "tensor": True,
        }
        args = Bunch(args_dict)
        run_send_feedback(args)
