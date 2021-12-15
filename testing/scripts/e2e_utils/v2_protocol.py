import requests
import grpc

from typing import Optional
from tenacity import (
    retry,
    retry_if_exception_type,
    stop_after_attempt,
    wait_exponential,
)
from google.protobuf.json_format import MessageToDict, ParseDict
from mlserver.grpc.dataplane_pb2_grpc import GRPCInferenceServiceStub
from mlserver.grpc.dataplane_pb2 import ModelInferRequest

from seldon_e2e_utils import API_AMBASSADOR


@retry(
    wait=wait_exponential(multiplier=1),
    stop=stop_after_attempt(3),
    retry=retry_if_exception_type(
        (requests.exceptions.ConnectionError, requests.exceptions.HTTPError)
    ),
)
def inference_request(
    deployment_name: str,
    namespace: str,
    payload: dict,
    model_name: Optional[str] = None,
    host: str = API_AMBASSADOR,
) -> dict:
    root_endpoint = f"http://{host}/seldon/{namespace}/{deployment_name}"

    endpoint = f"{root_endpoint}/v2/models/infer"
    if model_name:
        endpoint = f"{root_endpoint}/v2/models/{model_name}/infer"

    response = requests.post(endpoint, json=payload)
    response.raise_for_status()

    return response.json()


# TODO: Retry for gRPC errors
def inference_request_grpc(
    deployment_name: str,
    namespace: str,
    payload: dict,
    model_name: Optional[str] = None,
    host: str = API_AMBASSADOR,
) -> dict:
    if model_name:
        payload["model_name"] = model_name

    with grpc.insecure_channel(host) as channel:
        stub = GRPCInferenceServiceStub(channel)
        request = ParseDict(payload, ModelInferRequest())
        metadata = [("seldon", deployment_name), ("namespace", namespace)]
        response = stub.ModelInfer(request=request, metadata=metadata)
        return MessageToDict(response)
