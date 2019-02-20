import tensorflow as tf
from seldon_core.proto import prediction_pb2
from seldon_core.proto import prediction_pb2_grpc
from seldon_core.utils import array_to_grpc_datadef, array_to_rest_datadef, seldon_message_to_json, \
    json_to_seldon_message, json_to_feedback, feedback_to_json, seldon_messages_to_json
import numpy as np
import grpc
import requests
from requests.auth import HTTPBasicAuth
from typing import Tuple, Dict, Union, List
import json


class SeldonClientException(Exception):
    status_code = 400

    def __init__(self, message):
        Exception.__init__(self)
        self.message = message


class SeldonClientPrediction(object):

    def __init__(self, request: prediction_pb2.SeldonMessage, response: Union[prediction_pb2.SeldonMessage, None],
                 success: bool = True, msg: str = ""):
        self.request = request
        self.response = response
        self.success = success
        self.msg = msg

    def __repr__(self):
        return "Success:%s message:%s\nRequest:\n%s\nResponse:\n%s" % (
            self.success, self.msg, self.request, self.response)


class SeldonClientFeedback(object):
    def __init__(self, request: prediction_pb2.Feedback, response: prediction_pb2.SeldonMessage, success: bool = True,
                 msg: str = ""):
        self.request = request
        self.response = response
        self.success = success
        self.msg = msg

    def __repr__(self):
        return "Success:%s message:%s\nRequest:\n%s\nResponse:\n%s" % (
            self.success, self.msg, self.request, self.response)


class SeldonClientCombine(object):
    def __init__(self, request: prediction_pb2.SeldonMessageList, response: prediction_pb2.SeldonMessage,
                 success: bool = True, msg: str = ""):
        self.request = request
        self.response = response
        self.success = success
        self.msg = msg

    def __repr__(self):
        return "Success:%s message:%s\nRequest:\n%s\nResponse:\n%s" % (
            self.success, self.msg, self.request, self.response)


class SeldonClient(object):
    """
    A reference Seldon API Client
    """

    def __init__(self, gateway: str = "ambassador", transport: str = "rest", deployment_name: str = None,
                 payload_type: str = "tensor", oauth_key: str = None, oauth_secret: str = None,
                 seldon_rest_endpoint: str = "localhost:8002", seldon_grpc_endpoint: str = "localhost:8004",
                 ambassador_endpoint: str = "localhost:8003", microservice_endpoint: str = "localhost:5000",
                 shape: Tuple = (1, 1), namespace: str = None, data: np.ndarray = None, datas: List[np.ndarray] = None,
                 ndatas: int = None):
        """

        Parameters
        ----------
        gateway
           ambassador or seldon
        transport
           rest or grpc
        deployment_name
           the name of the Seldon Deployment
        payload_type
           tensor, ndarray or tftensor
        oauth_key
        oauth_secret
        seldon_rest_endpoint
        seldon_grpc_endpoint
        ambassador_endpoint
        shape
           Shape of random numpy array to create if data not provided
        namespace
           The namespace of the running Seldon Deployment
        """
        self.config = {}
        self.config["gateway"] = gateway
        self.config["transport"] = transport
        self.config["payload_type"] = payload_type
        self.config["oauth_key"] = oauth_key
        self.config["oauth_secret"] = oauth_secret
        self.config["seldon_rest_endpoint"] = seldon_rest_endpoint
        self.config["seldon_grpc_endpoint"] = seldon_grpc_endpoint
        self.config["ambassador_endpoint"] = ambassador_endpoint
        self.config["shape"] = shape
        self.config["namespace"] = namespace
        self.config["deployment_name"] = deployment_name
        self.config["data"] = data
        self.config["microservice_endpoint"] = microservice_endpoint

    def predict(self, **kwargs):
        k = {**self.config, **kwargs}
        if k["gateway"] == "ambassador":
            if k["transport"] == "rest":
                return rest_predict_ambassador(**k)
            elif k["transport"] == "grpc":
                return grpc_predict_ambassador(**k)
            else:
                raise SeldonClientException("Unknown transport " + k["transport"])
        elif k["gateway"] == "seldon":
            if k["transport"] == "rest":
                return rest_predict_seldon_oauth(**k)
            elif k["transport"] == "grpc":
                return grpc_predict_seldon_oauth(**k)
            else:
                raise SeldonClientException("Unknown transport " + k["transport"])
        else:
            raise SeldonClientException("Unknown gateway " + k["gateway"])

    def feedback(self, prediction_request: prediction_pb2.SeldonMessage = None,
                 prediction_response: prediction_pb2.SeldonMessage = None, reward: float = 0, **kwargs):
        k = {**self.config, **kwargs}
        if k["gateway"] == "ambassador":
            if k["transport"] == "rest":
                return rest_feedback_ambassador(prediction_request, prediction_response, reward, **k)
            elif k["transport"] == "grpc":
                return grpc_feedback_ambassador(prediction_request, prediction_response, reward, **k)
            else:
                raise SeldonClientException("Unknown transport " + k["transport"])
        elif k["gateway"] == "seldon":
            if k["transport"] == "rest":
                return rest_feedback_seldon_oauth(prediction_request, prediction_response, reward, **k)
            elif k["transport"] == "grpc":
                return grpc_feedback_seldon_oauth(prediction_request, prediction_response, reward, **k)
            else:
                raise SeldonClientException("Unknown transport " + k["transport"])
        else:
            raise SeldonClientException("Unknown gateway " + k["gateway"])

    def microservice(self, **kwargs):
        k = {**self.config, **kwargs}
        if k["transport"] == "rest":
            if k["method"] == "predict" or k["method"] == "transform-input" or k["method"] == "transform-output" or k[
                "method"] == "route":
                return microservice_api_rest_seldon_message(**k)
            elif k["method"] == "aggregate":
                return microservice_api_rest_aggregate(**k)
            else:
                raise SeldonClientException("Unknown method " + k["method"])
        elif k["transport"] == "grpc":
            if k["method"] == "predict" or k["method"] == "transform-input" or k["method"] == "transform-output" or k[
                "method"] == "route":
                return microservice_api_grpc_seldon_message(**k)
            elif k["method"] == "aggregate":
                return microservice_api_grpc_aggregate(**k)
            else:
                raise SeldonClientException("Unknown method " + k["method"])
        else:
            raise SeldonClientException("Unknown transport " + k["transport"])

    def microservice_feedback(self, prediction_request: prediction_pb2.SeldonMessage = None,
                              prediction_response: prediction_pb2.SeldonMessage = None, reward: float = 0, **kwargs):
        k = {**self.config, **kwargs}
        if k["transport"] == "rest":
            return microservice_api_rest_feedback(prediction_request, prediction_response, reward, **k)
        else:
            return microservice_api_grpc_feedback(prediction_request, prediction_response, reward, **k)


def microservice_api_rest_seldon_message(method: str = "predict", microservice_endpoint: str = "localhost:5000",
                                         shape: Tuple = (1, 1),
                                         data: object = None, payload_type: str = "tensor", **kwargs):
    if data is None:
        data = np.random.rand(*shape)
    datadef = array_to_grpc_datadef(payload_type, data)
    request = prediction_pb2.SeldonMessage(data=datadef)
    payload = seldon_message_to_json(request)
    response_raw = requests.post(
        "http://" + microservice_endpoint + "/" + method,
        data={"json": json.dumps(payload)})
    if response_raw.status_code == 200:
        success = True
        msg = ""
    else:
        success = False
        msg = response_raw.reason
    try:
        response = json_to_seldon_message(response_raw.json())
        return SeldonClientPrediction(request, response, success, msg)
    except:
        return SeldonClientPrediction(request, None, success, msg)


def microservice_api_rest_aggregate(method: str = "predict", microservice_endpoint: str = "localhost:5000",
                                    shape: Tuple = (1, 1),
                                    datas: List[np.ndarray] = None, ndatas: int = None, payload_type: str = "tensor",
                                    **kwargs):
    if datas is None:
        datas = []
        for n in range(ndatas):
            data = np.random.rand(*shape)
            datas.append(data)
    msgs = []
    for data in datas:
        datadef = array_to_grpc_datadef(payload_type, data)
        msgs.append(prediction_pb2.SeldonMessage(data=datadef))
    request = prediction_pb2.SeldonMessageList(seldonMessages=msgs)
    payload = seldon_messages_to_json(request)
    response_raw = requests.post(
        "http://" + microservice_endpoint + "/aggregate",
        data={"json": json.dumps(payload)})
    if response_raw.status_code == 200:
        success = True
        msg = ""
    else:
        success = False
        msg = response_raw.reason
    try:
        response = json_to_seldon_message(response_raw.json())
        return SeldonClientCombine(request, response, success, msg)
    except:
        return SeldonClientCombine(request, None, success, msg)


def microservice_api_rest_feedback(prediction_request: prediction_pb2.SeldonMessage = None,
                                   prediction_response: prediction_pb2.SeldonMessage = None, reward: float = 0,
                                   microservice_endpoint: str = None, **kwargs):
    request = prediction_pb2.Feedback(request=prediction_request, response=prediction_response, reward=reward)
    payload = feedback_to_json(request)
    response_raw = requests.post(
        "http://" + microservice_endpoint + "/send-feedback",
        data={"json": json.dumps(payload)})
    if response_raw.status_code == 200:
        success = True
        msg = ""
    else:
        success = False
        msg = response_raw.reason
    try:
        response = json_to_seldon_message(response_raw.json())
        return SeldonClientFeedback(request, response, success, msg)
    except:
        return SeldonClientFeedback(request, None, success, msg)


def microservice_api_grpc_seldon_message(method: str = "predict", microservice_endpoint: str = "localhost:5000",
                                         shape: Tuple = (1, 1),
                                         data: object = None, payload_type: str = "tensor", **kwargs):
    if data is None:
        data = np.random.rand(*shape)
    datadef = array_to_grpc_datadef(payload_type, data)
    request = prediction_pb2.SeldonMessage(data=datadef)
    channel = grpc.insecure_channel(microservice_endpoint)
    try:
        if method == "predict":
            stub = prediction_pb2_grpc.ModelStub(channel)
            response = stub.Predict(request=request)
        elif method == "transform-input":
            stub = prediction_pb2_grpc.GenericStub(channel)
            response = stub.TransformInput(request=request)
        elif method == "transform-output":
            stub = prediction_pb2_grpc.GenericStub(channel)
            response = stub.TransformOutput(request=request)
        elif method == "route":
            stub = prediction_pb2_grpc.GenericStub(channel)
            response = stub.Route(request=request)

        return SeldonClientPrediction(request, response, True, "")
    except Exception as e:
        return SeldonClientPrediction(request, None, False, str(e))


def microservice_api_grpc_aggregate(method: str = "predict", microservice_endpoint: str = "localhost:5000",
                                    shape: Tuple = (1, 1),
                                    datas: List[np.ndarray] = None, ndatas: int = None, payload_type: str = "tensor",
                                    **kwargs):
    if datas is None:
        datas = []
        for n in range(ndatas):
            data = np.random.rand(*shape)
            datas.append(data)
    msgs = []
    for data in datas:
        datadef = array_to_grpc_datadef(payload_type, data)
        msgs.append(prediction_pb2.SeldonMessage(data=datadef))
    request = prediction_pb2.SeldonMessageList(seldonMessages=msgs)
    try:
        channel = grpc.insecure_channel(microservice_endpoint)
        stub = prediction_pb2_grpc.GenericStub(channel)
        response = stub.Aggregate(request=request)
        return SeldonClientCombine(request, response, True, "")
    except Exception as e:
        print("what")
        return SeldonClientCombine(request, None, False, str(e))


def microservice_api_grpc_feedback(prediction_request: prediction_pb2.SeldonMessage = None,
                                   prediction_response: prediction_pb2.SeldonMessage = None, reward: float = 0,
                                   microservice_endpoint: str = None, **kwargs):
    request = prediction_pb2.Feedback(request=prediction_request, response=prediction_response, reward=reward)
    try:
        channel = grpc.insecure_channel(microservice_endpoint)
        stub = prediction_pb2_grpc.GenericStub(channel)
        response = stub.SendFeedback(request=request)
        return SeldonClientFeedback(request, response, True, "")
    except:
        return SeldonClientFeedback(request, None, False, "")


def get_token(oauth_key: str = "", oauth_secret: str = "", namespace: str = None,
              endpoint: str = "localhost:8002") -> str:
    """
    Get an OAUTH key from the Seldon Gateway
    Parameters
    ----------
    oauth_key
    oauth_secret
    namespace
    endpoint

    Returns
    -------
       The OAUTH token

    """
    payload = {'grant_type': 'client_credentials'}
    if namespace is None:
        key = oauth_key
    else:
        key = oauth_key + namespace
    response = requests.post(
        "http://" + endpoint + "/oauth/token",
        auth=HTTPBasicAuth(key, oauth_secret),
        data=payload)
    print(response.text)
    token = response.json()["access_token"]
    return token


def rest_predict_seldon_oauth(oauth_key: str, oauth_secret: str, namespace: str = None,
                              seldon_rest_endpoint: str = "localhost:8002", shape: Tuple = (1, 1),
                              data: object = None, payload_type: str = "tensor", **kwargs) -> SeldonClientPrediction:
    """
    Call Seldon API Gateway using REST
    Parameters
    ----------
    oauth_key
    oauth_secret
    namespace
    seldon_rest_endpoint
    shape
    data

    Returns
    -------
       Request library response

    """
    token = get_token(oauth_key, oauth_secret, namespace, seldon_rest_endpoint)
    if data is None:
        data = np.random.rand(*shape)
    headers = {'Authorization': 'Bearer ' + token}
    datadef = array_to_grpc_datadef(payload_type, data)
    request = prediction_pb2.SeldonMessage(data=datadef)
    payload = seldon_message_to_json(request)
    response_raw = requests.post(
        "http://" + seldon_rest_endpoint + "/api/v0.1/predictions",
        headers=headers,
        json=payload)
    if response_raw.status_code == 200:
        success = True
        msg = ""
    else:
        success = False
        msg = response_raw.reason
    try:
        response = json_to_seldon_message(response_raw.json())
        return SeldonClientPrediction(request, response, success, msg)
    except:
        return SeldonClientPrediction(request, None, success, msg)


def grpc_predict_seldon_oauth(oauth_key: str, oauth_secret: str, namespace: str = None,
                              seldon_rest_endpoint: str = "localhost:8002",
                              seldpon_grpc_endpoint: str = "localhost:8004", shape: Tuple[int] = (1, 1),
                              data: np.ndarray = None, payload_type: str = "tensor",
                              **kwargs) -> SeldonClientPrediction:
    """
    Call Seldon gRPC API Gateway endpoint
    Parameters
    ----------
    oauth_key
    oauth_secret
    namespace
    seldon_rest_endpoint
    seldpon_grpc_endpoint
    shape
    data
    payload_type

    Returns
    -------
       A SeldonMessage proto

    """
    token = get_token(oauth_key, oauth_secret, namespace, seldon_rest_endpoint)
    if data is None:
        data = np.random.rand(*shape)
    datadef = array_to_grpc_datadef(payload_type, data)
    request = prediction_pb2.SeldonMessage(data=datadef)
    channel = grpc.insecure_channel(seldpon_grpc_endpoint)
    stub = prediction_pb2_grpc.SeldonStub(channel)
    metadata = [('oauth_token', token)]
    try:
        response = stub.Predict(request=request, metadata=metadata)
        return SeldonClientPrediction(request, response, True, "")
    except Exception as e:
        return SeldonClientPrediction(request, None, False, str(e))


def rest_predict_ambassador(deployment_name: str, namespace: str = None, ambassador_endpoint: str = "localhost:8003",
                            shape: Tuple[int] = (1, 1),
                            data: np.ndarray = None, headers: Dict = None, prefix: str = None,
                            payload_type: str = "tensor", **kwargs) -> SeldonClientPrediction:
    """
    REST request to Seldon Ambassador Ingress
    Parameters
    ----------
    deployment_name
    namespace
    ambassador_endpoint
    shape
    data
    headers
    prefix

    Returns
    -------
       A requests Response object

    """
    if data is None:
        data = np.random.rand(*shape)
    datadef = array_to_grpc_datadef(payload_type, data)
    request = prediction_pb2.SeldonMessage(data=datadef)
    payload = seldon_message_to_json(request)
    if prefix is None:
        if namespace is None:
            response_raw = requests.post(
                "http://" + ambassador_endpoint + "/seldon/" + deployment_name + "/api/v0.1/predictions",
                json=payload,
                headers=headers)
        else:
            response_raw = requests.post(
                "http://" + ambassador_endpoint + "/seldon/" + namespace + "/" + deployment_name + "/api/v0.1/predictions",
                json=payload,
                headers=headers)
    else:
        response_raw = requests.post(
            "http://" + ambassador_endpoint + prefix + "/api/v0.1/predictions",
            json=payload,
            headers=headers)

    if response_raw.status_code == 200:
        success = True
        msg = ""
    else:
        success = False
        msg = response_raw.reason
    try:
        response = json_to_seldon_message(response_raw.json())
        return SeldonClientPrediction(request, response, success, msg)
    except:
        return SeldonClientPrediction(request, None, success, msg)


def rest_predict_ambassador_basicauth(deployment_name: str, username: str, password: str, namespace: str = None,
                                      ambassador_endpoint: str = "localhost:8003",
                                      shape: Tuple[int] = (1, 1), data: np.ndarray = None, payload_type: str = "tensor",
                                      **kwargs) -> SeldonClientPrediction:
    """
    REST request with Basic Auth to Seldon Ambassador Ingress
    Parameters
    ----------
    deployment_name
    username
    password
    namespace
    ambassador_endpoint
    shape
    data

    Returns
    -------

    """
    if data is None:
        data = np.random.rand(*shape)
    datadef = array_to_grpc_datadef(payload_type, data)
    request = prediction_pb2.SeldonMessage(data=datadef)
    payload = seldon_message_to_json(request)
    if namespace is None:
        response_raw = requests.post(
            "http://" + ambassador_endpoint + "/seldon/" + deployment_name + "/api/v0.1/predictions",
            json=payload,
            auth=HTTPBasicAuth(username, password))
    else:
        response_raw = requests.post(
            "http://" + ambassador_endpoint + "/seldon/" + namespace + "/" + deployment_name + "/api/v0.1/predictions",
            json=payload,
            auth=HTTPBasicAuth(username, password))
    if response_raw.status_code == 200:
        success = True
        msg = ""
    else:
        success = False
        msg = response_raw.reason
    try:
        response = json_to_seldon_message(response_raw.json())
        return SeldonClientPrediction(request, response, success, msg)
    except:
        return SeldonClientPrediction(request, None, success, msg)


def grpc_predict_ambassador(deployment_name: str, namespace: str = None, ambassador_endpoint: str = "localhost:8003",
                            shape: Tuple[int] = (1, 1),
                            data: np.ndarray = None,
                            headers: Dict = None, payload_type: str = "tensor", **kwargs) -> SeldonClientPrediction:
    """
    gRPC request to Seldon Ambassador Ingress
    Parameters
    ----------
    deployment_name
    namespace
    ambassador_endpoint
    shape
    data
    headers
      A Dict of key value pairs to add to gRPC HTTP Headers
    payload_type

    Returns
    -------
       A SeldonMessage proto response

    """
    if data is None:
        data = np.random.rand(*shape)
    datadef = array_to_grpc_datadef(payload_type, data)
    request = prediction_pb2.SeldonMessage(data=datadef)
    channel = grpc.insecure_channel(ambassador_endpoint)
    stub = prediction_pb2_grpc.SeldonStub(channel)
    if namespace is None:
        metadata = [('seldon', deployment_name)]
    else:
        metadata = [('seldon', deployment_name), ('namespace', namespace)]
    if not headers is None:
        for k in headers:
            metadata.append((k, headers[k]))
    try:
        response = stub.Predict(request=request, metadata=metadata)
        return SeldonClientPrediction(request, response, True, "")
    except Exception as e:
        return SeldonClientPrediction(request, None, False, str(e))


def rest_feedback_seldon_oauth(prediction_request: prediction_pb2.SeldonMessage = None,
                               prediction_response: prediction_pb2.SeldonMessage = None, reward: float = 0,
                               oauth_key: str = "", oauth_secret: str = "", namespace: str = None,
                               seldon_rest_endpoint: str = "localhost:8002", **kwargs) -> SeldonClientFeedback:
    """
    Send Feedback to Seldon API Gateway using REST
    Parameters
    ----------
    prediction_request
    prediction_response
    reward
    oauth_key
    oauth_secret
    namespace
    seldon_rest_endpoint
    kwargs

    Returns
    -------

    """
    token = get_token(oauth_key, oauth_secret, namespace, seldon_rest_endpoint)
    headers = {'Authorization': 'Bearer ' + token}
    request = prediction_pb2.Feedback(request=prediction_request, response=prediction_response, reward=reward)
    payload = feedback_to_json(request)
    response_raw = requests.post(
        "http://" + seldon_rest_endpoint + "/api/v0.1/feedback",
        headers=headers,
        json=payload)
    if response_raw.status_code == 200:
        success = True
        msg = ""
    else:
        success = False
        msg = response_raw.reason
    try:
        response = json_to_seldon_message(response_raw.json())
        return SeldonClientFeedback(request, response, success, "")
    except:
        return SeldonClientFeedback(request, None, success, msg)


def grpc_feedback_seldon_oauth(prediction_request: prediction_pb2.SeldonMessage = None,
                               prediction_response: prediction_pb2.SeldonMessage = None, reward: float = 0,
                               oauth_key: str = "", oauth_secret: str = "", namespace: str = None,
                               seldon_rest_endpoint: str = "localhost:8002",
                               seldpon_grpc_endpoint: str = "localhost:8004", **kwargs) -> SeldonClientFeedback:
    """
    Send feedback to Seldon API gateway via gRPC
    Parameters
    ----------
    prediction_request
    prediction_response
    reward
    oauth_key
    oauth_secret
    namespace
    seldon_rest_endpoint
    seldpon_grpc_endpoint
    kwargs

    Returns
    -------

    """
    token = get_token(oauth_key, oauth_secret, namespace, seldon_rest_endpoint)
    request = prediction_pb2.Feedback(request=prediction_request, response=prediction_response, reward=reward)
    channel = grpc.insecure_channel(seldpon_grpc_endpoint)
    stub = prediction_pb2_grpc.SeldonStub(channel)
    metadata = [('oauth_token', token)]
    try:
        response = stub.SendFeedback(request=request, metadata=metadata)
        return SeldonClientFeedback(request, response, True, "")
    except Exception as e:
        return SeldonClientFeedback(request, None, False, str(e))


def rest_feedback_ambassador(prediction_request: prediction_pb2.SeldonMessage = None,
                             prediction_response: prediction_pb2.SeldonMessage = None, reward: float = 0,
                             deployment_name: str = "", namespace: str = None,
                             ambassador_endpoint: str = "localhost:8003", headers: Dict = None, prefix: str = None,
                             **kwargs) -> SeldonClientFeedback:
    """
    Send Feedback to Seldon via Ambassador using REST
    Parameters
    ----------
    prediction_request
    prediction_response
    reward
    deployment_name
    namespace
    ambassador_endpoint
    headers
    prefix
    kwargs

    Returns
    -------

    """
    request = prediction_pb2.Feedback(request=prediction_request, response=prediction_response, reward=reward)
    payload = feedback_to_json(request)
    if prefix is None:
        if namespace is None:
            response_raw = requests.post(
                "http://" + ambassador_endpoint + "/seldon/" + deployment_name + "/api/v0.1/feedback",
                json=payload,
                headers=headers)
        else:
            response_raw = requests.post(
                "http://" + ambassador_endpoint + "/seldon/" + namespace + "/" + deployment_name + "/api/v0.1/feedback",
                json=payload,
                headers=headers)
    else:
        response_raw = requests.post(
            "http://" + ambassador_endpoint + prefix + "/api/v0.1/feedback",
            json=payload,
            headers=headers)

    if response_raw.status_code == 200:
        success = True
        msg = ""
    else:
        success = False
        msg = response_raw.reason
    try:
        response = json_to_seldon_message(response_raw.json())
        return SeldonClientFeedback(request, response, success, msg)
    except:
        return SeldonClientFeedback(request, None, success, msg)


def grpc_feedback_ambassador(prediction_request: prediction_pb2.SeldonMessage = None,
                             prediction_response: prediction_pb2.SeldonMessage = None, reward: float = 0,
                             deployment_name: str = "", namespace: str = None,
                             ambassador_endpoint: str = "localhost:8003",
                             headers: Dict = None, **kwargs) -> SeldonClientFeedback:
    """

    Parameters
    ----------
    prediction_request
    prediction_response
    reward
    deployment_name
    namespace
    ambassador_endpoint
    headers
    kwargs

    Returns
    -------

    """
    request = prediction_pb2.Feedback(request=prediction_request, response=prediction_response, reward=reward)
    channel = grpc.insecure_channel(ambassador_endpoint)
    stub = prediction_pb2_grpc.SeldonStub(channel)
    if namespace is None:
        metadata = [('seldon', deployment_name)]
    else:
        metadata = [('seldon', deployment_name), ('namespace', namespace)]
    if not headers is None:
        for k in headers:
            metadata.append((k, headers[k]))
    try:
        response = stub.SendFeedback(request=request, metadata=metadata)
        return SeldonClientFeedback(request, response, True, "")
    except Exception as e:
        return SeldonClientFeedback(request, None, False, str(e))
