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
import log_mapping
from collections.abc import Iterable
import array

MAX_PAYLOAD_BYTES = 300000
app = Flask(__name__)
print("starting logger")
sys.stdout.flush()

log = logging.getLogger("werkzeug")
log.setLevel(logging.ERROR)

es = log_helper.connect_elasticsearch()
log_mapping.init_api()


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

    error_output = "problem logging request "
    if request_id:
        error_output = error_output + request_id
    return Response(error_output, 500)


#below basically proxies to metadata service in deploy for diagnostic purposes
@app.route("/metadata", methods=["GET", "POST"])
def metadata():
    try:
        print('in metadata endpoint - this is for debugging')
        print(request.args)
        serving_engine = request.args.get('serving_engine','SeldonDeployment')
        if not serving_engine:
            serving_engine = 'SeldonDeployment'

        namespace = request.args.get('namespace','seldon')
        name = request.args.get('name','sklearn')
        predictor = request.args.get('predictor','default')

        print(namespace+'/'+name+'/'+predictor)

        metadata = log_mapping.fetch_metadata(namespace=namespace,serving_engine=serving_engine,
                                              inferenceservice_name=name,predictor_name=predictor)
        return str(metadata)
    except Exception as ex:
        print(ex)
    sys.stdout.flush()
    return Response("problem looking up metadata", 500)

def process_and_update_elastic_doc(
    elastic_object, message_type, message_body, request_id, headers, index_name
):

    if message_type == "unknown":
        print("UNKNOWN REQUEST TYPE FOR " + request_id + " - NOT PROCESSING")
        sys.stdout.flush()

    # first do any needed transformations
    new_content_part = process_content(message_type, message_body, headers)

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

        print('is this batch?')
        print(type(new_content_part["instance"]))
        print(new_content_part["instance"])

        if type(new_content_part["instance"]) == type([]):
            print('yes it is batch')
            # if we've a list then this is batch
            # we assume first dimension is always batch

            no_items_in_batch = len(new_content_part["instance"])
            index = 0
            elements = None
            if "elements" in new_content_part:
                elements = new_content_part["elements"]

            for num, item in enumerate(new_content_part["instance"],start=0):

                item_body = doc_body.copy()

                item_body[message_type]["instance"] = item

                if type(elements) == type([]) and len(elements) > num:
                    item_body[message_type]["elements"] = elements[num]

                item_request_id = build_request_id_batched(
                    request_id, no_items_in_batch, index
                )
                upsert_doc_to_elastic(
                    elastic_object, message_type, item_body, item_request_id, index_name
                )
                index = index + 1
        else:
            print('no it is not batch')
            #not batch so don't batch elements either
            if "elements" in new_content_part and type(new_content_part["elements"]) == type([]):
                new_content_part["elements"] = new_content_part["elements"][0]

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
    log_mapping.get_log_metadata(elastic_object, message_type, upsert_body, request_id, index_name)
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
def process_content(message_type, content, headers):

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
        requestCopy = extract_data_part(content, headers, message_type)

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

def extract_data_part(content, headers, message_type):
    copy = content.copy()

    namespace = log_helper.get_header(log_helper.NAMESPACE_HEADER_NAME, headers)
    inferenceservice_name = log_helper.get_header(log_helper.INFERENCESERVICE_HEADER_NAME, headers)
    endpoint_name = log_helper.get_header(log_helper.ENDPOINT_HEADER_NAME, headers)
    serving_engine = log_helper.serving_engine(headers)

    # if 'instances' in body then tensorflow request protocol
    # if 'predictions' then tensorflow response
    # if 'model_name' and 'outputs' then v2 dataplane response
    # if 'inputs' then v2 data plane request
    # otherwise can use seldon logic for parsing and inferring type (won't be in here if outlier)

    # V2 Data Plane Response
    if "model_name" in copy and "outputs" in copy:

        # first get schema if there is one
        metadata_schema = log_mapping.fetch_metadata(namespace, serving_engine, inferenceservice_name,
                                                     endpoint_name)

        sys.stdout.flush()

        metadata_dict = {}
        if metadata_schema:
            build_metadata_names_dict_basic(message_type, metadata_dict, metadata_schema, [])

        instance_elements_for_v2_protocol(copy, metadata_dict, "outputs")

        del copy["outputs"]
        #TODO: do we really need to remove below model_name and model_version from request?
        del copy["model_name"]
        del copy["model_version"]
    elif "inputs" in copy:
        #first get schema if there is one
        metadata_schema = log_mapping.fetch_metadata(namespace, serving_engine, inferenceservice_name,
                                                     endpoint_name)

        metadata_dict = {}
        if metadata_schema:
            build_metadata_names_dict_basic(message_type, metadata_dict, metadata_schema, [])

        instance_elements_for_v2_protocol(copy, metadata_dict, "inputs")

        del copy["inputs"]
    elif "instances" in copy:

        copy["instance"] = copy["instances"]
        content_np = np.array(copy["instance"])

        copy["dataType"] = "tabular"
        first_element = content_np.item(0)

        set_datatype_from_numpy(content_np, copy, first_element)
        elements = createElelmentsArray(content_np, None, namespace, serving_engine, inferenceservice_name, endpoint_name, message_type)
        copy["elements"] = elements

        del copy["instances"]
    elif "predictions" in copy:
        copy["instance"] = copy["predictions"]
        content_np = np.array(copy["predictions"])

        copy["dataType"] = "tabular"
        first_element = content_np.item(0)
        set_datatype_from_numpy(content_np, copy, first_element)
        elements = createElelmentsArray(content_np, None, namespace, serving_engine, inferenceservice_name, endpoint_name, message_type)
        copy["elements"] = elements

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

            elements = createElelmentsArray(req_features, list(req_datadef.names), namespace, serving_engine, inferenceservice_name, endpoint_name, message_type)
            copy["elements"] = elements

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


def instance_elements_for_v2_protocol(copy, metadata_dict, input_or_output):

    # instance should be an array of all the values
    # an array of arrays actually due to batch
    instance = []
    #TODO: or if it's batch it should be an array of array... but batch isn't clear for v2 so assuming not batch

    # elments a dict of k-v for names and resolved values
    # an array of dict actually due to batch
    elements = []

    # list of lists (columns-based)
    cols_list = []
    names = []
    col_shape = None
    data_type = None

    for element in copy[input_or_output]:

        data_type = element["datatype"]
        shape = element["shape"]
        if not shape:
            # if no shape then assume a single element
            shape = [1, 1]
        data = element["data"]
        names.append(element["name"])

        currentValue = None
        # will need a separete value for values looked up from metadata (categorical)
        resolvedValue = None

        if data_type == "BYTES":
            copy["dataType"] = "text"
            currentValue = array.array('B', data).tostring()
        else:
            arr = create_np_from_v2(data, data_type, shape)
            # assuming that if not bytes then tabular
            copy["dataType"] = "tabular"
            first_element = arr.item(0)
            col_shape = arr.shape
            set_datatype_from_numpy(arr, copy, first_element)
            currentValue = arr


        cols_list.append(currentValue)

    print('col_shape')
    print(col_shape)
    print('cols_list')
    print(cols_list)

    # only transform if tabular and not proba
    if len(col_shape) < 3 and (len(col_shape)<2 or col_shape[1] < 2):
        for col in cols_list:
            for num, row_val in enumerate(col):
                print(num)
                print('in col')
                print(col)
                if len(instance) == num:
                    instance.append([])
                if len(row_val) == 1:
                    row_val = row_val[0]
                instance[num].append(row_val)
    else:
        #treat it as image
        instance = create_np_from_v2(copy[input_or_output][0]["data"], data_type, copy[input_or_output][0]["shape"])
        if metadata_dict:
            names = list(metadata_dict.keys())

    #elements only apply for tabular or proba but need names
    if len(col_shape) < 3 and names:
        for row in instance:
            elements_dict = {}
            for num, column in enumerate(row):
                # could be categorical and need to look up
                if isinstance(column, list):

                    if len(column) == 1:
                        column = column[0]

                        if names[num] in metadata_dict:
                            column = lookupValueWithMetadata(names[num], metadata_dict, column)

                print('adding to elements dict')
                print(column)
                elements_dict[names[num]] = column
            elements.append(elements_dict)

    instance = np.array(instance)
    print(len(instance.shape))
    print(len(col_shape))

    print('instance shape')
    print(instance.shape)
    print('col shape')
    print(col_shape)

    #this hack is because we end up with an extra dimension when parsing numpy but actually do want an extra dim if proba
    if len(instance.shape) > len(col_shape) and input_or_output == "inputs":
        instance = instance[0]

    print('instance shape')
    print(instance.shape)
    print('col shape')
    print(col_shape)

    print('elements')
    print(elements)
    print('instance')
    print(instance)

    # might contain one_hot but we don't merge one_hot for v2, not yet clear what we'll do but prob something like proba

    #proba has special handling for v2 - will be single field containing multiple elements so find name for each
    #TODO: is it the same for one_hot? not clear for v2
    if elements:
        for elements_row in elements:
            for key, value in elements_row.items():

                if metadata_dict and key in metadata_dict:
                    metadata_elem = metadata_dict[key]

                    if metadata_elem['type'] == "PROBA":
                        print('type is proba')
                        #check if this element contains a list

                        if isinstance(value, np.ndarray):
                            print('its an ndarray')
                            value = value.tolist()

                        if isinstance(value, list):

                            #if we have metadata then replace proba list with k-v pairs
                            #to do that the schema needs to tell us what the names should be
                            if metadata_elem['schema']:
                                new_entries = {}
                                for num, proba_sub_elem in enumerate(metadata_elem['schema']):
                                    subelem_name = proba_sub_elem['name']
                                    #assume index of element in schema to correspond to ordering in body
                                    new_entries[subelem_name] = value[num]

                                elements_row[key] = new_entries

    copy["elements"] = elements
    copy["instance"] = instance.tolist()


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


def createElelmentsArray(X: np.ndarray, names: list, namespace_name, serving_engine, inferenceservice_name, endpoint_name, message_type):
    metadata_schema = None

    if namespace_name is not None and inferenceservice_name is not None and serving_engine is not None and endpoint_name is not None:
        metadata_schema = log_mapping.fetch_metadata(namespace_name, serving_engine, inferenceservice_name, endpoint_name)
    else:
        print('missing a param required for metadata lookup')
        sys.stdout.flush()

    results = None
    if not metadata_schema:
        results = createElementsNoMetadata(X, names, results)
    else:
        results = createElementsWithMetadata(X, names, results, metadata_schema, message_type)

    return results


def createElementsWithMetadata(X, names, results, metadata_schema, message_type):
    #we want field names from metadata if available - also build a dict of metadata by name for easy lookup
    metadata_dict = {}

    names = build_metadata_names_dict_probas(message_type, metadata_dict, metadata_schema, names)

    if isinstance(X, np.ndarray):
        if len(X.shape) == 1:
            temp_results = []
            results = []
            for i in range(X.shape[0]):
                d = {}
                for num, name in enumerate(names, start=0):
                    if isinstance(X[i], bytes):
                        d[name] = lookupValueWithMetadata(name,metadata_dict,X[i].decode("utf-8"))
                    else:
                        d[name] = lookupValueWithMetadata(name,metadata_dict,X[i])
                temp_results.append(d)
            results = mergeLinkedColumns(temp_results, metadata_dict)
        elif len(X.shape) >= 2:
            temp_results = []
            results = []
            for i in range(X.shape[0]):
                d = {}
                for num, name in enumerate(names, start=0):
                    d[name] = X[i, num].tolist()
                    if isinstance(d[name], Iterable):
                        newlist = []
                        for val in d[name]:
                            newlist.append(lookupValueWithMetadata(name,metadata_dict,val))
                        d[name] = newlist
                    else:
                        d[name] = lookupValueWithMetadata(name,metadata_dict,d[name])
                temp_results.append(d)
            results = mergeLinkedColumns(temp_results, metadata_dict)

    return results


def build_metadata_names_dict_basic(message_type, metadata_dict, metadata_schema, names):

    if 'requests' in metadata_schema and message_type == "request":
        names = []
        # we'll get names from metadata, assuming field order to match the request
        for elem in metadata_schema['requests']:
            if elem['name']:
                names.append(elem['name'])
                metadata_dict[elem['name']] = elem
    if 'responses' in metadata_schema and message_type == "response":
        names = []
        # we'll get names from metadata, assuming field order to match the response
        for elem in metadata_schema['responses']:
            if elem['name']:
                names.append(elem['name'])
                metadata_dict[elem['name']] = elem
    return names


def build_metadata_names_dict_probas(message_type, metadata_dict, metadata_schema, names):
    if 'requests' in metadata_schema and message_type == "request":
        names = []
        # we'll get names from metadata, assuming field order to match the request
        for elem in metadata_schema['requests']:
            if elem['name']:
                if elem['type'] == "ONE_HOT" or elem['type'] == "PROBA":
                    # don't just take element as name - instead each name entry in schema is a column name
                    # used later for merging columns
                    for subelem in elem['schema']:
                        names.append(subelem['name'])
                        metadata_dict[subelem['name']] = elem
                else:
                    # used for categorical lookups
                    names.append(elem['name'])
                    metadata_dict[elem['name']] = elem
    if 'responses' in metadata_schema and message_type == "response":
        names = []
        # we'll get names from metadata, assuming field order to match the response
        for elem in metadata_schema['responses']:
            if elem['name']:
                if elem['type'] == "PROBA" or elem['type'] == "ONE_HOT":
                    # don't just take element as name - instead each name entry in schema is a column name
                    for subelem in elem['schema']:
                        names.append(subelem['name'])
                        metadata_dict[subelem['name']] = elem
                else:
                    names.append(elem['name'])
                    metadata_dict[elem['name']] = elem
    return names


def mergeLinkedColumns(raw_list, metadata_dict):
    new_list = []

    #one_hot and proba elements need to be grouped
    #the names from the schema section of their entries in the metadata tell us which columns to group
    #they should be grouped under a new top column with the name of the top-level element in the metadata
    #e.g. for Income we have top-level proba element and subelements for greater or less than 50k
    #so we should end up with "elements":{"Income":{">$50K":0.14611811908359656,"<=$50K":0.8538818809164035}}}}}
    elems_to_group = {}

    for key, metadata_elem in metadata_dict.items():
        if metadata_elem['type'] == "ONE_HOT" or metadata_elem['type'] == "PROBA":
            for subelem in metadata_elem['schema']:
                sub_name = subelem['name']
                elems_to_group[sub_name] = metadata_elem['name']

    #raw list is actually a list containing a dict or set of dicts
    for dict in raw_list:

        #transform each dict to group one_hot and proba fields
        new_dict = {}

        for key, elem in dict.items():

            if not elems_to_group or key not in elems_to_group:
                #if element doesn't need to be grouped (e.g. it's a float) then just add it
                new_dict[key] = elem
            else:
                top_elem_name = elems_to_group[key]

                #if top_elem_name entry already in dict we're building then add to that
                if top_elem_name in new_dict:
                    new_dict[top_elem_name][key] = elem
                else:
                    #otherwise put in a new element
                    new_dict[top_elem_name] = {key: elem}

        if new_dict:
            new_list.append(new_dict)

    return new_list

def lookupValueWithMetadata(name, metadata_dict, raw_value):
    metadata_elem = metadata_dict[name]

    if not metadata_elem:
        return raw_value

    #categorical currently only case where we replace value
    if metadata_elem['type'] == "CATEGORICAL":

        if metadata_elem['data_type'] == 'INT':
            #need to convert raw vals back to ints as could have been floatified
            raw_value = int(raw_value)

        if metadata_elem['category_map'][str(raw_value)] is not None:
            return metadata_elem['category_map'][str(raw_value)]
        return raw_value

    return raw_value

def createElementsNoMetadata(X, names, results):
    if not names:
        return results
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
