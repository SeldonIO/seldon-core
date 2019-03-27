import json
from google.protobuf import json_format
from seldon_core.proto import prediction_pb2
from seldon_core.flask_utils import SeldonMicroserviceException
import numpy as np
import sys
import tensorflow as tf
from google.protobuf.struct_pb2 import ListValue
from seldon_core.user_model import client_class_names, client_custom_metrics, client_custom_tags, client_feature_names, \
    SeldonComponent
from typing import Tuple, Dict, Union, List, Optional, Iterable


def json_to_seldon_message(message_json: Dict) -> prediction_pb2.SeldonMessage:
    """
    Parses JSON input to a SeldonMessage proto

    Parameters
    ----------
    message_json
       JSON input

    Returns
    -------
      SeldonMessage
    """
    if message_json is None:
        message_json = {}
    message_proto = prediction_pb2.SeldonMessage()
    try:
        json_format.ParseDict(message_json, message_proto)
        return message_proto
    except json_format.ParseError as pbExc:
        raise SeldonMicroserviceException("Invalid JSON: " + str(pbExc))


def json_to_feedback(message_json: Dict) -> prediction_pb2.Feedback:
    """
    Parse a JSON message to a Feedback proto

    Parameters
    ----------
    message_json
       Input json message
    Returns
    -------
       A SeldonMessage
    """
    message_proto = prediction_pb2.Feedback()
    try:
        json_format.ParseDict(message_json, message_proto)
        return message_proto
    except json_format.ParseError as pbExc:
        raise SeldonMicroserviceException("Invalid JSON: " + str(pbExc))


def json_to_seldon_messages(message_json: Dict) -> prediction_pb2.SeldonMessageList:
    message_proto = prediction_pb2.SeldonMessageList()
    try:
        json_format.ParseDict(message_json, message_proto)
        return message_proto
    except json_format.ParseError as pbExc:
        raise SeldonMicroserviceException("Invalid JSON: " + str(pbExc))


def seldon_message_to_json(message_proto: prediction_pb2.SeldonMessage) -> Dict:
    """
    Convert a SeldonMessage proto to JSON Dict

    Parameters
    ----------
    message_proto
       SeldonMessage proto
    Returns
    -------
       JSON Dict
    """
    message_json = json_format.MessageToJson(message_proto)
    message_dict = json.loads(message_json)
    return message_dict


def seldon_messages_to_json(message_protos: prediction_pb2.SeldonMessageList) -> Dict:
    """
    Convert a SeldonMessage proto list to JSON Dict

    Parameters
    ----------
    message_protos
       SeldonMessage protos
    Returns
    -------
       JSON Dict
    """
    message_json = json_format.MessageToJson(message_protos)
    message_dict = json.loads(message_json)
    return message_dict


def feedback_to_json(message_proto: prediction_pb2.Feedback) -> Dict:
    """
    Convert a SeldonMessage proto to JSON Dict

    Parameters
    ----------
    message_proto
       SeldonMessage proto
    Returns
    -------
       JSON Dict
    """
    message_json = json_format.MessageToJson(message_proto)
    message_dict = json.loads(message_json)
    return message_dict


def get_data_from_proto(request: prediction_pb2.SeldonMessage) -> Union[np.ndarray, str, bytes]:
    """
    Extract the data payload from the SeldonMessage

    Parameters
    ----------
    request
       SeldonMessage

    Returns
    -------
       Data payload as numpy array or the raw message format. Numpy array will be returned if the "data" field was used.

    """
    data_type = request.WhichOneof("data_oneof")
    if data_type == "data":
        datadef = request.data
        return grpc_datadef_to_array(datadef)
    elif data_type == "binData":
        return request.binData
    elif data_type == "strData":
        return request.strData
    else:
        raise SeldonMicroserviceException("Unknown data in SeldonMessage")


def grpc_datadef_to_array(datadef: prediction_pb2.DefaultData) -> np.ndarray:
    """
    Convert a SeldonMessage DefaultData to a numpy array.

    Parameters
    ----------
    datadef
       SeldonMessage DefaultData

    Returns
    -------
       A numpy array

    """
    data_type = datadef.WhichOneof("data_oneof")
    if data_type == "tensor":
        if sys.version_info >= (3, 0):
            sz = np.prod(datadef.tensor.shape)  # get number of float64 entries
            c = datadef.tensor.SerializeToString()  # get bytes
            # create array from packed entries which are at end of bytes - assumes same endianness
            features = np.frombuffer(memoryview(
                c[-(sz * 8):]), dtype=np.float64, count=sz, offset=0)
            features = features.reshape(datadef.tensor.shape)
        else:
            # Python 2 version which is slower
            features = np.array(datadef.tensor.values).reshape(
                datadef.tensor.shape)
    elif data_type == "ndarray":
        features = np.array(datadef.ndarray)
    elif data_type == "tftensor":
        features = tf.make_ndarray(datadef.tftensor)
    else:
        features = np.array([])
    return features


def get_meta_from_proto(request: prediction_pb2.SeldonMessage) -> Dict:
    """
    Convert SeldonMessage proto meta into Dict

    Parameters
    ----------
    request
       SeldonMessage proto

    Returns
    -------
       Dict

    """
    meta = json_format.MessageToDict(request.meta)
    return meta


def array_to_rest_datadef(data_type: str, array: np.ndarray, names: Optional[List[str]] = []) -> Dict:
    """
    Construct a payload Dict from a numpy array

    Parameters
    ----------
    data_type
    array
    names

    Returns
    -------
       Dict representing Seldon payload

    """
    datadef: Dict = {"names": names}
    if data_type == "tensor":
        datadef["tensor"] = {
            "shape": array.shape,
            "values": array.ravel().tolist()
        }
    elif data_type == "ndarray":
        datadef["ndarray"] = array.tolist()
    elif data_type == "tftensor":
        tftensor = tf.make_tensor_proto(array)
        jStrTensor = json_format.MessageToJson(tftensor)
        jTensor = json.loads(jStrTensor)
        datadef["tftensor"] = jTensor
    else:
        datadef["ndarray"] = array.tolist()
    return datadef


def array_to_grpc_datadef(data_type: str, array: np.ndarray,
                          names: Optional[Iterable[str]] = []) -> prediction_pb2.DefaultData:
    """
    Convert numpy array and optional column names into a SeldonMessage DefaultData proto

    Parameters
    ----------
    array
       numpy array
    names
       column names
    data_type
       The SeldonMessage type to convert to

    Returns
    -------
       SeldonMessage DefaultData

    """
    if data_type == "tensor":
        datadef = prediction_pb2.DefaultData(
            names=names,
            tensor=prediction_pb2.Tensor(
                shape=array.shape,
                values=array.ravel().tolist()
            )
        )
    elif data_type == "ndarray":
        datadef = prediction_pb2.DefaultData(
            names=names,
            ndarray=array_to_list_value(array)
        )
    elif data_type == "tftensor":
        datadef = prediction_pb2.DefaultData(
            names=names,
            tftensor=tf.make_tensor_proto(array)
        )
    else:
        datadef = prediction_pb2.DefaultData(
            names=names,
            ndarray=array_to_list_value(array)
        )

    return datadef


def array_to_list_value(array: np.ndarray, lv: Optional[ListValue] = None) -> ListValue:
    """
    Construct a proto ListValue from numpy array

    Parameters
    ----------
    array
       Numpy array
    lv
       Proto buffer ListValue to extend

    Returns
    -------

    """
    if lv is None:
        lv = ListValue()
    if len(array.shape) == 1:
        lv.extend(array.tolist())
    else:
        for sub_array in array:
            sub_lv = lv.add_list()
            array_to_list_value(sub_array, sub_lv)
    return lv


def construct_response(user_model: SeldonComponent, is_request: bool, client_request: prediction_pb2.SeldonMessage,
                       client_raw_response: Union[np.ndarray, str, bytes]) -> prediction_pb2.SeldonMessage:
    """

    Parameters
    ----------
    user_model
       Client user class
    is_request
       Whether this is part of the request flow as opposed to the response flow
    client_request
       The request received
    client_raw_response
       The raw client response from their model

    Returns
    -------
       A SeldonMessage proto response

    """
    data_type = client_request.WhichOneof("data_oneof")
    meta = prediction_pb2.Meta()
    meta_json: Dict = {}
    tags = client_custom_tags(user_model)
    if tags:
        meta_json["tags"] = tags
    metrics = client_custom_metrics(user_model)
    if metrics:
        meta_json["metrics"] = metrics
    json_format.ParseDict(meta_json, meta)
    if isinstance(client_raw_response, np.ndarray) or isinstance(client_raw_response, list):
        client_raw_response = np.array(client_raw_response)
        if is_request:
            names = client_feature_names(user_model, client_request.data.names)
        else:
            names = client_class_names(user_model, client_raw_response)
        if data_type == "data":  # If request is using defaultdata then return what was sent if is numeric response else ndarray
            if np.issubdtype(client_raw_response.dtype, np.number):
                default_data_type = client_request.data.WhichOneof("data_oneof")
            else:
                default_data_type = "ndarray"
        else:  # If numeric response return as tensor else return as ndarray
            if np.issubdtype(client_raw_response.dtype, np.number):
                default_data_type = "tensor"
            else:
                default_data_type = "ndarray"
        data = array_to_grpc_datadef(default_data_type, client_raw_response, names)
        return prediction_pb2.SeldonMessage(data=data, meta=meta)
    elif isinstance(client_raw_response, str):
        return prediction_pb2.SeldonMessage(strData=client_raw_response, meta=meta)
    elif isinstance(client_raw_response, (bytes, bytearray)):
        return prediction_pb2.SeldonMessage(binData=client_raw_response, meta=meta)
    else:
        raise SeldonMicroserviceException("Unknown data type returned as payload:" + client_raw_response)


def extract_request_parts(request: prediction_pb2.SeldonMessage) -> Tuple[
    Union[np.ndarray, str, bytes], Dict, prediction_pb2.DefaultData, str]:
    """

    Parameters
    ----------
    request
       Input request

    Returns
    -------
       Key parts of the request extracted

    """
    features = get_data_from_proto(request)
    meta = get_meta_from_proto(request)
    datadef = request.data
    data_type = request.WhichOneof("data_oneof")
    return features, meta, datadef, data_type


def extract_feedback_request_parts(request: prediction_pb2.Feedback) -> Tuple[
    prediction_pb2.DefaultData, np.ndarray, np.ndarray, float]:
    """
    Extract key parts of the Feedback Message

    Parameters
    ----------
    request
       Feedback proto

    Returns
    -------
       Tuple of parts including extracted payloads

    """
    features = grpc_datadef_to_array(request.request.data)
    truth = grpc_datadef_to_array(request.truth.data)
    reward = request.reward
    return request.request.data, features, truth, reward
