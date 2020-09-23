from enum import Enum
from http import HTTPStatus
from typing import Dict, List, Union

import numpy as np
import tornado
from seldon_core.utils import extract_request_parts_json
from adserver.protocols.request_handler import (
    RequestHandler,
)  # pylint: disable=no-name-in-module


class SeldonPayload(Enum):
    TENSOR = 1
    NDARRAY = 2
    TFTENSOR = 3


def _extract_feedback_request(body: Dict) -> List:
    res = {}

    if "truth" in body:
        res["truth"] = extract_request_parts_json(body["truth"])
        if "meta" in body["truth"]:
            if "metrics" in body["truth"]["meta"]:
                res["metrics"] = body["truth"]["meta"]["metrics"]
    if "request" in body:
        res["request"] = extract_request_parts_json(body["request"])
    if "response" in body:
        res["response"] = extract_request_parts_json(body["response"])

    return res

def _get_request_ty(
    request: Dict
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


class SeldonFeedbackRequestHandler(RequestHandler):
    def __init__(self, request: Dict):  # pylint: disable=useless-super-delegation
        super().__init__(request)

    def _validate_seldon_message(self, message):
        if not "data" in message:
            raise tornado.web.HTTPError(
                status_code=HTTPStatus.BAD_REQUEST,
                reason='Expected key "data" in feedback request body',
            )
        ty = _get_request_ty(message)
        if not (
            ty == SeldonPayload.TENSOR
            or ty == SeldonPayload.NDARRAY
            or ty == SeldonPayload.TFTENSOR
        ):
            raise tornado.web.HTTPError(
                status_code=HTTPStatus.BAD_REQUEST,
                reason='"data" key should contain either "tensor","ndarray", or "tftensor"',
            )

    def validate(self):
        if not "truth" in self.request:
            raise tornado.web.HTTPError(
                status_code=HTTPStatus.BAD_REQUEST,
                reason='Expected key "truth" in request body',
            )
        self._validate_seldon_message(self.request["truth"])
        if "request" in self.request:
            self._validate_seldon_message(self.request["request"])
        if "response" in self.request:
            self._validate_seldon_message(self.request["response"])

    def extract_request(self) -> Union[List, Dict]:
        return _extract_feedback_request(self.request)

