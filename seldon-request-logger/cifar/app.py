from flask import Flask, request
import sys
import os
import numpy as np
import json
from elasticsearch import Elasticsearch
import logging
import datetime

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
print('starting cifar logger')
sys.stdout.flush()

log = logging.getLogger('werkzeug')
log.setLevel(logging.ERROR)


@app.route("/", methods=['GET','POST'])
def index():
    #try:

    body = request.get_json(force=True)
    if not type(body) is dict:
        body = json.loads(body)

    print('RECEIVED MESSAGE.')
    print(str(request.headers))
    print(str(body))
    print('----')
    sys.stdout.flush()

    es = connect_elasticsearch()

    type_header = request.headers.get(TYPE_HEADER_NAME)
    message_type = parse_message_type(type_header)
    index_name = build_index_name(request.headers)

    try:

        request_id = request.headers.get(REQUEST_ID_HEADER_NAME)
        if request_id is None:
            # TODO: need to fix this upstream
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
    #FIXME: upstream needs to actually send in this format
    if type_header == "io.seldon.serving.inference.outlier" or type_header == "org.kubeflow.serving.inference.outlier":
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

    if message_type == "request" or not '@timestamp' in content:
        timestamp = headers.get(TIMESTAMP_HEADER_NAME)
        if timestamp is None:
            timestamp = datetime.datetime.now(datetime.timezone.utc).isoformat()
        content['@timestamp'] = timestamp

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
    print('in process_and_update_elastic_doc')
    sys.stdout.flush()

    if message_type == 'unknown':
        print('UNKNOWN REQUEST TYPE FOR '+request_id+' - NOT PROCESSING')
        sys.stdout.flush()

    #first do any needed transformations
    new_content_part = process_content(message_type, message_body)

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
    if message_type == 'outlier':
        print(upsert_body)
        sys.stdout.flush()
    return str(new_content)



def connect_elasticsearch():
    _es = None
    elastic_host = os.getenv('ELASTICSEARCH_HOST', 'localhost')
    elastic_port = os.getenv('ELASTICSEARCH_PORT', 9200)

    _es = Elasticsearch([{'host': elastic_host, 'port': elastic_port}])
    if _es.ping():
        print('Connected to Elasticsearch')
        sys.stdout.flush()
    else:
        print('Could not connect to Elasticsearch')
        sys.stdout.flush()
    return _es


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