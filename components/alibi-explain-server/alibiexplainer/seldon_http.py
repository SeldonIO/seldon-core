from enum import Enum
from http import HTTPStatus
from typing import Dict, List, Union

import numpy as np
import requests
import tornado.web


class SeldonPayload(Enum):
    TENSOR = 1
    NDARRAY = 2
    TFTENSOR = 3


def _extract_list(body: Dict) -> List:
    data_def = body["data"]
    if "tensor" in data_def:
        arr = np.array(data_def.get("tensor").get("values")).reshape(
            data_def.get("tensor").get("shape")
        )
        return arr.tolist()
    elif "ndarray" in data_def:
        return data_def.get("ndarray")
    else:
        arr = np.array(data_def["tftensor"]["float_val"])
        shape = []
        for dim in data_def["tftensor"]["tensor_shape"]["dim"]:
            shape.append(dim["size"])
        arr = arr.reshape(shape)
        return arr


def _create_seldon_data_def(array: np.array, ty: SeldonPayload):
    datadef = {}
    if ty == SeldonPayload.TENSOR:
        datadef["tensor"] = {"shape": array.shape, "values": array.ravel().tolist()}
    elif ty == SeldonPayload.NDARRAY:
        datadef["ndarray"] = array.tolist()
    elif ty == SeldonPayload.TFTENSOR:
        raise NotImplementedError("Seldon payload %s not supported" % ty)
    else:
        raise Exception("Unknown Seldon payload %s" % ty)
    return datadef


def _get_request_ty(
    request: Dict,
) -> SeldonPayload:  # pylint: disable=inconsistent-return-statements
    data_def = request["data"]
    if "tensor" in data_def:
        return SeldonPayload.TENSOR
    elif "ndarray" in data_def:
        return SeldonPayload.NDARRAY
    elif "tftensor" in data_def:
        return SeldonPayload.TFTENSOR
    else:
        raise Exception("Unknown Seldon payload type %s" % data_def)


def create_request(arr: np.ndarray, ty: SeldonPayload) -> Dict:
    seldon_datadef = _create_seldon_data_def(arr, ty)
    return {"data": seldon_datadef}


class SeldonRequestHandler:
    def __init__(self, request: Dict):  # pylint: disable=useless-super-delegation
        self.request = request

    def validate(self):
        if not "data" in self.request:
            raise tornado.web.HTTPError(
                status_code=HTTPStatus.BAD_REQUEST,
                reason='Expected key "data" in request body',
            )
        ty = _get_request_ty(self.request)
        if not (ty == SeldonPayload.TENSOR or ty == SeldonPayload.NDARRAY):
            raise tornado.web.HTTPError(
                status_code=HTTPStatus.BAD_REQUEST,
                reason='"data" key should contain either "tensor","ndarray"',
            )

    def extract_request(self) -> List:
        return _extract_list(self.request)

    def wrap_response(self, response: List) -> Dict:
        arr = np.array(response)
        ty = _get_request_ty(self.request)
        seldon_datadef = _create_seldon_data_def(arr, ty)
        return {"data": seldon_datadef}

    @staticmethod
    def predict(inputs: Union[np.array, List], predictor_url: str) -> List:
        if isinstance(inputs, list):
            inputs = np.array(inputs)
        payload = create_request(inputs, SeldonPayload.NDARRAY)
        response_raw = requests.post(predictor_url, json=payload)
        if response_raw.status_code != 200:
            raise tornado.web.HTTPError(
                status_code=response_raw.status_code, reason=response_raw.reason
            )
        rh = SeldonRequestHandler(response_raw.json())
        response_list = rh.extract_request()
        return response_list
