import json
from google.protobuf import json_format
from google.protobuf.json_format import MessageToDict, ParseDict
from seldon_core.proto import prediction_pb2
from seldon_core.flask_utils import SeldonMicroserviceException
from tensorflow.core.framework.tensor_pb2 import TensorProto
import numpy as np
import sys
import tensorflow as tf
from google.protobuf.struct_pb2 import ListValue
from seldon_core.user_model import client_class_names, client_custom_metrics, client_custom_tags, client_feature_names, \
    SeldonComponent
from typing import Tuple, Dict, Union, List, Optional, Iterable
import base64


def json_to_seldon_message(message_json: Union[List, Dict]) -> prediction_pb2.SeldonMessage:
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


def get_data_from_proto(request: prediction_pb2.SeldonMessage) -> Union[np.ndarray, str, bytes, dict]:
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
    elif data_type == "jsonData":
        return MessageToDict(request.jsonData)
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
        py_arr = json_format.MessageToDict(datadef.ndarray)
        features = np.array(py_arr)
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

def construct_response_json(
        user_model: SeldonComponent,
        is_request: bool,
        client_request_raw: Union[List, Dict],
        client_raw_response: Union[np.ndarray, str, bytes, dict]) -> Union[List, Dict]:
    """
    This class converts a raw REST response into a JSON object that has the same structure as
    the SeldonMessage proto. This is necessary as the conversion using the SeldonMessage proto
    changes the Numeric types of all ints in a JSON into Floats.

    Parameters
    ----------
    user_model
       Client user class
    is_request
       Whether this is part of the request flow as opposed to the response flow
    client_request_raw
       The request received in JSON format
    client_raw_response
       The raw client response from their model

    Returns
    -------
       A SeldonMessage JSON response

    """
    response = {}

    if "jsonData" in client_request_raw:
        response["jsonData"] = client_raw_response
    elif isinstance(client_raw_response, (bytes, bytearray)):
        base64_data = base64.b64encode(client_raw_response)
        response["binData"] = base64_data.decode("utf-8")
    elif isinstance(client_raw_response, str):
        response["strData"] = client_raw_response
    else:
        is_np = isinstance(client_raw_response, np.ndarray)
        is_list = isinstance(client_raw_response, list)
        if not (is_np or is_list):
            raise SeldonMicroserviceException(
                "Unknown data type returned as payload (must be list or np array):"
                    + str(client_raw_response))
        if is_np:
            np_client_raw_response = client_raw_response
            list_client_raw_response = client_raw_response.tolist()
        else:
            np_client_raw_response = np.array(client_raw_response)
            list_client_raw_response = client_raw_response

        response["data"] = {}
        if "data" in client_request_raw:
            if np.issubdtype(np_client_raw_response.dtype, np.number):
                if "tensor" in client_request_raw["data"]:
                    default_data_type = "tensor"
                    result_client_response = {
                        "values": np_client_raw_response.ravel().tolist(),
                        "shape": np_client_raw_response.shape
                    }
                elif "tftensor" in client_request_raw["data"]:
                    default_data_type = "tftensor"
                    tf_json_str = json_format.MessageToJson(
                            tf.make_tensor_proto(np_client_raw_response))
                    result_client_response = json.loads(tf_json_str)
                else:
                    default_data_type = "ndarray"
                    result_client_response = list_client_raw_response
            else:
                default_data_type = "ndarray"
                result_client_response = list_client_raw_response
        else:
            if np.issubdtype(np_client_raw_response.dtype, np.number):
                default_data_type = "tensor"
                result_client_response = {
                    "values": np_client_raw_response.ravel().tolist(),
                    "shape": np_client_raw_response.shape
                }
            else:
                default_data_type = "ndarray"
                result_client_response = list_client_raw_response

        response["data"][default_data_type] = result_client_response

        if is_request:
            req_names = client_request_raw.get("data", {}).get("names", [])
            names = client_feature_names(user_model, req_names)
        else:
            names = client_class_names(user_model, np_client_raw_response)
        response["data"]["names"] = names

    response["meta"] = {}
    client_custom_tags(user_model)
    tags = client_custom_tags(user_model)
    if tags:
        response["meta"]["tags"] = tags
    metrics = client_custom_metrics(user_model)
    if metrics:
        response["meta"]["metrics"] = metrics
    puid = client_request_raw.get("meta", {}).get("puid", None)
    if puid:
        response["meta"]["puid"] = puid

    return response


def construct_response(user_model: SeldonComponent, is_request: bool, client_request: prediction_pb2.SeldonMessage,
                       client_raw_response: Union[np.ndarray, str, bytes, dict]) -> prediction_pb2.SeldonMessage:
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
    if client_request.meta:
        if client_request.meta.puid:
            meta_json["puid"] = client_request.meta.puid
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
    elif isinstance(client_raw_response, dict):
        jsonDataResponse = ParseDict(client_raw_response, prediction_pb2.SeldonMessage().jsonData)
        return prediction_pb2.SeldonMessage(jsonData=jsonDataResponse, meta=meta)
    elif isinstance(client_raw_response, (bytes, bytearray)):
        return prediction_pb2.SeldonMessage(binData=client_raw_response, meta=meta)
    else:
        raise SeldonMicroserviceException("Unknown data type returned as payload:" + client_raw_response)


def extract_request_parts_json(request: Union[Dict, List]
       ) -> Tuple[
           Union[np.ndarray, str, bytes, Dict, List],
           Union[Dict, None],
           Union[np.ndarray, str, bytes, Dict, List, None],
           str]:
    """

    Parameters
    ----------
    request
       Input request in JSON format

    Returns
    -------
       Key parts of the request extracted

    """
    if not isinstance(request, dict):
        raise SeldonMicroserviceException(f"Invalid request data type: {request}")
    meta = request.get("meta", None)
    datadef_type = None
    datadef = None

    if "data" in request:
        data_type = "data"
        datadef = request["data"]
        if "tensor" in datadef:
            datadef_type = "tensor"
            tensor = datadef["tensor"]
            features = np.array(tensor["values"]).reshape(tensor["shape"])
        elif "ndarray" in datadef:
            datadef_type = "ndarray"
            features = np.array(datadef["ndarray"])
        elif "tftensor" in datadef:
            datadef_type = "tftensor"
            tf_proto = TensorProto()
            json_format.ParseDict(datadef["tftensor"], tf_proto)
            features = tf.make_ndarray(tf_proto)
        else:
            features = np.array([])
    elif "jsonData" in request:
        data_type = "jsonData"
        features = request["jsonData"]
    elif "strData" in request:
        data_type = "strData"
        features = request["strData"]
    elif "binData" in request:
        data_type = "binData"
        features = bytes(request["binData"], "utf8")
    else:
        raise SeldonMicroserviceException(f"Invalid request data type: {request}")

    return features, meta, datadef, data_type

def extract_request_parts(request: prediction_pb2.SeldonMessage) -> Tuple[
    Union[np.ndarray, str, bytes, dict], Dict, prediction_pb2.DefaultData, str]:
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
