import grpc
from concurrent import futures

from flask import jsonify, Flask, send_from_directory
from flask_cors import CORS
import numpy as np
import logging

from seldon_core.proto import prediction_pb2, prediction_pb2_grpc
from seldon_core.microservice import extract_message, sanity_check_request, rest_datadef_to_array, \
    array_to_rest_datadef, grpc_datadef_to_array, array_to_grpc_datadef, \
    SeldonMicroserviceException, ANNOTATION_GRPC_MAX_MSG_SIZE

logger = logging.getLogger(__name__)

# ---------------------------
# Interaction with user model
# ---------------------------


def score(user_model, features, feature_names):
    # Returns a numpy array of floats that corresponds to the outlier scores for each point in the batch
    return user_model.score(features, feature_names)

# ----------------------------
# REST
# ----------------------------


def get_rest_microservice(user_model, debug=False):
    logger = logging.getLogger(__name__ + '.get_rest_microservice')

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
        return send_from_directory("openapi", "seldon.json")

    @app.route("/transform-input", methods=["GET", "POST"])
    def TransformInput():
        request = extract_message()
        sanity_check_request(request)

        datadef = request.get("data")
        features = rest_datadef_to_array(datadef)

        outlier_scores = score(user_model, features, datadef.get("names"))
        # TODO: check that predictions is 2 dimensional

        request["meta"].setdefault("tags", {})
        request["meta"]["tags"]["outlierScore"] = list(outlier_scores)

        return jsonify(request)

    return app


# ----------------------------
# GRPC
# ----------------------------

class SeldonTransformerGRPC(object):
    def __init__(self, user_model):
        self.user_model = user_model

    def TransformInput(self, request, context):
        datadef = request.data
        features = grpc_datadef_to_array(datadef)

        outlier_scores = score(self.user_model, features, datadef.names)

        request.meta.tags["outlierScore"] = list(outlier_scores)

        return request


def get_grpc_server(user_model, debug=False, annotations={}):
    seldon_model = SeldonTransformerGRPC(user_model)
    options = []
    if ANNOTATION_GRPC_MAX_MSG_SIZE in annotations:
        max_msg = int(annotations[ANNOTATION_GRPC_MAX_MSG_SIZE])
        logger.info("Setting grpc max message to %d", max_msg)
        options.append(('grpc.max_message_length', max_msg))

    server = grpc.server(futures.ThreadPoolExecutor(
        max_workers=10), options=options)
    prediction_pb2_grpc.add_ModelServicer_to_server(seldon_model, server)

    return server
