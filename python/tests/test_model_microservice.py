import base64
import io
import json
import logging
from unittest import mock

import numpy as np
from google.protobuf import json_format
from PIL import Image

from seldon_core.flask_utils import SeldonMicroserviceException
from seldon_core.imports_helper import _TF_PRESENT
from seldon_core.metrics import SeldonMetrics
from seldon_core.proto import prediction_pb2
from seldon_core.user_model import SeldonComponent
from seldon_core.utils import json_to_seldon_message, seldon_message_to_json
from seldon_core.wrapper import SeldonModelGRPC, get_grpc_server, get_rest_microservice

from .utils import skipif_tf_missing

if _TF_PRESENT:
    import tensorflow as tf
    from tensorflow.core.framework.tensor_pb2 import TensorProto

HEALTH_PING_URL = "/health/ping"
HEALTH_STATUS_URL = "/health/status"
METADATA_URL = "/metadata"

"""
 Checksum of bytes. Used to check data integrity of binData passed in multipart/form-data request

 Parameters
 ----------
  the_bytes
    Input bytes

  Returns
  -------
  the checksum
"""


def rs232_checksum(the_bytes):
    return b"%02X" % (sum(the_bytes) & 0xFF)


class UserObject(SeldonComponent):
    HEALTH_STATUS_REPONSE = [0.123]
    METADATA_RESPONSE = {
        "name": "my-model-name",
        "versions": ["model-version"],
        "platform": "model-platform",
        "inputs": [{"name": "input", "datatype": "BYTES", "shape": [1]}],
        "outputs": [{"name": "output", "datatype": "BYTES", "shape": [1]}],
        "custom": {},
    }

    def __init__(self, metrics_ok=True, ret_nparray=False, ret_meta=False):
        self.metrics_ok = metrics_ok
        self.ret_nparray = ret_nparray
        self.nparray = np.array([1, 2, 3])
        self.ret_meta = ret_meta

        self.rewards = []

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
            logging.info("Predict called - will run identity function")
            logging.info(X)
            return X

    def send_feedback(self, features, feature_names, reward, truth, routing=None):
        self.rewards.append(reward)
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

    def health_status(self):
        return self.predict(self.HEALTH_STATUS_REPONSE, ["some_float"])

    def init_metadata(self):
        return self.METADATA_RESPONSE


class UserObjectLowLevel(SeldonComponent):
    HEALTH_STATUS_RAW_RESPONSE = [123.456, 7.89]

    def __init__(self, metrics_ok=True, ret_nparray=False):
        self.metrics_ok = metrics_ok
        self.ret_nparray = ret_nparray
        self.nparray = np.array([1, 2, 3])

    def predict_rest(self, request):
        return {"data": {"ndarray": [9, 9]}}

    def predict_grpc(self, request):
        arr = np.array([9, 9])
        datadef = prediction_pb2.DefaultData(
            tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
        )
        request = prediction_pb2.SeldonMessage(data=datadef)
        return request

    def send_feedback_rest(self, request):
        logging.info("Feedback called")

    def send_feedback_grpc(self, request):
        logging.info("Feedback called")

    def health_status_raw(self):
        return {"data": {"ndarray": self.HEALTH_STATUS_RAW_RESPONSE}}


class UserObjectLowLevelWithStatusInResponse(SeldonComponent):
    def __init__(self, metrics_ok=True, ret_nparray=False):
        self.metrics_ok = metrics_ok
        self.ret_nparray = ret_nparray
        self.nparray = np.array([1, 2, 3])

    def predict_rest(self, request):
        return {
            "data": {"ndarray": [9, 9]},
            "status": {"code": 400, "status": "FAILURE"},
        }

    def predict_grpc(self, request):
        arr = np.array([9, 9])
        datadef = prediction_pb2.DefaultData(
            tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
        )
        request = prediction_pb2.SeldonMessage(data=datadef)
        return request

    def send_feedback_rest(self, request):
        logging.info("Feedback called")

    def send_feedback_grpc(self, request):
        logging.info("Feedback called")


class UserObjectLowLevelWithStatusInResponseWithPredictRaw(SeldonComponent):
    def __init__(self, check_name):
        self.check_name = check_name

    def predict_raw(self, msg):
        msg = json_to_seldon_message(msg)
        if self.check_name == "img":
            file_data = msg.binData
            img = Image.open(io.BytesIO(file_data))
            img.verify()
            return {
                "meta": seldon_message_to_json(msg.meta),
                "data": {"ndarray": [rs232_checksum(file_data).decode("utf-8")]},
                "status": {"code": 400, "status": "FAILURE"},
            }
        elif self.check_name == "txt":
            file_data = msg.binData
            return {
                "meta": seldon_message_to_json(msg.meta),
                "data": {"ndarray": [file_data.decode("utf-8")]},
                "status": {"code": 400, "status": "FAILURE"},
            }
        elif self.check_name == "strData":
            file_data = msg.strData
            return {
                "meta": seldon_message_to_json(msg.meta),
                "data": {"ndarray": [file_data]},
                "status": {"code": 400, "status": "FAILURE"},
            }


class UserObjectLowLevelWithPredictRaw(SeldonComponent):
    def __init__(self, check_name):
        self.check_name = check_name

    def predict_raw(self, msg):
        msg = json_to_seldon_message(msg)
        if self.check_name == "img":
            file_data = msg.binData
            img = Image.open(io.BytesIO(file_data))
            img.verify()
            return {
                "meta": seldon_message_to_json(msg.meta),
                "data": {"ndarray": [rs232_checksum(file_data).decode("utf-8")]},
            }
        elif self.check_name == "txt":
            file_data = msg.binData
            return {
                "meta": seldon_message_to_json(msg.meta),
                "data": {"ndarray": [file_data.decode("utf-8")]},
            }
        elif self.check_name == "strData":
            file_data = msg.strData
            return {
                "meta": seldon_message_to_json(msg.meta),
                "data": {"ndarray": [file_data]},
            }


class UserObjectLowLevelGrpc(SeldonComponent):
    def __init__(self, metrics_ok=True, ret_nparray=False):
        self.metrics_ok = metrics_ok
        self.ret_nparray = ret_nparray
        self.nparray = np.array([1, 2, 3])

    def predict_grpc(self, request):
        arr = np.array([9, 9])
        datadef = prediction_pb2.DefaultData(
            tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
        )
        request = prediction_pb2.SeldonMessage(data=datadef)
        return request

    def send_feedback_rest(self, request):
        logging.info("Feedback called")

    def send_feedback_grpc(self, request):
        logging.info("Feedback called")


def test_model_ok():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/predict?json={"data":{"names":["a","b"],"ndarray":[[1,2]]}}')
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]
    assert j["data"]["names"] == ["t:0", "t:1"]
    assert j["data"]["ndarray"] == [[1.0, 2.0]]


def test_model_v01_ok():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()

    payload = {"data": {"names": ["a", "b"], "ndarray": [[1, 2]]}}
    rv = client.post("/api/v0.1/predictions", json=payload)
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]
    assert j["data"]["names"] == ["t:0", "t:1"]
    assert j["data"]["ndarray"] == [[1.0, 2.0]]


def test_model_v10_ok():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()

    payload = {"data": {"names": ["a", "b"], "ndarray": [[1, 2]]}}
    rv = client.post("/api/v1.0/predictions", json=payload)
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]
    assert j["data"]["names"] == ["t:0", "t:1"]
    assert j["data"]["ndarray"] == [[1.0, 2.0]]


def test_model_puid_ok():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get(
        '/predict?json={"meta":{"puid":"123"},"data":{"names":["a","b"],"ndarray":[[1,2]]}}'
    )
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]
    assert j["data"]["names"] == ["t:0", "t:1"]
    assert j["data"]["ndarray"] == [[1.0, 2.0]]
    assert j["meta"]["puid"] == "123"


@mock.patch("seldon_core.utils.model_name", "my-test-model")
@mock.patch("seldon_core.utils.image_name", "my-test-model-image")
def test_requestPath_ok():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get(
        '/predict?json={"meta":{"puid":"123"},"data":{"names":["a","b"],"ndarray":[[1,2]]}}'
    )
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["meta"]["requestPath"] == {"my-test-model": "my-test-model-image"}


@mock.patch("seldon_core.utils.model_name", "my-test-model")
@mock.patch("seldon_core.utils.image_name", "my-test-model-image")
def test_requestPath_2nd_node_ok():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get(
        '/predict?json={"meta":{"requestPath":{"earlier-node": "earlier-image"}},"data":{"names":["a","b"],"ndarray":[[1,2]]}}'
    )
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["meta"]["requestPath"] == {
        "my-test-model": "my-test-model-image",
        "earlier-node": "earlier-image",
    }


@mock.patch("seldon_core.utils.model_name", "my-test-model")
@mock.patch("seldon_core.utils.image_name", "my-test-model-image")
def test_proto_requestPath_ok():
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
    resp = app.Predict(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    logging.info(j)
    assert j["meta"]["requestPath"] == {"my-test-model": "my-test-model-image"}


@mock.patch("seldon_core.utils.model_name", "my-test-model")
@mock.patch("seldon_core.utils.image_name", "my-test-model-image")
def test_proto_requestPath_2nd_node_ok():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = SeldonModelGRPC(user_object, seldon_metrics)
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
    )
    meta = prediction_pb2.Meta()
    json_format.ParseDict({"requestPath": {"earlier-node": "earlier-image"}}, meta)
    request = prediction_pb2.SeldonMessage(data=datadef, meta=meta)
    resp = app.Predict(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    logging.info(j)
    assert j["meta"]["requestPath"] == {
        "my-test-model": "my-test-model-image",
        "earlier-node": "earlier-image",
    }


def test_model_lowlevel_ok():
    user_object = UserObjectLowLevel()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/predict?json={"data":{"ndarray":[1,2]}}')
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["data"]["ndarray"] == [9, 9]


def test_model_lowlevel_multi_form_data_text_file_ok():
    user_object = UserObjectLowLevelWithPredictRaw("txt")
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.post(
        "/predict",
        data={
            "meta": '{"puid":"1234"}',
            "binData": (f"./tests/resources/test.txt", "test.txt"),
        },
        content_type="multipart/form-data",
    )
    j = json.loads(rv.data)
    assert rv.status_code == 200
    assert j["meta"]["puid"] == "1234"
    assert (
        j["data"]["ndarray"][0]
        == "this is test file for testing multipart/form-data input\n"
    )


def test_model_lowlevel_multi_form_data_img_file_ok():
    user_object = UserObjectLowLevelWithPredictRaw("img")
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.post(
        "/predict",
        data={
            "meta": '{"puid":"1234"}',
            "binData": (f"./tests/resources/test.png", "test.png"),
        },
        content_type="multipart/form-data",
    )
    j = json.loads(rv.data)
    assert rv.status_code == 200
    assert j["meta"]["puid"] == "1234"
    with open("./tests/resources/test.png", "rb") as f:
        img_data = f.read()
    assert j["data"]["ndarray"][0] == rs232_checksum(img_data).decode("utf-8")


def test_model_lowlevel_multi_form_data_strData_ok():
    user_object = UserObjectLowLevelWithPredictRaw("strData")
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.post(
        "/predict",
        data={
            "meta": '{"puid":"1234"}',
            "strData": (f"./tests/resources/test.txt", "test.txt"),
        },
        content_type="multipart/form-data",
    )
    j = json.loads(rv.data)
    assert rv.status_code == 200
    assert j["meta"]["puid"] == "1234"
    assert (
        j["data"]["ndarray"][0]
        == "this is test file for testing multipart/form-data input\n"
    )


def test_model_lowlevel_multi_form_data_strData_non200status():
    user_object = UserObjectLowLevelWithStatusInResponseWithPredictRaw("strData")
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.post(
        "/predict",
        data={
            "meta": '{"puid":"1234"}',
            "strData": (f"./tests/resources/test.txt", "test.txt"),
        },
        content_type="multipart/form-data",
    )
    j = json.loads(rv.data)
    assert rv.status_code == 400
    assert j["meta"]["puid"] == "1234"
    assert (
        j["data"]["ndarray"][0]
        == "this is test file for testing multipart/form-data input\n"
    )


def test_model_multi_form_data_ok():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.post(
        "/predict",
        data={"data": '{"names":["a","b"],"ndarray":[[1,2]]}'},
        content_type="multipart/form-data",
    )
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]
    assert j["data"]["names"] == ["t:0", "t:1"]
    assert j["data"]["ndarray"] == [[1.0, 2.0]]


def test_model_feedback_ok():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get(
        '/send-feedback?json={"request":{"data":{"ndarray":[]}},"reward":1.0}'
    )
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert user_object.rewards[-1] == 1.0


def test_feedback_v10_ok():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()

    payload = {
        "request": {"data": {"names": ["a", "b"], "ndarray": [[1, 2]]}},
        "reward": 1.0,
    }
    rv = client.post("/api/v1.0/feedback", json=payload)
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200


def test_feedback_v01_ok():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()

    payload = {
        "request": {"data": {"names": ["a", "b"], "ndarray": [[1, 2]]}},
        "reward": 1.0,
    }
    rv = client.post("/api/v0.1/feedback", json=payload)
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200


def test_model_feedback_lowlevel_ok():
    user_object = UserObjectLowLevel()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get(
        '/send-feedback?json={"request":{"data":{"ndarray":[]}},"reward":1.0}'
    )
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200


def test_model_non200status_lowlevel():
    user_object = UserObjectLowLevelWithStatusInResponse()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/predict?json={"request":{"data":{"ndarray":[]}},"reward":1.0}')
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 400


@skipif_tf_missing
def test_model_tftensor_ok():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(tftensor=tf.make_tensor_proto(arr))
    request = prediction_pb2.SeldonMessage(data=datadef)
    jStr = json_format.MessageToJson(request)
    rv = client.get("/predict?json=" + jStr)
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]
    assert "tftensor" in j["data"]
    tfp = TensorProto()
    json_format.ParseDict(j["data"].get("tftensor"), tfp, ignore_unknown_fields=False)
    arr2 = tf.make_ndarray(tfp)
    assert np.array_equal(arr, arr2)


def test_model_ok_with_names():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/predict?json={"data":{"names":["a","b"],"ndarray":[[1,2]]}}')
    j = json.loads(rv.data)
    assert rv.status_code == 200
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]


def test_model_bin_data():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    bdata = b"123"
    bdata_base64 = base64.b64encode(bdata).decode("utf-8")
    rv = client.get('/predict?json={"binData":"' + bdata_base64 + '"}')
    j = json.loads(rv.data)
    assert rv.status_code == 200
    assert j["binData"] == bdata_base64
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]


def test_model_bin_data_nparray():
    user_object = UserObject(ret_nparray=True)
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    encoded = base64.b64encode(b"1234").decode("utf-8")
    rv = client.get('/predict?json={"binData":"' + encoded + '"}')
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["data"]["tensor"]["values"] == [1, 2, 3]
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]


def test_model_str_data():
    user_object = UserObject(ret_nparray=True)
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/predict?json={"strData":"my data"}')
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["data"]["tensor"]["values"] == [1, 2, 3]
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]


def test_model_str_data_identity():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/predict?json={"strData":"my data"}')
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 200
    assert j["strData"] == "my data"
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]


def test_model_no_json():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    uo = UserObject()
    rv = client.get("/predict?")
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 400


def test_model_bad_metrics():
    user_object = UserObject(metrics_ok=False)
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/predict?json={"data":{"ndarray":[]}}')
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 500


def test_model_error_status_code():
    class ErrorUserObject:
        def predict(self, X, features_names, **kwargs):
            raise SeldonMicroserviceException("foo", status_code=403)

    user_object = ErrorUserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    uo = UserObject()
    rv = client.get('/predict?json={"strData":"my data"}')
    j = json.loads(rv.data)
    logging.info(j)
    assert rv.status_code == 403


def test_model_gets_meta():
    user_object = UserObject(ret_meta=True)
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/predict?json={"meta":{"puid": "abc"},"data":{"ndarray":[]}}')
    j = json.loads(rv.data)
    logging.info(j)

    assert rv.status_code == 200
    assert j["meta"]["tags"] == {"inc_meta": {"puid": "abc"}}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]


def test_model_passes_through_tags():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get(
        '/predict?json={"meta":{"tags":{"foo":"bar"}},"data":{"ndarray":[]}}'
    )
    j = json.loads(rv.data)
    logging.info(j)

    assert rv.status_code == 200
    assert j["meta"]["tags"] == {"foo": "bar", "mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]


def test_model_passes_through_metrics():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get(
        '/predict?json={"meta":{"metrics":[{"key":"request_gauge","type":"GAUGE","value":100}]},"data":{"ndarray":[]}}'
    )
    j = json.loads(rv.data)
    logging.info(j)

    assert rv.status_code == 200
    assert j["meta"]["metrics"][0]["key"] == "request_gauge"
    assert j["meta"]["metrics"][0]["value"] == 100

    assert j["meta"]["metrics"][1]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][1]["value"] == user_object.metrics()[0]["value"]


def test_model_seldon_json_ok():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get("/seldon.json")
    assert rv.status_code == 200


def test_model_health_ping():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get(HEALTH_PING_URL)
    assert rv.status_code == 200
    assert rv.data == b"pong"


def test_model_health_status():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get(HEALTH_STATUS_URL)
    assert rv.status_code == 200
    j = json.loads(rv.data)
    logging.info(j)
    assert j["data"]["tensor"]["values"] == UserObject.HEALTH_STATUS_REPONSE


def test_model_health_status_raw():
    user_object = UserObjectLowLevel()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get(HEALTH_STATUS_URL)
    assert rv.status_code == 200
    j = json.loads(rv.data)
    assert j["data"]["ndarray"] == UserObjectLowLevel.HEALTH_STATUS_RAW_RESPONSE


def test_model_metadata():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get(METADATA_URL)
    assert rv.status_code == 200
    j = json.loads(rv.data)
    logging.info(j)
    assert j == UserObject.METADATA_RESPONSE


def test_proto_ok():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = SeldonModelGRPC(user_object, seldon_metrics)
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
    )
    request = prediction_pb2.SeldonMessage(data=datadef)
    resp = app.Predict(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    logging.info(j)
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]
    assert j["data"]["tensor"]["shape"] == [2, 1]
    assert j["data"]["tensor"]["values"] == [1, 2]


def test_proto_passes_through_tags():
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
    resp = app.Predict(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    logging.info(j)
    assert j["meta"]["tags"] == {"foo": "bar", "mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]
    assert j["data"]["tensor"]["shape"] == [2, 1]
    assert j["data"]["tensor"]["values"] == [1, 2]


def test_proto_passes_through_metrics():
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
    resp = app.Predict(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    logging.info(j)

    assert j["meta"]["metrics"][0]["key"] == "request_gauge"
    assert j["meta"]["metrics"][0]["value"] == 100

    assert j["meta"]["metrics"][1]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][1]["value"] == user_object.metrics()[0]["value"]

    assert j["data"]["tensor"]["shape"] == [2, 1]
    assert j["data"]["tensor"]["values"] == [1, 2]


def test_proto_lowlevel():
    user_object = UserObjectLowLevelGrpc()
    seldon_metrics = SeldonMetrics()
    app = SeldonModelGRPC(user_object, seldon_metrics)
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
    )
    request = prediction_pb2.SeldonMessage(data=datadef)
    resp = app.Predict(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    logging.info(j)
    assert j["data"]["tensor"]["shape"] == [2, 1]
    assert j["data"]["tensor"]["values"] == [9, 9]


def test_proto_feedback():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = SeldonModelGRPC(user_object, seldon_metrics)
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
    )
    request = prediction_pb2.SeldonMessage(data=datadef)
    feedback = prediction_pb2.Feedback(request=request, reward=1.0)
    resp = app.SendFeedback(feedback, None)


def test_proto_feedback_custom():
    user_object = UserObjectLowLevel()
    seldon_metrics = SeldonMetrics()
    app = SeldonModelGRPC(user_object, seldon_metrics)
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(shape=(2, 1), values=arr)
    )
    request = prediction_pb2.SeldonMessage(data=datadef)
    feedback = prediction_pb2.Feedback(request=request, reward=1.0)
    resp = app.SendFeedback(feedback, None)


@skipif_tf_missing
def test_proto_tftensor_ok():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = SeldonModelGRPC(user_object, seldon_metrics)
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(tftensor=tf.make_tensor_proto(arr))
    request = prediction_pb2.SeldonMessage(data=datadef)
    resp = app.Predict(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    logging.info(j)
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]
    arr2 = tf.make_ndarray(resp.data.tftensor)
    assert np.array_equal(arr, arr2)


def test_proto_bin_data():
    user_object = UserObject()
    seldon_metrics = SeldonMetrics()
    app = SeldonModelGRPC(user_object, seldon_metrics)
    bdata = b"123"
    bdata_base64 = base64.b64encode(bdata)
    request = prediction_pb2.SeldonMessage(binData=bdata_base64)
    resp = app.Predict(request, None)
    assert resp.binData == bdata_base64


def test_proto_bin_data_nparray():
    user_object = UserObject(ret_nparray=True)
    seldon_metrics = SeldonMetrics()
    app = SeldonModelGRPC(user_object, seldon_metrics)
    binData = b"\0\1"
    request = prediction_pb2.SeldonMessage(binData=binData)
    resp = app.Predict(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    logging.info(j)
    assert j["data"]["tensor"]["values"] == list(user_object.nparray.flatten())


def test_get_grpc_server():
    user_object = UserObject(ret_nparray=True)
    seldon_metrics = SeldonMetrics()
    server = get_grpc_server(user_object, seldon_metrics)


def test_proto_gets_meta():
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
    resp = app.Predict(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    logging.info(j)
    assert j["meta"]["tags"] == {"inc_meta": {"puid": "abc"}}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]
    assert j["data"]["tensor"]["shape"] == [2, 1]
    assert j["data"]["tensor"]["values"] == [1, 2]


def test_unimplemented_predict_raw_on_seldon_component():
    class CustomSeldonComponent(SeldonComponent):
        def predict(self, X, features_names, **kwargs):
            return X * 2

    user_object = CustomSeldonComponent()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/predict?json={"data":{"names":["a","b"],"ndarray":[[1,2]]}}')
    j = json.loads(rv.data)

    logging.info(j)
    assert rv.status_code == 200
    assert j["data"]["ndarray"] == [[2.0, 4.0]]


def test_unimplemented_predict_raw():
    class CustomObject:
        def predict(self, X, features_names, **kwargs):
            return X * 2

    user_object = CustomObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get('/predict?json={"data":{"names":["a","b"],"ndarray":[[1,2]]}}')
    j = json.loads(rv.data)

    logging.info(j)
    assert rv.status_code == 200
    assert j["data"]["ndarray"] == [[2.0, 4.0]]


def test_unimplemented_feedback_raw_on_seldon_component():
    class CustomSeldonComponent(SeldonComponent):
        def feedback(self, features, feature_names, reward, truth):
            logging.info("Feedback called")

    user_object = CustomSeldonComponent()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get(
        '/send-feedback?json={"request":{"data":{"ndarray":[]}},"reward":1.0}'
    )
    j = json.loads(rv.data)

    logging.info(j)
    assert rv.status_code == 200


def test_unimplemented_feedback_raw():
    class CustomObject:
        def feedback(self, features, feature_names, reward, truth):
            logging.info("Feedback called")

    user_object = CustomObject()
    seldon_metrics = SeldonMetrics()
    app = get_rest_microservice(user_object, seldon_metrics)
    client = app.test_client()
    rv = client.get(
        '/send-feedback?json={"request":{"data":{"ndarray":[]}},"reward":1.0}'
    )
    j = json.loads(rv.data)

    logging.info(j)
    assert rv.status_code == 200
