import logging
import os
from concurrent import futures

import grpc
from flask import Flask, Response, request, send_from_directory
from flask_cors import CORS
from grpc_reflection.v1alpha import reflection

import seldon_core.seldon_methods
from seldon_core.env_utils import get_model_name
from seldon_core.flask_utils import (
    ANNOTATION_GRPC_MAX_MSG_SIZE,
    SeldonMicroserviceException,
    get_request,
    jsonify,
)
from seldon_core.proto import prediction_pb2, prediction_pb2_grpc
from seldon_core.utils import (
    getenv_as_bool,
    json_to_feedback,
    json_to_seldon_model_metadata,
    seldon_message_to_json,
)

logger = logging.getLogger(__name__)

PRED_UNIT_ID = get_model_name("0")
METRICS_ENDPOINT = os.environ.get("PREDICTIVE_UNIT_METRICS_ENDPOINT", "/metrics")
PAYLOAD_PASSTHROUGH = getenv_as_bool("PAYLOAD_PASSTHROUGH", default=False)


def get_rest_microservice(user_model, seldon_metrics):
    app = Flask(__name__, static_url_path="")
    CORS(app)

    _set_flask_app_configs(app)

    # dict representing the validated model metadata
    # None value will represent a validation error
    metadata_data = seldon_core.seldon_methods.init_metadata(user_model)

    if hasattr(user_model, "model_error_handler"):
        logger.info("Registering the custom error handler...")
        app.register_blueprint(user_model.model_error_handler)

    @app.errorhandler(Exception)
    def handle_generic_exception(e):
        error = SeldonMicroserviceException(
            message=str(e), status_code=500, reason="MICROSERVICE_INTERNAL_ERROR"
        )
        response = jsonify(error.to_dict())
        logger.error("%s", error.to_dict())
        response.status_code = error.status_code
        return response

    @app.errorhandler(SeldonMicroserviceException)
    def handle_invalid_usage(error):
        response = jsonify(error.to_dict())
        logger.error("%s", error.to_dict())
        response.status_code = error.status_code
        return response

    @app.route("/seldon.json", methods=["GET"])
    def openAPI():
        return send_from_directory("", "openapi/seldon.json")

    @app.route("/predict", methods=["GET", "POST"])
    @app.route("/api/v1.0/predictions", methods=["POST"])
    @app.route("/api/v0.1/predictions", methods=["POST"])
    def Predict():
        requestJson = get_request(skip_decoding=PAYLOAD_PASSTHROUGH)
        logger.debug("REST Request: %s", request)
        response = seldon_core.seldon_methods.predict(
            user_model, requestJson, seldon_metrics
        )

        json_response = jsonify(response, skip_encoding=PAYLOAD_PASSTHROUGH)
        if (
            isinstance(response, dict)
            and "status" in response
            and "code" in response["status"]
        ):
            json_response.status_code = response["status"]["code"]

        logger.debug("REST Response: %s", response)
        return json_response

    @app.route("/send-feedback", methods=["GET", "POST"])
    @app.route("/api/v1.0/feedback", methods=["POST"])
    @app.route("/api/v0.1/feedback", methods=["POST"])
    def SendFeedback():
        requestJson = get_request()
        logger.debug("REST Request: %s", request)
        requestProto = json_to_feedback(requestJson)
        logger.debug("Proto Request: %s", requestProto)
        responseProto = seldon_core.seldon_methods.send_feedback(
            user_model, requestProto, PRED_UNIT_ID, seldon_metrics
        )
        jsonDict = seldon_message_to_json(responseProto)
        return jsonify(jsonDict)

    @app.route("/transform-input", methods=["GET", "POST"])
    def TransformInput():
        requestJson = get_request()
        logger.debug("REST Request: %s", request)
        response = seldon_core.seldon_methods.transform_input(
            user_model, requestJson, seldon_metrics
        )
        logger.debug("REST Response: %s", response)
        return jsonify(response)

    @app.route("/transform-output", methods=["GET", "POST"])
    def TransformOutput():
        requestJson = get_request()
        logger.debug("REST Request: %s", request)
        response = seldon_core.seldon_methods.transform_output(
            user_model, requestJson, seldon_metrics
        )
        logger.debug("REST Response: %s", response)
        return jsonify(response)

    @app.route("/route", methods=["GET", "POST"])
    def Route():
        requestJson = get_request()
        logger.debug("REST Request: %s", request)
        response = seldon_core.seldon_methods.route(
            user_model, requestJson, seldon_metrics
        )
        logger.debug("REST Response: %s", response)
        return jsonify(response)

    @app.route("/aggregate", methods=["GET", "POST"])
    def Aggregate():
        requestJson = get_request()
        logger.debug("REST Request: %s", request)
        response = seldon_core.seldon_methods.aggregate(
            user_model, requestJson, seldon_metrics
        )
        logger.debug("REST Response: %s", response)
        return jsonify(response)

    @app.route("/health/ping", methods=["GET"])
    def HealthPing():
        """
        Lightweight endpoint to check the liveness of the REST endpoint
        """
        return "pong"

    @app.route("/health/status", methods=["GET"])
    @app.route("/api/v1.0/health/status", methods=["GET"])
    def HealthStatus():
        logger.debug("REST Health Status Request")
        response = seldon_core.seldon_methods.health_status(user_model, seldon_metrics)
        logger.debug("REST Health Status Response: %s", response)
        return jsonify(response)

    @app.route("/metadata", methods=["GET"])
    @app.route("/api/v1.0/metadata", methods=["GET"])
    def Metadata():
        if metadata_data is None:
            # None value represents validation error in current implementation
            # if user_model would not define init_metadata than metadata_data
            # would just contain a default values
            raise SeldonMicroserviceException(
                "Model metadata unavailable",
                status_code=500,
                reason="MICROSERVICE_BAD_METADATA",
            )
        logger.debug("REST Metadata Request")
        logger.debug("REST Metadata Response: %s", metadata_data)
        return jsonify(metadata_data)

    return app


def get_metrics_microservice(seldon_metrics):
    app = Flask(__name__, static_url_path="")
    CORS(app)

    _set_flask_app_configs(app)

    @app.route(METRICS_ENDPOINT, methods=["GET"])
    def Metrics():
        logger.debug("REST Metrics Request")
        metrics, mimetype = seldon_metrics.generate_metrics()
        return Response(metrics, mimetype=mimetype)

    return app


def _set_flask_app_configs(app):
    """
    Set the configs for the flask app based on environment variables
    See https://flask.palletsprojects.com/config/#builtin-configuration-values
    :param app:
    :return:
    """
    FLASK_CONFIG_IDENTIFIER = "FLASK_"
    FLASK_CONFIGS_ALLOWED = [
        "DEBUG",
        "EXPLAIN_TEMPLATE_LOADING",
        "JSONIFY_PRETTYPRINT_REGULAR",
        "JSON_SORT_KEYS",
        "PROPAGATE_EXCEPTIONS",
        "PRESERVE_CONTEXT_ON_EXCEPTION",
        "SESSION_COOKIE_HTTPONLY",
        "SESSION_COOKIE_SECURE",
        "SESSION_REFRESH_EACH_REQUEST",
        "TEMPLATES_AUTO_RELOAD",
        "TESTING",
        "TRAP_HTTP_EXCEPTIONS",
        "TRAP_BAD_REQUEST_ERRORS",
        "USE_X_SENDFILE",
    ]

    for flask_config in FLASK_CONFIGS_ALLOWED:
        flask_config_value = getenv_as_bool(
            f"{FLASK_CONFIG_IDENTIFIER}{flask_config}", default=None
        )
        if flask_config_value is None:
            continue
        app.config[flask_config] = flask_config_value
    logger.info(f"App Config:  {app.config}")


# ----------------------------
# GRPC
# ----------------------------


class SeldonModelGRPC:
    def __init__(self, user_model, seldon_metrics):
        self.user_model = user_model
        self.seldon_metrics = seldon_metrics

        self.metadata_data = seldon_core.seldon_methods.init_metadata(user_model)

    def Predict(self, request_grpc, context):
        scope = self.user_model.tracer.scope_manager.activate(context.get_active_span(), True)
        predict = seldon_core.seldon_methods.predict(
            self.user_model, request_grpc, self.seldon_metrics
        )
        scope.close()
        return predict

    def SendFeedback(self, feedback_grpc, context):
        return seldon_core.seldon_methods.send_feedback(
            self.user_model, feedback_grpc, PRED_UNIT_ID, self.seldon_metrics
        )

    def TransformInput(self, request_grpc, context):
        return seldon_core.seldon_methods.transform_input(
            self.user_model, request_grpc, self.seldon_metrics
        )

    def TransformOutput(self, request_grpc, context):
        return seldon_core.seldon_methods.transform_output(
            self.user_model, request_grpc, self.seldon_metrics
        )

    def Route(self, request_grpc, context):
        return seldon_core.seldon_methods.route(
            self.user_model, request_grpc, self.seldon_metrics
        )

    def Aggregate(self, request_grpc, context):
        return seldon_core.seldon_methods.aggregate(
            self.user_model, request_grpc, self.seldon_metrics
        )

    def Metadata(self, request_grpc, context):
        """Metadata method of rpc Model service"""
        return json_to_seldon_model_metadata(self.metadata_data)

    def ModelMetadata(self, request_grpc, context):
        """ModelMetadata method of rpc Seldon service"""
        return json_to_seldon_model_metadata(self.metadata_data)

    def GraphMetadata(self, request_grpc, context):
        """GraphMetadata method of rpc Seldon service"""
        raise NotImplementedError("GraphMetadata not available on the Model level.")


def get_grpc_server(
    user_model, seldon_metrics, annotations={}, trace_interceptor=None, num_threads=1
):
    seldon_model = SeldonModelGRPC(user_model, seldon_metrics)
    options = [
        ("grpc.so_reuseport", 1),
    ]
    if ANNOTATION_GRPC_MAX_MSG_SIZE in annotations:
        max_msg = int(annotations[ANNOTATION_GRPC_MAX_MSG_SIZE])
        logger.info("Setting grpc max message and receive length to %d", max_msg)
        options.append(("grpc.max_message_length", max_msg))
        options.append(("grpc.max_send_message_length", max_msg))
        options.append(("grpc.max_receive_message_length", max_msg))

    server = grpc.server(
        futures.ThreadPoolExecutor(max_workers=num_threads), options=options
    )

    if trace_interceptor:
        from grpc_opentracing.grpcext import intercept_server

        server = intercept_server(server, trace_interceptor)

    prediction_pb2_grpc.add_GenericServicer_to_server(seldon_model, server)
    prediction_pb2_grpc.add_ModelServicer_to_server(seldon_model, server)
    prediction_pb2_grpc.add_TransformerServicer_to_server(seldon_model, server)
    prediction_pb2_grpc.add_OutputTransformerServicer_to_server(seldon_model, server)
    prediction_pb2_grpc.add_CombinerServicer_to_server(seldon_model, server)
    prediction_pb2_grpc.add_RouterServicer_to_server(seldon_model, server)
    prediction_pb2_grpc.add_SeldonServicer_to_server(seldon_model, server)

    SERVICE_NAMES = (
        prediction_pb2.DESCRIPTOR.services_by_name["Generic"].full_name,
        prediction_pb2.DESCRIPTOR.services_by_name["Model"].full_name,
        prediction_pb2.DESCRIPTOR.services_by_name["Router"].full_name,
        prediction_pb2.DESCRIPTOR.services_by_name["Transformer"].full_name,
        prediction_pb2.DESCRIPTOR.services_by_name["OutputTransformer"].full_name,
        prediction_pb2.DESCRIPTOR.services_by_name["Combiner"].full_name,
        prediction_pb2.DESCRIPTOR.services_by_name["Seldon"].full_name,
        reflection.SERVICE_NAME,
    )
    reflection.enable_server_reflection(SERVICE_NAMES, server)

    return server
