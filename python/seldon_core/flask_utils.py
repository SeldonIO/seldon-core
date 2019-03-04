from flask import request
import json
from typing import Dict


def get_request() -> Dict:
    """
    Parse a request to get JSON dict

    Returns
    -------
       JSON Dict

    """
    j_str = request.form.get("json")
    if j_str:
        message = json.loads(j_str)
    else:
        j_str = request.args.get('json')
        if j_str:
            message = json.loads(j_str)
        else:
            raise SeldonMicroserviceException("Empty json parameter in data")
    if message is None:
        raise SeldonMicroserviceException("Invalid Data Format")
    return message


class SeldonMicroserviceException(Exception):
    status_code = 400

    def __init__(self, message, status_code=None, payload=None, reason="MICROSERVICE_BAD_DATA"):
        Exception.__init__(self)
        self.message = message
        if status_code is not None:
            self.status_code = status_code
        self.payload = payload
        self.reason = reason

    def to_dict(self):
        rv = {"status": {"status": 1, "info": self.message,
                         "code": -1, "reason": self.reason}}
        return rv


ANNOTATIONS_FILE = "/etc/podinfo/annotations"
ANNOTATION_GRPC_MAX_MSG_SIZE = 'seldon.io/grpc-max-message-size'
