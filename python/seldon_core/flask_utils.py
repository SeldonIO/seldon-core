from flask import request
import json
from seldon_core.microservice import SeldonMicroserviceException

def get_request():
    """ Parse a REST request into a SeldonMessage proto buffer
    """
    jStr = request.form.get("json")
    if jStr:
        message = json.loads(jStr)
    else:
        jStr = request.args.get('json')
        if jStr:
            message = json.loads(jStr)
        else:
            raise SeldonMicroserviceException("Empty json parameter in data")
    if message is None:
        raise SeldonMicroserviceException("Invalid Data Format")
    return message