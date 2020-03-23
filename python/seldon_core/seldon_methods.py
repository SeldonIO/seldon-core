import logging
from seldon_core.utils import (
    extract_request_parts,
    construct_response,
    json_to_seldon_message,
    seldon_message_to_json,
    construct_response_json,
    extract_request_parts_json,
    extract_feedback_request_parts,
)
from seldon_core.user_model import (
    client_predict,
    client_aggregate,
    client_route,
    client_custom_metrics,
    client_transform_output,
    client_transform_input,
    client_send_feedback,
    client_health_status,
    SeldonNotImplementedError,
)
from seldon_core.flask_utils import SeldonMicroserviceException
from seldon_core.metrics import SeldonMetrics
from google.protobuf import json_format
from seldon_core.proto import prediction_pb2
from typing import Any, Union, List, Dict
import numpy as np

logger = logging.getLogger(__name__)


def predict(
    user_model: Any,
    request: Union[prediction_pb2.SeldonMessage, List, Dict],
    seldon_metrics: SeldonMetrics,
) -> Union[prediction_pb2.SeldonMessage, List, Dict]:
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
                if is_proto:
                    metrics = seldon_message_to_json(response.meta).get("metrics", [])
                else:
                    metrics = response.get("meta", {}).get("metrics", [])
                seldon_metrics.update(metrics)
                return response
            except SeldonNotImplementedError:
                pass

        if is_proto:
            (features, meta, datadef, data_type) = extract_request_parts(request)
            client_response = client_predict(
                user_model, features, datadef.names, meta=meta
            )

            metrics = client_custom_metrics(user_model)
            if seldon_metrics is not None:
                seldon_metrics.update(metrics)

            return construct_response(
                user_model, False, request, client_response, meta, metrics
            )
        else:
            (features, meta, datadef, data_type) = extract_request_parts_json(request)
            class_names = datadef["names"] if datadef and "names" in datadef else []
            client_response = client_predict(
                user_model, features, class_names, meta=meta
            )

            metrics = client_custom_metrics(user_model)
            if seldon_metrics is not None:
                seldon_metrics.update(metrics)

            return construct_response_json(
                user_model, False, request, client_response, meta, metrics
            )


def send_feedback(
    user_model: Any, request: prediction_pb2.Feedback, predictive_unit_id: str
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
                return user_model.send_feedback_raw(request)
            except SeldonNotImplementedError:
                pass

        (datadef_request, features, truth, reward) = extract_feedback_request_parts(
            request
        )
        routing = request.response.meta.routing.get(predictive_unit_id)
        client_response = client_send_feedback(
            user_model, features, datadef_request.names, reward, truth, routing
        )

        if client_response is None:
            client_response = np.array([])
        else:
            client_response = np.array(client_response)
        return construct_response(user_model, False, request.request, client_response)


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
                if is_proto:
                    metrics = seldon_message_to_json(response.meta).get("metrics", [])
                else:
                    metrics = response.get("meta", {}).get("metrics", [])
                seldon_metrics.update(metrics)
                return response
            except SeldonNotImplementedError:
                pass

        if is_proto:
            (features, meta, datadef, data_type) = extract_request_parts(request)
            client_response = client_transform_input(
                user_model, features, datadef.names, meta=meta
            )

            metrics = client_custom_metrics(user_model)
            if seldon_metrics is not None:
                seldon_metrics.update(metrics)

            return construct_response(
                user_model, False, request, client_response, meta, metrics
            )
        else:
            (features, meta, datadef, data_type) = extract_request_parts_json(request)
            class_names = datadef["names"] if datadef and "names" in datadef else []
            client_response = client_transform_input(
                user_model, features, class_names, meta=meta
            )

            metrics = client_custom_metrics(user_model)
            if seldon_metrics is not None:
                seldon_metrics.update(metrics)

            return construct_response_json(
                user_model, False, request, client_response, meta, metrics
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
                if is_proto:
                    metrics = seldon_message_to_json(response.meta).get("metrics", [])
                else:
                    metrics = response.get("meta", {}).get("metrics", [])
                seldon_metrics.update(metrics)
                return response
            except SeldonNotImplementedError:
                pass

        if is_proto:
            (features, meta, datadef, data_type) = extract_request_parts(request)
            client_response = client_transform_output(
                user_model, features, datadef.names, meta=meta
            )

            metrics = client_custom_metrics(user_model)
            if seldon_metrics is not None:
                seldon_metrics.update(metrics)

            return construct_response(
                user_model, False, request, client_response, meta, metrics
            )
        else:
            (features, meta, datadef, data_type) = extract_request_parts_json(request)
            class_names = datadef["names"] if datadef and "names" in datadef else []
            client_response = client_transform_output(
                user_model, features, class_names, meta=meta
            )

            metrics = client_custom_metrics(user_model)
            if seldon_metrics is not None:
                seldon_metrics.update(metrics)

            return construct_response_json(
                user_model, False, request, client_response, meta, metrics
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
                if is_proto:
                    metrics = seldon_message_to_json(response.meta).get("metrics", [])
                else:
                    metrics = response.get("meta", {}).get("metrics", [])
                seldon_metrics.update(metrics)
                return response
            except SeldonNotImplementedError:
                pass

        if is_proto:
            (features, meta, datadef, data_type) = extract_request_parts(request)
            client_response = client_route(
                user_model, features, datadef.names, meta=meta
            )
            if not isinstance(client_response, int):
                raise SeldonMicroserviceException(
                    "Routing response must be int but got " + str(client_response)
                )
            client_response_arr = np.array([[client_response]])

            metrics = client_custom_metrics(user_model)
            if seldon_metrics is not None:
                seldon_metrics.update(metrics)

            return construct_response(
                user_model, False, request, client_response_arr, None, metrics
            )
        else:
            (features, meta, datadef, data_type) = extract_request_parts_json(request)
            class_names = datadef["names"] if datadef and "names" in datadef else []
            client_response = client_route(user_model, features, class_names, meta=meta)
            if not isinstance(client_response, int):
                raise SeldonMicroserviceException(
                    "Routing response must be int but got " + str(client_response)
                )
            client_response_arr = np.array([[client_response]])

            metrics = client_custom_metrics(user_model)
            if seldon_metrics is not None:
                seldon_metrics.update(metrics)

            return construct_response_json(
                user_model, False, request, client_response_arr, None, metrics
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

    Returns
    -------
       Aggregated SeldonMessage proto

    """

    def merge_meta(meta_list):
        tags = {}
        for meta in meta_list:
            if meta:
                tags.update(meta.get("tags", {}))
        return {"tags": tags}

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
                if is_proto:
                    metrics = seldon_message_to_json(response.meta).get("metrics", [])
                else:
                    metrics = response.get("meta", {}).get("metrics", [])
                seldon_metrics.update(metrics)
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

            metrics = client_custom_metrics(user_model)
            if seldon_metrics is not None:
                seldon_metrics.update(metrics)

            return construct_response(
                user_model,
                False,
                request.seldonMessages[0],
                client_response,
                merge_meta(meta_list),
                merge_metrics(meta_list, metrics),
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

            metrics = client_custom_metrics(user_model)
            if seldon_metrics is not None:
                seldon_metrics.update(metrics)

            return construct_response_json(
                user_model,
                False,
                msgs[0],
                client_response,
                merge_meta(meta_list),
                merge_metrics(meta_list, metrics),
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
    metrics = client_custom_metrics(user_model)
    if seldon_metrics is not None:
        seldon_metrics.update(metrics)

    return construct_response_json(
        user_model, False, {}, client_response, None, metrics
    )


def metadata(user_model: Any) -> Dict:
    """
    Call the user model to get the model metadata

    Parameters
    ----------
    user_model
       User defined class instance
    Returns
    -------
      Model Metadata
    """
    if hasattr(user_model, "metadata"):
        return user_model.metadata()
    else:
        return {}
