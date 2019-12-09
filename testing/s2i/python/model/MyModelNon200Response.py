from flask.json import jsonify


class MyModelNon200Response(object):
    def __init__(self, metrics_ok=True):
        print("Init called")

    def predict_raw(self, message):
        status = {
            "code": 400,
            "reason": "exception message",
            "status": "FAILURE",
            "info": "exception caught",
        }

        return {"data": {"names": ["score"], "ndarray": []}, "status": status}
