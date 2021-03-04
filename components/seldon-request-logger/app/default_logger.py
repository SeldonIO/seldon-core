from flask import Flask, request, Response
from seldon_core.utils import (
    json_to_seldon_message,
    extract_request_parts,
    array_to_grpc_datadef,
    seldon_message_to_json,
)
from seldon_core.proto import prediction_pb2
import numpy as np
import json
import logging
import sys
import log_helper
from collections.abc import Iterable
import array

MAX_PAYLOAD_BYTES = 300000
app = Flask(__name__)
print("starting logger")
sys.stdout.flush()

log = logging.getLogger("werkzeug")
log.setLevel(logging.ERROR)

es = log_helper.connect_elasticsearch()


@app.route("/", methods=["GET", "POST"])
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
        too_large_message = (
            "body too large for "
            + index_name
            + "/"
            + (log_helper.DOC_TYPE_NAME if log_helper.DOC_TYPE_NAME != None else "_doc")
            + "/"
            + request_id
            + " adding "
            + message_type
        )
        print(too_large_message)
        sys.stdout.flush()
        return too_large_message

    if not type(body) is dict:
        body = json.loads(body)

    # print('RECEIVED MESSAGE.')
    # print(str(request.headers))
    # print(str(body))
    # print('----')
    # sys.stdout.flush()

    try:

        # now process and update the doc
        process_and_update_elastic_doc(
            es, message_type, body, request_id, request.headers, index_name
        )

        return ""
    except Exception as ex:
        print(ex)
    sys.stdout.flush()
    return Response("problem logging request", 500)


def process_and_update_elastic_doc(
    elastic_object, message_type, message_body, request_id, headers, index_name
):

    if message_type == "unknown":
        print("UNKNOWN REQUEST TYPE FOR " + request_id + " - NOT PROCESSING")
        sys.stdout.flush()

    # first do any needed transformations
    new_content_part = process_content(message_type, message_body)

    # set metadata to go just in this part (request or response) and not top-level
    log_helper.field_from_header(
        content=new_content_part, header_name="ce-time", headers=headers
    )
    log_helper.field_from_header(
        content=new_content_part, header_name="ce-source", headers=headers
    )

    doc_body = {message_type: new_content_part}

    log_helper.set_metadata(doc_body, headers, message_type, request_id)

    # req or res might be batches of instances so split out into individual docs
    if "instance" in new_content_part:

        if type(new_content_part["instance"]) == type([]):
            # if we've a list then this is batch
            # we assume first dimension is always batch

            no_items_in_batch = len(new_content_part["instance"])
            index = 0
            for item in new_content_part["instance"]:

                item_body = doc_body.copy()

                item_body[message_type]["instance"] = item
                item_request_id = build_request_id_batched(
                    request_id, no_items_in_batch, index
                )
                upsert_doc_to_elastic(
                    elastic_object, message_type, item_body, item_request_id, index_name
                )
                index = index + 1
        else:
            item_request_id = build_request_id_batched(request_id, 1, 0)
            upsert_doc_to_elastic(
                elastic_object, message_type, doc_body, item_request_id, index_name
            )
    elif message_type == "feedback":
        item_request_id = build_request_id_batched(request_id, 1, 0)
        upsert_doc_to_elastic(elastic_object, message_type, doc_body, item_request_id, index_name)
    elif "data" in new_content_part and message_type == "outlier":
        no_items_in_batch = len(doc_body[message_type]["data"]["is_outlier"])
        index = 0
        for item in doc_body[message_type]["data"]["is_outlier"]:
            item_body = doc_body.copy()
            item_body[message_type]["data"]["is_outlier"] = item
            if (
                "feature_score" in item_body[message_type]["data"]
                and item_body[message_type]["data"]["feature_score"] is not None
                and len(item_body[message_type]["data"]["feature_score"])
                == no_items_in_batch
            ):
                item_body[message_type]["data"]["feature_score"] = item_body[
                    message_type
                ]["data"]["feature_score"][index]
            if (
                "instance_score" in item_body[message_type]["data"]
                and item_body[message_type]["data"]["instance_score"] is not None
                and len(item_body[message_type]["data"]["instance_score"])
                == no_items_in_batch
            ):
                item_body[message_type]["data"]["instance_score"] = item_body[
                    message_type
                ]["data"]["instance_score"][index]
            item_request_id = build_request_id_batched(
                request_id, no_items_in_batch, index
            )
            upsert_doc_to_elastic(
                elastic_object, message_type, item_body, item_request_id, index_name
            )
            index = index + 1
    else:
        print("unexpected data format")
        print(new_content_part)
    return


def build_request_id_batched(request_id, no_items_in_batch, item_index):
    item_request_id = request_id
    if no_items_in_batch > 1:
        item_request_id = item_request_id + "-item-" + str(item_index)
    return item_request_id


def upsert_doc_to_elastic(
    elastic_object, message_type, upsert_body, request_id, index_name
):
    upsert_doc = {
        "doc_as_upsert": True,
        "doc": upsert_body,
    }
    new_content = elastic_object.update(
        index=index_name,
        doc_type=log_helper.DOC_TYPE_NAME,
        id=request_id,
        body=upsert_doc,
        retry_on_conflict=3,
        refresh=True,
        timeout="60s",
    )
    print(
        "upserted to doc "
        + index_name
        + "/"
        + (log_helper.DOC_TYPE_NAME if log_helper.DOC_TYPE_NAME != None else "_doc")
        + "/"
        + request_id
        + " adding "
        + message_type
    )
    sys.stdout.flush()
    return str(new_content)


# take request or response part and process it by deriving metadata
def process_content(message_type, content):

    if content is None:
        print("content is empty")
        sys.stdout.flush()
        return content

    # if we have dataType then have already parsed before
    if "dataType" in content:
        print("dataType already in content")
        sys.stdout.flush()
        return content

    requestCopy = content.copy()

    # extract data part out and process for req or resp - handle differently later for outlier
    if message_type == "request" or message_type == "response":
        requestCopy = extract_data_part(content)

    return requestCopy

def create_np_from_v2(data: list,ty: str, shape: list) -> np.array:
    npty = np.float
    if ty == "BOOL":
        npty = np.bool
    elif ty ==  "UINT8":
        npty = np.uint8
    elif ty == "UINT16":
        npty = np.uint16
    elif ty == "UINT32":
        npty = np.uint32
    elif ty == "UINT64":
        npty = np.uint64
    elif ty == "INT8":
        npty = np.int8
    elif ty == "INT16":
        npty = np.int16
    elif ty == "INT32":
        npty = np.int32
    elif ty == "INT64":
        npty = np.int64
    elif ty == "FP16":
        npty = np.float32
    elif ty == "FP32":
        npty = np.float32
    elif ty == "FP64":
        npty = np.float64
    else:
        raise ValueError(f"V2 unknown type or type that can't be coerced {ty}")

    arr = np.array(data, dtype=npty)
    arr.shape = tuple(shape)
    return arr

def extract_data_part(content):
    copy = content.copy()

    # if 'instances' in body then tensorflow request protocol
    # if 'predictions' then tensorflow response
    # if 'model_name' and 'outputs' then v2 dataplane response
    # if 'inputs' then v2 data plane request
    # otherwise can use seldon logic for parsing and inferring type (won't be in here if outlier)

    # V2 Data Plane Response
    if "model_name" in copy and "outputs" in copy:
        # assumes single output
        output = copy["outputs"][0]
        data_type = output["datatype"]
        shape = output["shape"]
        data = output["data"]

        if data_type == "BYTES":
            copy["dataType"] = "text"
            copy["instance"] = array.array('B', data).tostring()
        else:
            arr = create_np_from_v2(data, data_type, shape)
            copy["dataType"] = "tabular"
            first_element = arr.item(0)
            set_datatype_from_numpy(arr, copy, first_element)
            copy["instance"] = arr.tolist()

        del copy["outputs"]
        del copy["model_name"]
        del copy["model_version"]
    elif "inputs" in copy:
        # assumes single input
        inputs = copy["inputs"][0]
        data_type = inputs["datatype"]
        shape = inputs["shape"]
        data = inputs["data"]

        if data_type == "BYTES":
            copy["dataType"] = "text"
            copy["instance"] = array.array('B', data).tostring()
        else:
            arr = create_np_from_v2(data, data_type, shape)
            copy["dataType"] = "tabular"
            first_element = arr.item(0)
            set_datatype_from_numpy(arr, copy, first_element)
            copy["instance"] = arr.tolist()

        del copy["inputs"]
    elif "instances" in copy:

        copy["instance"] = copy["instances"]
        content_np = np.array(copy["instance"])

        copy["dataType"] = "tabular"
        first_element = content_np.item(0)

        set_datatype_from_numpy(content_np, copy, first_element)
        del copy["instances"]
    elif "predictions" in copy:
        copy["instance"] = copy["predictions"]
        content_np = np.array(copy["predictions"])

        copy["dataType"] = "tabular"
        first_element = content_np.item(0)
        set_datatype_from_numpy(content_np, copy, first_element)

        del copy["predictions"]
    else:
        requestMsg = json_to_seldon_message(copy)

        (req_features, _, req_datadef, req_datatype) = extract_request_parts(requestMsg)

        # set sensible defaults for non-tabular dataTypes
        # tabular should be iterable and get inferred through later block
        if req_datatype == "strData":
            copy["dataType"] = "text"
        if req_datatype == "jsonData":
            copy["dataType"] = "json"
        if req_datatype == "binData":
            copy["dataType"] = "image"

        if isinstance(req_features, Iterable):

            elements = createElelmentsArray(req_features, list(req_datadef.names))

            if isinstance(elements, Iterable):

                for i, e in enumerate(elements):
                    reqJson = extractRow(
                        i, requestMsg, req_datatype, req_features, req_datadef
                    )
                    reqJson["elements"] = e
                    copy = reqJson

        copy["instance"] = json.loads(
            json.dumps(req_features, cls=log_helper.NumpyEncoder)
        )

        if isinstance(req_features, np.ndarray):
            set_datatype_from_numpy(req_features, copy, req_features.item(0))

    # copy names into its own section of request
    if "data" in content:
        if "names" in content["data"]:
            copy["names"] = content["data"]["names"]

    # should now have processed content of seldon message so don't want its various constructs on top-level anymore
    if "data" in copy:
        del copy["data"]
    if "strData" in copy:
        del copy["strData"]
    if "jsonData" in copy:
        del copy["jsonData"]
    if "binData" in copy:
        del copy["binData"]

    copy["payload"] = content

    return copy


def set_datatype_from_numpy(content_np, copy, first_element):

    if first_element is not None and not isinstance(first_element, (int, float)):
        copy["dataType"] = "text"
    if first_element is not None and isinstance(first_element, (int, float)):
        copy["dataType"] = "number"
    if content_np.shape is not None and len(content_np.shape) > 1:
        # first dim is batch so second reveals whether instance is array
        if content_np.shape[1] > 1:
            copy["dataType"] = "tabular"
    if len(content_np.shape) > 2:
        copy["dataType"] = "tabular"
    if len(content_np.shape) > 3:
        copy["dataType"] = "image"


def extractRow(
    i: int,
    requestMsg: prediction_pb2.SeldonMessage,
    req_datatype: str,
    req_features: np.ndarray,
    req_datadef: "prediction_pb2.SeldonMessage.data",
):
    datatyReq = "ndarray"
    dataType = "tabular"
    if len(req_features.shape) == 2:
        dataReq = array_to_grpc_datadef(
            datatyReq, np.expand_dims(req_features[i], axis=0), req_datadef.names
        )
    else:
        if len(req_features.shape) > 2:
            dataType = "image"
        else:
            dataType = "text"
            req_features = np.char.decode(req_features.astype("S"), "utf-8")
        dataReq = array_to_grpc_datadef(
            datatyReq, np.expand_dims(req_features[i], axis=0), req_datadef.names
        )
    requestMsg2 = prediction_pb2.SeldonMessage(data=dataReq, meta=requestMsg.meta)
    reqJson = {}
    reqJson["payload"] = seldon_message_to_json(requestMsg2)
    # setting dataType here temporarily so calling method will be able to access it
    # don't want to set it at the payload level
    reqJson["dataType"] = dataType
    return reqJson


def createElelmentsArray(X: np.ndarray, names: list):
    results = None
    if isinstance(X, np.ndarray):
        if len(X.shape) == 1:
            results = []
            for i in range(X.shape[0]):
                d = {}
                for num, name in enumerate(names, start=0):
                    if isinstance(X[i], bytes):
                        d[name] = X[i].decode("utf-8")
                    else:
                        d[name] = X[i]
                results.append(d)
        elif len(X.shape) >= 2:
            results = []
            for i in range(X.shape[0]):
                d = {}
                for num, name in enumerate(names, start=0):
                    d[name] = X[i, num].tolist()
                results.append(d)
    return results


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8080)
