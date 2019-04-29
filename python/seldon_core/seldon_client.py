from seldon_core.proto import prediction_pb2
from seldon_core.proto import prediction_pb2_grpc
from seldon_core.utils import array_to_grpc_datadef, seldon_message_to_json, \
    json_to_seldon_message, feedback_to_json, seldon_messages_to_json
import numpy as np
import grpc
import requests
from requests.auth import HTTPBasicAuth
from typing import Tuple, Dict, Union, List, Optional, Iterable
import json
import logging

logger = logging.getLogger(__name__)


class SeldonClientException(Exception):
    """
    Seldon Client Exception
    """
    status_code = 400

    def __init__(self, message):
        Exception.__init__(self)
        self.message = message


class SeldonClientPrediction(object):
    """
    Data class to return from Seldon Client
    """

    def __init__(self, request: Optional[prediction_pb2.SeldonMessage],
                 response: Optional[prediction_pb2.SeldonMessage],
                 success: bool = True, msg: str = ""):
        self.request = request
        self.response = response
        self.success = success
        self.msg = msg

    def __repr__(self):
        return "Success:%s message:%s\nRequest:\n%s\nResponse:\n%s" % (
            self.success, self.msg, self.request, self.response)


class SeldonClientFeedback(object):
    """
    Data class to return from Seldon Client for feedback calls
    """

    def __init__(self, request: Optional[prediction_pb2.Feedback], response: Optional[prediction_pb2.SeldonMessage],
                 success: bool = True,
                 msg: str = ""):
        self.request = request
        self.response = response
        self.success = success
        self.msg = msg

    def __repr__(self):
        return "Success:%s message:%s\nRequest:\n%s\nResponse:\n%s" % (
            self.success, self.msg, self.request, self.response)


class SeldonClientCombine(object):
    """
    Data class to return from Seldon Client for aggregate calls
    """

    def __init__(self, request: Optional[prediction_pb2.SeldonMessageList],
                 response: Optional[prediction_pb2.SeldonMessage],
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

    def __init__(self, gateway: str = "ambassador", transport: str = "rest", namespace: str = None,
                 deployment_name: str = None,
                 payload_type: str = "tensor", oauth_key: str = None, oauth_secret: str = None,
                 seldon_rest_endpoint: str = "localhost:8002", seldon_grpc_endpoint: str = "localhost:8004",
                 ambassador_endpoint: str = "localhost:8003", microservice_endpoint: str = "localhost:5000",
                 grpc_max_send_message_length: int = 4 * 1024 * 1024,
                 grpc_max_receive_message_length: int = 4 * 1024 * 1024):
        """

        Parameters
        ----------
        gateway
           API Gateway - either ambassador or seldon
        transport
           API transport - grpc or rest
        namespace
           k8s namespace of running deployment
        deployment_name
           name of seldon deployment
        payload_type
           pyalod - tensor, ndarray or tftensor
        oauth_key
           OAUTH key (if using seldon api server)
        oauth_secret
           OAUTH secret (if using seldon api server)
        seldon_rest_endpoint
           REST endpoint to seldon api server
        seldon_grpc_endpoint
           gRPC endpoint to seldon api server
        ambassador_endpoint
           Ambassador endpoint
        microservice_endpoint
           Running microservice endpoint
        grpc_max_send_message_length
           Max grpc send message size in bytes
        grpc_max_receive_message_length
           Max grpc receive message size in bytes
        """
        self.config = locals()
        del self.config["self"]
        logging.debug("Configuration:" + str(self.config))

    def _gather_args(self, **kwargs):

        c2 = {**self.config}
        c2.update({k: v for k, v in kwargs.items() if v is not None})
        return c2

    def _validate_args(self, gateway: str = None, transport: str = None,
                       method: str = None, data: np.ndarray = None, **kwargs):
        """
        Internal method to validate parameters

        Parameters
        ----------
        gateway
           API gateway
        transport
           API transport
        method
           The method to call
        data
           Numpy data to send
        kwargs

        Returns
        -------

        """
        if not (gateway == "ambassador" or gateway == "seldon"):
            raise SeldonClientException("Valid values for gateway are 'ambassador' or 'seldon'")
        if not (transport == "rest" or transport == "grpc"):
            raise SeldonClientException("Valid values for transport are 'rest' or 'grpc'")
        if not (method == "predict" or method == "route" or method == "aggregate" or method == "transform-input" or
                method == "transform-output" or method == "send-feedback" or method is None):
            raise SeldonClientException(
                "Valid values for method are 'predict', 'route', 'transform-input', 'transform-output', 'aggregate' or None")
        if not (data is None or isinstance(data, np.ndarray)):
            raise SeldonClientException("Valid values for data are None or numpy array")

    def predict(self, gateway: str = None, transport: str = None, deployment_name: str = None,
                payload_type: str = None, oauth_key: str = None, oauth_secret: str = None,
                seldon_rest_endpoint: str = None, seldon_grpc_endpoint: str = None,
                ambassador_endpoint: str = None, microservice_endpoint: str = None,
                method: str = None, shape: Tuple = (1, 1), namespace: str = None, data: np.ndarray = None,
                bin_data: Union[bytes, bytearray] = None, str_data: str = None, names: Iterable[str] = None,
                ambassador_prefix: str = None, headers: Dict = None) -> SeldonClientPrediction:
        """

        Parameters
        ----------
        gateway
           API Gateway - either ambassador or seldon
        transport
           API transport - grpc or rest
        namespace
           k8s namespace of running deployment
        deployment_name
           name of seldon deployment
        payload_type
           pyalod - tensor, ndarray or tftensor
        oauth_key
           OAUTH key (if using seldon api server)
        oauth_secret
           OAUTH secret (if using seldon api server)
        seldon_rest_endpoint
           REST endpoint to seldon api server
        seldon_grpc_endpoint
           gRPC endpoint to seldon api server
        ambassador_endpoint
           Ambassador endpoint
        microservice_endpoint
           Running microservice endpoint
        grpc_max_send_message_length
           Max grpc send message size in bytes
        grpc_max_receive_message_length
           Max grpc receive message size in bytes
        data
           Numpy Array Payload to send
        bin_data
           Binary payload to send - will override data
        str_data
           String payload to send - will override data
        names
           Column names
        ambassador_prefix
           prefix path for Ambassador URL endpoint
        headers
           Headers to add to request

        Returns
        -------

        """
        k = self._gather_args(gateway=gateway, transport=transport, deployment_name=deployment_name,
                              payload_type=payload_type, oauth_key=oauth_key,
                              oauth_secret=oauth_secret, seldon_rest_endpoint=seldon_rest_endpoint,
                              seldon_grpc_endpoint=seldon_grpc_endpoint, ambassador_endpoint=ambassador_endpoint,
                              microservice_endpoint=microservice_endpoint, method=method, shape=shape,
                              namespace=namespace, names=names,
                              data=data, bin_data=bin_data, str_data=str_data,
                              ambassador_prefix=ambassador_prefix, headers=headers)
        self._validate_args(**k)
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
                 prediction_response: prediction_pb2.SeldonMessage = None, reward: float = 0,
                 gateway: str = None, transport: str = None, deployment_name: str = None,
                 payload_type: str = None, oauth_key: str = None, oauth_secret: str = None,
                 seldon_rest_endpoint: str = None, seldon_grpc_endpoint: str = None,
                 ambassador_endpoint: str = None, microservice_endpoint: str = None,
                 method: str = None, shape: Tuple = (1, 1), namespace: str = None,
                 ambassador_prefix: str = None) -> SeldonClientFeedback:
        """

        Parameters
        ----------
        prediction_request
           Previous prediction request
        prediction_response
           Previous prediction response
        reward
           A reward to send in feedback
        gateway
           API Gateway - either ambassador or seldon
        transport
           API transport - grpc or rest
        deployment_name
           name of seldon deployment
        payload_type
           payload - tensor, ndarray or tftensor
        oauth_key
           OAUTH key (if using seldon api server)
        oauth_secret
           OAUTH secret (if using seldon api server)
        seldon_rest_endpoint
           REST endpoint to seldon api server
        seldon_grpc_endpoint
           gRPC endpoint to seldon api server
        ambassador_endpoint
           Ambassador endpoint
        microservice_endpoint
           Running microservice endpoint
        grpc_max_send_message_length
           Max grpc send message size in bytes
        grpc_max_receive_message_length
           Max grpc receive message size in bytes
        method
           The microservice method to call
        shape
           The shape of the data to send
        namespace
           k8s namespace of running deployment

        Returns
        -------

        """
        k = self._gather_args(gateway=gateway, transport=transport, deployment_name=deployment_name,
                              payload_type=payload_type, oauth_key=oauth_key, oauth_secret=oauth_secret,
                              seldon_rest_endpoint=seldon_rest_endpoint
                              , seldon_grpc_endpoint=seldon_grpc_endpoint, ambassador_endpoint=ambassador_endpoint,
                              microservice_endpoint=microservice_endpoint, method=method, shape=shape,
                              namespace=namespace, ambassador_prefix=ambassador_prefix)
        self._validate_args(**k)
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

    def microservice(self, gateway: str = None, transport: str = None, deployment_name: str = None,
                     payload_type: str = None, oauth_key: str = None, oauth_secret: str = None,
                     seldon_rest_endpoint: str = None, seldon_grpc_endpoint: str = None,
                     ambassador_endpoint: str = None, microservice_endpoint: str = None,
                     method: str = None, shape: Tuple = (1, 1), namespace: str = None, data: np.ndarray = None,
                     datas: List[np.ndarray] = None, ndatas: int = None, bin_data: Union[bytes, bytearray] = None,
                     str_data: str = None, names: Iterable[str] = None) -> Union[SeldonClientPrediction, SeldonClientCombine]:
        """

        Parameters
        ----------
        gateway
           API Gateway - either ambassador or seldon
        transport
           API transport - grpc or rest
        deployment_name
           name of seldon deployment
        payload_type
           payload - tensor, ndarray or tftensor
        oauth_key
           OAUTH key (if using seldon api server)
        oauth_secret
           OAUTH secret (if using seldon api server)
        seldon_rest_endpoint
           REST endpoint to seldon api server
        seldon_grpc_endpoint
           gRPC endpoint to seldon api server
        ambassador_endpoint
           Ambassador endpoint
        microservice_endpoint
           Running microservice endpoint
        grpc_max_send_message_length
           Max grpc send message size in bytes
        grpc_max_receive_message_length
           Max grpc receive message size in bytes
        method
           The microservice method to call
        shape
           The shape of the data to send
        namespace
           k8s namespace of running deployment
        data
           Numpy Array Payload to send
        bin_data
           Binary payload to send - will override data
        str_data
           String payload to send - will override data
        ndatas
           Multiple numpy arrays to send for aggregation
        bin_data
           Binary data payload
        str_data
           String data payload
        names
           Column names

        Returns
        -------
           A prediction result

        """
        k = self._gather_args(gateway=gateway, transport=transport, deployment_name=deployment_name,
                              payload_type=payload_type, oauth_key=oauth_key,
                              oauth_secret=oauth_secret, seldon_rest_endpoint=seldon_rest_endpoint,
                              seldon_grpc_endpoint=seldon_grpc_endpoint, ambassador_endpoint=ambassador_endpoint,
                              microservice_endpoint=microservice_endpoint, method=method, shape=shape,
                              namespace=namespace, datas=datas, ndatas=ndatas, names=names,
                              data=data, bin_data=bin_data, str_data=str_data)
        self._validate_args(**k)
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
                              prediction_response: prediction_pb2.SeldonMessage = None, reward: float = 0,
                              gateway: str = None, transport: str = None, deployment_name: str = None,
                              payload_type: str = None, oauth_key: str = None, oauth_secret: str = None,
                              seldon_rest_endpoint: str = None,
                              seldon_grpc_endpoint: str = None,
                              ambassador_endpoint: str = None,
                              microservice_endpoint: str = None,
                              method: str = None, shape: Tuple = (1, 1), namespace: str = None) -> SeldonClientFeedback:
        """

        Parameters
        ----------
        prediction_request
           Previous prediction request
        prediction_response
           Previous prediction response
        reward
           A reward to send in feedback
        gateway
           API Gateway - either ambassador or seldon
        transport
           API transport - grpc or rest
        deployment_name
           name of seldon deployment
        payload_type
           payload - tensor, ndarray or tftensor
        oauth_key
           OAUTH key (if using seldon api server)
        oauth_secret
           OAUTH secret (if using seldon api server)
        seldon_rest_endpoint
           REST endpoint to seldon api server
        seldon_grpc_endpoint
           gRPC endpoint to seldon api server
        ambassador_endpoint
           Ambassador endpoint
        microservice_endpoint
           Running microservice endpoint
        grpc_max_send_message_length
           Max grpc send message size in bytes
        grpc_max_receive_message_length
           Max grpc receive message size in bytes
        method
           The microservice method to call
        shape
           The shape of the data to send
        namespace
           k8s namespace of running deployment

        Returns
        -------
           A client response

        """
        k = self._gather_args(gateway=gateway, transport=transport, deployment_name=deployment_name,
                              payload_type=payload_type, oauth_key=oauth_key, oauth_secret=oauth_secret,
                              seldon_rest_endpoint=seldon_rest_endpoint
                              , seldon_grpc_endpoint=seldon_grpc_endpoint, ambassador_endpoint=ambassador_endpoint,
                              microservice_endpoint=microservice_endpoint, method=method, shape=shape,
                              namespace=namespace)
        self._validate_args(**k)
        if k["transport"] == "rest":
            return microservice_api_rest_feedback(prediction_request, prediction_response, reward, **k)
        else:
            return microservice_api_grpc_feedback(prediction_request, prediction_response, reward, **k)


def microservice_api_rest_seldon_message(method: str = "predict", microservice_endpoint: str = "localhost:5000",
                                         shape: Tuple = (1, 1),
                                         data: object = None, payload_type: str = "tensor",
                                         bin_data: Union[bytes, bytearray] = None, str_data: str = None,
                                         names: Iterable[str] = None,
                                         **kwargs) -> SeldonClientPrediction:
    """
    Call Seldon microservice REST API

    Parameters
    ----------
    method
       The microservice method to call
    microservice_endpoint
       Running microservice endpoint
    grpc_max_send_message_length
       Max grpc send message size in bytes
    grpc_max_receive_message_length
       Max grpc receive message size in bytes
    method
       The microservice method to call
    shape
       The shape of the data to send
    namespace
       k8s namespace of running deployment
    shape
       Shape of the data to send
    data
       Numpy array data to send
    payload_type
       payload - tensor, ndarray or tftensor
    bin_data
       Binary data payload
    str_data
       String data payload
    names
       Column names
    kwargs

    Returns
    -------
      A SeldonClientPrediction data response
    """
    if bin_data is not None:
        request = prediction_pb2.SeldonMessage(binData=bin_data)
    elif str_data is not None:
        request = prediction_pb2.SeldonMessage(strData=str_data)
    else:
        if data is None:
            data = np.random.rand(*shape)
        datadef = array_to_grpc_datadef(payload_type, data, names=names)
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
    except Exception as e:
        return SeldonClientPrediction(request, None, success, str(e))


def microservice_api_rest_aggregate(microservice_endpoint: str = "localhost:5000",
                                    shape: Tuple = (1, 1),
                                    datas: List[np.ndarray] = None, ndatas: int = None, payload_type: str = "tensor",
                                    names: Iterable[str] = None,
                                    **kwargs) -> SeldonClientCombine:
    """
    Call Seldon microservice REST API aggregate endpoint

    Parameters
    ----------
    microservice_endpoint
       Running microservice endpoint
    shape
       The shape of the data to send
    datas
       List of Numpy array data to send
    ndatas
       Multiple numpy arrays to send for aggregation
    payload_type
       payload - tensor, ndarray or tftensor
    names
       Column names
    kwargs

    Returns
    -------
       A SeldonClientPrediction

    """
    if datas is None:
        datas = []
        for n in range(ndatas):
            data = np.random.rand(*shape)
            datas.append(data)
    msgs = []
    for data in datas:
        if isinstance(data, (bytes, bytearray)):
            msgs.append(prediction_pb2.SeldonMessage(binData=data))
        elif isinstance(data, str):
            msgs.append(prediction_pb2.SeldonMessage(strData=data))
        else:
            datadef = array_to_grpc_datadef(payload_type, data, names)
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
    except Exception as e:
        return SeldonClientCombine(request, None, success, str(e))


def microservice_api_rest_feedback(prediction_request: prediction_pb2.SeldonMessage = None,
                                   prediction_response: prediction_pb2.SeldonMessage = None, reward: float = 0,
                                   microservice_endpoint: str = None, **kwargs) -> SeldonClientFeedback:
    """
    Call Seldon microserice REST API to send feedback

    Parameters
    ----------
    prediction_request
       Previous prediction request
    prediction_response
       Previous prediction response
    reward
       A reward to send in feedback
    microservice_endpoint
       Running microservice endpoint
    kwargs

    Returns
    -------
       A SeldonClientFeedback
    """
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
    except Exception as e:
        return SeldonClientFeedback(request, None, success, str(e))


def microservice_api_grpc_seldon_message(method: str = "predict", microservice_endpoint: str = "localhost:5000",
                                         shape: Tuple = (1, 1),
                                         data: object = None, payload_type: str = "tensor",
                                         bin_data: Union[bytes, bytearray] = None, str_data: str = None,
                                         grpc_max_send_message_length: int = 4 * 1024 * 1024,
                                         grpc_max_receive_message_length: int = 4 * 1024 * 1024,
                                         names: Iterable[str] = None,
                                         **kwargs) -> SeldonClientPrediction:
    """
    Call Seldon microservice gRPC API

    Parameters
    ----------
    method
       Method to call
    microservice_endpoint
       Running microservice endpoint
    shape
       The shape of the data to send
    data
       Numpy array data to send
    payload_type
       payload - tensor, ndarray or tftensor
    bin_data
       Binary data to send
    str_data
       String data to send
    grpc_max_send_message_length
       Max grpc send message size in bytes
    grpc_max_receive_message_length
       Max grpc receive message size in bytes
    names
       column names
    kwargs

    Returns
    -------
      SeldonClientPrediction
    """
    if bin_data is not None:
        request = prediction_pb2.SeldonMessage(binData=bin_data)
    elif str_data is not None:
        request = prediction_pb2.SeldonMessage(strData=str_data)
    else:
        if data is None:
            data = np.random.rand(*shape)
        datadef = array_to_grpc_datadef(payload_type, data, names=names)
        request = prediction_pb2.SeldonMessage(data=datadef)
    channel = grpc.insecure_channel(microservice_endpoint, options=[
        ('grpc.max_send_message_length', grpc_max_send_message_length),
        ('grpc.max_receive_message_length', grpc_max_receive_message_length)])
    try:
        if method == "predict":
            stub_model = prediction_pb2_grpc.ModelStub(channel)
            response = stub_model.Predict(request=request)
        elif method == "transform-input":
            stub = prediction_pb2_grpc.GenericStub(channel)
            response = stub.TransformInput(request=request)
        elif method == "transform-output":
            stub = prediction_pb2_grpc.GenericStub(channel)
            response = stub.TransformOutput(request=request)
        elif method == "route":
            stub = prediction_pb2_grpc.GenericStub(channel)
            response = stub.Route(request=request)
        else:
            raise SeldonClientException("Unknown method:" + method)

        return SeldonClientPrediction(request, response, True, "")
    except Exception as e:
        return SeldonClientPrediction(request, None, False, str(e))


def microservice_api_grpc_aggregate(microservice_endpoint: str = "localhost:5000",
                                    shape: Tuple = (1, 1),
                                    datas: List[np.ndarray] = None, ndatas: int = None, payload_type: str = "tensor",
                                    grpc_max_send_message_length: int = 4 * 1024 * 1024,
                                    grpc_max_receive_message_length: int = 4 * 1024 * 1024,
                                    names: Iterable[str] = None,
                                    **kwargs) -> SeldonClientCombine:
    """
    Call Seldon microservice gRPC API aggregate

    Parameters
    ----------
    microservice_endpoint
       Microservice API endpoint
    shape
       Shape of the data to send
    datas
       List of Numpy array data to send
    ndatas
       Multiple numpy arrays to send for aggregation
    payload_type
       payload - tensor, ndarray or tftensor
    grpc_max_send_message_length
       Max grpc send message size in bytes
    grpc_max_receive_message_length
       Max grpc receive message size in bytes
    names
       Column names
    kwargs

    Returns
    -------
       SeldonClientCombine

    """
    if datas is None:
        datas = []
        for n in range(ndatas):
            data = np.random.rand(*shape)
            datas.append(data)
    msgs = []
    for data in datas:
        if isinstance(data, (bytes, bytearray)):
            msgs.append(prediction_pb2.SeldonMessage(binData=data))
        elif isinstance(data, str):
            msgs.append(prediction_pb2.SeldonMessage(strData=data))
        else:
            datadef = array_to_grpc_datadef(payload_type, data, names=names)
            msgs.append(prediction_pb2.SeldonMessage(data=datadef))
    request = prediction_pb2.SeldonMessageList(seldonMessages=msgs)
    try:
        channel = grpc.insecure_channel(microservice_endpoint, options=[
            ('grpc.max_send_message_length', grpc_max_send_message_length),
            ('grpc.max_receive_message_length', grpc_max_receive_message_length)])
        stub = prediction_pb2_grpc.GenericStub(channel)
        response = stub.Aggregate(request=request)
        return SeldonClientCombine(request, response, True, "")
    except Exception as e:
        return SeldonClientCombine(request, None, False, str(e))


def microservice_api_grpc_feedback(prediction_request: prediction_pb2.SeldonMessage = None,
                                   prediction_response: prediction_pb2.SeldonMessage = None, reward: float = 0,
                                   microservice_endpoint: str = None,
                                   grpc_max_send_message_length: int = 4 * 1024 * 1024,
                                   grpc_max_receive_message_length: int = 4 * 1024 * 1024,
                                   **kwargs) -> SeldonClientFeedback:
    """
    Call Seldon gRPC

    Parameters
    ----------
    prediction_request
       Previous prediction request
    prediction_response
       Previous prediction response
    reward
       A reward to send in feedback
    microservice_endpoint
       Running microservice endpoint
    kwargs

    Returns
    -------

    """
    request = prediction_pb2.Feedback(request=prediction_request, response=prediction_response, reward=reward)
    try:
        channel = grpc.insecure_channel(microservice_endpoint, options=[
            ('grpc.max_send_message_length', grpc_max_send_message_length),
            ('grpc.max_receive_message_length', grpc_max_receive_message_length)])
        stub = prediction_pb2_grpc.GenericStub(channel)
        response = stub.SendFeedback(request=request)
        return SeldonClientFeedback(request, response, True, "")
    except Exception as e:
        return SeldonClientFeedback(request, None, False, str(e))


#
# External API
#

def get_token(oauth_key: str = "", oauth_secret: str = "", namespace: str = None,
              endpoint: str = "localhost:8002") -> str:
    """
    Get an OAUTH key from the Seldon Gateway

    Parameters
    ----------
    oauth_key
       OAUTH key
    oauth_secret
       OAUTH secret
    namespace
       k8s namespace of running deployment
    endpoint
       The host:port of the endpoint for the OAUTH API server
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
    token = response.json()["access_token"]
    return token


def rest_predict_seldon_oauth(oauth_key: str, oauth_secret: str, namespace: str = None,
                              seldon_rest_endpoint: str = "localhost:8002", shape: Tuple = (1, 1),
                              data: object = None, payload_type: str = "tensor",
                              bin_data: Union[bytes, bytearray] = None, str_data: str = None,
                              names: Iterable[str] = None,
                              **kwargs) -> SeldonClientPrediction:
    """
    Call Seldon API Gateway using REST

    Parameters
    ----------
    oauth_key
       OAUTH key
    oauth_secret
       OAUTH secret
    namespace
       k8s namespace of running deployment
    seldon_rest_endpoint
       Endpoint of REST endpoint
    shape
       Shape of endpoint
    data
       Data to send
    payload_type
       payload - tensor, ndarray or tftensor
    bin_data
       Binary data to send
    str_data
       String data to send
    names
       column names
    kwargs

    Returns
    -------
       Seldon Client Prediction

    """
    token = get_token(oauth_key, oauth_secret, namespace, seldon_rest_endpoint)
    if bin_data is not None:
        request = prediction_pb2.SeldonMessage(binData=bin_data)
    elif str_data is not None:
        request = prediction_pb2.SeldonMessage(strData=str_data)
    else:
        if data is None:
            data = np.random.rand(*shape)
        datadef = array_to_grpc_datadef(payload_type, data, names=names)
        request = prediction_pb2.SeldonMessage(data=datadef)
    headers = {'Authorization': 'Bearer ' + token}
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
        msg = str(response_raw.status_code) + ":" + response_raw.reason
    try:
        if len(response_raw.text) > 0:
            try:
                response = json_to_seldon_message(response_raw.json())
            except:
                response = None
        else:
            response = None
        return SeldonClientPrediction(request, response, success, msg)
    except Exception as e:
        return SeldonClientPrediction(request, None, False, str(e))


def grpc_predict_seldon_oauth(oauth_key: str, oauth_secret: str, namespace: str = None,
                              seldon_rest_endpoint: str = "localhost:8002",
                              seldon_grpc_endpoint: str = "localhost:8004", shape: Tuple[int, int] = (1, 1),
                              data: np.ndarray = None, payload_type: str = "tensor",
                              bin_data: Union[bytes, bytearray] = None, str_data: str = None,
                              grpc_max_send_message_length: int = 4 * 1024 * 1024,
                              grpc_max_receive_message_length: int = 4 * 1024 * 1024,
                              names: Iterable[str] = None,
                              **kwargs) -> SeldonClientPrediction:
    """
    Call Seldon gRPC API Gateway endpoint

    Parameters
    ----------
    oauth_key
       OAUTH key
    oauth_secret
       OAUTH secret
    namespace
       k8s namespace of running deployment
    seldon_rest_endpoint
       Endpoint of REST endpoint
    shape
       Shape of endpoint
    data
       Data to send
    payload_type
       payload - tensor, ndarray or tftensor
    bin_data
       Binary data to send
    str_data
       String data to send
    grpc_max_send_message_length
       Max grpc send message size in bytes
    grpc_max_receive_message_length
       Max grpc receive message size in bytes
    names
       Column names
    kwargs

    Returns
    -------
       A SeldonMessage proto

    """
    token = get_token(oauth_key, oauth_secret, namespace, seldon_rest_endpoint)
    if bin_data is not None:
        request = prediction_pb2.SeldonMessage(binData=bin_data)
    elif str_data is not None:
        request = prediction_pb2.SeldonMessage(strData=str_data)
    else:
        if data is None:
            data = np.random.rand(*shape)
        datadef = array_to_grpc_datadef(payload_type, data, names=names)
        request = prediction_pb2.SeldonMessage(data=datadef)
    channel = grpc.insecure_channel(seldon_grpc_endpoint, options=[
        ('grpc.max_send_message_length', grpc_max_send_message_length),
        ('grpc.max_receive_message_length', grpc_max_receive_message_length)])
    stub = prediction_pb2_grpc.SeldonStub(channel)
    metadata = [('oauth_token', token)]
    try:
        response = stub.Predict(request=request, metadata=metadata)
        return SeldonClientPrediction(request, response, True, "")
    except Exception as e:
        return SeldonClientPrediction(request, None, False, str(e))


def rest_predict_ambassador(deployment_name: str, namespace: str = None, ambassador_endpoint: str = "localhost:8003",
                            shape: Tuple[int, int] = (1, 1),
                            data: np.ndarray = None, headers: Dict = None, ambassador_prefix: str = None,
                            payload_type: str = "tensor",
                            bin_data: Union[bytes, bytearray] = None, str_data: str = None,
                            names: Iterable[str] = None,
                            **kwargs) -> SeldonClientPrediction:
    """
    REST request to Seldon Ambassador Ingress

    Parameters
    ----------
    deployment_name
       The name of the Seldon Deployment
    namespace
       k8s namespace of running deployment
    ambassador_endpoint
       The host:port of Ambassador
    shape
       The shape of the data to send
    data
       The numpy data to send
    headers
       Headers to add to request
    ambassador_prefix
       The prefix path to add to the request
    payload_type
       payload - tensor, ndarray or tftensor
    bin_data
       Binary data to send
    str_data
       String data to send
    names
       Column names

    Returns
    -------
       A requests Response object

    """
    if bin_data is not None:
        request = prediction_pb2.SeldonMessage(binData=bin_data)
    elif str_data is not None:
        request = prediction_pb2.SeldonMessage(strData=str_data)
    else:
        if data is None:
            data = np.random.rand(*shape)
        datadef = array_to_grpc_datadef(payload_type, data, names=names)
        request = prediction_pb2.SeldonMessage(data=datadef)
    payload = seldon_message_to_json(request)
    if ambassador_prefix is None:
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
            "http://" + ambassador_endpoint + ambassador_prefix + "/api/v0.1/predictions",
            json=payload,
            headers=headers)

    if response_raw.status_code == 200:
        success = True
        msg = ""
    else:
        success = False
        msg = str(response_raw.status_code) + ":" + response_raw.reason
    try:
        if len(response_raw.text) > 0:
            try:
                response = json_to_seldon_message(response_raw.json())
            except:
                response = None
        else:
            response = None
        return SeldonClientPrediction(request, response, success, msg)
    except Exception as e:
        return SeldonClientPrediction(request, None, False, str(e))


def rest_predict_ambassador_basicauth(deployment_name: str, username: str, password: str, namespace: str = None,
                                      ambassador_endpoint: str = "localhost:8003",
                                      shape: Tuple[int, int] = (1, 1), data: np.ndarray = None,
                                      payload_type: str = "tensor",
                                      bin_data: Union[bytes, bytearray] = None, str_data: str = None,
                                      names: Iterable[str] = None,
                                      **kwargs) -> SeldonClientPrediction:
    """
    REST request with Basic Auth to Seldon Ambassador Ingress

    Parameters
    ----------
    deployment_name
       The name of the running deployment
    username
       Username for basic auth
    password
       Password for basic auth
    namespace
       The namespace of the running deployment
    ambassador_endpoint
       The host:port of ambassador
    shape
       The shape of data
    data
       The numpy data to send
    payload_type
       payload - tensor, ndarray or tftensor
    bin_data
       Binary data to send
    str_data
    names
       Column names

    Returns
    -------

    """
    if bin_data is not None:
        request = prediction_pb2.SeldonMessage(binData=bin_data)
    elif str_data is not None:
        request = prediction_pb2.SeldonMessage(strData=str_data)
    else:
        if data is None:
            data = np.random.rand(*shape)
        datadef = array_to_grpc_datadef(payload_type, data, names)
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
        msg = str(response_raw.status_code) + ":" + response_raw.reason
    try:
        if len(response_raw.text) > 0:
            try:
                response = json_to_seldon_message(response_raw.json())
            except:
                response = None
        else:
            response = None
        return SeldonClientPrediction(request, response, success, msg)
    except Exception as e:
        return SeldonClientPrediction(request, None, False, str(e))


def grpc_predict_ambassador(deployment_name: str, namespace: str = None, ambassador_endpoint: str = "localhost:8003",
                            shape: Tuple[int, int] = (1, 1),
                            data: np.ndarray = None,
                            headers: Dict = None, payload_type: str = "tensor",
                            bin_data: Union[bytes, bytearray] = None, str_data: str = None,
                            grpc_max_send_message_length: int = 4 * 1024 * 1024,
                            grpc_max_receive_message_length: int = 4 * 1024 * 1024,
                            names: Iterable[str] = None,
                            **kwargs) -> SeldonClientPrediction:
    """
    gRPC request to Seldon Ambassador Ingress

    Parameters
    ----------
    deployment_name
       Deployment name of Seldon Deployment
    namespace
       The namespace the Seldon Deployment is running in
    ambassador_endpoint
       The endpoint for Ambassador
    shape
       The shape of the data
    data
       The numpy array data to send
    headers
      A Dict of key value pairs to add to gRPC HTTP Headers
    payload_type
       payload - tensor, ndarray or tftensor
    bin_data
       Binary data to send
    str_data
       String data to send
    grpc_max_send_message_length
       Max grpc send message size in bytes
    grpc_max_receive_message_length
       Max grpc receive message size in bytes
    names
       Column names

    Returns
    -------
       A SeldonMessage proto response

    """
    if bin_data is not None:
        request = prediction_pb2.SeldonMessage(binData=bin_data)
    elif str_data is not None:
        request = prediction_pb2.SeldonMessage(strData=str_data)
    else:
        if data is None:
            data = np.random.rand(*shape)
        datadef = array_to_grpc_datadef(payload_type, data, names=names)
        request = prediction_pb2.SeldonMessage(data=datadef)
    channel = grpc.insecure_channel(ambassador_endpoint, options=[
        ('grpc.max_send_message_length', grpc_max_send_message_length),
        ('grpc.max_receive_message_length', grpc_max_receive_message_length)])
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
       Previous prediction request
    prediction_response
       Previous prediction response
    reward
       A reward to send in feedback
    oauth_key
       OAUTH key
    oauth_secret
       OAUTH secret
    namespace
       k8s namespace of running deployment
    seldon_rest_endpoint
       Endpoint of REST endpoint
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
        msg = str(response_raw.status_code) + ":" + response_raw.reason
    try:
        if len(response_raw.text) > 0:
            try:
                response = json_to_seldon_message(response_raw.json())
            except:
                response = None
        else:
            response = None
        return SeldonClientFeedback(request, response, success, msg)
    except Exception as e:
        return SeldonClientFeedback(request, None, False, str(e))


def grpc_feedback_seldon_oauth(prediction_request: prediction_pb2.SeldonMessage = None,
                               prediction_response: prediction_pb2.SeldonMessage = None, reward: float = 0,
                               oauth_key: str = "", oauth_secret: str = "", namespace: str = None,
                               seldon_rest_endpoint: str = "localhost:8002",
                               seldon_grpc_endpoint: str = "localhost:8004",
                               grpc_max_send_message_length: int = 4 * 1024 * 1024,
                               grpc_max_receive_message_length: int = 4 * 1024 * 1024,
                               **kwargs) -> SeldonClientFeedback:
    """
    Send feedback to Seldon API gateway via gRPC

    Parameters
    ----------
    prediction_request
       Previous prediction request
    prediction_response
       Previous prediction response
    reward
       A reward to send in feedback
    oauth_key
       OAUTH key
    oauth_secret
       OAUTH secret
    namespace
       k8s namespace of running deployment
    seldon_rest_endpoint
       Endpoint of REST endpoint
    seldon_grpc_endpoint
       Endpoint for Seldon grpc
    grpc_max_send_message_length
       Max grpc send message size in bytes
    grpc_max_receive_message_length
       Max grpc receive message size in bytes
    kwargs

    Returns
    -------

    """
    token = get_token(oauth_key, oauth_secret, namespace, seldon_rest_endpoint)
    request = prediction_pb2.Feedback(request=prediction_request, response=prediction_response, reward=reward)
    channel = grpc.insecure_channel(seldon_grpc_endpoint, options=[
        ('grpc.max_send_message_length', grpc_max_send_message_length),
        ('grpc.max_receive_message_length', grpc_max_receive_message_length)])
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
                             ambassador_endpoint: str = "localhost:8003", headers: Dict = None, ambassador_prefix: str = None,
                             **kwargs) -> SeldonClientFeedback:
    """
    Send Feedback to Seldon via Ambassador using REST

    Parameters
    ----------
    prediction_request
       Previous prediction request
    prediction_response
       Previous prediction response
    reward
       A reward to send in feedback
    deployment_name
       The name of the running Seldon deployment
    namespace
       k8s namespace of running deployment
    ambassador_endpoint
       The ambassador host:port endpoint
    headers
       Headers to add to the request
    ambassador_prefix
      The prefix to add to the request path for Ambassador
    kwargs

    Returns
    -------
      A Seldon Feedback Response

    """
    request = prediction_pb2.Feedback(request=prediction_request, response=prediction_response, reward=reward)
    payload = feedback_to_json(request)
    if ambassador_prefix is None:
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
            "http://" + ambassador_endpoint + ambassador_prefix + "/api/v0.1/feedback",
            json=payload,
            headers=headers)

    if response_raw.status_code == 200:
        success = True
        msg = ""
    else:
        success = False
        msg = str(response_raw.status_code) + ":" + response_raw.reason
    try:
        if len(response_raw.text) > 0:
            try:
                response = json_to_seldon_message(response_raw.json())
            except:
                response = None
        else:
            response = None
        return SeldonClientFeedback(request, response, success, msg)
    except Exception as e:
        return SeldonClientFeedback(request, None, False, str(e))


def grpc_feedback_ambassador(prediction_request: prediction_pb2.SeldonMessage = None,
                             prediction_response: prediction_pb2.SeldonMessage = None, reward: float = 0,
                             deployment_name: str = "", namespace: str = None,
                             ambassador_endpoint: str = "localhost:8003",
                             headers: Dict = None,
                             grpc_max_send_message_length: int = 4 * 1024 * 1024,
                             grpc_max_receive_message_length: int = 4 * 1024 * 1024,
                             **kwargs) -> SeldonClientFeedback:
    """

    Parameters
    ----------
    prediction_request
       Previous prediction request
    prediction_response
       Previous prediction response
    reward
       A reward to send in feedback
    deployment_name
       The name of the running Seldon deployment
    namespace
       k8s namespace of running deployment
    ambassador_endpoint
       The ambassador host:port endpoint
    headers
       Headers to add to the request
    grpc_max_send_message_length
       Max grpc send message size in bytes
    grpc_max_receive_message_length
       Max grpc receive message size in bytes
    kwargs

    Returns
    -------

    """
    request = prediction_pb2.Feedback(request=prediction_request, response=prediction_response, reward=reward)
    channel = grpc.insecure_channel(ambassador_endpoint, options=[
        ('grpc.max_send_message_length', grpc_max_send_message_length),
        ('grpc.max_receive_message_length', grpc_max_receive_message_length)])
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
