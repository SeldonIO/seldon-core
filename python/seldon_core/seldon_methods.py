import logging
import os
from typing import Any, Dict, List, Tuple, Union

import numpy as np
import yaml
from google.protobuf import json_format

from seldon_core.flask_utils import SeldonMicroserviceException
from seldon_core.metadata import SeldonInvalidMetadataError, validate_model_metadata
from seldon_core.metrics import (
    AGGREGATE_METRIC_METHOD_TAG,
    FEEDBACK_METRIC_METHOD_TAG,
    HEALTH_METRIC_METHOD_TAG,
    INPUT_TRANSFORM_METRIC_METHOD_TAG,
    OUTPUT_TRANSFORM_METRIC_METHOD_TAG,
    PREDICT_METRIC_METHOD_TAG,
    ROUTER_METRIC_METHOD_TAG,
    SeldonMetrics,
)
from seldon_core.proto import prediction_pb2
from seldon_core.user_model import (
    INCLUDE_METRICS_IN_CLIENT_RESPONSE,
    SeldonNotImplementedError,
    SeldonResponse,
    client_aggregate,
    client_custom_metrics,
    client_health_status,
    client_predict,
    client_route,
    client_send_feedback,
    client_transform_input,
    client_transform_output,
)
from seldon_core.utils import (
    construct_response,
    construct_response_json,
    extract_feedback_request_parts,
    extract_request_parts,
    extract_request_parts_json,
    getenv_as_bool,
    json_to_seldon_message,
    seldon_message_to_json,
)

logger = logging.getLogger(__name__)


def handle_raw_custom_metrics(
    msg: Union[prediction_pb2.SeldonMessage, Dict],
    seldon_metrics: SeldonMetrics,
    is_proto: bool,
    method: str,
):
    """
    Update SeldonMetrics object with custom metrics from raw methods.
    If INCLUDE_METRICS_IN_CLIENT_RESPONSE environmental variable is set to "true"
    metrics will be dropped from msg.
    """
    metrics = []
    if is_proto:
        # proto to json extracts dictionary in no particular order
        # sorting tags it here to ensure unique key in metrics.py _generate_tags_key
        metrics = seldon_message_to_json(msg.meta).get("metrics", [])
        for metric in metrics:
            metric["tags"] = dict(sorted(metric["tags"].items(), key=lambda item: item[1]))
        if metrics and not INCLUDE_METRICS_IN_CLIENT_RESPONSE:
            del msg.meta.metrics[:]
    elif isinstance(msg, dict):
        metrics = msg.get("meta", {}).get("metrics", [])
        if metrics and not INCLUDE_METRICS_IN_CLIENT_RESPONSE:
            del msg["meta"]["metrics"]
    seldon_metrics.update(metrics, method)


def predict(
    user_model: Any,
    request: Union[prediction_pb2.SeldonMessage, List, Dict, bytes],
    seldon_metrics: SeldonMetrics,
) -> Union[prediction_pb2.SeldonMessage, List, Dict, bytes]:
    """
    Call the user model to get a prediction and package the response

    Parameters
    ----------
    user_model
       User defined class instance
    request
       The incoming request

    Returns
    -------
      The prediction
    """
    # TODO: Find a way to choose predict_rest or predict_grpc when payload is
    # not decoded
    is_proto = isinstance(request, prediction_pb2.SeldonMessage)

    if hasattr(user_model, "predict_rest") and not is_proto:
        logger.warning("predict_rest is deprecated. Please use predict_raw")
        return user_model.predict_rest(request)
    elif hasattr(user_model, "predict_grpc") and is_proto:
        logger.warning("predict_grpc is deprecated. Please use predict_raw")
        return user_model.predict_grpc(request)
    else:
        if hasattr(user_model, "predict_raw"):
            try:
                response = user_model.predict_raw(request)
                handle_raw_custom_metrics(
                    response, seldon_metrics, is_proto, PREDICT_METRIC_METHOD_TAG
                )
                return response
            except SeldonNotImplementedError:
                pass

        if is_proto:
            (features, meta, datadef, data_type) = extract_request_parts(request)

            client_response = client_predict(
                user_model, features, datadef.names, meta=meta
            )

            metrics = client_custom_metrics(
                user_model,
                seldon_metrics,
                PREDICT_METRIC_METHOD_TAG,
                client_response.metrics,
            )

            return construct_response(
                user_model,
                False,
                request,
                client_response.data,
                meta,
                metrics,
                client_response.tags,
            )
        else:
            (features, meta, datadef, data_type) = extract_request_parts_json(request)
            class_names = datadef["names"] if datadef and "names" in datadef else []

            client_response = client_predict(
                user_model, features, class_names, meta=meta
            )

            metrics = client_custom_metrics(
                user_model,
                seldon_metrics,
                PREDICT_METRIC_METHOD_TAG,
                client_response.metrics,
            )

            return construct_response_json(
                user_model,
                False,
                request,
                client_response.data,
                meta,
                metrics,
                client_response.tags,
            )


def send_feedback(
    user_model: Any,
    request: prediction_pb2.Feedback,
    predictive_unit_id: str,
    seldon_metrics: SeldonMetrics,
) -> prediction_pb2.SeldonMessage:
    """

    Parameters
    ----------
    user_model
       A Seldon user model
    request
       SeldonMesage proto
    predictive_unit_id
       The ID of the enclosing container predictive unit. Will be taken from environment.

    Returns
    -------

    """
    seldon_metrics.update_reward(request.reward)

    if hasattr(user_model, "send_feedback_rest"):
        logger.warning("send_feedback_rest is deprecated. Please use send_feedback_raw")
        request_json = json_format.MessageToJson(request)
        response_json = user_model.send_feedback_rest(request_json)
        return json_to_seldon_message(response_json)
    elif hasattr(user_model, "send_feedback_grpc"):
        logger.warning("send_feedback_grpc is deprecated. Please use send_feedback_raw")
        response_json = user_model.send_feedback_grpc(request)
        return json_to_seldon_message(response_json)
    else:
        if hasattr(user_model, "send_feedback_raw"):
            try:
                response = user_model.send_feedback_raw(request)
                handle_raw_custom_metrics(
                    response, seldon_metrics, True, FEEDBACK_METRIC_METHOD_TAG
                )
                return response
            except SeldonNotImplementedError:
                pass

        (datadef_request, features, truth, reward) = extract_feedback_request_parts(
            request
        )
        routing = request.response.meta.routing.get(predictive_unit_id)

        client_response = client_send_feedback(
            user_model, features, datadef_request.names, reward, truth, routing
        )

        metrics = client_custom_metrics(
            user_model,
            seldon_metrics,
            FEEDBACK_METRIC_METHOD_TAG,
            client_response.metrics,
        )

        if client_response.data is None:
            client_response.data = np.array([])

        return construct_response(
            user_model,
            False,
            request.request,
            client_response.data,
            None,
            metrics,
            client_response.tags,
        )


def transform_input(
    user_model: Any,
    request: Union[prediction_pb2.SeldonMessage, List, Dict],
    seldon_metrics: SeldonMetrics,
) -> Union[prediction_pb2.SeldonMessage, List, Dict]:
    """

    Parameters
    ----------
    user_model
       User defined class to handle transform input
    request
       The incoming request

    Returns
    -------
       The transformed request

    """
    is_proto = isinstance(request, prediction_pb2.SeldonMessage)

    if hasattr(user_model, "transform_input_rest"):
        logger.warning(
            "transform_input_rest is deprecated. Please use transform_input_raw"
        )
        return user_model.transform_input_rest(request)
    elif hasattr(user_model, "transform_input_grpc"):
        logger.warning(
            "transform_input_grpc is deprecated. Please use transform_input_raw"
        )
        return user_model.transform_input_grpc(request)
    else:
        if hasattr(user_model, "transform_input_raw"):
            try:
                response = user_model.transform_input_raw(request)
                handle_raw_custom_metrics(
                    response,
                    seldon_metrics,
                    is_proto,
                    INPUT_TRANSFORM_METRIC_METHOD_TAG,
                )
                return response
            except SeldonNotImplementedError:
                pass

        if is_proto:
            (features, meta, datadef, data_type) = extract_request_parts(request)

            client_response = client_transform_input(
                user_model, features, datadef.names, meta=meta
            )

            metrics = client_custom_metrics(
                user_model,
                seldon_metrics,
                INPUT_TRANSFORM_METRIC_METHOD_TAG,
                client_response.metrics,
            )

            return construct_response(
                user_model,
                False,
                request,
                client_response.data,
                meta,
                metrics,
                client_response.tags,
            )
        else:
            (features, meta, datadef, data_type) = extract_request_parts_json(request)
            class_names = datadef["names"] if datadef and "names" in datadef else []

            client_response = client_transform_input(
                user_model, features, class_names, meta=meta
            )

            metrics = client_custom_metrics(
                user_model,
                seldon_metrics,
                INPUT_TRANSFORM_METRIC_METHOD_TAG,
                client_response.metrics,
            )

            return construct_response_json(
                user_model,
                False,
                request,
                client_response.data,
                meta,
                metrics,
                client_response.tags,
            )


def transform_output(
    user_model: Any,
    request: Union[prediction_pb2.SeldonMessage, List, Dict],
    seldon_metrics: SeldonMetrics,
) -> Union[prediction_pb2.SeldonMessage, List, Dict]:
    """

    Parameters
    ----------
    user_model
       User defined class to handle transform input
    request
       The incoming request

    Returns
    -------
       The transformed request

    """
    is_proto = isinstance(request, prediction_pb2.SeldonMessage)

    if hasattr(user_model, "transform_output_rest"):
        logger.warning(
            "transform_input_rest is deprecated. Please use transform_input_raw"
        )
        return user_model.transform_output_rest(request)
    elif hasattr(user_model, "transform_output_grpc"):
        logger.warning(
            "transform_input_grpc is deprecated. Please use transform_input_raw"
        )
        return user_model.transform_output_grpc(request)
    else:
        if hasattr(user_model, "transform_output_raw"):
            try:
                response = user_model.transform_output_raw(request)
                handle_raw_custom_metrics(
                    response,
                    seldon_metrics,
                    is_proto,
                    OUTPUT_TRANSFORM_METRIC_METHOD_TAG,
                )
                return response
            except SeldonNotImplementedError:
                pass

        if is_proto:
            (features, meta, datadef, data_type) = extract_request_parts(request)

            client_response = client_transform_output(
                user_model, features, datadef.names, meta=meta
            )

            metrics = client_custom_metrics(
                user_model,
                seldon_metrics,
                OUTPUT_TRANSFORM_METRIC_METHOD_TAG,
                client_response.metrics,
            )

            return construct_response(
                user_model,
                False,
                request,
                client_response.data,
                meta,
                metrics,
                client_response.tags,
            )
        else:
            (features, meta, datadef, data_type) = extract_request_parts_json(request)
            class_names = datadef["names"] if datadef and "names" in datadef else []

            client_response = client_transform_output(
                user_model, features, class_names, meta=meta
            )

            metrics = client_custom_metrics(
                user_model,
                seldon_metrics,
                OUTPUT_TRANSFORM_METRIC_METHOD_TAG,
                client_response.metrics,
            )

            return construct_response_json(
                user_model,
                False,
                request,
                client_response.data,
                meta,
                metrics,
                client_response.tags,
            )


def route(
    user_model: Any,
    request: Union[prediction_pb2.SeldonMessage, List, Dict],
    seldon_metrics: SeldonMetrics,
) -> Union[prediction_pb2.SeldonMessage, List, Dict]:
    """
    Parameters
    ----------
    user_model
       A Seldon user model
    request
       A SelodonMessage proto
    seldon_metrics
        A SeldonMetrics instance
    Returns
    -------
    """
    is_proto = isinstance(request, prediction_pb2.SeldonMessage)

    if hasattr(user_model, "route_rest"):
        logger.warning("route_rest is deprecated. Please use route_raw")
        return user_model.route_rest(request)
    elif hasattr(user_model, "route_grpc"):
        logger.warning("route_grpc is deprecated. Please use route_raw")
        return user_model.route_grpc(request)
    else:
        if hasattr(user_model, "route_raw"):
            try:
                response = user_model.route_raw(request)
                handle_raw_custom_metrics(
                    response, seldon_metrics, is_proto, ROUTER_METRIC_METHOD_TAG
                )
                return response
            except SeldonNotImplementedError:
                pass

        if is_proto:
            (features, meta, datadef, data_type) = extract_request_parts(request)
            client_response = client_route(
                user_model, features, datadef.names, meta=meta
            )
            if not isinstance(client_response.data, int):
                raise SeldonMicroserviceException(
                    "Routing response must be int but got " + str(client_response.data)
                )
            client_response_arr = np.array([[client_response.data]])

            metrics = client_custom_metrics(
                user_model,
                seldon_metrics,
                ROUTER_METRIC_METHOD_TAG,
                client_response.metrics,
            )

            return construct_response(
                user_model,
                False,
                request,
                client_response_arr,
                None,
                metrics,
                client_response.tags,
            )
        else:
            (features, meta, datadef, data_type) = extract_request_parts_json(request)
            class_names = datadef["names"] if datadef and "names" in datadef else []
            client_response = client_route(user_model, features, class_names, meta=meta)
            if not isinstance(client_response.data, int):
                raise SeldonMicroserviceException(
                    "Routing response must be int but got " + str(client_response.data)
                )
            client_response_arr = np.array([[client_response.data]])

            metrics = client_custom_metrics(
                user_model,
                seldon_metrics,
                ROUTER_METRIC_METHOD_TAG,
                client_response.metrics,
            )

            return construct_response_json(
                user_model,
                False,
                request,
                client_response_arr,
                None,
                metrics,
                client_response.tags,
            )


def aggregate(
    user_model: Any,
    request: Union[prediction_pb2.SeldonMessageList, List, Dict],
    seldon_metrics: SeldonMetrics,
) -> Union[prediction_pb2.SeldonMessage, List, Dict]:
    """
    Aggregate a list of payloads

    Parameters
    ----------
    user_model
       A Seldon user model
    request
       SeldonMessage proto
    seldon_metrics
        A SeldonMetrics instance

    Returns
    -------
       Aggregated SeldonMessage proto

    """

    def merge_meta(meta_list):
        tags = {}
        requestPath = {}
        for meta in meta_list:
            if meta:
                tags.update(meta.get("tags", {}))
                requestPath.update(meta.get("requestPath", {}))
        return {"tags": tags, "requestPath": requestPath}

    def merge_metrics(meta_list, custom_metrics):
        metrics = []
        for meta in meta_list:
            if meta:
                metrics.extend(meta.get("metrics", []))
        metrics.extend(custom_metrics)
        return metrics

    is_proto = isinstance(request, prediction_pb2.SeldonMessageList)

    if hasattr(user_model, "aggregate_rest"):
        logger.warning("aggregate_rest is deprecated. Please use aggregate_raw")
        return user_model.aggregate_rest(request)
    elif hasattr(user_model, "aggregate_grpc"):
        logger.warning("aggregate_grpc is deprecated. Please use aggregate_raw")
        return user_model.aggregate_grpc(request)
    else:
        if hasattr(user_model, "aggregate_raw"):
            try:
                response = user_model.aggregate_raw(request)
                handle_raw_custom_metrics(
                    response, seldon_metrics, is_proto, AGGREGATE_METRIC_METHOD_TAG
                )
                return response
            except SeldonNotImplementedError:
                pass

        if is_proto:
            features_list = []
            names_list = []
            meta_list = []

            for msg in request.seldonMessages:
                (features, meta, datadef, data_type) = extract_request_parts(msg)
                features_list.append(features)
                names_list.append(datadef.names)
                meta_list.append(meta)

            client_response = client_aggregate(user_model, features_list, names_list)

            metrics = client_custom_metrics(
                user_model,
                seldon_metrics,
                AGGREGATE_METRIC_METHOD_TAG,
                client_response.metrics,
            )

            return construct_response(
                user_model,
                False,
                request.seldonMessages[0],
                client_response.data,
                merge_meta(meta_list),
                merge_metrics(meta_list, metrics),
                client_response.tags,
            )
        else:
            features_list = []
            names_list = []

            if isinstance(request, list):
                msgs = request
            elif "seldonMessages" in request and isinstance(
                request["seldonMessages"], list
            ):
                msgs = request["seldonMessages"]
            else:
                raise SeldonMicroserviceException(
                    f"Invalid request data type: {request}"
                )

            meta_list = []
            for msg in msgs:
                (features, meta, datadef, data_type) = extract_request_parts_json(msg)
                class_names = datadef["names"] if datadef and "names" in datadef else []
                features_list.append(features)
                names_list.append(class_names)
                meta_list.append(meta)

            client_response = client_aggregate(user_model, features_list, names_list)

            metrics = client_custom_metrics(
                user_model,
                seldon_metrics,
                AGGREGATE_METRIC_METHOD_TAG,
                client_response.metrics,
            )

            return construct_response_json(
                user_model,
                False,
                msgs[0],
                client_response.data,
                merge_meta(meta_list),
                merge_metrics(meta_list, metrics),
                client_response.tags,
            )


def health_status(
    user_model: Any, seldon_metrics: SeldonMetrics
) -> Union[prediction_pb2.SeldonMessage, List, Dict]:
    """
    Call the user model to check the health of the model

    Parameters
    ----------
    user_model
       User defined class instance
    seldon_metrics
        A SeldonMetrics instance

    Returns
    -------
      Health check output
    """

    if hasattr(user_model, "health_status_raw"):
        try:
            return user_model.health_status_raw()
        except SeldonNotImplementedError:
            pass

    client_response = client_health_status(user_model)
    metrics = client_custom_metrics(
        user_model, seldon_metrics, HEALTH_METRIC_METHOD_TAG
    )

    return construct_response_json(
        user_model, False, {}, client_response, None, metrics
    )


def init_metadata(user_model: Any) -> Dict:
    """
    Call the user model to get the model init_metadata

    Parameters
    ----------
    user_model
        User defined class instance

    Returns
    -------
        Validated model metadata
    """
    # meta_user: load metadata defined in the user_model instance
    if hasattr(user_model, "init_metadata"):
        try:
            meta_user = user_model.init_metadata()
        except SeldonNotImplementedError:
            meta_user = {}
            pass
    else:
        meta_user = {}

    if not isinstance(meta_user, dict):
        logger.error("init_metadata must return dict")
        meta_user = {}

    # meta_env: load metadata from environmental variable
    try:
        meta_env = yaml.safe_load(os.environ.get("MODEL_METADATA", "{}"))
    except yaml.YAMLError as e:
        logger.error(f"Reading metadata from MODEL_METADATA env variable failed: {e}")
        meta_env = {}

    meta = {**meta_user, **meta_env}

    try:
        return validate_model_metadata(meta)
    except SeldonInvalidMetadataError as e:
        logger.error(f"Metadata validation error\n{e}")
        logger.error(f"Failed to validate metadata {meta}")
        return None
