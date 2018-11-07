import grpc
import numpy
import tensorflow as tf

from tensorflow.python.saved_model import signature_constants
from tensorflow_serving.apis import predict_pb2
from tensorflow_serving.apis import prediction_service_pb2_grpc

import requests
import json

class TensorflowServerError(Exception):

    def __init__(self, message):
        self.message = message

'''
A basic tensorflow serving proxy
'''
class TfServingProxy(object):

    def __init__(self,rest_endpoint=None,grpc_endpoint=None,model_name=None,signature_name=None,model_input=None,model_output=None):
        print("rest_endpoint:",rest_endpoint)
        print("grpc_endpoint:",grpc_endpoint)
        if not grpc_endpoint is None:
            self.grpc = True
            channel = grpc.insecure_channel(grpc_endpoint)
            self.stub = prediction_service_pb2_grpc.PredictionServiceStub(channel)
        else:
            self.grpc = False
            self.rest_endpoint = rest_endpoint
        self.model_name = model_name
        if signature_name is None:
            self.signature_name = signature_constants.DEFAULT_SERVING_SIGNATURE_DEF_KEY
        else:
            self.signature_name = signature_name
        self.model_input = model_input        
        self.model_output = model_output
        
    def predict(self,X,features_names):
        if self.grpc:
            request = predict_pb2.PredictRequest()
            request.model_spec.name = self.model_name
            request.model_spec.signature_name = self.signature_name
            request.inputs[self.model_input].CopyFrom(tf.contrib.util.make_tensor_proto(X.tolist(), shape=X.shape))
            print(request)
            result = self.stub.Predict(request)
            print(result)
            response = numpy.array(result.outputs[self.model_output].float_val)
            if len(response.shape) == 1:
                response = numpy.expand_dims(response, axis=0)
            return response
        else:
            print(self.rest_endpoint)
            data = {"instances":X.tolist()}
            if not self.signature_name is None:
                data["signature_name"] = self.signature_name
            print(data)
            response = requests.post(
                self.rest_endpoint,
                data = json.dumps(data))
            if response.status_code == 200:
                result = numpy.array(response.json()["predictions"])
                if len(result.shape) == 1:
                    result = numpy.expand_dims(result, axis=0)
                return result
            else:
                print("Error from server:",response)
                raise TensorflowServerError(response.json())
