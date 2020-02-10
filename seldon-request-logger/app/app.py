from flask import Flask, request
import sys
import dict_digger
import backoff
import requests
import json
import time
import os
from seldon_core.utils import json_to_seldon_message, extract_request_parts, array_to_grpc_datadef, seldon_message_to_json
from seldon_core.proto import prediction_pb2
import numpy as np
from elasticsearch import Elasticsearch
import logging


TYPE_HEADER_NAME = "Ce-Type"
REQUEST_ID_HEADER_NAME = "Ce-Requestid"
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

    es = connect_elasticsearch()

    type_header = request.headers.get(TYPE_HEADER_NAME)
    message_type = parse_message_type(type_header)
    index_name = build_index_name(request.headers)
    print('index is '+index_name)
    sys.stdout.flush()

    try:
        #first ensure there is an elastic doc as we need something to lock against
        #use req id as doc id (if None then elastic should generate one but then req & res won't be linked)
        request_id = request.headers.get(REQUEST_ID_HEADER_NAME)
        update_elastic_doc(es,message_type,{}, request_id, request.headers, index_name)
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
    if type_header == "io.seldon.serving.inference.request":
        return 'request'
    if type_header == "io.seldon.serving.inference.response":
        return 'response'
    return 'unknown'


def set_metadata(content, headers, message_type):
    serving_engine_name = serving_engine(headers)
    content['ServingEngine'] = serving_engine_name

    # TODO: provide a way for custom headers to be passed on too?
    field_from_header(content, INFERENCESERVICE_HEADER_NAME, headers)
    field_from_header(content, PREDICTOR_HEADER_NAME, headers)
    field_from_header(content, NAMESPACE_HEADER_NAME, headers)
    field_from_header(content, MODELID_HEADER_NAME, headers)

    if message_type == "request":
       content['@timestamp'] = headers.get(TIMESTAMP_HEADER_NAME)
       content['RequestId'] = headers.get(REQUEST_ID_HEADER_NAME)
    return


def serving_engine(headers):
    type_header = request.headers.get(TYPE_HEADER_NAME)
    if type_header.startswith('io.seldon.serving'):
        return 'seldon'
    elif type_header.startswith('org.kubeflow.serving'):
        return 'inferenceservice'


def field_from_header(content, header_name, headers):
    if not request.headers.get(header_name) is None:
        content[header_name] = headers.get(header_name)


def process_and_update_elastic_doc(elastic_object, message_type, message_body, request_id, headers, index_name):
    if message_type == 'unknown':
        print('UNKNOWN REQUEST TYPE FOR '+request_id+' - NOT PROCESSING')

    #first do any needed transformations
    new_content_part = process_content(message_body)
    #set metadata specific to this part (request or response)
    field_from_header(content=new_content_part,header_name='ce-time',headers=headers)
    field_from_header(content=new_content_part, header_name='ce-source', headers=headers)

    new_content = update_elastic_doc(elastic_object, message_type, new_content_part, request_id, headers, index_name)
    return str(new_content)


@backoff.on_exception(backoff.fibo,
                      Exception,
                      max_time=60,
                      jitter=backoff.random_jitter)
def update_elastic_doc(elastic_object, message_type, new_content_part, request_id, headers, index_name):
    # now ready to upsert
    #TODO: might put inferenceservices under a different doc type and not 'seldonrequest' (use env vars?)
    doc = retrieve_doc(elastic_object, index_name, DOC_TYPE_NAME, request_id)
    # req and response will come through separately and we need enrich the doc with response
    # doc can have existing content - should have (processed) request content already
    # JITTERED BACKOFFS ARE NEEDED TO HANDLE CONCURRENT UPDATES
    new_content = {}
    seq_no = None
    primary_term = None
    if not doc is None:
        new_content = doc['_source']
        # need seq_no for elastic optimistic locking
        seq_no = doc['_seq_no']
        primary_term = doc['_primary_term']

    # add the new content under its key
    new_content[message_type] = new_content_part
    # ensure any top-level metadata is set
    set_metadata(new_content,headers, message_type)

    store_record(elastic_object, index_name, new_content, request_id, DOC_TYPE_NAME, seq_no, primary_term)
    return new_content


def connect_elasticsearch():
    _es = None
    elastic_host = os.getenv('ELASTICSEARCH_HOST', 'localhost')
    elastic_port = os.getenv('ELASTICSEARCH_PORT', 9200)

    _es = Elasticsearch([{'host': elastic_host, 'port': elastic_port}])
    if _es.ping():
        print('Connected to Elasticsearch')
    else:
        print('Could not connect to Elasticsearch')
    return _es


def store_record(elastic_object, index_name, record, req_id, record_doc_type, seq_no, primary_term):
    doc = None

    # don't already have a seq_no for optimistic lock so get one
    if seq_no is None or primary_term is None:
        doc = retrieve_doc(elastic_object, index_name, record_doc_type, req_id)
        if not doc is None:
            seq_no = doc['_seq_no']
            primary_term = doc['_primary_term']

    try:
        # see https://elasticsearch-py.readthedocs.io/en/master/api.html#elasticsearch.Elasticsearch.index
        if doc is None and req_id is not None:
            print('doc '+index_name+'/'+record_doc_type+'/'+req_id+' does not exist')
        elif req_id is not None:
            print('doc '+index_name+'/'+record_doc_type+'/'+req_id+' exists')
            print(str(doc))

        sys.stdout.flush()
        outcome = elastic_object.index(index=index_name, doc_type=record_doc_type, id=req_id, body=record, if_seq_no=seq_no, if_primary_term=primary_term, refresh='wait_for',request_timeout=60.0)

        if outcome is not None:
            print('doc '+outcome['_index']+'/'+outcome['_type']+'/'+outcome['_id']+' indexed')
            sys.stdout.flush()

    except Exception as ex:
        print('Error in indexing data')
        print(str(ex))
        sys.stdout.flush()
        raise


def retrieve_doc(elastic_object, index_name, record_doc_type, req_id):
    doc = None
    try:
        # see https://elasticsearch-py.readthedocs.io/en/master/api.html#elasticsearch.Elasticsearch.get
        doc = elastic_object.get(index=index_name, doc_type=record_doc_type, id=req_id)
    except:
        pass
    return doc

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