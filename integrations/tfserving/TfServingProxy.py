import grpc
import numpy
import tensorflow as tf

from tensorflow.python.saved_model import signature_constants
from tensorflow_serving.apis import predict_pb2
from tensorflow_serving.apis import prediction_service_pb2_grpc
from seldon_core.utils import get_data_from_proto, array_to_grpc_datadef, json_to_seldon_message
from seldon_core.proto import prediction_pb2
from google.protobuf.json_format import ParseError

import requests
import json
import numpy as np

import logging

log = logging.getLogger()

class TfServingProxy(object):
    """
    A basic tensorflow serving proxy
    """

    def __init__(
            self,
            rest_endpoint=None,
            grpc_endpoint=None,
            model_name=None,
            signature_name=None,
            model_input=None,
            model_output=None):
        log.debug("rest_endpoint:",rest_endpoint)
        log.debug("grpc_endpoint:",grpc_endpoint)
        if not grpc_endpoint is None:
            self.grpc = True
            channel = grpc.insecure_channel(grpc_endpoint)
            self.stub = prediction_service_pb2_grpc.PredictionServiceStub(channel)
        else:
            self.grpc = False
            self.rest_endpoint = rest_endpoint+"/v1/models/"+model_name+":predict"
        self.model_name = model_name
        if signature_name is None:
            self.signature_name = signature_constants.DEFAULT_SERVING_SIGNATURE_DEF_KEY
        else:
            self.signature_name = signature_name
        self.model_input = model_input
        self.model_output = model_output

    def predict_grpc(self,request):
        """
        predict_grpc will be called only when there is a GRPC call to the server
        which in this case, the request will be sent to the TFServer directly.
        """
        log.debug("Preprocessing contents for predict function")
        request_data_type = request.WhichOneof("data_oneof")
        default_data_type = request.data.WhichOneof("data_oneof")
        log.debug(f"Request data type: {request_data_type}, Default data type: {default_data_type}")

        if request_data_type != "data":
            raise Exception("strData, binData and jsonData not supported.")

        request = predict_pb2.PredictRequest()
        request.model_spec.name = self.model_name
        request.model_spec.signature_name = self.signature_name

        # For GRPC case, if we have a TFTensor message we can pass it directly
        if default_data_type == "tftensor":
            request.inputs[self.model_input].CopyFrom(request.data.tftensor)
            result = self.stub.Predict(request)
            datadef = prediction_pb2.DefaultData(
                tftensor=result.outputs[self.model_output]
            )
            return prediction_pb2.SeldonMessage(data=datadef)

        else:
            request.inputs[self.model_input].CopyFrom(
                tf.contrib.util.make_tensor_proto(
                    request.data.tolist(), 
                    shape=request.data.shape))
            result = self.stub.Predict(request)
            result_arr = numpy.array(result.outputs[self.model_output].float_val)
            if len(response.shape) == 1:
                response = numpy.expand_dims(response, axis=0)
            class_names = []
            data = array_to_grpc_datadef(
                result_arr, class_names, default_data_type)
            return prediction_pb2.SeldonMessage(data=data)

    def predict(self, X, features_names=[]):
        """
        This predict function will only be called when the server is called with a REST request.
        The REST request is able to support any configuration required.
        """
        if type(X) is dict:
            log.debug(f"JSON Request: {X}")
            data = X
        else:
            log.debug(f"Data Request: {X}")
            data = {"instances":X.tolist()}
            if not self.signature_name is None:
                data["signature_name"] = self.signature_name

        response = requests.post(self.rest_endpoint, data=json.dumps(data))

        if response.status_code == 200:
            log.debug(response.text)
            if type(X) is dict:
                try:
                    return response.json()
                except ValueError:
                    return response.text
            else:
                result = numpy.array(response.json()["predictions"])
                if len(result.shape) == 1:
                    result = numpy.expand_dims(result, axis=0)
                return result
        else:
            log.warning("Error from server: "+ str(response) + " content: " + str(response.text))
            try:
                return response.json()
            except ValueError:
                return response.text

