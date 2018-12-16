import grpc
from concurrent import futures

from flask import jsonify, Flask, send_from_directory
from flask_cors import CORS
import numpy as np
import os
import logging
from google.protobuf import json_format

from seldon_core.proto import prediction_pb2, prediction_pb2_grpc
from seldon_core.microservice import extract_message, sanity_check_request, rest_datadef_to_array, \
    array_to_rest_datadef, grpc_datadef_to_array, array_to_grpc_datadef, \
    SeldonMicroserviceException, get_custom_tags, ANNOTATION_GRPC_MAX_MSG_SIZE
from seldon_core.metrics import get_custom_metrics

PRED_UNIT_ID = os.environ.get("PREDICTIVE_UNIT_ID","0")

logger = logging.getLogger(__name__)

# ---------------------------
# Interaction with user router
# ---------------------------


def route(user_router, features, feature_names):
    return user_router.route(features, feature_names)


def send_feedback(user_router, features, feature_names, routing, reward, truth):
    return user_router.send_feedback(features, feature_names, routing, reward, truth)

# ----------------------------
# REST
# ----------------------------


def get_rest_microservice(user_router, debug=False):

    app = Flask(__name__, static_url_path='')
    CORS(app)

    @app.errorhandler(SeldonMicroserviceException)
    def handle_invalid_usage(error):
        response = jsonify(error.to_dict())
        response.status_code = 400
        return response

    @app.route("/seldon.json", methods=["GET"])
    def openAPI():
        return send_from_directory("openapi", "seldon.json")

    @app.route("/route", methods=["GET", "POST"])
    def Route():

        request = extract_message()
        logger.debug("Request: %s", request)

        sanity_check_request(request)

        if hasattr(user_router, "route_rest"):
            return jsonify(user_router.route_rest(request))
        else:
            datadef = request.get("data")
            features = rest_datadef_to_array(datadef)

            routing = np.array(
                [[route(user_router, features, datadef.get("names"))]])
            # TODO: check that predictions is 2 dimensional
            class_names = []

            data = array_to_rest_datadef(routing, class_names, datadef)

            response = {"data": data, "meta": {}}
            tags = get_custom_tags(user_router)
            if tags:
                response["meta"]["tags"] = tags
            metrics = get_custom_metrics(user_router)
            if metrics:
                response["meta"]["metrics"] = metrics
            return jsonify(response)

    @app.route("/send-feedback", methods=["GET", "POST"])
    def SendFeedback():
        feedback = extract_message()

        logger.debug("Feedback received: %s", feedback)

        if hasattr(user_router, "send_feedback_rest"):
            return jsonify(user_router.send_feedback_rest(feedback))
        else:
            datadef_request = feedback.get("request", {}).get("data", {})
            features = rest_datadef_to_array(datadef_request)

            datadef_truth = feedback.get("truth", {}).get("data", {})
            truth = rest_datadef_to_array(datadef_truth)
            reward = feedback.get("reward")

            try:
                routing = feedback.get("response").get(
                    "meta").get("routing").get(PRED_UNIT_ID)
            except AttributeError:
                raise SeldonMicroserviceException(
                    "Router feedback must contain a routing dictionary in the response metadata")

            send_feedback(user_router, features, datadef_request.get(
                "names"), routing, reward, truth)
            return jsonify({})

    return app


# ----------------------------
# GRPC
# ----------------------------

class SeldonRouterGRPC(object):
    def __init__(self, user_model):
        self.user_model = user_model

    def Route(self, request, context):
        if hasattr(self.user_model, "route_grpc"):
            return self.user_model.route_grpc(request)
        else:
            datadef = request.data
            features = grpc_datadef_to_array(datadef)

            routing = np.array([[route(self.user_model, features, datadef.names)]])
            # TODO: check that predictions is 2 dimensional
            class_names = []

            data = array_to_grpc_datadef(
                routing, class_names, request.data.WhichOneof("data_oneof"))

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

            return prediction_pb2.SeldonMessage(data=data, meta=meta)

    def SendFeedback(self, feedback, context):
        if hasattr(self.user_model, "send_feedback_grpc"):
            self.user_model.send_feedback_grpc(feedback)
        else:
            datadef_request = feedback.request.data
            features = grpc_datadef_to_array(datadef_request)

            truth = grpc_datadef_to_array(feedback.truth)
            reward = feedback.reward
            routing = feedback.response.meta.routing.get(PRED_UNIT_ID)

            send_feedback(self.user_model, features,
                          datadef_request.names, routing, reward, truth)

            return prediction_pb2.SeldonMessage()


def get_grpc_server(user_model, debug=False, annotations={}):
    seldon_router = SeldonRouterGRPC(user_model)
    options = []
    if ANNOTATION_GRPC_MAX_MSG_SIZE in annotations:
        max_msg = int(annotations[ANNOTATION_GRPC_MAX_MSG_SIZE])
        logger.info("Setting grpc max message to %d", max_msg)
        options.append(('grpc.max_message_length', max_msg))

    server = grpc.server(futures.ThreadPoolExecutor(
        max_workers=10), options=options)
    prediction_pb2_grpc.add_RouterServicer_to_server(seldon_router, server)

    return server
