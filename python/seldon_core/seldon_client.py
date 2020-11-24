from seldon_core.proto import prediction_pb2
from seldon_core.proto import prediction_pb2_grpc
from seldon_core.utils import (
    array_to_grpc_datadef,
    seldon_message_to_json,
    json_to_feedback,
    json_to_seldon_message,
    feedback_to_json,
    seldon_messages_to_json,
)
import numpy as np
import grpc
import requests
from typing import Tuple, Dict, Union, List, Optional, Iterable
import json
import logging
import http.client as http_client
from google.protobuf import any_pb2, json_format

logger = logging.getLogger(__name__)


class SeldonClientException(Exception):
    """
    Seldon Client Exception
    """

    status_code = 400

    def __init__(self, message):
        Exception.__init__(self)
        self.message = message


class SeldonChannelCredentials(object):
    """
    Channel credentials.

    Presently just denotes an SSL connection.
    For GRPC in order to be properly implemented, you need to provide *either*
    the root_certificate_files, *or* all the file paths.
    The verify attribute currently is used to avoid SSL verification in REST
    however for GRPC it is recommended that you provide a path at least for the
    root_certificates_file otherwise it may not work as expected.
    """

    def __init__(
        self,
        verify: bool = True,
        root_certificates_file: str = None,
        private_key_file: str = None,
        certificate_chain_file: str = None,
    ):
        self.verify = verify
        self.root_certificates_file = root_certificates_file
        self.private_key_file = private_key_file
        self.certificate_chain_file = certificate_chain_file


class SeldonCallCredentials(object):
    """
    Credentials for each call, currently implements the ability to provide
        an OAuth token which is currently made available through REST via
        the X-Auth-Token header, and via GRPC via the metadata call creds.
    """

    def __init__(self, token: str = None):
        self.token = token


class SeldonClientPrediction(object):
    """
    Data class to return from Seldon Client
    """

    def __init__(
        self,
        request: Optional[Union[prediction_pb2.SeldonMessage, Dict]],
        response: Optional[Union[prediction_pb2.SeldonMessage, Dict]],
        success: bool = True,
        msg: str = "",
    ):
        self.request = request
        self.response = response
        self.success = success
        self.msg = msg

    def __repr__(self):
        return "Success:%s message:%s\nRequest:\n%s\nResponse:\n%s" % (
            self.success,
            self.msg,
            self.request,
            self.response,
        )


class SeldonClientFeedback(object):
    """
    Data class to return from Seldon Client for feedback calls
    """

    def __init__(
        self,
        request: Optional[prediction_pb2.Feedback],
        response: Optional[Union[prediction_pb2.SeldonMessage, Dict]],
        success: bool = True,
        msg: str = "",
    ):
        self.request = request
        self.response = response
        self.success = success
        self.msg = msg

    def __repr__(self):
        return "Success:%s message:%s\nRequest:\n%s\nResponse:\n%s" % (
            self.success,
            self.msg,
            self.request,
            self.response,
        )


class SeldonClientCombine(object):
    """
    Data class to return from Seldon Client for aggregate calls
    """

    def __init__(
        self,
        request: Optional[prediction_pb2.SeldonMessageList],
        response: Optional[prediction_pb2.SeldonMessage],
        success: bool = True,
        msg: str = "",
    ):
        self.request = request
        self.response = response
        self.success = success
        self.msg = msg

    def __repr__(self):
        return "Success:%s message:%s\nRequest:\n%s\nResponse:\n%s" % (
            self.success,
            self.msg,
            self.request,
            self.response,
        )


class SeldonClient(object):
    """
    A reference Seldon API Client
    """

    def __init__(
        self,
        gateway: str = "ambassador",
        transport: str = "rest",
        namespace: str = None,
        deployment_name: str = None,
        payload_type: str = "tensor",
        seldon_rest_endpoint: str = "localhost:8002",
        seldon_grpc_endpoint: str = "localhost:8004",
        gateway_endpoint: str = "localhost:8003",
        microservice_endpoint: str = "localhost:5000",
        grpc_max_send_message_length: int = 4 * 1024 * 1024,
        grpc_max_receive_message_length: int = 4 * 1024 * 1024,
        channel_credentials: SeldonChannelCredentials = None,
        call_credentials: SeldonCallCredentials = None,
        debug: bool = False,
        client_return_type: str = "dict",
    ):
        """

        Parameters
        ----------
        gateway
           API Gateway - either ambassador, istio or seldon
        transport
           API transport - grpc or rest
        namespace
           k8s namespace of running deployment
        deployment_name
           name of seldon deployment
        payload_type
           type of payload - tensor, ndarray or tftensor
        seldon_rest_endpoint
           REST endpoint to seldon api server
        seldon_grpc_endpoint
           gRPC endpoint to seldon api server
        gateway_endpoint
           Gateway endpoint
        microservice_endpoint
           Running microservice endpoint
        grpc_max_send_message_length
           Max grpc send message size in bytes
        grpc_max_receive_message_length
           Max grpc receive message size in bytes
        client_return_type
            the return type of all functions can be either dict or proto
        """
        if debug:
            logger.setLevel(logging.DEBUG)
            http_client.HTTPConnection.debuglevel = 1
        self.config = locals().copy()
        del self.config["self"]
        logger.debug("Configuration:" + str(self.config))

    def _gather_args(self, **kwargs):
        c2 = {**self.config}
        c2.update({k: v for k, v in kwargs.items() if v is not None})
        return c2

    def _validate_args(
        self,
        gateway: str = None,
        transport: str = None,
        method: str = None,
        data: np.ndarray = None,
        client_return_type: str = "dict",
        **kwargs,
    ):
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
        if not (gateway == "ambassador" or gateway == "seldon" or gateway == "istio"):
            raise SeldonClientException(
                "Valid values for gateway are 'ambassador', 'istio', or 'seldon'"
            )
        if not (transport == "rest" or transport == "grpc"):
            raise SeldonClientException(
                "Valid values for transport are 'rest' or 'grpc'"
            )
        if not (
            method == "predict"
            or method == "route"
            or method == "aggregate"
            or method == "transform-input"
            or method == "transform-output"
            or method == "send-feedback"
            or method is None
        ):
            raise SeldonClientException(
                "Valid values for method are 'predict', 'route', 'transform-input', 'transform-output', 'aggregate' or None"
            )
        if not (data is None or isinstance(data, np.ndarray)):
            raise SeldonClientException("Valid values for data are None or numpy array")
        if not (client_return_type == "proto" or client_return_type == "dict"):
            raise SeldonClientException(
                "Valid values for client_return_type are proto or dict"
            )

    def predict(
        self,
        gateway: str = None,
        transport: str = None,
        deployment_name: str = None,
        payload_type: str = None,
        seldon_rest_endpoint: str = None,
        seldon_grpc_endpoint: str = None,
        gateway_endpoint: str = None,
        microservice_endpoint: str = None,
        method: str = None,
        shape: Tuple = (1, 1),
        namespace: str = None,
        data: np.ndarray = None,
        bin_data: Union[bytes, bytearray] = None,
        str_data: str = None,
        json_data: Union[str, List, Dict] = None,
        custom_data: any_pb2.Any = None,
        names: Iterable[str] = None,
        gateway_prefix: str = None,
        headers: Dict = None,
        http_path: str = None,
        meta: Dict = None,
        client_return_type: str = None,
    ) -> SeldonClientPrediction:
        """

        Parameters
        ----------
        gateway
           API Gateway - either ambassador, istio or seldon
        transport
           API transport - grpc or rest
        namespace
           k8s namespace of running deployment
        deployment_name
           name of seldon deployment
        payload_type
           type of payload - tensor, ndarray or tftensor
        seldon_rest_endpoint
           REST endpoint to seldon api server
        seldon_grpc_endpoint
           gRPC endpoint to seldon api server
        gateway_endpoint
           Gateway endpoint
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
        json_data
           JSON payload to send - will override data
        custom_data
           Custom payload to send - will override data
        names
           Column names
        gateway_prefix
           prefix path for gateway URL endpoint
        headers
           Headers to add to request
        http_path:
           Custom http path for predict call to use
        meta:
           Custom meta map
        client_return_type
            the return type of all functions can be either dict or proto

        Returns
        -------

        """
        k = self._gather_args(
            gateway=gateway,
            transport=transport,
            deployment_name=deployment_name,
            payload_type=payload_type,
            seldon_rest_endpoint=seldon_rest_endpoint,
            seldon_grpc_endpoint=seldon_grpc_endpoint,
            gateway_endpoint=gateway_endpoint,
            microservice_endpoint=microservice_endpoint,
            method=method,
            shape=shape,
            namespace=namespace,
            names=names,
            data=data,
            bin_data=bin_data,
            str_data=str_data,
            json_data=json_data,
            custom_data=custom_data,
            gateway_prefix=gateway_prefix,
            headers=headers,
            http_path=http_path,
            meta=meta,
            client_return_type=client_return_type,
        )
        self._validate_args(**k)
        if k["gateway"] == "ambassador" or k["gateway"] == "istio":
            if k["transport"] == "rest":
                return rest_predict_gateway(**k)
            elif k["transport"] == "grpc":
                return grpc_predict_gateway(**k)
            else:
                raise SeldonClientException("Unknown transport " + k["transport"])
        elif k["gateway"] == "seldon":
            if k["transport"] == "rest":
                return rest_predict_seldon(**k)
            elif k["transport"] == "grpc":
                return grpc_predict_seldon(**k)
            else:
                raise SeldonClientException("Unknown transport " + k["transport"])
        else:
            raise SeldonClientException("Unknown gateway " + k["gateway"])

    def feedback(
        self,
        prediction_request: prediction_pb2.SeldonMessage = None,
        prediction_response: prediction_pb2.SeldonMessage = None,
        prediction_truth: prediction_pb2.SeldonMessage = None,
        reward: float = 0,
        gateway: str = None,
        transport: str = None,
        deployment_name: str = None,
        payload_type: str = None,
        seldon_rest_endpoint: str = None,
        seldon_grpc_endpoint: str = None,
        gateway_endpoint: str = None,
        microservice_endpoint: str = None,
        method: str = None,
        shape: Tuple = (1, 1),
        namespace: str = None,
        gateway_prefix: str = None,
        client_return_type: str = None,
        raw_request: dict = None,
    ) -> SeldonClientFeedback:
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
           API Gateway - either ambassador, istio or seldon
        transport
           API transport - grpc or rest
        deployment_name
           name of seldon deployment
        payload_type
           payload - tensor, ndarray or tftensor
        seldon_rest_endpoint
           REST endpoint to seldon api server
        seldon_grpc_endpoint
           gRPC endpoint to seldon api server
        gateway_endpoint
           Gateway endpoint
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
        client_return_type
            the return type of all functions can be either dict or proto

        Returns
        -------

        """
        k = self._gather_args(
            gateway=gateway,
            transport=transport,
            deployment_name=deployment_name,
            payload_type=payload_type,
            seldon_rest_endpoint=seldon_rest_endpoint,
            seldon_grpc_endpoint=seldon_grpc_endpoint,
            gateway_endpoint=gateway_endpoint,
            microservice_endpoint=microservice_endpoint,
            method=method,
            shape=shape,
            namespace=namespace,
            gateway_prefix=gateway_prefix,
            client_return_type=client_return_type,
            raw_request=raw_request,
        )
        self._validate_args(**k)
        if k["gateway"] == "ambassador" or k["gateway"] == "istio":
            if k["transport"] == "rest":
                return rest_feedback_gateway(
                    prediction_request,
                    prediction_response,
                    prediction_truth,
                    reward,
                    **k,
                )
            elif k["transport"] == "grpc":
                return grpc_feedback_gateway(
                    prediction_request,
                    prediction_response,
                    prediction_truth,
                    reward,
                    **k,
                )
            else:
                raise SeldonClientException("Unknown transport " + k["transport"])
        elif k["gateway"] == "seldon":
            if k["transport"] == "rest":
                return rest_feedback_seldon(
                    prediction_request,
                    prediction_response,
                    prediction_truth,
                    reward,
                    **k,
                )
            elif k["transport"] == "grpc":
                return grpc_feedback_seldon(
                    prediction_request,
                    prediction_response,
                    prediction_truth,
                    reward,
                    **k,
                )
            else:
                raise SeldonClientException("Unknown transport " + k["transport"])
        else:
            raise SeldonClientException("Unknown gateway " + k["gateway"])

    def explain(
        self,
        gateway: str = None,
        transport: str = None,
        deployment_name: str = None,
        payload_type: str = None,
        gateway_endpoint: str = None,
        shape: Tuple = (1, 1),
        namespace: str = None,
        data: np.ndarray = None,
        bin_data: Union[bytes, bytearray] = None,
        str_data: str = None,
        json_data: str = None,
        names: Iterable[str] = None,
        gateway_prefix: str = None,
        headers: Dict = None,
        http_path: str = None,
        client_return_type: str = None,
        predictor: str = None,
    ) -> Dict:
        """

        Parameters
        ----------
        gateway
           API Gateway - either ambassador, istio or seldon
        transport
           API transport - grpc or rest
        namespace
           k8s namespace of running deployment
        deployment_name
           name of seldon deployment
        payload_type
           type of payload - tensor, ndarray or tftensor
        seldon_rest_endpoint
           REST endpoint to seldon api server
        seldon_grpc_endpoint
           gRPC endpoint to seldon api server
        gateway_endpoint
           Gateway endpoint
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
        json_data
           JSON payload to send - will override data
        names
           Column names
        gateway_prefix
           prefix path for gateway URL endpoint
        headers
           Headers to add to request
        http_path:
           Custom http path for predict call to use
        client_return_type
            the return type of all functions can be either dict or proto
        predictor
            The name of the predictor to send the explanations to

        Returns
        -------

        """
        k = self._gather_args(
            gateway=gateway,
            transport=transport,
            deployment_name=deployment_name,
            payload_type=payload_type,
            gateway_endpoint=gateway_endpoint,
            shape=shape,
            namespace=namespace,
            names=names,
            data=data,
            bin_data=bin_data,
            str_data=str_data,
            json_data=json_data,
            gateway_prefix=gateway_prefix,
            headers=headers,
            http_path=http_path,
            client_return_type=client_return_type,
            predictor=predictor,
        )
        self._validate_args(**k)
        if k["gateway"] == "ambassador" or k["gateway"] == "istio":
            if k["transport"] == "rest":
                return explain_predict_gateway(**k)
            elif k["transport"] == "grpc":
                raise SeldonClientException("gRPC not supported for explain")
            else:
                raise SeldonClientException("Unknown transport " + k["transport"])
        else:
            raise SeldonClientException("Unknown gateway " + k["gateway"])

    def microservice(
        self,
        gateway: str = None,
        transport: str = None,
        deployment_name: str = None,
        payload_type: str = None,
        seldon_rest_endpoint: str = None,
        seldon_grpc_endpoint: str = None,
        gateway_endpoint: str = None,
        microservice_endpoint: str = None,
        method: str = None,
        shape: Tuple = (1, 1),
        namespace: str = None,
        data: np.ndarray = None,
        datas: List[np.ndarray] = None,
        ndatas: int = None,
        bin_data: Union[bytes, bytearray] = None,
        str_data: str = None,
        json_data: Union[str, List, Dict] = None,
        custom_data: any_pb2.Any = None,
        names: Iterable[str] = None,
    ) -> Union[SeldonClientPrediction, SeldonClientCombine]:
        """

        Parameters
        ----------
        gateway
           API Gateway - either ambassador, istio or seldon
        transport
           API transport - grpc or rest
        deployment_name
           name of seldon deployment
        payload_type
           payload - tensor, ndarray or tftensor
        seldon_rest_endpoint
           REST endpoint to seldon api server
        seldon_grpc_endpoint
           gRPC endpoint to seldon api server
        gateway_endpoint
           Gateway endpoint
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
        json_data
           String payload to send - will override data
        custom_data
           Custom payload to send - will override data
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
        k = self._gather_args(
            gateway=gateway,
            transport=transport,
            deployment_name=deployment_name,
            payload_type=payload_type,
            seldon_rest_endpoint=seldon_rest_endpoint,
            seldon_grpc_endpoint=seldon_grpc_endpoint,
            gateway_endpoint=gateway_endpoint,
            microservice_endpoint=microservice_endpoint,
            method=method,
            shape=shape,
            namespace=namespace,
            datas=datas,
            ndatas=ndatas,
            names=names,
            data=data,
            bin_data=bin_data,
            str_data=str_data,
            json_data=json_data,
            custom_data=custom_data,
        )
        self._validate_args(**k)
        if k["transport"] == "rest":
            if (
                k["method"] == "predict"
                or k["method"] == "transform-input"
                or k["method"] == "transform-output"
                or k["method"] == "route"
            ):
                return microservice_api_rest_seldon_message(**k)
            elif k["method"] == "aggregate":
                return microservice_api_rest_aggregate(**k)
            else:
                raise SeldonClientException("Unknown method " + k["method"])
        elif k["transport"] == "grpc":
            if (
                k["method"] == "predict"
                or k["method"] == "transform-input"
                or k["method"] == "transform-output"
                or k["method"] == "route"
            ):
                return microservice_api_grpc_seldon_message(**k)
            elif k["method"] == "aggregate":
                return microservice_api_grpc_aggregate(**k)
            else:
                raise SeldonClientException("Unknown method " + k["method"])
        else:
            raise SeldonClientException("Unknown transport " + k["transport"])

    def microservice_feedback(
        self,
        prediction_request: prediction_pb2.SeldonMessage = None,
        prediction_response: prediction_pb2.SeldonMessage = None,
        reward: float = 0,
        gateway: str = None,
        transport: str = None,
        deployment_name: str = None,
        payload_type: str = None,
        seldon_rest_endpoint: str = None,
        seldon_grpc_endpoint: str = None,
        gateway_endpoint: str = None,
        microservice_endpoint: str = None,
        method: str = None,
        shape: Tuple = (1, 1),
        namespace: str = None,
    ) -> SeldonClientFeedback:
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
           API Gateway - either Gateway or seldon
        transport
           API transport - grpc or rest
        deployment_name
           name of seldon deployment
        payload_type
           payload - tensor, ndarray or tftensor
        seldon_rest_endpoint
           REST endpoint to seldon api server
        seldon_grpc_endpoint
           gRPC endpoint to seldon api server
        gateway_endpoint
           Gateway endpoint
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
        k = self._gather_args(
            gateway=gateway,
            transport=transport,
            deployment_name=deployment_name,
            payload_type=payload_type,
            seldon_rest_endpoint=seldon_rest_endpoint,
            seldon_grpc_endpoint=seldon_grpc_endpoint,
            gateway_endpoint=gateway_endpoint,
            microservice_endpoint=microservice_endpoint,
            method=method,
            shape=shape,
            namespace=namespace,
        )
        self._validate_args(**k)
        if k["transport"] == "rest":
            return microservice_api_rest_feedback(
                prediction_request, prediction_response, reward, **k
            )
        else:
            return microservice_api_grpc_feedback(
                prediction_request, prediction_response, reward, **k
            )


def microservice_api_rest_seldon_message(
    method: str = "predict",
    microservice_endpoint: str = "localhost:5000",
    shape: Tuple = (1, 1),
    data: object = None,
    payload_type: str = "tensor",
    bin_data: Union[bytes, bytearray] = None,
    str_data: str = None,
    json_data: Union[str, List, Dict] = None,
    names: Iterable[str] = None,
    **kwargs,
) -> SeldonClientPrediction:
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
    json_data
       JSON data payload
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
    elif json_data is not None:
        request = json_to_seldon_message({"jsonData": json_data})
    else:
        if data is None:
            data = np.random.rand(*shape)
        datadef = array_to_grpc_datadef(payload_type, data, names=names)
        request = prediction_pb2.SeldonMessage(data=datadef)
    payload = seldon_message_to_json(request)
    response_raw = requests.post(
        "http://" + microservice_endpoint + "/" + method,
        data={"json": json.dumps(payload)},
    )
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


def microservice_api_rest_aggregate(
    microservice_endpoint: str = "localhost:5000",
    shape: Tuple = (1, 1),
    datas: List[np.ndarray] = None,
    ndatas: int = None,
    payload_type: str = "tensor",
    names: Iterable[str] = None,
    **kwargs,
) -> SeldonClientCombine:
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
        data={"json": json.dumps(payload)},
    )
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


def microservice_api_rest_feedback(
    prediction_request: prediction_pb2.SeldonMessage = None,
    prediction_response: prediction_pb2.SeldonMessage = None,
    reward: float = 0,
    microservice_endpoint: str = None,
    **kwargs,
) -> SeldonClientFeedback:
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
    request = prediction_pb2.Feedback(
        request=prediction_request, response=prediction_response, reward=reward
    )
    payload = feedback_to_json(request)
    response_raw = requests.post(
        "http://" + microservice_endpoint + "/send-feedback",
        data={"json": json.dumps(payload)},
    )
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


def microservice_api_grpc_seldon_message(
    method: str = "predict",
    microservice_endpoint: str = "localhost:5000",
    shape: Tuple = (1, 1),
    data: object = None,
    payload_type: str = "tensor",
    bin_data: Union[bytes, bytearray] = None,
    str_data: str = None,
    json_data: Union[str, List, Dict] = None,
    custom_data: any_pb2.Any = None,
    grpc_max_send_message_length: int = 4 * 1024 * 1024,
    grpc_max_receive_message_length: int = 4 * 1024 * 1024,
    names: Iterable[str] = None,
    **kwargs,
) -> SeldonClientPrediction:
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
    json_data
        JSON data to send
    custom_data
        Custom data to send
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
    elif json_data is not None:
        request = json_to_seldon_message({"jsonData": json_data})
    elif custom_data is not None:
        request = prediction_pb2.SeldonMessage(customData=custom_data)
    else:
        if data is None:
            data = np.random.rand(*shape)
        datadef = array_to_grpc_datadef(payload_type, data, names=names)
        request = prediction_pb2.SeldonMessage(data=datadef)
    channel = grpc.insecure_channel(
        microservice_endpoint,
        options=[
            ("grpc.max_send_message_length", grpc_max_send_message_length),
            ("grpc.max_receive_message_length", grpc_max_receive_message_length),
        ],
    )
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

        channel.close()
        return SeldonClientPrediction(request, response, True, "")
    except Exception as e:
        channel.close()
        return SeldonClientPrediction(request, None, False, str(e))


def microservice_api_grpc_aggregate(
    microservice_endpoint: str = "localhost:5000",
    shape: Tuple = (1, 1),
    datas: List[np.ndarray] = None,
    ndatas: int = None,
    payload_type: str = "tensor",
    grpc_max_send_message_length: int = 4 * 1024 * 1024,
    grpc_max_receive_message_length: int = 4 * 1024 * 1024,
    names: Iterable[str] = None,
    **kwargs,
) -> SeldonClientCombine:
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
        elif isinstance(data, any_pb2.Any):
            msgs.append(prediction_pb2.SeldonMessage(customData=data))
        else:
            datadef = array_to_grpc_datadef(payload_type, data, names=names)
            msgs.append(prediction_pb2.SeldonMessage(data=datadef))
    request = prediction_pb2.SeldonMessageList(seldonMessages=msgs)
    try:
        channel = grpc.insecure_channel(
            microservice_endpoint,
            options=[
                ("grpc.max_send_message_length", grpc_max_send_message_length),
                ("grpc.max_receive_message_length", grpc_max_receive_message_length),
            ],
        )
        stub = prediction_pb2_grpc.GenericStub(channel)
        response = stub.Aggregate(request=request)
        channel.close()
        return SeldonClientCombine(request, response, True, "")
    except Exception as e:
        return SeldonClientCombine(request, None, False, str(e))


def microservice_api_grpc_feedback(
    prediction_request: prediction_pb2.SeldonMessage = None,
    prediction_response: prediction_pb2.SeldonMessage = None,
    reward: float = 0,
    microservice_endpoint: str = None,
    grpc_max_send_message_length: int = 4 * 1024 * 1024,
    grpc_max_receive_message_length: int = 4 * 1024 * 1024,
    **kwargs,
) -> SeldonClientFeedback:
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
    request = prediction_pb2.Feedback(
        request=prediction_request, response=prediction_response, reward=reward
    )
    try:
        channel = grpc.insecure_channel(
            microservice_endpoint,
            options=[
                ("grpc.max_send_message_length", grpc_max_send_message_length),
                ("grpc.max_receive_message_length", grpc_max_receive_message_length),
            ],
        )
        stub = prediction_pb2_grpc.GenericStub(channel)
        response = stub.SendFeedback(request=request)
        channel.close()
        return SeldonClientFeedback(request, response, True, "")
    except Exception as e:
        return SeldonClientFeedback(request, None, False, str(e))


#
# External API
#


def rest_predict_seldon(
    namespace: str = None,
    gateway_endpoint: str = "localhost:8002",
    seldon_rest_endpoint: str = "localhost:8002",
    shape: Tuple = (1, 1),
    data: object = None,
    payload_type: str = "tensor",
    bin_data: Union[bytes, bytearray] = None,
    str_data: str = None,
    json_data: Union[str, List, Dict] = None,
    names: Iterable[str] = None,
    client_return_type: str = "proto",
    **kwargs,
) -> SeldonClientPrediction:
    """
    Call Seldon API Gateway using REST

    Parameters
    ----------
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
    json_data
        JSON data to send
    names
       column names
    client_return_type
        the return type of all functions can be either dict or proto
    kwargs

    Returns
    -------
       Seldon Client Prediction

    """
    if bin_data is not None:
        request = prediction_pb2.SeldonMessage(binData=bin_data)
    elif str_data is not None:
        request = prediction_pb2.SeldonMessage(strData=str_data)
    elif json_data is not None:
        request = json_to_seldon_message({"jsonData": json_data})
    else:
        if data is None:
            data = np.random.rand(*shape)
        datadef = array_to_grpc_datadef(payload_type, data, names=names)
        request = prediction_pb2.SeldonMessage(data=datadef)
    payload = seldon_message_to_json(request)

    rest_endpoint = gateway_endpoint or seldon_rest_endpoint
    response_raw = requests.post(
        "http://" + rest_endpoint + "/api/v0.1/predictions", json=payload,
    )
    if response_raw.status_code == 200:
        success = True
        msg = ""
    else:
        success = False
        msg = str(response_raw.status_code) + ":" + response_raw.reason
    try:
        if len(response_raw.text) > 0:
            try:
                if client_return_type == "proto":
                    response = json_to_seldon_message(response_raw.json())
                elif client_return_type == "dict":
                    response = response_raw.json()
                else:
                    SeldonClientException("Invalid client_return_type")
            except:
                response = None
        else:
            response = None
        return SeldonClientPrediction(request, response, success, msg)
    except Exception as e:
        return SeldonClientPrediction(request, None, False, str(e))


def grpc_predict_seldon(
    namespace: str = None,
    gateway_endpoint: str = "localhost:8004",
    seldon_grpc_endpoint: str = "localhost:8004",
    shape: Tuple[int, int] = (1, 1),
    data: np.ndarray = None,
    payload_type: str = "tensor",
    bin_data: Union[bytes, bytearray] = None,
    str_data: str = None,
    json_data: Union[str, List, Dict] = None,
    custom_data: any_pb2.Any = None,
    grpc_max_send_message_length: int = 4 * 1024 * 1024,
    grpc_max_receive_message_length: int = 4 * 1024 * 1024,
    names: Iterable[str] = None,
    client_return_type: str = "proto",
    **kwargs,
) -> SeldonClientPrediction:
    """
    Call Seldon gRPC API Gateway endpoint

    Parameters
    ----------
    namespace
       k8s namespace of running deployment
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
    json_data
       JSON data to send
    custom_data
       Custom data to send
    grpc_max_send_message_length
       Max grpc send message size in bytes
    grpc_max_receive_message_length
       Max grpc receive message size in bytes
    names
       Column names
    client_return_type
        the return type of all functions can be either dict or proto
    kwargs

    Returns
    -------
       A SeldonMessage proto

    """
    if bin_data is not None:
        request = prediction_pb2.SeldonMessage(binData=bin_data)
    elif str_data is not None:
        request = prediction_pb2.SeldonMessage(strData=str_data)
    elif json_data is not None:
        request = json_to_seldon_message({"jsonData": json_data})
    elif custom_data is not None:
        request = prediction_pb2.SeldonMessage(customData=custom_data)
    else:
        if data is None:
            data = np.random.rand(*shape)
        datadef = array_to_grpc_datadef(payload_type, data, names=names)
        request = prediction_pb2.SeldonMessage(data=datadef)

    grpc_endpoint = gateway_endpoint or seldon_grpc_endpoint
    channel = grpc.insecure_channel(
        grpc_endpoint,
        options=[
            ("grpc.max_send_message_length", grpc_max_send_message_length),
            ("grpc.max_receive_message_length", grpc_max_receive_message_length),
        ],
    )
    stub = prediction_pb2_grpc.SeldonStub(channel)
    try:
        response = stub.Predict(request=request)
        channel.close()
        if client_return_type == "dict":
            request = seldon_message_to_json(request)
            response = seldon_message_to_json(response)
        elif client_return_type != "proto":
            raise SeldonClientException("Invalid client_return_type")
        return SeldonClientPrediction(request, response, True, "")
    except Exception as e:
        channel.close()
        return SeldonClientPrediction(request, None, False, str(e))


def rest_predict_gateway(
    deployment_name: str,
    namespace: str = None,
    gateway_endpoint: str = "localhost:8003",
    shape: Tuple[int, int] = (1, 1),
    data: np.ndarray = None,
    headers: Dict = None,
    gateway_prefix: str = None,
    payload_type: str = "tensor",
    bin_data: Union[bytes, bytearray] = None,
    str_data: str = None,
    json_data: Union[str, Dict, List] = None,
    names: Iterable[str] = None,
    call_credentials: SeldonCallCredentials = None,
    channel_credentials: SeldonChannelCredentials = None,
    http_path: str = None,
    meta: Dict = {},
    client_return_type: str = "proto",
    **kwargs,
) -> SeldonClientPrediction:
    """
    REST request to Gateway Ingress

    Parameters
    ----------
    deployment_name
       The name of the Seldon Deployment
    namespace
       k8s namespace of running deployment
    gateway_endpoint
       The host:port of gateway
    shape
       The shape of the data to send
    data
       The numpy data to send
    headers
       Headers to add to request
    gateway_prefix
       The prefix path to add to the request
    payload_type
       payload - tensor, ndarray or tftensor
    bin_data
       Binary data to send
    str_data
       String data to send
    json_data
       JSON data to send as str, dict or list
    names
       Column names
    call_credentials
       Call credentials - see SeldonCallCredentials
    channel_credentials
       Channel credentials - see SeldonChannelCredentials
    http_path
       Custom http path
    meta
       Custom meta map
    client_return_type
        the return type of all functions can be either dict or proto

    Returns
    -------
       A requests Response object

    """
    # Create meta data
    metaKV = prediction_pb2.Meta()
    metaJson = {"tags": meta}
    json_format.ParseDict(metaJson, metaKV)

    if bin_data is not None:
        request = prediction_pb2.SeldonMessage(binData=bin_data, meta=metaKV)
    elif str_data is not None:
        request = prediction_pb2.SeldonMessage(strData=str_data, meta=metaKV)
    elif json_data is not None:
        request = json_to_seldon_message({"jsonData": json_data})
    else:
        if data is None:
            data = np.random.rand(*shape)
        datadef = array_to_grpc_datadef(payload_type, data, names=names)
        request = prediction_pb2.SeldonMessage(data=datadef, meta=metaKV)
    payload = seldon_message_to_json(request)

    if not headers is None:
        req_headers = headers.copy()
    else:
        req_headers = {}
    if call_credentials is None:
        scheme = "http"
    else:
        scheme = "https"
        if not call_credentials is None:
            if not call_credentials.token is None:
                req_headers["X-Auth-Token"] = call_credentials.token
    if http_path is not None:
        url = url = (
            scheme
            + "://"
            + gateway_endpoint
            + "/seldon/"
            + namespace
            + "/"
            + deployment_name
            + http_path
        )
    else:
        if gateway_prefix is None:
            if namespace is None:
                url = (
                    scheme
                    + "://"
                    + gateway_endpoint
                    + "/seldon/"
                    + deployment_name
                    + "/api/v1.0/predictions"
                )
            else:
                url = (
                    scheme
                    + "://"
                    + gateway_endpoint
                    + "/seldon/"
                    + namespace
                    + "/"
                    + deployment_name
                    + "/api/v1.0/predictions"
                )
        else:
            url = (
                scheme
                + "://"
                + gateway_endpoint
                + gateway_prefix
                + "/api/v1.0/predictions"
            )
    verify = True
    cert = None
    if not channel_credentials is None:
        if not channel_credentials.certificate_chain_file is None:
            verify = channel_credentials.certificate_chain_file
        else:
            verify = channel_credentials.verify
        if not channel_credentials.private_key_file is None:
            cert = (
                channel_credentials.root_certificates_file,
                channel_credentials.private_key_file,
            )
    logger.debug("URL is " + url)
    response_raw = requests.post(
        url, json=payload, headers=req_headers, verify=verify, cert=cert
    )
    if response_raw.status_code == 200:
        success = True
        msg = ""
    else:
        success = False
        msg = str(response_raw.status_code) + ":" + response_raw.reason
    try:
        if len(response_raw.text) > 0:
            try:
                logger.debug("Raw response: %s", response_raw.text)
                if client_return_type == "proto":
                    response = json_to_seldon_message(response_raw.json())
                elif client_return_type == "dict":
                    response = response_raw.json()
                else:
                    raise SeldonClientException("Invalid client_return_type")
            except:
                response = None
        else:
            response = None
        return SeldonClientPrediction(request, response, success, msg)
    except Exception as e:
        return SeldonClientPrediction(request, None, False, str(e))


def explain_predict_gateway(
    deployment_name: str,
    namespace: str = None,
    gateway_endpoint: str = "localhost:8003",
    gateway: str = None,
    transport: str = "rest",
    shape: Tuple[int, int] = (1, 1),
    data: np.ndarray = None,
    headers: Dict = None,
    gateway_prefix: str = None,
    payload_type: str = "tensor",
    bin_data: Union[bytes, bytearray] = None,
    str_data: str = None,
    json_data: Union[str, List, Dict] = None,
    names: Iterable[str] = None,
    call_credentials: SeldonCallCredentials = None,
    channel_credentials: SeldonChannelCredentials = None,
    http_path: str = None,
    client_return_type: str = "dict",
    predictor: str = None,
    **kwargs,
) -> SeldonClientPrediction:
    """
    REST explain request to Gateway Ingress

    Parameters
    ----------
    deployment_name
       The name of the Seldon Deployment
    namespace
       k8s namespace of running deployment
    gateway_endpoint
       The host:port of gateway
    gateway
       The type of gateway which can be seldon or ambassador/istio
    transport
       The type of transport, in this case only rest is supported
    shape
       The shape of the data to send
    data
       The numpy data to send
    headers
       Headers to add to request
    gateway_prefix
       The prefix path to add to the request
    payload_type
       payload - tensor, ndarray or tftensor
    bin_data
       Binary data to send
    str_data
       String data to send
    json_data
       JSON data to send
    names
       Column names
    call_credentials
       Call credentials - see SeldonCallCredentials
    channel_credentials
       Channel credentials - see SeldonChannelCredentials
    http_path
       Custom http path
    client_return_type
        the return type of all functions can be either dict or proto

    Returns
    -------
       A JSON Dict

    """
    if transport != "rest":
        raise SeldonClientException("Only supported transport is REST for explanations")

    if bin_data is not None:
        request = prediction_pb2.SeldonMessage(binData=bin_data)
    elif str_data is not None:
        request = prediction_pb2.SeldonMessage(strData=str_data)
    elif json_data is not None:
        request = json_to_seldon_message({"jsonData": json_data})
    else:
        if data is None:
            data = np.random.rand(*shape)
        datadef = array_to_grpc_datadef(payload_type, data, names=names)
        request = prediction_pb2.SeldonMessage(data=datadef)
    payload = seldon_message_to_json(request)

    if not headers is None:
        req_headers = headers.copy()
    else:
        req_headers = {}
    if channel_credentials is None:
        scheme = "http"
    else:
        scheme = "https"
        if not call_credentials is None:
            if not call_credentials.token is None:
                req_headers["X-Auth-Token"] = call_credentials.token
    if http_path is not None:
        url = (
            scheme
            + "://"
            + gateway_endpoint
            + "/seldon/"
            + namespace
            + "/"
            + deployment_name
            + "-explainer"
            + "/"
            + predictor
            + http_path
        )
    elif gateway == "seldon":
        url = scheme + "://" + gateway_endpoint + "/api/v1.0/explain"
    else:
        if not predictor:
            raise SeldonClientException(
                "Predictor parameter must be provided to talk through explainer via gateway"
            )

        if gateway_prefix is None:
            if namespace is None:
                url = (
                    scheme
                    + "://"
                    + gateway_endpoint
                    + "/seldon/"
                    + deployment_name
                    + "-explainer"
                    + "/"
                    + predictor
                    + "/api/v1.0/explain"
                )
            else:
                url = (
                    scheme
                    + "://"
                    + gateway_endpoint
                    + "/seldon/"
                    + namespace
                    + "/"
                    + deployment_name
                    + "-explainer"
                    + "/"
                    + predictor
                    + "/api/v1.0/explain"
                )
        else:
            url = (
                scheme + "://" + gateway_endpoint + gateway_prefix + "/api/v1.0/explain"
            )
    verify = True
    cert = None
    if not channel_credentials is None:
        if not channel_credentials.certificate_chain_file is None:
            verify = channel_credentials.certificate_chain_file
        else:
            verify = channel_credentials.verify
        if not channel_credentials.private_key_file is None:
            cert = (
                channel_credentials.root_certificates_file,
                channel_credentials.private_key_file,
            )
    logger.debug("URL is " + url)
    response_raw = requests.post(
        url, json=payload, headers=req_headers, verify=verify, cert=cert
    )
    if response_raw.status_code == 200:
        if client_return_type == "dict":
            ret_request = payload
            ret_response = response_raw.json()
        else:
            raise SeldonClientException("Invalid client_return_type")
        return SeldonClientPrediction(ret_request, ret_response, True, "")
    else:
        return SeldonClientPrediction(
            payload,
            response_raw,
            False,
            f"Unsuccessful request with status code: {response_raw.status_code}",
        )


def grpc_predict_gateway(
    deployment_name: str,
    namespace: str = None,
    gateway_endpoint: str = "localhost:8003",
    shape: Tuple[int, int] = (1, 1),
    data: np.ndarray = None,
    headers: Dict = None,
    payload_type: str = "tensor",
    bin_data: Union[bytes, bytearray] = None,
    str_data: str = None,
    json_data: Union[str, List, Dict] = None,
    custom_data: any_pb2.Any = None,
    grpc_max_send_message_length: int = 4 * 1024 * 1024,
    grpc_max_receive_message_length: int = 4 * 1024 * 1024,
    names: Iterable[str] = None,
    call_credentials: SeldonCallCredentials = None,
    channel_credentials: SeldonChannelCredentials = None,
    meta: Dict = {},
    client_return_type: str = "proto",
    **kwargs,
) -> SeldonClientPrediction:
    """
    gRPC request to Gateway Ingress

    Parameters
    ----------
    deployment_name
       Deployment name of Seldon Deployment
    namespace
       The namespace the Seldon Deployment is running in
    gateway_endpoint
       The endpoint for gateway
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
    json_data
       JSON data to send
    custom_data
       Custom data to send
    grpc_max_send_message_length
       Max grpc send message size in bytes
    grpc_max_receive_message_length
       Max grpc receive message size in bytes
    names
       Column names
    call_credentials
       Call credentials - see SeldonCallCredentials
    channel_credentials
       Channel credentials - see SeldonChannelCredentials
    meta
       Custom meta data map
    client_return_type
        the return type of all functions can be either dict or proto


    Returns
    -------
       A SeldonMessage proto response

    """

    # Create meta data
    metaKV = prediction_pb2.Meta()
    metaJson = {"tags": meta}
    json_format.ParseDict(metaJson, metaKV)

    if bin_data is not None:
        request = prediction_pb2.SeldonMessage(binData=bin_data, meta=metaKV)
    elif str_data is not None:
        request = prediction_pb2.SeldonMessage(strData=str_data, meta=metaKV)
    elif json_data is not None:
        request = json_to_seldon_message({"jsonData": json_data})
    elif custom_data is not None:
        request = prediction_pb2.SeldonMessage(customData=custom_data, meta=metaKV)
    else:
        if data is None:
            data = np.random.rand(*shape)
        datadef = array_to_grpc_datadef(payload_type, data, names=names)
        request = prediction_pb2.SeldonMessage(data=datadef, meta=metaKV)
    options = [
        ("grpc.max_send_message_length", grpc_max_send_message_length),
        ("grpc.max_receive_message_length", grpc_max_receive_message_length),
    ]
    if channel_credentials is None:
        channel = grpc.insecure_channel(gateway_endpoint, options)
    else:
        # If one of root cert & cert chain are provided, both must be provided
        #   otherwise there is a null pointer exception in the Go underlying impl
        if (
            channel_credentials.private_key_file
            and channel_credentials.root_certificates_file
            and channel_credentials.certificate_chain_file
        ):
            grpc_channel_credentials = grpc.ssl_channel_credentials(
                root_certificates=open(
                    channel_credentials.root_certificates_file, "rb"
                ).read(),
                private_key=open(channel_credentials.private_key_file, "rb").read(),
                certificate_chain=open(
                    channel_credentials.certificate_chain_file, "rb"
                ).read(),
            )
        # For most usecases only providing the root cert file is enough
        elif channel_credentials.root_certificates_file:
            grpc_channel_credentials = grpc.ssl_channel_credentials(
                root_certificates=open(
                    channel_credentials.root_certificates_file, "rb"
                ).read()
            )
        # This piece also allows for blank SSL Channel credentials in case this is required
        else:
            grpc_channel_credentials = grpc.ssl_channel_credentials()
        if channel_credentials.verify == False:
            # If Verify is set to false then we add the SSL Target Name Override option
            options += [
                ("grpc.ssl_target_name_override", gateway_endpoint.split(":")[0])
            ]

        if not call_credentials is None:
            grpc_call_credentials = grpc.metadata_call_credentials(
                lambda context, callback: callback(
                    (("x-auth-token", call_credentials.token),), None
                )
            )
            credentials = grpc.composite_channel_credentials(
                grpc_channel_credentials, grpc_call_credentials
            )
        else:
            credentials = grpc_channel_credentials
        logger.debug(f"Sending GRPC Request to endpoint: {gateway_endpoint}")
        channel = grpc.secure_channel(gateway_endpoint, credentials, options)
    stub = prediction_pb2_grpc.SeldonStub(channel)
    if namespace is None:
        metadata = [("seldon", deployment_name)]
    else:
        metadata = [("seldon", deployment_name), ("namespace", namespace)]
    if not headers is None:
        for k in headers:
            metadata.append((k, headers[k]))
    try:
        response = stub.Predict(request=request, metadata=metadata)
        channel.close()
        if client_return_type == "dict":
            request = seldon_message_to_json(request)
            response = seldon_message_to_json(response)
        elif client_return_type != "proto":
            raise SeldonClientException("Invalid client_return_type")
        return SeldonClientPrediction(request, response, True, "")
    except Exception as e:
        channel.close()
        return SeldonClientPrediction(request, None, False, str(e))


def rest_feedback_seldon(
    prediction_request: prediction_pb2.SeldonMessage = None,
    prediction_response: prediction_pb2.SeldonMessage = None,
    prediction_truth: prediction_pb2.SeldonMessage = None,
    reward: float = 0,
    namespace: str = None,
    gateway_endpoint: str = "localhost:8002",
    seldon_rest_endpoint: str = "localhost:8002",
    client_return_type: str = "proto",
    raw_request: dict = None,
    **kwargs,
) -> SeldonClientFeedback:
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
    namespace
       k8s namespace of running deployment
    seldon_rest_endpoint
       Endpoint of REST endpoint
    client_return_type
        the return type of all functions can be either dict or proto
    kwargs

    Returns
    -------

    """

    if raw_request:
        request = json_to_feedback(raw_request)
        payload = raw_request
    else:
        request = prediction_pb2.Feedback(
            request=prediction_request,
            response=prediction_response,
            reward=reward,
            truth=prediction_truth,
        )
        payload = feedback_to_json(request)

    rest_endpoint = gateway_endpoint or seldon_rest_endpoint
    response_raw = requests.post(
        "http://" + rest_endpoint + "/api/v1.0/feedback", json=payload,
    )
    if response_raw.status_code == 200:
        success = True
        msg = ""
    else:
        success = False
        msg = str(response_raw.status_code) + ":" + response_raw.reason
    try:
        if len(response_raw.text) > 0:
            try:
                if client_return_type == "proto":
                    response = json_to_seldon_message(response_raw.json())
                elif client_return_type == "dict":
                    response = response_raw.json()
                else:
                    raise SeldonClientException("Invalid client_return_type")
            except:
                response = None
        else:
            response = None
        return SeldonClientFeedback(request, response, success, msg)
    except Exception as e:
        return SeldonClientFeedback(request, None, False, str(e))


def grpc_feedback_seldon(
    prediction_request: prediction_pb2.SeldonMessage = None,
    prediction_response: prediction_pb2.SeldonMessage = None,
    prediction_truth: prediction_pb2.SeldonMessage = None,
    reward: float = 0,
    namespace: str = None,
    gateway_endpoint: str = "localhost:8004",
    seldon_grpc_endpoint: str = "localhost:8004",
    grpc_max_send_message_length: int = 4 * 1024 * 1024,
    grpc_max_receive_message_length: int = 4 * 1024 * 1024,
    client_return_type: str = "proto",
    raw_request: dict = None,
    **kwargs,
) -> SeldonClientFeedback:
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
    namespace
       k8s namespace of running deployment
    seldon_grpc_endpoint
       Endpoint for Seldon grpc
    grpc_max_send_message_length
       Max grpc send message size in bytes
    grpc_max_receive_message_length
       Max grpc receive message size in bytes
    client_return_type
        the return type of all functions can be either dict or proto
    kwargs

    Returns
    -------

    """

    if isinstance(raw_request, prediction_pb2.Feedback):
        request = raw_request
    elif raw_request:
        request = json_to_feedback(raw_request)
    else:
        request = prediction_pb2.Feedback(
            request=prediction_request,
            response=prediction_response,
            reward=reward,
            truth=prediction_truth,
        )

    grpc_endpoint = gateway_endpoint or seldon_grpc_endpoint
    channel = grpc.insecure_channel(
        grpc_endpoint,
        options=[
            ("grpc.max_send_message_length", grpc_max_send_message_length),
            ("grpc.max_receive_message_length", grpc_max_receive_message_length),
        ],
    )
    stub = prediction_pb2_grpc.SeldonStub(channel)
    try:
        response = stub.SendFeedback(request=request)
        channel.close()
        if client_return_type == "dict":
            request = seldon_message_to_json(request)
            response = seldon_message_to_json(response)
        elif client_return_type != "proto":
            raise SeldonClientException("Invalid client_return_type")
        return SeldonClientFeedback(request, response, True, "")
    except Exception as e:
        channel.close()
        return SeldonClientFeedback(request, None, False, str(e))


def rest_feedback_gateway(
    prediction_request: prediction_pb2.SeldonMessage = None,
    prediction_response: prediction_pb2.SeldonMessage = None,
    prediction_truth: prediction_pb2.SeldonMessage = None,
    reward: float = 0,
    deployment_name: str = "",
    namespace: str = None,
    gateway_endpoint: str = "localhost:8003",
    headers: Dict = None,
    gateway_prefix: str = None,
    client_return_type: str = "proto",
    raw_request: dict = None,
    **kwargs,
) -> SeldonClientFeedback:
    """
    Send Feedback to Seldon via gateway using REST

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
    gateway_endpoint
       The gateway host:port endpoint
    headers
       Headers to add to the request
    gateway_prefix
      The prefix to add to the request path for gateway
    client_return_type
        the return type of all functions can be either dict or proto
    kwargs

    Returns
    -------
      A Seldon Feedback Response

    """
    if raw_request:
        request = json_to_feedback(raw_request)
        payload = raw_request
    else:
        request = prediction_pb2.Feedback(
            request=prediction_request,
            response=prediction_response,
            reward=reward,
            truth=prediction_truth,
        )
        payload = feedback_to_json(request)
    if gateway_prefix is None:
        if namespace is None:
            response_raw = requests.post(
                "http://"
                + gateway_endpoint
                + "/seldon/"
                + deployment_name
                + "/api/v1.0/feedback",
                json=payload,
                headers=headers,
            )
        else:
            response_raw = requests.post(
                "http://"
                + gateway_endpoint
                + "/seldon/"
                + namespace
                + "/"
                + deployment_name
                + "/api/v1.0/feedback",
                json=payload,
                headers=headers,
            )
    else:
        response_raw = requests.post(
            "http://" + gateway_endpoint + gateway_prefix + "/api/v1.0/feedback",
            json=payload,
            headers=headers,
        )

    if response_raw.status_code == 200:
        success = True
        msg = ""
    else:
        success = False
        msg = str(response_raw.status_code) + ":" + response_raw.reason
    try:
        if len(response_raw.text) > 0:
            try:
                if client_return_type == "proto":
                    response = json_to_seldon_message(response_raw.json())
                elif client_return_type == "dict":
                    response = response_raw.json()
                else:
                    raise SeldonClientException("Invalid client_return_type")
            except:
                response = None
        else:
            response = None
        return SeldonClientFeedback(request, response, success, msg)
    except Exception as e:
        return SeldonClientFeedback(request, None, False, str(e))


def grpc_feedback_gateway(
    prediction_request: prediction_pb2.SeldonMessage = None,
    prediction_response: prediction_pb2.SeldonMessage = None,
    prediction_truth: prediction_pb2.SeldonMessage = None,
    reward: float = 0,
    deployment_name: str = "",
    namespace: str = None,
    gateway_endpoint: str = "localhost:8003",
    headers: Dict = None,
    grpc_max_send_message_length: int = 4 * 1024 * 1024,
    grpc_max_receive_message_length: int = 4 * 1024 * 1024,
    client_return_type: str = "proto",
    raw_request: dict = None,
    **kwargs,
) -> SeldonClientFeedback:
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
    gateway_endpoint
       The gateway host:port endpoint
    headers
       Headers to add to the request
    grpc_max_send_message_length
       Max grpc send message size in bytes
    grpc_max_receive_message_length
       Max grpc receive message size in bytes
    client_return_type
        the return type of all functions can be either dict or proto
    kwargs

    Returns
    -------

    """
    if isinstance(raw_request, prediction_pb2.Feedback):
        request = raw_request
    elif raw_request:
        request = json_to_feedback(raw_request)
    else:
        request = prediction_pb2.Feedback(
            request=prediction_request,
            response=prediction_response,
            reward=reward,
            truth=prediction_truth,
        )
    channel = grpc.insecure_channel(
        gateway_endpoint,
        options=[
            ("grpc.max_send_message_length", grpc_max_send_message_length),
            ("grpc.max_receive_message_length", grpc_max_receive_message_length),
        ],
    )
    stub = prediction_pb2_grpc.SeldonStub(channel)
    if namespace is None:
        metadata = [("seldon", deployment_name)]
    else:
        metadata = [("seldon", deployment_name), ("namespace", namespace)]
    if not headers is None:
        for k in headers:
            metadata.append((k, headers[k]))
    try:
        response = stub.SendFeedback(request=request, metadata=metadata)
        channel.close()
        if client_return_type == "dict":
            request = seldon_message_to_json(request)
            response = seldon_message_to_json(response)
        elif client_return_type != "proto":
            raise SeldonClientException("Invalid client_return_type")
        return SeldonClientFeedback(request, response, True, "")
    except Exception as e:
        channel.close()
        return SeldonClientFeedback(request, None, False, str(e))
