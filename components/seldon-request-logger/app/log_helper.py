import os
import datetime
import sys
import json
import numpy as np
from elasticsearch import Elasticsearch, RequestsHttpConnection

TYPE_HEADER_NAME = "Ce-Type"
REQUEST_ID_HEADER_NAME = "Ce-Requestid"
CLOUD_EVENT_ID = "Ce-id"
# in seldon case modelid is node in graph as graph can have multiple models
MODELID_HEADER_NAME = "Ce-Modelid"
NAMESPACE_HEADER_NAME = "Ce-Namespace"
# endpoint distinguishes default, canary, shadow, A/B etc.
ENDPOINT_HEADER_NAME = "Ce-Endpoint"
TIMESTAMP_HEADER_NAME = "CE-Time"
# inferenceservicename is k8s resource name for SeldonDeployment or InferenceService
INFERENCESERVICE_HEADER_NAME = "Ce-Inferenceservicename"
LENGTH_HEADER_NAME = "Content-Length"
DOC_TYPE_NAME = None


def get_max_payload_bytes(default_value):
    max_payload_bytes = os.getenv("MAX_PAYLOAD_BYTES")
    if not max_payload_bytes:
        max_payload_bytes = default_value
    return max_payload_bytes


def extract_request_id(headers):
    request_id = headers.get(REQUEST_ID_HEADER_NAME)
    if not request_id:
        # TODO: need to fix this upstream - https://github.com/kubeflow/kfserving/pull/699/files#diff-de6e9737c409666fc6c48dbcb50363faR18
        request_id = headers.get(CLOUD_EVENT_ID)
    return request_id


def build_index_name(headers):
    # use a fixed index name if user chooses to do so
    index_name = os.getenv("INDEX_NAME")
    if index_name:
        return index_name
    
    # Adding seldon_environment (dev/test/staging/prod) to index_name if defined as a environment variable
    seldon_environment = os.getenv("SELDON_ENVIRONMENT")
    if seldon_environment:
       index_name = "inference-log-" + seldon_environment + "-" + serving_engine(headers)
    else:
       index_name = "inference-log-" + serving_engine(headers)
    
    # otherwise create an index per deployment
    # index_name = "inference-log-" + serving_engine(headers)
    namespace = clean_header(NAMESPACE_HEADER_NAME, headers)
    if not namespace:
        index_name = index_name + "-unknown-namespace"
    else:
        index_name = index_name + "-" + namespace
    inference_service_name = clean_header(INFERENCESERVICE_HEADER_NAME, headers)
    # won't get inference service name for older kfserving versions i.e. prior to https://github.com/kubeflow/kfserving/pull/699/
    if not inference_service_name:
        inference_service_name = clean_header(MODELID_HEADER_NAME, headers)

    if not inference_service_name:
        index_name = index_name + "-unknown-inferenceservice"
    else:
        index_name = index_name + "-" + inference_service_name

    endpoint_name = clean_header(ENDPOINT_HEADER_NAME, headers)
    if not endpoint_name:
        index_name = index_name + "-unknown-endpoint"
    else:
        index_name = index_name + "-" + endpoint_name

    return index_name


def parse_message_type(type_header):
    if (
        type_header == "io.seldon.serving.inference.request"
        or type_header == "org.kubeflow.serving.inference.request"
    ):
        return "request"
    if (
        type_header == "io.seldon.serving.inference.response"
        or type_header == "org.kubeflow.serving.inference.response"
    ):
        return "response"
    if (
        type_header == "io.seldon.serving.feedback"
        or type_header == "org.kubeflow.serving.feedback"
    ):
        return "feedback"
    # FIXME: upstream needs to actually send in this format
    if (
        type_header == "io.seldon.serving.inference.outlier"
        or type_header == "org.kubeflow.serving.inference.outlier"
    ):
        return "outlier"
    return "unknown"


def set_metadata(content, headers, message_type, request_id):
    serving_engine_name = serving_engine(headers)
    content["ServingEngine"] = serving_engine_name

    # TODO: provide a way for custom headers to be passed on too?
    field_from_header(content, INFERENCESERVICE_HEADER_NAME, headers)
    field_from_header(content, ENDPOINT_HEADER_NAME, headers)
    field_from_header(content, NAMESPACE_HEADER_NAME, headers)
    field_from_header(content, MODELID_HEADER_NAME, headers)

    inference_service_name = content.get(INFERENCESERVICE_HEADER_NAME)
    # kfserving won't set inferenceservice header
    if not inference_service_name:
        content[INFERENCESERVICE_HEADER_NAME] = clean_header(
            MODELID_HEADER_NAME, headers
        )

    if message_type == "request" or not "@timestamp" in content:
        timestamp = headers.get(TIMESTAMP_HEADER_NAME)
        if not timestamp:
            timestamp = datetime.datetime.now(datetime.timezone.utc).isoformat()
        content["@timestamp"] = timestamp

    content["RequestId"] = request_id
    return


def serving_engine(headers):
    type_header = clean_header(TYPE_HEADER_NAME, headers)
    if type_header.startswith("io.seldon.serving") or type_header.startswith("seldon"):
        return "seldon"
    elif type_header.startswith("org.kubeflow.serving"):
        return "inferenceservice"

def get_header(header_name, headers):
    if headers.get(header_name):
        return clean_header(header_name, headers)

def field_from_header(content, header_name, headers):
    if headers.get(header_name):
        content[header_name] = clean_header(header_name, headers)


def clean_header(header_name, headers):
    header_val = headers.get(header_name)
    if header_val:
        header_val = header_val.translate({ord(c): None for c in '!@#$"<>/?'})
    return header_val


def connect_elasticsearch():
    _es = None
    elastic_host = os.getenv("ELASTICSEARCH_HOST", "localhost")
    elastic_port = os.getenv("ELASTICSEARCH_PORT", 9200)
    elastic_protocol = os.getenv("ELASTICSEARCH_PROTOCOL", "http")
    elastic_user = os.getenv("ELASTICSEARCH_USER")
    elastic_pass = os.getenv("ELASTICSEARCH_PASS")
    elastic_token = os.getenv("ELASTICSEARCH_TOKEN")

    connection_string = elastic_protocol + "://"

    if elastic_user and elastic_pass:
        connection_string = connection_string + elastic_user + ":" + elastic_pass + "@"

    connection_string = connection_string + elastic_host + ":" + str(elastic_port)
    headers = None

    if elastic_token:
        headers = {"Authorization": "Bearer " + elastic_token}

    _es = Elasticsearch(
        connection_string,
        verify_certs=False,
        connection_class=RequestsHttpConnection,
        headers=headers,
        retry_on_timeout=True,
        timeout=30,
    )
    if _es.ping():
        print("Connected to Elasticsearch")
        sys.stdout.flush()
    else:
        print("Could not connect to Elasticsearch")
        sys.stdout.flush()
        sys.exit()
    return _es


class NumpyEncoder(json.JSONEncoder):
    def default(self, obj):  # pylint: disable=arguments-differ,method-hidden
        if isinstance(
            obj,
            (
                np.int_,
                np.intc,
                np.intp,
                np.int8,
                np.int16,
                np.int32,
                np.int64,
                np.uint8,
                np.uint16,
                np.uint32,
                np.uint64,
            ),
        ):
            return int(obj)
        elif isinstance(obj, (np.float_, np.float16, np.float32, np.float64)):
            return float(obj)
        elif isinstance(obj, (np.ndarray,)):
            return obj.tolist()
        return json.JSONEncoder.default(self, obj)
