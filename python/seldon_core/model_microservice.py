import grpc
from concurrent import futures
from google.protobuf import json_format

from flask import jsonify, Flask, send_from_directory
from flask_cors import CORS
import numpy as np
import logging

from tornado.tcpserver import TCPServer
from tornado.iostream import StreamClosedError
from tornado import gen
import tornado.ioloop
import struct
import traceback
import os

from seldon_core.proto import prediction_pb2, prediction_pb2_grpc
from seldon_core.microservice import extract_message, sanity_check_request, rest_datadef_to_array, \
    array_to_rest_datadef, grpc_datadef_to_array, array_to_grpc_datadef, \
    SeldonMicroserviceException, get_custom_tags, get_data_from_json, get_data_from_proto, ANNOTATION_GRPC_MAX_MSG_SIZE
from seldon_core.metrics import get_custom_metrics
from seldon_core.seldon_flatbuffers import SeldonRPCToNumpyArray, NumpyArrayToSeldonRPC, CreateErrorMsg

logger = logging.getLogger(__name__)

# ---------------------------
# Interaction with user model
# ---------------------------


def predict(user_model, features, feature_names):
    return user_model.predict(features, feature_names)


def send_feedback(user_model, features, feature_names, reward, truth):
    if hasattr(user_model, "send_feedback"):
        user_model.send_feedback(features, feature_names, reward, truth)


def get_class_names(user_model, n_targets):
    if hasattr(user_model, "class_names"):
        return user_model.class_names
    else:
        return ["t:{}".format(i) for i in range(n_targets)]


# ----------------------------
# REST
# ----------------------------

def get_rest_microservice(user_model, debug=False):

    app = Flask(__name__, static_url_path='')
    CORS(app)

    @app.errorhandler(SeldonMicroserviceException)
    def handle_invalid_usage(error):
        response = jsonify(error.to_dict())
        logger.error("%s", error.to_dict())
        response.status_code = 400
        return response

    @app.route("/seldon.json", methods=["GET"])
    def openAPI():
        return send_from_directory('', "seldon.json")

    @app.route("/predict", methods=["GET", "POST"])
    def Predict():
        request = extract_message()
        logger.debug("Request: %s", request)

        sanity_check_request(request)

        if hasattr(user_model, "predict_rest"):
            return jsonify(user_model.predict_rest(request))
        else:
            features = get_data_from_json(request)
            names = request.get("data", {}).get("names")

            predictions = predict(user_model, features, names)
            logger.debug("Predictions: %s", predictions)

            # If predictions is an numpy array or we used the default data then return as numpy array
            if isinstance(predictions, np.ndarray) or "data" in request:
                predictions = np.array(predictions)
                if len(predictions.shape) > 1:
                    class_names = get_class_names(user_model, predictions.shape[1])
                else:
                    class_names = []
                data = array_to_rest_datadef(
                    predictions, class_names, request.get("data", {}))
                response = {"data": data, "meta": {}}
            else:
                response = {"binData": predictions, "meta": {}}

            tags = get_custom_tags(user_model)
            if tags:
                response["meta"]["tags"] = tags
            metrics = get_custom_metrics(user_model)
            if metrics:
                response["meta"]["metrics"] = metrics
            return jsonify(response)

    @app.route("/send-feedback", methods=["GET", "POST"])
    def SendFeedback():
        feedback = extract_message()
        logger.debug("Feedback received: %s", feedback)
        
        if hasattr(user_model, "send_feedback_rest"):
            return jsonify(user_model.send_feedback_rest(feedback))
        else:
            datadef_request = feedback.get("request", {}).get("data", {})
            features = rest_datadef_to_array(datadef_request)

            datadef_truth = feedback.get("truth", {}).get("data", {})
            truth = rest_datadef_to_array(datadef_truth)

            reward = feedback.get("reward")

            send_feedback(user_model, features,
                          datadef_request.get("names"), reward, truth)
            return jsonify({})

    return app


# ----------------------------
# GRPC
# ----------------------------

class SeldonModelGRPC(object):
    def __init__(self, user_model):
        self.user_model = user_model

    def Predict(self, request, context):
        if hasattr(self.user_model, "predict_grpc"):
            return self.user_model.predict_grpc(request)
        else:
            features = get_data_from_proto(request)
            datadef = request.data
            data_type = request.WhichOneof("data_oneof")
            predictions = predict(self.user_model, features, datadef.names)

            # Construct meta data
            meta = prediction_pb2.Meta()
            metaJson = {}
            tags = get_custom_tags(self.user_model)
            if tags:
                metaJson["tags"] = tags
            metrics = get_custom_metrics(self.user_model)
            if metrics:
                metaJson["metrics"] = metrics
            json_format.ParseDict(metaJson, meta)

            if isinstance(predictions, np.ndarray) or data_type == "data":
                predictions = np.array(predictions)
                if len(predictions.shape) > 1:
                    class_names = get_class_names(
                        self.user_model, predictions.shape[1])
                else:
                    class_names = []

                if data_type == "data":
                    default_data_type = request.data.WhichOneof("data_oneof")
                else:
                    default_data_type = "tensor"
                data = array_to_grpc_datadef(
                    predictions, class_names, default_data_type)
                return prediction_pb2.SeldonMessage(data=data, meta=meta)
            else:
                return prediction_pb2.SeldonMessage(binData=predictions, meta=meta)

    def SendFeedback(self, feedback, context):
        if hasattr(self.user_model, "send_feedback_grpc"):
            self.user_model.send_feedback_grpc(feedback)
        else:
            datadef_request = feedback.request.data
            features = grpc_datadef_to_array(datadef_request)

            truth = grpc_datadef_to_array(feedback.truth)
            reward = feedback.reward

            send_feedback(self.user_model, features,
                          datadef_request.names, truth, reward)

        return prediction_pb2.SeldonMessage()



def get_grpc_server(user_model, debug=False, annotations={}):
    seldon_model = SeldonModelGRPC(user_model)
    options = []
    if ANNOTATION_GRPC_MAX_MSG_SIZE in annotations:
        max_msg = int(annotations[ANNOTATION_GRPC_MAX_MSG_SIZE])
        logger.info(
            "Setting grpc max message and receive length to %d", max_msg)
        options.append(('grpc.max_message_length', max_msg))
        options.append(('grpc.max_receive_message_length', max_msg))

    server = grpc.server(futures.ThreadPoolExecutor(
        max_workers=10), options=options)
    prediction_pb2_grpc.add_ModelServicer_to_server(seldon_model, server)

    return server


# ----------------------------
# Flatbuffers (experimental)
# ----------------------------

class SeldonFlatbuffersServer(TCPServer):
    def __init__(self, user_model):
        super(SeldonFlatbuffersServer, self).__init__()
        self.user_model = user_model

    @gen.coroutine
    def handle_stream(self, stream, address):
        while True:
            try:
                data = yield stream.read_bytes(4)
                obj = struct.unpack('<i', data)
                len_msg = obj[0]
                data = yield stream.read_bytes(len_msg)
                try:
                    features, names = SeldonRPCToNumpyArray(data)
                    predictions = np.array(
                        predict(self.user_model, features, names))
                    if len(predictions.shape) > 1:
                        class_names = get_class_names(
                            self.user_model, predictions.shape[1])
                    else:
                        class_names = []
                    outData = NumpyArrayToSeldonRPC(predictions, class_names)
                    yield stream.write(outData)
                except StreamClosedError:
                    logger.exception(
                        "Stream closed during processing:", address)
                    break
                except Exception:
                    tb = traceback.format_exc()
                    logger.exception(
                        "Caught exception during processing:", address, tb)
                    outData = CreateErrorMsg(tb)
                    yield stream.write(outData)
                    stream.close()
                    break
            except StreamClosedError:
                logger.exception(
                    "Stream closed during data inputstream read:", address)
                break


def run_flatbuffers_server(user_model, port, debug=False):
    server = SeldonFlatbuffersServer(user_model)
    server.listen(port)
    logger.info("Tornado Server listening on port %s", port)
    tornado.ioloop.IOLoop.current().start()
