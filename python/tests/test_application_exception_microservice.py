import base64
import json
import logging

import flask
import numpy as np
from flask import jsonify
from google.protobuf import json_format

from seldon_core.metrics import SeldonMetrics
from seldon_core.proto import prediction_pb2
from seldon_core.user_model import SeldonComponent
from seldon_core.wrapper import SeldonModelGRPC, get_grpc_server, get_rest_microservice


class UserCustomException(Exception):
    status_code = 404

    def __init__(self, message, application_error_code, http_status_code):
        Exception.__init__(self)
        self.message = message
        if http_status_code is not None:
            self.status_code = http_status_code
        self.application_error_code = application_error_code

    def to_dict(self):
        rv = {
            "status": {
                "status": self.status_code,
                "message": self.message,
                "app_code": self.application_error_code,
            }
        }
        return rv


class UserObject(SeldonComponent):

    model_error_handler = flask.Blueprint("error_handlers", __name__)

    @model_error_handler.app_errorhandler(UserCustomException)
    def handleCustomError(error):
        response = jsonify(error.to_dict())
        response.status_code = error.status_code
        return response

    def __init__(self, metrics_ok=True, ret_nparray=False, ret_meta=False):
        self.metrics_ok = metrics_ok
        self.ret_nparray = ret_nparray
        self.nparray = np.array([1, 2, 3])
        self.ret_meta = ret_meta

    def predict(self, X, features_names, **kwargs):
        raise UserCustomException("Test-Error-Msg", 1402, 402)
        return X


class UserObjectLowLevel(SeldonComponent):

    model_error_handler = flask.Blueprint("error_handlers", __name__)

    @model_error_handler.app_errorhandler(UserCustomException)
    def handleCustomError(error):
        response = jsonify(error.to_dict())
        response.status_code = error.status_code
        return response

    def __init__(self, metrics_ok=True, ret_nparray=False):
        self.metrics_ok = metrics_ok
        self.ret_nparray = ret_nparray
        self.nparray = np.array([1, 2, 3])

    def predict_rest(self, request):
        raise UserCustomException("Test-Error-Msg", 1402, 402)
        return {"data": {"ndarray": [9, 9]}}


def test_raise_exception():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/predict?json={"data":{"names":["a","b"],"ndarray":[[1,2]]}}')
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 402
    assert j["status"]["app_code"] == 1402


def test_raise_eception_lowlevel():
    user_object = UserObjectLowLevel()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/predict?json={"data":{"ndarray":[1,2]}}')
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 402
    assert j["status"]["app_code"] == 1402
