import logging
import os
import numpy as np

from seldon_core.api_tester import run_predict, run_send_feedback
from unittest import mock
from seldon_core.utils import array_to_grpc_datadef, seldon_message_to_json
from seldon_core.proto import prediction_pb2

from .conftest import RESOURCES_PATH


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


class Bunch:
    def __init__(self, adict):
        self.__dict__.update(adict)


@mock.patch("requests.post", side_effect=mocked_requests_post_success)
def test_predict_rest(mock_post):
    filename = os.path.join(RESOURCES_PATH, "model-template-app", "contract.json")
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
        "deployment": "abc",
        "namespace": None,
    }
    args = Bunch(args_dict)
    run_predict(args)
    logging.info(mock_post.call_args)
    assert mock_post.call_args[1]["json"]["data"]["names"] == [
        "sepal_length",
        "sepal_width",
        "petal_length",
        "petal_width",
    ]


@mock.patch("requests.post", side_effect=mocked_requests_post_success)
def test_feedback_rest(mock_post):
    filename = os.path.join(RESOURCES_PATH, "model-template-app", "contract.json")
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
        "deployment": "abc",
        "namespace": None,
    }
    args = Bunch(args_dict)
    run_send_feedback(args)
