from flask import Flask, request
import sys
import dict_digger
import json
from seldon_core.utils import json_to_seldon_message, extract_request_parts, array_to_grpc_datadef, seldon_message_to_json
from seldon_core.proto import prediction_pb2
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
        (req_features, _, req_datadef, req_datatype) = extract_request_parts(requestMsg)
        req_elements = createElelmentsArray(req_features,list(req_datadef.names))

    responsePart = dict_digger.dig(content,'response')
    res_elements = None
    if not responsePart is None:
        responseCopy = responsePart.copy()
        if "date" in responseCopy:
            del responseCopy["date"]
        responseMsg = json_to_seldon_message(responseCopy)
        (res_features, _, res_datadef, res_datatype) = extract_request_parts(responseMsg)
        res_elements = createElelmentsArray(res_features,list(res_datadef.names))
    
    if not req_elements is None and not res_elements is None:
        for i,(a,b) in enumerate(zip(req_elements,res_elements)):
            # merged = {**a, **b}
            content["request_elements"] = a
            content["response_elements"] = b            
            reqJson = extractRow(i, requestMsg, req_datatype, req_features, req_datadef)
            resJson = extractRow(i, responseMsg, res_datatype, res_features, res_datadef)
            content["request"] = {"dataType":reqJson["dataType"]}
            content["request"][reqJson["dataType"]] = reqJson
            content["response"] = {"dataType":resJson["dataType"]}
            content["response"][resJson["dataType"]] = resJson
            #log formatted json to stdout for fluentd collection
            print(str(json.dumps(content)))
    elif not req_elements is None:
        for i,e in enumerate(req_elements):
            content["request_elements"] = e
            reqJson = extractRow(i, requestMsg, req_datatype, req_features, req_datadef)
            content["request"] = {"dataType":reqJson["dataType"]}
            content["request"][reqJson["dataType"]] = reqJson
            print(str(json.dumps(content)))
    elif not res_elements is None:
        for i,e in enumerate(res_elements):
            content["response_elements"] = e
            resJson = extractRow(i, responseMsg, res_datatype, res_features, res_datadef)
            content["response"] = {"dataType":resJson["dataType"]}
            content["response"][resJson["dataType"]] = resJson
            print(str(json.dumps(content)))
    else:
        if "strData" in requestPart:
            content["request"]["dataType"] = "text"
        print(str(json.dumps(content)))

    sys.stdout.flush()

    return str(content)
    #except Exception as e:
    #    print(e, file=sys.stderr)
    #    return 'Error processing input'


def extractRow(i:int,requestMsg: prediction_pb2.SeldonMessage,req_datatype: str,req_features: np.ndarray,req_datadef: "prediction_pb2.SeldonMessage.data"):
    datatyReq = "ndarray"
    dataType = "tabular"
    if len(req_features.shape) == 2:
        dataReq = array_to_grpc_datadef(datatyReq, np.expand_dims(req_features[i], axis=0), req_datadef.names)
    else:
        if len(req_features.shape) > 2:
            dataType="image"
        else:
            dataType="text"
            req_features= np.char.decode(req_features.astype('S'),"utf-8")
        dataReq = array_to_grpc_datadef(datatyReq, np.expand_dims(req_features[i], axis=0), req_datadef.names)  
    requestMsg2 = prediction_pb2.SeldonMessage(data=dataReq, meta=requestMsg.meta)
    reqJson = seldon_message_to_json(requestMsg2)
    reqJson["dataType"] = dataType
    return reqJson


def createElelmentsArray(X: np.ndarray,names: list):
    results = None
    if isinstance(X,np.ndarray):
        if len(X.shape) == 1:
            results = []
            for i in range(X.shape[0]):
                d = {}
                for num, name in enumerate(names, start=0):
                    if isinstance(X[i],bytes):
                        d[name] = X[i].decode("utf-8")
                    else:
                        d[name] = X[i]    
                results.append(d)
        elif len(X.shape) >= 2:
            results = []
            for i in range(X.shape[0]):
                d = {}
                for num, name in enumerate(names, start=0):
                    d[name] = np.expand_dims(X[i,num], axis=0).tolist()
                results.append(d)      
    return results

if __name__ == "__main__":
    app.run(host='0.0.0.0', port=8080)