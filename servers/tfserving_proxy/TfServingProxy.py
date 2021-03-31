import json
import logging

import grpc
import numpy
import requests
import tensorflow as tf
from google.protobuf.any_pb2 import Any
from seldon_core.proto import prediction_pb2
from seldon_core.utils import grpc_datadef_to_array
from tensorflow.python.saved_model import signature_constants
from tensorflow_serving.apis import predict_pb2
from tensorflow_serving.apis import prediction_service_pb2_grpc

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
        log.debug("rest_endpoint:", rest_endpoint)
        log.debug("grpc_endpoint:", grpc_endpoint)

        # grpc
        max_msg = 1000000000
        options = [('grpc.max_message_length', max_msg),
                   ('grpc.max_send_message_length', max_msg),
                   ('grpc.max_receive_message_length', max_msg)]
        channel = grpc.insecure_channel(grpc_endpoint, options)
        self.stub = prediction_service_pb2_grpc.PredictionServiceStub(channel)

        # rest
        self.rest_endpoint = rest_endpoint + "/v1/models/" + model_name + ":predict"
        self.model_name = model_name
        if signature_name is None:
            self.signature_name = signature_constants.DEFAULT_SERVING_SIGNATURE_DEF_KEY
        else:
            self.signature_name = signature_name
        self.model_input = model_input
        self.model_output = model_output

    def predict_grpc(self, request):
        """
        predict_grpc will be called only when there is a GRPC call to the server
        which in this case, the request will be sent to the TFServer directly.
        """
        log.debug("Preprocessing contents for predict function")
        request_data_type = request.WhichOneof("data_oneof")
        default_data_type = request.data.WhichOneof("data_oneof")
        log.debug(f"Request data type: {request_data_type}, Default data type: {default_data_type}")

        if request_data_type not in ["data", "customData"]:
            raise Exception("strData, binData and jsonData not supported.")

        tfrequest = predict_pb2.PredictRequest()

        # handle inputs
        if request_data_type == "data":
            # For GRPC case, if we have a TFTensor message we can pass it directly
            if default_data_type == "tftensor":
                tfrequest.inputs[self.model_input].CopyFrom(request.data.tftensor)
            else:
                data_arr = grpc_datadef_to_array(request.data)
                tfrequest.inputs[self.model_input].CopyFrom(
                    tf.make_tensor_proto(
                        data_arr.tolist(),
                        shape=data_arr.shape))

        else:
            # Unpack custom data into tfrequest - taking raw inputs prepared by the user.
            # This allows the use case when the model's input is not a single tftensor
            # but a map of tensors like defined in predict.proto:
            # PredictRequest.inputs: map<string, TensorProto>
            request.customData.Unpack(tfrequest)

        # handle prediction
        tfrequest.model_spec.name = self.model_name
        tfrequest.model_spec.signature_name = self.signature_name
        tfresponse = self.stub.Predict(tfrequest)

        # handle result
        if request_data_type == "data":
            datadef = prediction_pb2.DefaultData(
                tftensor=tfresponse.outputs[self.model_output]
            )
            result = prediction_pb2.SeldonMessage(data=datadef)
        else:
            # Pack tfresponse into the SeldonMessage's custom data - letting user handle
            # raw outputs. This allows the case when the model's output is not a single tftensor
            # but a map of tensors like defined in predict.proto:
            # PredictResponse: map<string, TensorProto>
            custom_data = Any()
            custom_data.Pack(tfresponse)
            result = prediction_pb2.SeldonMessage(customData=custom_data)

        return result

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
            data = {"instances": X.tolist()}
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
            log.warning("Error from server: " + str(response) + " content: " + str(response.text))
            try:
                return response.json()
            except ValueError:
                return response.text
