from TfServingProxy import TfServingProxy
import tensorflow as tf
from seldon_core.proto import prediction_pb2
from seldon_core.utils import get_data_from_proto, array_to_grpc_datadef, json_to_seldon_message
from tensorflow_serving.apis import predict_pb2
import numpy as np
from unittest import mock

ARR_REQUEST_VALUE=np.random.rand(1,1)
ARR_RESPONSE_VALUE=np.random.rand(1,1)

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
    tf = TfServingProxy(
        grpc_endpoint="localhost:8080",
        model_name="newmodel",
        signature_name="signame",
        model_input="images",
        model_output="scores")

    data = ARR_REQUEST_VALUE
    datadef = array_to_grpc_datadef("tensor", data)
    request = prediction_pb2.SeldonMessage(data=datadef)
    response = tf.predict_grpc(request)
    resp_data = get_data_from_proto(response)
    assert resp_data[0] == ARR_RESPONSE_VALUE[0]




