import grpc
from concurrent import futures
from flask import jsonify, Flask, send_from_directory, request
from flask_cors import CORS
import logging
from seldon_core.utils import get_request, json_to_seldonMessage, seldonMessage_to_json, json_to_feedback
import seldon_core.seldon_methods
from seldon_core.microservice import SeldonMicroserviceException, ANNOTATION_GRPC_MAX_MSG_SIZE
from seldon_core.proto import prediction_pb2_grpc

logger = logging.getLogger(__name__)


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
        requestJson = get_request()
        logger.debug("REST Request: %s", request)
        requestProto = json_to_seldonMessage(requestJson)
        logger.debug("Proto Request: %s", requestProto)
        responseProto = seldon_core.seldon_methods.predict(user_model, requestProto)
        jsonDict = seldonMessage_to_json(responseProto)
        return jsonify(jsonDict)

    @app.route("/send-feedback", methods=["GET", "POST"])
    def SendFeedback():
        requestJson = get_request()
        logger.debug("REST Request: %s", request)
        requestProto = json_to_feedback(requestJson)
        logger.debug("Proto Request: %s", requestProto)
        responseProto = seldon_core.seldon_methods.send_feedback(user_model, requestProto)
        jsonDict = seldonMessage_to_json(responseProto)
        return jsonify(jsonDict)

    @app.route("/transform-input", methods=["GET", "POST"])
    def TransformInput():
        requestJson = get_request()
        logger.debug("REST Request: %s", request)
        requestProto = json_to_seldonMessage(requestJson)
        logger.debug("Proto Request: %s", request)
        responseProto = seldon_core.seldon_methods.transform_input(user_model, requestProto)
        jsonDict = seldonMessage_to_json(responseProto)
        return jsonify(jsonDict)

    @app.route("/transform-output", methods=["GET", "POST"])
    def TransformOutput():
        requestJson = get_request()
        logger.debug("REST Request: %s", request)
        requestProto = json_to_seldonMessage(requestJson)
        logger.debug("Proto Request: %s", request)
        responseProto = seldon_core.seldon_methods.transform_output(user_model, requestProto)
        jsonDict = seldonMessage_to_json(responseProto)
        return jsonify(jsonDict)

    return app

# ----------------------------
# GRPC
# ----------------------------

class SeldonModelGRPC(object):
    def __init__(self, user_model):
        self.user_model = user_model

    def Predict(self, request, context):
        return seldon_core.seldon_methods.predict(self.user_model,request)

    def SendFeedback(self, feedback, context):
       return seldon_core.seldon_methods.send_feedback(self.user_model,feedback)



def get_grpc_server(user_model, debug=False, annotations={}, trace_interceptor=None):
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

    if trace_interceptor:
        from grpc_opentracing.grpcext import intercept_server
        server = intercept_server(server, trace_interceptor)

    prediction_pb2_grpc.add_ModelServicer_to_server(seldon_model, server)

    return server

