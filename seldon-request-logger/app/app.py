from flask import Flask, request
import sys
import dict_digger
import json
import time
from seldon_core.utils import json_to_seldon_message, extract_request_parts, array_to_grpc_datadef, seldon_message_to_json
from seldon_core.proto import prediction_pb2
import numpy as np
from elasticsearch import Elasticsearch
app = Flask(__name__)

import logging
log = logging.getLogger('werkzeug')
log.setLevel(logging.ERROR)

@app.route("/", methods=['GET','POST'])
def index():
    #try:

    body = request.get_json(force=True)

    print('RAW LOGMESSAGE')
    print(str(request.headers))
    print(str(body))
    print('----------')
    sys.stdout.flush()
    es = connect_elasticsearch()

    # TODO: use env vars for index and doc type
    # TODO: see executor code for proper headers - need model and SeldonDeployment name
    type_header = request.headers.get('Ce-Type')
    content = separate_request_response_sections(es, type_header, body, request.headers.get('Seldon-Puid'))

    # req and response will come through separately and we need enrich the doc with response
    content = extract_content(content)
    print('TRANSFORMED LOGMESSAGE')
    print(str(content))
    print('----------')
    sys.stdout.flush()
    # TODO: that means getting the existing record and joining

    store_record_with_retry(es, 'seldon', content, request.headers.get('Seldon-Puid'), 'seldonrequest')
    #store_record_with_retry(es, 'seldon', content, '7f70cbb5-70d0-42c2-a6b4-561edef3ccba', 'seldonrequest')
    sys.stdout.flush()

    return str(content)
    #except Exception as e:
    #    print(e, file=sys.stderr)
    #    return 'Error processing input'

def separate_request_response_sections(elastic_object, type_header, content, request_id):
    new_content = {}
    if type_header == "io.seldon.serving.inference.request":
        print('SETTING CONTENT FOR REQUEST')
        sys.stdout.flush()
        # put whole dict under 'request' and return it
        new_content['request'] = content
    elif type_header == "io.seldon.serving.inference.response":

        # TODO: waiting for req to be in first... do a wait with retries or allow response to be first?
        time.sleep(1)
        doc = retrieve_doc(elastic_object, 'seldon', 'seldonrequest', request_id)

        # build a new doc that contains the request and response
        new_content['request'] = {}
        # TODO: but the request has already been transformed before insertion so have to untransform it
        # have 'untransform' it because whole thing will go through transformation again in extract_content
        new_content['request']['data'] = doc['_source']['request']['tabular']['data']

        new_content['response'] = content

        # TODO: also not nice that for response we retrieve doc and don't retain the '_seq_no' and '_primary_term' so we do an unnecessary retry on posting to elastic

    else:
        new_content = content

    return new_content

def connect_elasticsearch():
    _es = None
    #TODO: use env vars as host will change
    _es = Elasticsearch([{'host': 'localhost', 'port': 9200}])
    if _es.ping():
        print('Connected to Elasticsearch')
    else:
        print('Could not connect to Elasticsearch')
    return _es

def store_record_with_retry(elastic_object, index_name, record, req_id, record_doc_type):
    try:
        store_record(elastic_object, index_name, record, req_id, record_doc_type)
    except Exception as ex:
        time.sleep(1)
        print('retrying indexing of doc '+req_id)
        store_record(elastic_object, index_name, record, req_id, record_doc_type)

def store_record(elastic_object, index_name, record, req_id, record_doc_type):
    doc = retrieve_doc(elastic_object, index_name, record_doc_type, req_id)

    try:
        # see https://elasticsearch-py.readthedocs.io/en/master/api.html#elasticsearch.Elasticsearch.index
        if doc is None:
            print('doc '+req_id+' does not exist')
            outcome = elastic_object.index(index=index_name, doc_type=record_doc_type, id=req_id, body=record)
        else:
            print('doc '+req_id+' exists')
            print(str(doc))

            outcome = elastic_object.index(index=index_name, doc_type=record_doc_type, id=req_id, body=record, if_seq_no=doc['_seq_no'], if_primary_term=doc['_primary_term'])
    except Exception as ex:
        print('Error in indexing data')
        print(str(ex))
        raise


def retrieve_doc(elastic_object, index_name, record_doc_type, req_id):
    doc = None
    try:
        # see https://elasticsearch-py.readthedocs.io/en/master/api.html#elasticsearch.Elasticsearch.get
        doc = elastic_object.get(index=index_name, doc_type=record_doc_type, id=req_id)
    except:
        pass
    return doc


def extract_content(content):
    requestPart = dict_digger.dig(content, 'request')
    req_elements = None
    if not requestPart is None:
        requestCopy = requestPart.copy()
        if "date" in requestCopy:
            del requestCopy["date"]
        requestMsg = json_to_seldon_message(requestCopy)
        (req_features, _, req_datadef, req_datatype) = extract_request_parts(requestMsg)
        req_elements = createElelmentsArray(req_features, list(req_datadef.names))
    responsePart = dict_digger.dig(content, 'response')
    res_elements = None
    if not responsePart is None:
        responseCopy = responsePart.copy()
        if "date" in responseCopy:
            del responseCopy["date"]
        responseMsg = json_to_seldon_message(responseCopy)
        (res_features, _, res_datadef, res_datatype) = extract_request_parts(responseMsg)
        res_elements = createElelmentsArray(res_features, list(res_datadef.names))
    if not req_elements is None and not res_elements is None:
        for i, (a, b) in enumerate(zip(req_elements, res_elements)):
            # merged = {**a, **b}
            content["request_elements"] = a
            content["response_elements"] = b
            reqJson = extractRow(i, requestMsg, req_datatype, req_features, req_datadef)
            resJson = extractRow(i, responseMsg, res_datatype, res_features, res_datadef)
            content["request"] = {"dataType": reqJson["dataType"]}
            content["request"][reqJson["dataType"]] = reqJson
            content["response"] = {"dataType": resJson["dataType"]}
            content["response"][resJson["dataType"]] = resJson
            # log formatted json to stdout for fluentd collection
            return json.dumps(content)
    elif not req_elements is None:
        for i, e in enumerate(req_elements):
            content["request_elements"] = e
            reqJson = extractRow(i, requestMsg, req_datatype, req_features, req_datadef)
            content["request"] = {"dataType": reqJson["dataType"]}
            content["request"][reqJson["dataType"]] = reqJson
            return json.dumps(content)
    elif not res_elements is None:
        for i, e in enumerate(res_elements):
            content["response_elements"] = e
            resJson = extractRow(i, responseMsg, res_datatype, res_features, res_datadef)
            content["response"] = {"dataType": resJson["dataType"]}
            content["response"][resJson["dataType"]] = resJson
            return json.dumps(content)
    else:
        if "strData" in requestPart:
            content["request"]["dataType"] = "text"
        return json.dumps(content)


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