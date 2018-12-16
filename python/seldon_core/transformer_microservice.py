import grpc
from concurrent import futures

from flask import jsonify, Flask, send_from_directory
from flask_cors import CORS
import numpy as np
from google.protobuf import json_format
import logging

from seldon_core.proto import prediction_pb2, prediction_pb2_grpc
from seldon_core.microservice import extract_message, sanity_check_request, rest_datadef_to_array, \
    array_to_rest_datadef, grpc_datadef_to_array, array_to_grpc_datadef, \
    SeldonMicroserviceException, get_custom_tags, get_data_from_json, get_data_from_proto, ANNOTATION_GRPC_MAX_MSG_SIZE
from seldon_core.metrics import get_custom_metrics

logger = logging.getLogger(__name__)

# ---------------------------
# Interaction with user model
# ---------------------------


def transform_input(user_model, features, feature_names):
    if hasattr(user_model, "transform_input"):
        return user_model.transform_input(features, feature_names)
    else:
        return features


def transform_output(user_model, features, feature_names):
    if hasattr(user_model, "transform_output"):
        return user_model.transform_output(features, feature_names)
    else:
        return features


def get_feature_names(user_model, original):
    if hasattr(user_model, "feature_names"):
        return user_model.feature_names
    else:
        return original


def get_class_names(user_model, original):
    if hasattr(user_model, "class_names"):
        return user_model.class_names
    else:
        return original


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

    @app.route("/transform-input", methods=["GET", "POST"])
    def TransformInput():
        request = extract_message()
        logger.debug("Request: %s", request)

        sanity_check_request(request)

        if hasattr(user_model, "transform_input_rest"):
            return jsonify(user_model.transform_input_rest(request))
        else:
            features = get_data_from_json(request)
            names = request.get("data", {}).get("names")

            transformed = transform_input(user_model, features, names)
            logger.debug("Transformed: %s", transformed)

            # If predictions is an numpy array or we used the default data then return as numpy array
            if isinstance(transformed, np.ndarray) or "data" in request:
                new_feature_names = get_feature_names(user_model, names)
                transformed = np.array(transformed)
                data = array_to_rest_datadef(
                    transformed, new_feature_names, request.get("data", {}))
                response = {"data": data, "meta": {}}
            else:
                response = {"binData": transformed, "meta": {}}

            tags = get_custom_tags(user_model)
            if tags:
                response["meta"]["tags"] = tags
            metrics = get_custom_metrics(user_model)
            if metrics:
                response["meta"]["metrics"] = metrics
            return jsonify(response)

    @app.route("/transform-output", methods=["GET", "POST"])
    def TransformOutput():
        request = extract_message()
        logger.debug("Request: %s", request)

        sanity_check_request(request)

        if hasattr(user_model, "transform_output_rest"):
            return jsonify(user_model.transform_output_rest(request))
        else:
            features = get_data_from_json(request)
            names = request.get("data", {}).get("names")

            transformed = transform_output(user_model, features, names)
            logger.debug("Transformed: %s", transformed)

            if isinstance(transformed, np.ndarray) or "data" in request:
                new_class_names = get_class_names(user_model, names)
                data = array_to_rest_datadef(
                    transformed, new_class_names, request.get("data", {}))
                response = {"data": data, "meta": {}}
            else:
                response = {"binData": transformed, "meta": {}}

            tags = get_custom_tags(user_model)
            if tags:
                response["meta"]["tags"] = tags
            metrics = get_custom_metrics(user_model)
            if metrics:
                response["meta"]["metrics"] = metrics
            return jsonify(response)

    return app


# ----------------------------
# GRPC
# ----------------------------

class SeldonTransformerGRPC(object):
    def __init__(self, user_model):
        self.user_model = user_model

    def TransformInput(self, request, context):
        if hasattr(self.user_model, "transform_input_grpc"):
            return self.user_model.transform_input_grpc(request)
        else:
            features = get_data_from_proto(request)
            datadef = request.data
            data_type = request.WhichOneof("data_oneof")

            transformed = transform_input(self.user_model, features, datadef.names)

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

            if isinstance(transformed, np.ndarray) or data_type == "data":
                transformed = np.array(transformed)
                feature_names = get_feature_names(self.user_model, datadef.names)
                if data_type == "data":
                    default_data_type = request.data.WhichOneof("data_oneof")
                else:
                    default_data_type = "tensor"
                data = array_to_grpc_datadef(
                    transformed, feature_names, default_data_type)
                return prediction_pb2.SeldonMessage(data=data, meta=meta)
            else:
                return prediction_pb2.SeldonMessage(binData=transformed, meta=meta)

    def TransformOutput(self, request, context):
        if hasattr(self.user_model, "transform_output_grpc"):
            return self.user_model.transform_output_grpc(request)
        else:
            features = get_data_from_proto(request)
            datadef = request.data
            data_type = request.WhichOneof("data_oneof")

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

            transformed = transform_output(
                self.user_model, features, datadef.names)

            if isinstance(transformed, np.ndarray) or data_type == "data":
                transformed = np.array(transformed)
                class_names = get_class_names(self.user_model, datadef.names)
                if data_type == "data":
                    default_data_type = request.data.WhichOneof("data_oneof")
                else:
                    default_data_type = "tensor"
                data = array_to_grpc_datadef(
                    transformed, class_names, default_data_type)
                return prediction_pb2.SeldonMessage(data=data, meta=meta)
            else:
                return prediction_pb2.SeldonMessage(binData=transformed, meta=meta)



def get_grpc_server(user_model, debug=False, annotations={}):
    seldon_model = SeldonTransformerGRPC(user_model)
    options = []
    if ANNOTATION_GRPC_MAX_MSG_SIZE in annotations:
        max_msg = int(annotations[ANNOTATION_GRPC_MAX_MSG_SIZE])
        logger.info("Setting grpc max message to %d", max_msg)
        options.append(('grpc.max_message_length', max_msg))

    server = grpc.server(futures.ThreadPoolExecutor(
        max_workers=10), options=options)
    prediction_pb2_grpc.add_TransformerServicer_to_server(seldon_model, server)
    prediction_pb2_grpc.add_OutputTransformerServicer_to_server(seldon_model, server)    

    return server
