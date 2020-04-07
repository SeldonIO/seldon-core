from http import HTTPStatus
from typing import Dict, List

import tornado
from adserver.protocols.request_handler import (
    RequestHandler,
)  # pylint: disable=no-name-in-module


class TensorflowRequestHandler(RequestHandler):
    def __init__(self, request: Dict):  # pylint: disable=useless-super-delegation
        super().__init__(request)

    def validate(self):
        if "instances" not in self.request:
            raise tornado.web.HTTPError(
                status_code=HTTPStatus.BAD_REQUEST,
                reason='Expected key "instances" in request body',
            )

    def extract_request(self) -> List:
        return self.request["instances"]
