import os
import datetime
import sys
from elasticsearch import Elasticsearch, RequestsHttpConnection

TYPE_HEADER_NAME = "Ce-Type"
REQUEST_ID_HEADER_NAME = "Ce-Requestid"
CLOUD_EVENT_ID = "Ce-id"
MODELID_HEADER_NAME = 'Ce-Modelid'
NAMESPACE_HEADER_NAME = 'Ce-Namespace'
ENDPOINT_HEADER_NAME = 'Ce-Endpoint'
TIMESTAMP_HEADER_NAME = 'CE-Time'
INFERENCESERVICE_HEADER_NAME = 'Ce-Inferenceservicename'
LENGTH_HEADER_NAME = 'Content-Length'
DOC_TYPE_NAME = 'inferencerequest'

def get_max_payload_bytes(default_value):
    max_payload_bytes = os.getenv('MAX_PAYLOAD_BYTES')
    if not max_payload_bytes:
        max_payload_bytes = default_value
    return max_payload_bytes

def extract_request_id(headers):
    request_id = headers.get(REQUEST_ID_HEADER_NAME)
    if request_id is None:
        # TODO: need to fix this upstream
        request_id = headers.get(CLOUD_EVENT_ID)
    return request_id

def build_index_name(headers):
    #use a fixed index name if user chooses to do so
    index_name = os.getenv('INDEX_NAME')
    if index_name:
        return index_name

    #otherwise create an index per deployment
    index_name = "inference-log-" + serving_engine(headers)
    namespace = clean_header(NAMESPACE_HEADER_NAME, headers)
    if namespace is None:
        index_name = index_name + "-unknown-namespace"
    else:
        index_name = index_name + "-" + namespace
    inference_service_name = clean_header(INFERENCESERVICE_HEADER_NAME, headers)
    #won't get inference service name for kfserving versions prior to https://github.com/kubeflow/kfserving/pull/699/
    if inference_service_name is None or not inference_service_name:
        inference_service_name = clean_header(MODELID_HEADER_NAME, headers)

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
    field_from_header(content, ENDPOINT_HEADER_NAME, headers)
    field_from_header(content, NAMESPACE_HEADER_NAME, headers)
    field_from_header(content, MODELID_HEADER_NAME, headers)

    inference_service_name = content.get(INFERENCESERVICE_HEADER_NAME)
    #kfserving won't set inferenceservice header
    if inference_service_name is None or not inference_service_name:
        content[INFERENCESERVICE_HEADER_NAME] = clean_header(MODELID_HEADER_NAME, headers)

    if message_type == "request" or not '@timestamp' in content:
        timestamp = headers.get(TIMESTAMP_HEADER_NAME)
        if timestamp is None:
            timestamp = datetime.datetime.now(datetime.timezone.utc).isoformat()
        content['@timestamp'] = timestamp

    content['RequestId'] = request_id
    return


def serving_engine(headers):
    type_header = clean_header(TYPE_HEADER_NAME,headers)
    if type_header.startswith('io.seldon.serving') or type_header.startswith('seldon'):
        return 'seldon'
    elif type_header.startswith('org.kubeflow.serving'):
        return 'inferenceservice'


def field_from_header(content, header_name, headers):
    if not headers.get(header_name) is None:
        content[header_name] = clean_header(header_name, headers)


def clean_header(header_name, headers):
    header_val = headers.get(header_name)
    if not header_val is None:
        header_val = header_val.translate({ord(c): None for c in '!@#$"<>/?'})
    return header_val


def connect_elasticsearch():
    _es = None
    elastic_host = os.getenv('ELASTICSEARCH_HOST', 'localhost')
    elastic_port = os.getenv('ELASTICSEARCH_PORT', 9200)
    elastic_protocol = os.getenv('ELASTICSEARCH_PROTOCOL', 'http')
    elastic_user = os.getenv('ELASTICSEARCH_USER')
    elastic_pass = os.getenv('ELASTICSEARCH_PASS')

    connection_string = elastic_protocol +'://'

    if elastic_user and elastic_pass:
        connection_string = connection_string + elastic_user + ':' + elastic_pass + '@'

    connection_string = connection_string + elastic_host + ':' + str(elastic_port)

    _es = Elasticsearch(connection_string,verify_certs=False,connection_class=RequestsHttpConnection)
    if _es.ping():
        print('Connected to Elasticsearch')
        sys.stdout.flush()
    else:
        print('Could not connect to Elasticsearch')
        sys.stdout.flush()
    return _es


