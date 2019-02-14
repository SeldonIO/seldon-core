import logging
from seldon_core.utils import *
from seldon_core.user_model import *
from google.protobuf import json_format
from seldon_core.proto import prediction_pb2

logger = logging.getLogger(__name__)


def predict(user_model: object, request: prediction_pb2.SeldonMessage) -> prediction_pb2.SeldonMessage:
    '''
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
    '''
    if hasattr(user_model, "predict_rest"):
        logger.warning("predict_rest is deprecated. Please use predict_raw")
        requestJson = json_format.MessageToJson(request)
        responseJson = user_model.predict_rest(requestJson)
        return json_to_seldonMessage(responseJson)
    elif hasattr(user_model, "predict_grpc"):
        logger.warning("predict_grpc is deprecated. Please use predict_raw")
        return user_model.predict_grpc(request)
    elif hasattr(user_model, "predict_raw"):
        return user_model.predict_raw(request)
    else:
        (features, meta, datadef, data_type) = extract_request_parts(request)
        client_response = client_predict(user_model, features, datadef.names, meta=meta)

        return construct_response(user_model, True, request, client_response)


def send_feedback(user_model: object, request: prediction_pb2.Feedback) -> prediction_pb2.SeldonMessage:
    '''

    Parameters
    ----------
    user_model
    request

    Returns
    -------

    '''
    if hasattr(user_model, "send_feedback_rest"):
        logger.warning("send_feedback_rest is deprecated. Please use send_feedback_raw")
        requestJson = json_format.MessageToJson(request)
        responseJson = user_model.send_feedback_rest(requestJson)
        return json_to_seldonMessage(responseJson)
    elif hasattr(user_model, "send_feedback_raw"):
        responseJson = user_model.send_feedback_raw(request)
        return json_to_seldonMessage(responseJson)
    else:
        (datadef_request,features,truth,reward) = extract_feedback_request_parts(request)

        client_response = client_send_feedback(user_model, features, datadef_request.names, reward, truth)
        if client_response is None:
            client_response = np.array([])
        return construct_response(user_model, True, request.request, client_response)


def transform_input(user_model: object, request: prediction_pb2.SeldonMessage) -> prediction_pb2.SeldonMessage:
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
    if hasattr(user_model, "transform_input_rest"):
        logger.warning("transform_input_rest is deprecated. Please use transform_input_raw")
        requestJson = json_format.MessageToJson(request)
        responseJson = user_model.transform_input_rest(requestJson)
        return json_to_seldonMessage(responseJson)
    if hasattr(user_model, "transform_input_grpc"):
        logger.warning("transform_input_grpc is deprecated. Please use transform_input_raw")
        return user_model.transform_input_grpc(request)
    if hasattr(user_model, "transform_input_raw"):
        return user_model.transform_input_grpc(request)
    else:
        (features, meta, datadef, data_type) = extract_request_parts(request)
        client_response = client_transform_input(user_model, features, datadef.names, meta=meta)

        return construct_response(user_model, False, request,client_response)


def transform_output(user_model: object, request: prediction_pb2.SeldonMessage) -> prediction_pb2.SeldonMessage:
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
    if hasattr(user_model, "transform_output_rest"):
        logger.warning("transform_input_rest is deprecated. Please use transform_input_raw")
        requestJson = json_format.MessageToJson(request)
        responseJson = user_model.transform_output_rest(requestJson)
        return json_to_seldonMessage(responseJson)
    if hasattr(user_model, "transform_output_grpc"):
        logger.warning("transform_input_grpc is deprecated. Please use transform_input_raw")
        return user_model.transform_output_grpc(request)
    if hasattr(user_model, "transform_output_raw"):
        return user_model.transform_output_grpc(request)
    else:
        (features, meta, datadef, data_type) = extract_request_parts(request)
        client_response = client_transform_output(user_model, features, datadef.names, meta=meta)

        return construct_response(user_model, False, request, client_response)