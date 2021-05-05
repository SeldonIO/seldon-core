import log_helper

default_mapping = {
    "properties": {
        "@timestamp": {
            "type": "date"
        },
        "Ce-Endpoint": {
            "type": "keyword"
        },
        "Ce-Inferenceservicename": {
            "type": "keyword"
        },
        "Ce-Modelid": {
            "type": "keyword"
        },
        "Ce-Namespace": {
            "type": "keyword"
        },
        "RequestId": {
            "type": "keyword"
        },
        "ServingEngine": {
            "type": "keyword"
        },
        "request": {
            "properties": {
                "ce-source": {
                    "type": "keyword",
                    "index": "false"
                },
                "ce-time": {
                    "type": "date",
                    "index": "false"
                },
                "dataType": {
                    "type": "keyword"
                }
            }
        },
        "response": {
            "properties": {
                "ce-source": {
                    "type": "keyword",
                    "index": "false"
                },
                "ce-time": {
                    "type": "date",
                    "index": "false"
                },
                "dataType": {
                    "type": "keyword"
                }
            }
        }
    }
}


def get_log_metadata(elastic_object, message_type, upsert_body, request_id, index_name):
    index_exist = elastic_object.indices.exists(index=index_name)
    if not index_exist:
        print("Index doesn't exists. Creating index with mapping for ", index_name)
        try:
            mapping_body = get_index_mapping(index_name, upsert_body)
            elastic_object.indices.create(
                index=index_name,
                body={"mappings": mapping_body}
            )
            print("Created index with mapping ->", index_name)
        except Exception as ex:
            print(ex)


def get_index_mapping(index_name, upsert_body):
    index_mapping = default_mapping.copy()
    inferenceservice_name = upsert_body[log_helper.INFERENCESERVICE_HEADER_NAME] if log_helper.INFERENCESERVICE_HEADER_NAME in upsert_body else ""
    namespace_name = upsert_body[log_helper.NAMESPACE_HEADER_NAME] if log_helper.NAMESPACE_HEADER_NAME in upsert_body else ""
    serving_engine = upsert_body["ServingEngine"] if "ServingEngine" in upsert_body else "seldon"

    metadata = fetch_metadata(
        namespace_name, serving_engine, inferenceservice_name)
    if not metadata:
        return index_mapping
    else:
        print("Retrieved metadata for index", index_name)
        if "requests" in metadata:
            req_mapping = get_field_mapping(metadata["requests"])
            if req_mapping != None:
                index_mapping["properties"]["request"]["properties"]["elements"] = req_mapping

        if "responses" in metadata:
            resp_mapping = get_field_mapping(metadata["responses"])
            if resp_mapping != None:
                index_mapping["properties"]["response"]["properties"]["elements"] = resp_mapping
        return index_mapping


def get_field_mapping(metadata):
    props = {}
    if not metadata:
        return None
    else:
        for elm in metadata:
            props = update_props_element(props, elm)
    return None if not props else {"properties": props}


def fetch_metadata(namespace, serving_engine, inferenceservice_name):
    # Fetch real metadata
    print("Fetching predictions schema for", namespace+"/" +
          serving_engine+"/"+inferenceservice_name)
    return None


def update_props_element(props, elm):
    if not ("type" in elm):
        props[elm["name"]] = {
            "type": "float"  # Use data type if available
        }
        return props

    else:
        if elm["type"] == "CATEGORICAL":
            props[elm["name"]] = {
                "type": "keyword"
            }
            return props
        if elm["type"] == "TEXT":
            props[elm["name"]] = {
                "type": "text"
            }
            return props
        if elm["type"] == "PROBA" or elm["type"] == "ONE_HOT":
            props[elm["name"]] = get_field_mapping(elm["schema"])
            return props
        else:  # For REAL,TENSOR
            props[elm["name"]] = {
                "type": "float"  # Use data type if available
            }
            return props
