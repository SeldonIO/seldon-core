from flask import Flask, request
import sys
import os
from seldon_core.utils import json_to_seldon_message, extract_request_parts, array_to_grpc_datadef, seldon_message_to_json
from seldon_core.proto import prediction_pb2
import numpy as np
from elasticsearch import Elasticsearch
import logging


TYPE_HEADER_NAME = "Ce-Type"
REQUEST_ID_HEADER_NAME = "Ce-Requestid"
CLOUD_EVENT_ID = "Ce-id"
MODELID_HEADER_NAME = 'Ce-Modelid'
NAMESPACE_HEADER_NAME = 'Ce-Namespace'
PREDICTOR_HEADER_NAME = 'Ce-Predictor'
TIMESTAMP_HEADER_NAME = 'CE-Time'
INFERENCESERVICE_HEADER_NAME = 'Ce-Inferenceservicename'
DOC_TYPE_NAME = 'inferencerequest'

app = Flask(__name__)


log = logging.getLogger('werkzeug')
log.setLevel(logging.ERROR)


@app.route("/", methods=['GET','POST'])
def index():
    #try:

    body = request.get_json(force=True)
    # print('RECEIVED MESSAGE.')
    # print(str(request.headers))
    # print(str(body))
    # print('----')
    # sys.stdout.flush()
    #TODO: limit size of body with env var (100KB)
    # this way main logger will by default ignore larger messages as they probably require custom logger

    es = connect_elasticsearch()

    type_header = request.headers.get(TYPE_HEADER_NAME)
    message_type = parse_message_type(type_header)
    index_name = build_index_name(request.headers)

    try:

        request_id = request.headers.get(REQUEST_ID_HEADER_NAME)

        # TODO: need to fix this upstream
        if request_id is None:
            request_id = request.headers.get(CLOUD_EVENT_ID)

        #now process and update the doc
        doc = process_and_update_elastic_doc(es, message_type, body, request_id,request.headers, index_name)
        return str(doc)
    except Exception as ex:
        print(ex)
    sys.stdout.flush()
    return 'problem logging request'


def build_index_name(headers):
    #use a fixed index name if user chooses to do so
    index_name = os.getenv('INDEX_NAME')
    if index_name:
        return index_name

    #otherwise create an index per deployment
    index_name = "inference-log-" + serving_engine(headers)
    namespace = request.headers.get(NAMESPACE_HEADER_NAME)
    if namespace is None:
        index_name = index_name + "-unknown-namespace"
    else:
        index_name = index_name + "-" + namespace
    inference_service_name = request.headers.get(INFERENCESERVICE_HEADER_NAME)
    if inference_service_name is None:
        index_name = index_name + "-unknown-inferenceservice"
    else:
        index_name = index_name + "-" + inference_service_name
    return index_name


def parse_message_type(type_header):
    if type_header == "io.seldon.serving.inference.request" or type_header == "org.kubeflow.serving.inference.request":
        return 'request'
    if type_header == "io.seldon.serving.inference.response" or type_header == "org.kubeflow.serving.inference.response":
        return 'response'
    if type_header == 'seldon.outlier':
        return 'outlier'
    return 'unknown'


def set_metadata(content, headers, message_type, request_id):
    serving_engine_name = serving_engine(headers)
    content['ServingEngine'] = serving_engine_name

    # TODO: provide a way for custom headers to be passed on too?
    field_from_header(content, INFERENCESERVICE_HEADER_NAME, headers)
    field_from_header(content, PREDICTOR_HEADER_NAME, headers)
    field_from_header(content, NAMESPACE_HEADER_NAME, headers)
    field_from_header(content, MODELID_HEADER_NAME, headers)

    if message_type == "request":
       content['@timestamp'] = headers.get(TIMESTAMP_HEADER_NAME)

    content['RequestId'] = request_id
    return


def serving_engine(headers):
    type_header = request.headers.get(TYPE_HEADER_NAME)
    if type_header.startswith('io.seldon.serving') or type_header.startswith('seldon'):
        return 'seldon'
    elif type_header.startswith('org.kubeflow.serving'):
        return 'inferenceservice'


def field_from_header(content, header_name, headers):
    if not request.headers.get(header_name) is None:
        content[header_name] = headers.get(header_name)


def process_and_update_elastic_doc(elastic_object, message_type, message_body, request_id, headers, index_name):

    if message_type == 'unknown':
        print('UNKNOWN REQUEST TYPE FOR '+request_id+' - NOT PROCESSING')
        sys.stdout.flush()

    #first do any needed transformations
    new_content_part = process_content(message_body)
    #set metadata specific to this part (request or response)
    field_from_header(content=new_content_part,header_name='ce-time',headers=headers)
    field_from_header(content=new_content_part, header_name='ce-source', headers=headers)

    upsert_body= {
        "doc_as_upsert": True,
        "doc": {
            message_type: new_content_part
        }
    }

    set_metadata(upsert_body['doc'],headers,message_type,request_id)

    new_content = elastic_object.update(index=index_name,doc_type=DOC_TYPE_NAME,id=request_id,body=upsert_body,retry_on_conflict=3,refresh=True,timeout="60s")
    print('upserted to doc '+index_name+"/"+DOC_TYPE_NAME+"/"+ request_id+ ' adding '+message_type)
    sys.stdout.flush()
    return str(new_content)



def connect_elasticsearch():
    _es = None
    elastic_host = os.getenv('ELASTICSEARCH_HOST', 'localhost')
    elastic_port = os.getenv('ELASTICSEARCH_PORT', 9200)

    _es = Elasticsearch([{'host': elastic_host, 'port': elastic_port}])
    if not _es.ping():
        print('Could not connect to Elasticsearch')
    return _es


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