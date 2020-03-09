from flask import Flask, request
import sys
from seldon_core.utils import json_to_seldon_message, extract_request_parts, array_to_grpc_datadef, seldon_message_to_json
from seldon_core.proto import prediction_pb2
import numpy as np
import json
import logging
import log_helper

MAX_PAYLOAD_BYTES = 100000
app = Flask(__name__)


log = logging.getLogger('werkzeug')
log.setLevel(logging.ERROR)


@app.route("/", methods=['GET','POST'])
def index():

    request_id = log_helper.extract_request_id(request.headers)
    type_header = request.headers.get(log_helper.TYPE_HEADER_NAME)
    message_type = log_helper.parse_message_type(type_header)
    index_name = log_helper.build_index_name(request.headers)

    # max size is configurable with env var or defaults to constant
    max_payload_bytes = log_helper.get_max_payload_bytes(MAX_PAYLOAD_BYTES)

    body_length = request.headers.get(log_helper.LENGTH_HEADER_NAME)
    if body_length and int(body_length) > int(max_payload_bytes):
        too_large_message = 'body too large for '+index_name+"/"+log_helper.DOC_TYPE_NAME+"/"+ request_id+ ' adding '+message_type
        print()
        sys.stdout.flush()
        return too_large_message

    body = request.get_json(force=True)
    if not type(body) is dict:
        body = json.loads(body)

    # print('RECEIVED MESSAGE.')
    # print(str(request.headers))
    # print(str(body))
    # print('----')
    # sys.stdout.flush()

    es = log_helper.connect_elasticsearch()

    try:

        #now process and update the doc
        doc = process_and_update_elastic_doc(es, message_type, body, request_id,request.headers, index_name)
        return str(doc)
    except Exception as ex:
        print(ex)
    sys.stdout.flush()
    return 'problem logging request'


def process_and_update_elastic_doc(elastic_object, message_type, message_body, request_id, headers, index_name):

    if message_type == 'unknown':
        print('UNKNOWN REQUEST TYPE FOR '+request_id+' - NOT PROCESSING')
        sys.stdout.flush()

    #first do any needed transformations
    new_content_part = process_content(message_body)
    #set metadata specific to this part (request or response)
    log_helper.field_from_header(content=new_content_part,header_name='ce-time',headers=headers)
    log_helper.field_from_header(content=new_content_part, header_name='ce-source', headers=headers)

    upsert_body= {
        "doc_as_upsert": True,
        "doc": {
            message_type: new_content_part
        }
    }

    log_helper.set_metadata(upsert_body['doc'],headers,message_type,request_id)

    new_content = elastic_object.update(index=index_name,doc_type=log_helper.DOC_TYPE_NAME,id=request_id,body=upsert_body,retry_on_conflict=3,refresh=True,timeout="60s")
    print('upserted to doc '+index_name+"/"+log_helper.DOC_TYPE_NAME+"/"+ request_id+ ' adding '+message_type)
    sys.stdout.flush()
    return str(new_content)



# take request or response part and process it by deriving metadata
def process_content(content):

    if content is None:
        return content

    #no transformation using strData
    if "strData" in content:
        content["dataType"] = "text"
        return content

    #if we have dataType then have already parsed before
    if "dataType" in content:
        print('dataType already in content')
        sys.stdout.flush()
        return content

    requestCopy = content.copy()
    if "date" in requestCopy:
        del requestCopy["date"]
    requestMsg = json_to_seldon_message(requestCopy)
    (req_features, _, req_datadef, req_datatype) = extract_request_parts(requestMsg)
    elements = createElelmentsArray(req_features, list(req_datadef.names))
    for i, e in enumerate(elements):
        reqJson = extractRow(i, requestMsg, req_datatype, req_features, req_datadef)
        reqJson["elements"] = e
        content = reqJson

    return content

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
    if len(req_datadef.names) > 0:
        dataType="tabular"
    requestMsg2 = prediction_pb2.SeldonMessage(data=dataReq, meta=requestMsg.meta)
    reqJson = {}
    reqJson["payload"] = seldon_message_to_json(requestMsg2)
    # setting dataType here temporarily so calling method will be able to access it
    # don't want to set it at the payload level
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