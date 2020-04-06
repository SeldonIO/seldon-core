from flask import Flask, request
import numpy as np
import json
import logging
import sys
import log_helper

MAX_PAYLOAD_BYTES = 300000
app = Flask(__name__)
print('starting cifar logger')
sys.stdout.flush()

log = logging.getLogger('werkzeug')
log.setLevel(logging.ERROR)


@app.route("/", methods=['GET','POST'])
def index():

    request_id = log_helper.extract_request_id(request.headers)
    type_header = request.headers.get(log_helper.TYPE_HEADER_NAME)
    message_type = log_helper.parse_message_type(type_header)
    index_name = log_helper.build_index_name(request.headers)

    body = request.get_json(force=True)

    # max size is configurable with env var or defaults to constant
    max_payload_bytes = log_helper.get_max_payload_bytes(MAX_PAYLOAD_BYTES)

    body_length = request.headers.get(log_helper.LENGTH_HEADER_NAME)
    if body_length and int(body_length) > int(max_payload_bytes):
        too_large_message = 'body too large for '+index_name+"/"+log_helper.DOC_TYPE_NAME+"/"+ request_id+ ' adding '+message_type
        print()
        sys.stdout.flush()
        return too_large_message

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
    new_content_part = process_content(message_type, message_body)

    #set metadata specific to this part (request or response)
    log_helper.field_from_header(content=new_content_part,header_name='ce-time',headers=headers)
    log_helper.field_from_header(content=new_content_part, header_name='ce-source', headers=headers)

    doc_body = {
            message_type: new_content_part
    }

    log_helper.set_metadata(doc_body,headers,message_type,request_id)

    # req or res might be batches of instances so split out into individual docs
    if "instance" in new_content_part:
        no_items_in_batch = len(new_content_part["instance"])
        index = 0
        for item in new_content_part["instance"]:
            item_body = doc_body.copy()

            item_body[message_type]['instance'] = item
            item_request_id = build_request_id_batched(request_id,no_items_in_batch,index)
            upsert_doc_to_elastic(elastic_object,message_type,item_body,item_request_id,index_name)
            index = index + 1
    elif "data" in new_content_part and message_type == 'outlier':
        no_items_in_batch = len(doc_body[message_type]["data"]["is_outlier"])
        index = 0
        for item in doc_body[message_type]["data"]["is_outlier"]:
            item_body = doc_body.copy()
            item_body[message_type]["data"]["is_outlier"] = item
            if "feature_score" in item_body[message_type]["data"] and item_body[message_type]["data"]["feature_score"] is not None and len(item_body[message_type]["data"]["feature_score"]) == no_items_in_batch:
                item_body[message_type]["data"]["feature_score"] = item_body[message_type]["data"]["feature_score"][index]
            if "instance_score" in item_body[message_type]["data"] and item_body[message_type]["data"]["instance_score"] is not None and len(item_body[message_type]["data"]["instance_score"]) == no_items_in_batch:
                item_body[message_type]["data"]["instance_score"] = item_body[message_type]["data"]["instance_score"][index]
            item_request_id = build_request_id_batched(request_id, no_items_in_batch, index)
            upsert_doc_to_elastic(elastic_object, message_type, item_body, item_request_id, index_name)
            index = index + 1
    else:
        print('unexpected data format')
        print(new_content_part)
    return

def build_request_id_batched(request_id, no_items_in_batch, item_index):
    item_request_id = request_id
    if no_items_in_batch > 1:
        item_request_id = item_request_id + "-item-" + item_index
    return item_request_id

def upsert_doc_to_elastic(elastic_object, message_type, upsert_body, request_id, index_name):
    upsert_doc = {
        "doc_as_upsert": True,
        "doc": upsert_body,
    }
    new_content = elastic_object.update(index=index_name, doc_type=log_helper.DOC_TYPE_NAME, id=request_id,
                                        body=upsert_doc, retry_on_conflict=3, refresh=True, timeout="60s")
    print('upserted to doc ' + index_name + "/" + log_helper.DOC_TYPE_NAME + "/" + request_id + ' adding ' + message_type)
    sys.stdout.flush()
    return str(new_content)

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
        requestCopy = extract_data_part(content)
        # we know this is a cifar10 image
        requestCopy["dataType"] = "image"


    if message_type == 'response':
        requestCopy = extract_data_part(content)
        # we know cifar10 response is tabular
        requestCopy["dataType"] = "tabular"

    #don't do extraction in same way for outlier
    return requestCopy


def extract_data_part(content):
    copy = content.copy()
    copy['payload'] = content

    if "instances" in copy:
        copy["instance"] = copy["instances"]
        del copy["instances"]
    if "data" in copy:
        if "tensor" in copy["data"]:
            copy["instance"] = copy["data"]["tensor"]["values"]
        if "ndarray" in copy["data"]:
            copy["instance"] = copy["data"]["ndarray"]
        del copy["data"]
    if "predictions" in copy:
        copy["instance"] = copy["predictions"]
        del copy["predictions"]

    return copy

if __name__ == "__main__":
    app.run(host='0.0.0.0', port=8080)