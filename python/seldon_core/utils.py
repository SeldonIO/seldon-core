import json
from flask import request
from google.protobuf import json_format
from seldon_core.proto import prediction_pb2
from seldon_core.microservice import SeldonMicroserviceException
import numpy as np
import sys
import tensorflow as tf
from google.protobuf.struct_pb2 import ListValue
from seldon_core.user_model import client_class_names, client_custom_metrics, client_custom_tags, client_feature_names
from typing import Tuple, Dict


def get_request():
    """ Parse a REST request into a SeldonMessage proto buffer
    """
    jStr = request.form.get("json")
    if jStr:
        message = json.loads(jStr)
    else:
        jStr = request.args.get('json')
        if jStr:
            message = json.loads(jStr)
        else:
            raise SeldonMicroserviceException("Empty json parameter in data")
    if message is None:
        raise SeldonMicroserviceException("Invalid Data Format")
    return message

def json_to_seldonMessage(messageJson):
    if messageJson is None:
        messageJson = {}
    messageProto = prediction_pb2.SeldonMessage()
    try:
        json_format.ParseDict(messageJson, messageProto)
        return messageProto
    except json_format.ParseError as pbExc:
        raise SeldonMicroserviceException("Invalid JSON: "+str(pbExc)) 

def json_to_feedback(messageJson):
    messageProto = prediction_pb2.Feedback()
    try:
        json_format.ParseDict(messageJson, messageProto)
        return messageProto
    except json_format.ParseError as pbExc:
        raise SeldonMicroserviceException("Invalid JSON: "+str(pbExc))


def seldonMessage_to_json(messageProto):
    messageJson = json_format.MessageToJson(messageProto)
    messageDict = json.loads(messageJson)
    return messageDict


def get_data_from_proto(request: prediction_pb2.SeldonMessage) -> object:
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

def grpc_datadef_to_array(datadef):
    data_type = datadef.WhichOneof("data_oneof")
    if data_type == "tensor":
        if (sys.version_info >= (3, 0)):
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


def get_meta_from_proto(request):
    meta = json_format.MessageToDict(request.meta)
    return meta


def array_to_grpc_datadef(array, names, data_type):
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

def array_to_list_value(array, lv=None):
    if lv is None:
        lv = ListValue()
    if len(array.shape) == 1:
        lv.extend(array)
    else:
        for sub_array in array:
            sub_lv = lv.add_list()
            array_to_list_value(sub_array, sub_lv)
    return lv


def construct_response(user_model: object, is_request: bool, client_request: prediction_pb2.SeldonMessage, client_raw_response: np.ndarray) -> prediction_pb2.SeldonMessage:
    '''

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

    '''
    data_type = client_request.WhichOneof("data_oneof")
    meta = prediction_pb2.Meta()
    metaJson = {}
    tags = client_custom_tags(user_model)
    if tags:
        metaJson["tags"] = tags
    metrics = client_custom_metrics(user_model)
    if metrics:
        metaJson["metrics"] = metrics
    json_format.ParseDict(metaJson, meta)
    if isinstance(client_raw_response, np.ndarray) or data_type == "data":
        client_raw_response = np.array(client_raw_response)
        if is_request:
            #names = client_request.get("data", {}).get("names")
            names = client_feature_names(user_model, client_request.data.names)
        else:
            names = client_class_names(user_model, client_raw_response)
        if data_type == "data":
            default_data_type = client_request.data.WhichOneof("data_oneof")
        else:
            default_data_type = "tensor"
        data = array_to_grpc_datadef(
            client_raw_response, names, default_data_type)
        return prediction_pb2.SeldonMessage(data=data, meta=meta)
    else:
        if isinstance(client_raw_response, str):
            return prediction_pb2.SeldonMessage(strData=client_raw_response, meta=meta)
        else:
            return prediction_pb2.SeldonMessage(binData=client_raw_response, meta=meta)


def extract_request_parts(request: prediction_pb2.SeldonMessage) -> Tuple[object,Dict,object,str]:
    '''

    Parameters
    ----------
    request
       Input request

    Returns
    -------
       Key parts of the request extracted

    '''
    features = get_data_from_proto(request)
    meta = get_meta_from_proto(request)
    datadef = request.data
    data_type = request.WhichOneof("data_oneof")
    return (features,meta,datadef,data_type)


def extract_feedback_request_parts(request: prediction_pb2.Feedback) -> Tuple[object,np.ndarray,prediction_pb2.SeldonMessage,float]:
    datadef_request = request.request.data
    features = grpc_datadef_to_array(datadef_request)
    truth = grpc_datadef_to_array(request.truth)
    reward = request.reward
    return (datadef_request,features,truth,reward)
