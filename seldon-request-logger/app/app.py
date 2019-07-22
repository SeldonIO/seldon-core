from flask import Flask, request
import sys
import dict_digger
import json
from seldon_core.utils import json_to_seldon_message, extract_request_parts
import numpy as np

app = Flask(__name__)

import logging
log = logging.getLogger('werkzeug')
log.setLevel(logging.ERROR)

@app.route("/", methods=['GET','POST'])
def index():
    #try:
    content = request.get_json(force=True)

    requestPart = dict_digger.dig(content,'request')
    req_elements = None
    if not requestPart is None:
        requestCopy = requestPart.copy()
        if "date" in requestCopy:
            del requestCopy["date"]
        requestMsg = json_to_seldon_message(requestCopy)
        (req_features, _, req_datadef, _) = extract_request_parts(requestMsg)
        req_elements = createElelmentsArray(req_features,list(req_datadef.names))

    responsePart = dict_digger.dig(content,'response')
    res_elements = None
    if not responsePart is None:
        responseCopy = responsePart.copy()
        if "date" in responseCopy:
            del responseCopy["date"]
        responseMsg = json_to_seldon_message(responseCopy)
        (res_features, _, res_datadef, _) = extract_request_parts(responseMsg)
        res_elements = createElelmentsArray(res_features,list(res_datadef.names))

    if not req_elements is None and not res_elements is None:
        for (a,b) in zip(req_elements,res_elements):
            merged = {**a, **b}
            content["elements"] = merged
            #log formatted json to stdout for fluentd collection
            print(str(json.dumps(content)))
    elif not req_elements is None:
        for e in req_elements:
            content["elements"] = e
            print(str(json.dumps(content)))
    elif not res_elements is None:
        for e in res_elements:
            content["elements"] = e
            print(str(json.dumps(content)))
    else:
        print(str(json.dumps(content)))

    sys.stdout.flush()

    return str(content)
    #except Exception as e:
    #    print(e, file=sys.stderr)
    #    return 'Error processing input'

def createElelmentsArray(X: np.ndarray,names: list):
    results = []
    if isinstance(X,np.ndarray):
        if len(X.shape) == 1:
            d = {}
            for num, name in enumerate(names, start=0):
                d[name] = X[num]
                results.append(d)
        elif len(X.shape) == 2:
            for i in range(X.shape[0]):
                d = {}
                for num, name in enumerate(names, start=0):
                    d[name] = X[i,num]
                results.append(d)
    return results

if __name__ == "__main__":
    app.run(host='0.0.0.0', port=8080)