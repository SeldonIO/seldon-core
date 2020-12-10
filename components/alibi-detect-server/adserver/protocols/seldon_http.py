from enum import Enum
from http import HTTPStatus
from typing import Dict, List

import numpy as np
import tornado
from adserver.protocols.request_handler import (
    RequestHandler,
)  # pylint: disable=no-name-in-module


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
    elif "tftensor" in data_def:
        arr = np.array(data_def["tftensor"]["float_val"])
        shape = []
        for dim in data_def["tftensor"]["tensor_shape"]["dim"]:
            shape.append(dim["size"])
        arr = arr.reshape(shape)
        return arr.tolist()
    else:
        raise Exception("Unknown Seldon payload %s" % body)


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
        raise Exception("Unknown Seldon payload %s" % data_def)


def create_request(arr: np.ndarray, ty: SeldonPayload) -> Dict:
    seldon_datadef = _create_seldon_data_def(arr, ty)
    return {"data": seldon_datadef}


class SeldonRequestHandler(RequestHandler):
    def __init__(self, request: Dict):  # pylint: disable=useless-super-delegation
        super().__init__(request)

    def validate(self):
        if not "data" in self.request:
            raise tornado.web.HTTPError(
                status_code=HTTPStatus.BAD_REQUEST,
                reason='Expected key "data" in request body',
            )
        ty = _get_request_ty(self.request)
        if not (
            ty == SeldonPayload.TENSOR
            or ty == SeldonPayload.NDARRAY
            or ty == SeldonPayload.TFTENSOR
        ):
            raise tornado.web.HTTPError(
                status_code=HTTPStatus.BAD_REQUEST,
                reason='"data" key should contain either "tensor","ndarray", or "tftensor"',
            )

    def extract_request(self) -> List:
        return _extract_list(self.request)
