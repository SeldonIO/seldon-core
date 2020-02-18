from flask import Flask, request
import sys
import dict_digger
import backoff
import requests
import json
import time
import random
import os
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

    print('RECEIVED MESSAGE.')
    print(str(request.headers))
    #print(str(body))
    print('----')
    sys.stdout.flush()

    es = connect_elasticsearch()

    type_header = request.headers.get(TYPE_HEADER_NAME)
    message_type = parse_message_type(type_header)
    index_name = build_index_name(request.headers)
    print('index is '+index_name)
    sys.stdout.flush()

    try:

        request_id = request.headers.get(REQUEST_ID_HEADER_NAME)
        if request_id is None:
            # TODO: need to fix this upstream
            request_id = request.headers.get(CLOUD_EVENT_ID)

        if message_type != 'request':
            #can reduce overall number of elastic queries if we wait for req first
            #locking involves contention and retries so want to spread out to increase success %
            time.sleep(random.uniform(3.0,6.0))

        print('type is '+message_type)
        sys.stdout.flush()

        #first ensure there is an elastic doc as we'll need something to lock against
        #use req id as doc id (if None then elastic should generate one but then req & res won't be linked)
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
    if type_header == "io.seldon.serving.inference.request" or type_header == "org.kubeflow.serving.inference.request":
        return 'request'
    if type_header == "io.seldon.serving.inference.response" or type_header == "org.kubeflow.serving.inference.response":
        return 'response'
    if type_header == 'seldon.outlier':
        return 'outlier'
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
    new_content_part = process_content(message_type, message_body)

    #set metadata specific to this part (request or response)
    field_from_header(content=new_content_part,header_name='ce-time',headers=headers)
    field_from_header(content=new_content_part, header_name='ce-source', headers=headers)

    new_content = update_elastic_doc(elastic_object, message_type, new_content_part, request_id, headers, index_name)
    return str(new_content)


@backoff.on_exception(backoff.expo,
                      Exception,
                      max_time=240,
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
        print('seq_no')
        print(seq_no)
        print('primary_term')
        print(primary_term)

    # add the new content under its key
    new_content[message_type] = new_content_part

    if message_type == 'outlier':
        print('outlier content')
        print(new_content)
        sys.stdout.flush()

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

    #FIXME: first time through this loop for any existing doc we will have seq_no and req_id ... but code won't look it up!
    # it won't error either - it will try to overwrite it!

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
def process_content(message_type,content):

    if content is None:
        print('content is empty')
        sys.stdout.flush()
        return content

    #if we have dataType then have already parsed before
    if "dataType" in content:
        print('dataType already in content')
        sys.stdout.flush()
        return content

    requestCopy = content.copy()

    print('in process_content for '+message_type)
    sys.stdout.flush()

    if message_type == 'request':
        # we know this is a cifar10 image
        content["dataType"] = "image"
        requestCopy["image"] = decode(content)
        if "instances" in requestCopy:
            del requestCopy["instances"]

    return requestCopy


def decode(X):
    X=np.array(X["instances"])
    X=np.transpose(X, (0,2, 3, 1))
    img = X/2.0 + 0.5
    return img.tolist()

if __name__ == "__main__":
    app.run(host='0.0.0.0', port=8080)