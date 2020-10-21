from TfServingProxy import TfServingProxy
from seldon_core.proto import prediction_pb2
from seldon_core.utils import get_data_from_proto, array_to_grpc_datadef, json_to_seldon_message
from tensorflow_serving.apis import predict_pb2
import numpy as np
from unittest import mock
import tensorflow as tf
import requests

ARR_REQUEST_VALUE=np.random.rand(1,1).astype(np.float32)
ARR_RESPONSE_VALUE=np.random.rand(1,1).astype(np.float32)

class FakeStub(object):

    def __init__(self, channel):
        self.channel = channel

    @staticmethod
    def Predict(*args, **kwargs):
        data = ARR_RESPONSE_VALUE
        tensor_proto = tf.contrib.util.make_tensor_proto(
                data.tolist(),
                shape=data.shape)
        tfresponse = predict_pb2.PredictResponse()
        tfresponse.model_spec.name = "newmodel"
        tfresponse.model_spec.signature_name = "signame"
        tfresponse.outputs["scores"].CopyFrom(
            tensor_proto)
        return tfresponse

@mock.patch("tensorflow_serving.apis.prediction_service_pb2_grpc.PredictionServiceStub", new=FakeStub)
def test_grpc_predict_function_tensor():
    t = TfServingProxy(
        grpc_endpoint="localhost:8080",
        model_name="newmodel",
        signature_name="signame",
        model_input="images",
        model_output="scores")

    data = ARR_REQUEST_VALUE
    datadef = array_to_grpc_datadef("tensor", data)
    request = prediction_pb2.SeldonMessage(data=datadef)
    response = t.predict_grpc(request)
    resp_data = get_data_from_proto(response)
    assert resp_data == ARR_RESPONSE_VALUE


@mock.patch("tensorflow_serving.apis.prediction_service_pb2_grpc.PredictionServiceStub", new=FakeStub)
def test_grpc_predict_function_tftensor():
    t = TfServingProxy(
        grpc_endpoint="localhost:8080",
        model_name="newmodel",
        signature_name="signame",
        model_input="images",
        model_output="scores")

    data = ARR_REQUEST_VALUE
    tensor_proto = tf.contrib.util.make_tensor_proto(
            data.tolist(),
            shape=data.shape)
    datadef = prediction_pb2.DefaultData(
        tftensor=tensor_proto
    )
    request = prediction_pb2.SeldonMessage(data=datadef)
    response = t.predict_grpc(request)
    resp_data = get_data_from_proto(response)
    assert resp_data == ARR_RESPONSE_VALUE


@mock.patch("tensorflow_serving.apis.prediction_service_pb2_grpc.PredictionServiceStub", new=FakeStub)
def test_grpc_predict_function_ndarray():
    t = TfServingProxy(
        grpc_endpoint="localhost:8080",
        model_name="newmodel",
        signature_name="signame",
        model_input="images",
        model_output="scores")

    data = ARR_REQUEST_VALUE
    datadef = array_to_grpc_datadef(
        "ndarray", data, [])
    request = prediction_pb2.SeldonMessage(data=datadef)
    response = t.predict_grpc(request)
    resp_data = get_data_from_proto(response)
    assert resp_data == ARR_RESPONSE_VALUE


@mock.patch.object(requests, "post")
def test_rest_predict_function_json(mock_request_post):

    data = {"jsonData": ARR_RESPONSE_VALUE.tolist() }
    def res():
        r = requests.Response()
        r.status_code = 200
        type(r).text = mock.PropertyMock(return_value="text")  # property mock
        def json_func():
            return data
        r.json = json_func
        return r
    mock_request_post.return_value = res()

    t = TfServingProxy(
        rest_endpoint="http://localhost:8080",
        model_name="newmodel",
        signature_name="signame",
        model_input="images",
        model_output="scores")

    request = { "jsonData": ARR_REQUEST_VALUE.tolist() }
    response = t.predict(request)
    assert response == data

@mock.patch.object(requests, "post")
def test_rest_predict_function_ndarray(mock_request_post):

    data = {"data": { "ndarray": ARR_RESPONSE_VALUE.tolist(), "names": [] } }
    def res():
        r = requests.Response()
        r.status_code = 200
        type(r).text = mock.PropertyMock(return_value="text")  # property mock
        def json_func():
            return data
        r.json = json_func
        return r
    mock_request_post.return_value = res()

    t = TfServingProxy(
        rest_endpoint="http://localhost:8080",
        model_name="newmodel",
        signature_name="signame",
        model_input="images",
        model_output="scores")

    request = { "data": { "ndarray": ARR_REQUEST_VALUE.tolist(), "names": [] }}
    response = t.predict(request)
    assert response == data

