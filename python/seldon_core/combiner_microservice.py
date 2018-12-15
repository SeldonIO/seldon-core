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


def aggregate(user_model, features_list, feature_names_list):
    if hasattr(user_model, "aggregate"):
        return user_model.aggregate(features_list, feature_names_list)
    else:
        return features_list[0]

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


def sanity_check_seldon_message_list(request):
    if not "seldonMessages" in request:
         raise SeldonMicroserviceException("Request must contain seldonMessages field")
    msgs = request["seldonMessages"]
    if not type(msgs) in [list,tuple]:
        raise SeldonMicroserviceException("seldonMessages field is not a list")
    if len(msgs) == 0:
        raise SeldonMicroserviceException("seldonMessages field is empty")
    for idx, msg in enumerate(msgs):
        try:
            sanity_check_request(msg)
        except SeldonMicroserviceException as err:
            raise SeldonMicroserviceException("Invalid SeldonMessage at index {0} : {1}".format(idx,err.message))
            
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

    @app.route("/aggregate", methods=["GET", "POST"])
    def Aggregate():
        request = extract_message()
        logger.debug("Request: %s", request)

        sanity_check_seldon_message_list(request)

        if hasattr(user_model, "aggregate_rest"):
            return jsonify(user_model.aggregate_rest(request))
        else:
            features_list = []
            names_list = []

            for msg in request["seldonMessages"]:
                features = get_data_from_json(msg)
                names = msg.get("data", {}).get("names")

                features_list.append(features)
                names_list.append(names)
                
            aggregated = aggregate(user_model, features_list, names_list)
            logger.debug("Aggregated: %s", aggregated)

            # If predictions is a numpy array or we used the default data then return as numpy array
            if isinstance(aggregated, np.ndarray) or "data" in request["seldonMessages"][0]:
                new_feature_names = get_feature_names(user_model, names_list[0])
                aggregated = np.array(aggregated)
                data = array_to_rest_datadef(
                    aggregated, new_feature_names, request["seldonMessages"][0].get("data", {}))
                response = {"data": data, "meta": {}}
            else:
                response = {"binData": aggregated, "meta": {}}

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

class SeldonCombinerGRPC(object):
    def __init__(self, user_model):
        self.user_model = user_model

    def Aggregate(self, request, context):
        if hasattr(self.user_model, "aggregate_grpc"):
            return self.user_model.aggregate_grpc(request)
        else:
            features_list = []
            names_list = []
            
            for msg in request.seldonMessages:
                features = get_data_from_proto(msg)
                features_list.append(features)
                names_list.append(msg.data.names)

            data_type = request.seldonMessages[0].WhichOneof("data_oneof")

            aggregated = aggregate(self.user_model, features_list, names_list)

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

            if isinstance(aggregated, np.ndarray) or data_type == "data":
                aggregated = np.array(aggregated)
                feature_names = get_feature_names(self.user_model, [])
                if data_type == "data":
                    default_data_type = request.seldonMessages[0].data.WhichOneof("data_oneof")
                else:
                    default_data_type = "tensor"
                data = array_to_grpc_datadef(
                    aggregated, feature_names, default_data_type)
                return prediction_pb2.SeldonMessage(data=data, meta=meta)
            else:
                return prediction_pb2.SeldonMessage(binData=aggregated, meta=meta)


def get_grpc_server(user_model, debug=False, annotations={}):
    seldon_model = SeldonCombinerGRPC(user_model)
    options = []
    if ANNOTATION_GRPC_MAX_MSG_SIZE in annotations:
        max_msg = int(annotations[ANNOTATION_GRPC_MAX_MSG_SIZE])
        logger.info("Setting grpc max message to %d", max_msg)
        options.append(('grpc.max_message_length', max_msg))

    server = grpc.server(futures.ThreadPoolExecutor(
        max_workers=10), options=options)
    prediction_pb2_grpc.add_CombinerServicer_to_server(seldon_model, server)

    return server
