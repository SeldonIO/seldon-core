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


'''
A basic tensorflow serving proxy
'''
class TfServingProxy(object):

    def __init__(
            self,
            rest_endpoint=None,
            grpc_endpoint=None,
            model_name=None,
            signature_name=None,
            model_input=None,
            model_output=None):
        log.warning("rest_endpoint:",rest_endpoint)
        log.warning("grpc_endpoint:",grpc_endpoint)
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


    # if we have a TFTensor message we got directly without converting the message otherwise we go the usual route
    def predict_raw(self,request):
        log.debug("Predict raw")
        request_data_type = request.WhichOneof("data_oneof")
        default_data_type = request.data.WhichOneof("data_oneof")
        log.debug(str(request_data_type), str(default_data_type))
        if default_data_type == "tftensor" and self.grpc:
            tfrequest = predict_pb2.PredictRequest()
            tfrequest.model_spec.name = self.model_name
            tfrequest.model_spec.signature_name = self.signature_name
            tfrequest.inputs[self.model_input].CopyFrom(request.data.tftensor)
            result = self.stub.Predict(tfrequest)
            log.debug(result)
            datadef = prediction_pb2.DefaultData(
                tftensor=result.outputs[self.model_output]
            )
            return prediction_pb2.SeldonMessage(data=datadef)

        elif request_data_type == "jsonData":
            features = get_data_from_proto(request)
            predictions = self.predict(features, features_names=[])
            try:
                sm = json_to_seldon_message({"jsonData": predictions})
            except ParseError as e:
                sm = prediction_pb2.SeldonMessage(strData=predictions)
            return sm
        
        else:
            features = get_data_from_proto(request)
            datadef = request.data
            predictions = self.predict(features, datadef.names)
            predictions = np.array(predictions)

            if request_data_type is not "data":
                default_data_type = "tensor"

            class_names = []
            data = array_to_grpc_datadef(
                predictions, class_names, default_data_type)

            return prediction_pb2.SeldonMessage(data=data)



    def predict(self,X,features_names=[]):
        if self.grpc and type(X) is not dict:
            request = predict_pb2.PredictRequest()
            request.model_spec.name = self.model_name
            request.model_spec.signature_name = self.signature_name
            request.inputs[self.model_input].CopyFrom(tf.contrib.util.make_tensor_proto(X.tolist(), shape=X.shape))
            result = self.stub.Predict(request)
            log.debug("GRPC Response: ", str(result))
            response = numpy.array(result.outputs[self.model_output].float_val)
            if len(response.shape) == 1:
                response = numpy.expand_dims(response, axis=0)
            return response
        else:
            log.debug(self.rest_endpoint)
            if type(X) is dict:
                log.debug("JSON Request")
                data = X
            else:
                log.debug("Data Request")
                data = {"instances":X.tolist()}
                if not self.signature_name is None:
                    data["signature_name"] = self.signature_name
            log.debug(str(data))

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

