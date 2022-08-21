import base64
import json
from http import HTTPStatus
from typing import Dict, Union

from flask import current_app
from flask import jsonify as flask_jsonify
from flask import request
from werkzeug.exceptions import BadRequest


def get_multi_form_data_request() -> Dict:
    """
    Parses a request submitted with Content-type:multipart/form-data
    all the keys under SeldonMessage are accepted as form input
    binData can only be passed as file input
    strData can be passed as file or text input
    the file input is base64 encoded

    Returns
    -------
       JSON Dict

    """
    req_dict = {}
    for key in request.form:
        if key == "strData":
            req_dict[key] = request.form.get(key)
        else:
            req_dict[key] = json.loads(request.form.get(key))
    for fileKey in request.files:
        """
        The bytes data needs to be base64 encode because the protobuf trys to do base64 decode for bytes
        """
        if fileKey == "binData":
            req_dict[fileKey] = base64.b64encode(request.files[fileKey].read())
        else:
            """
            This is the case when strData can be passed as file as well
            """
            req_dict[fileKey] = request.files[fileKey].read().decode("utf-8")
    return req_dict


def get_request(skip_decoding=False) -> Union[Dict, bytes]:
    """
    Parse a request to get JSON dict

    Returns
    -------
       JSON Dict

    """

    if (
        request.content_type is not None
        and "multipart/form-data" in request.content_type
    ):
        return get_multi_form_data_request()

    j_str = request.form.get("json")
    if not j_str:
        j_str = request.args.get("json")

    if j_str:
        if skip_decoding:
            return j_str

        return json.loads(j_str)

    if skip_decoding:
        try:
            data = request.get_data()
        except BadRequest:
            raise SeldonMicroserviceException("Can't find data")
        if data is None:
            raise SeldonMicroserviceException("Can't find data")

        return data

    try:
        message = request.get_json()
    except BadRequest:
        raise SeldonMicroserviceException("Can't find JSON in data")
    if message is None:
        raise SeldonMicroserviceException("Can't find JSON in data")

    return message


def jsonify(response, skip_encoding=False):
    if skip_encoding:
        return current_app.response_class(
            response=response, status=HTTPStatus.OK, mimetype="application/json"
        )

    return flask_jsonify(response)


class SeldonMicroserviceException(Exception):
    status_code = 400

    def __init__(
        self, message, status_code=None, payload=None, reason="MICROSERVICE_BAD_DATA"
    ):
        Exception.__init__(self)
        self.message = message
        if status_code is not None:
            self.status_code = status_code
        self.payload = payload
        self.reason = reason

    def to_dict(self):
        rv = {
            "status": {
                "status": 1,
                "info": self.message,
                "code": -1,
                "reason": self.reason,
            }
        }
        return rv


ANNOTATIONS_FILE = "/etc/podinfo/annotations"
ANNOTATION_GRPC_MAX_MSG_SIZE = "seldon.io/grpc-max-message-size"
DEFAULT_ANNOTATION_REST_TIMEOUT = 5000
ANNOTATION_REST_TIMEOUT = "seldon.io/rest-timeout"
