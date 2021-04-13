import log_helper
import example_metadata

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
        if "request" in metadata:
            req_mapping = get_field_mapping(metadata["request"])
            if req_mapping != None:
                index_mapping["properties"]["request"]["properties"]["elements"] = req_mapping

        if "response" in metadata:
            resp_mapping = get_field_mapping(metadata["response"])
            if resp_mapping != None:
                index_mapping["properties"]["response"]["properties"]["elements"] = resp_mapping

        return index_mapping


def get_field_mapping(metadata):
    props = {}
    if not metadata or not ("schema" in metadata):
        return None
    else:
        for elm in metadata["schema"]:
            props[elm["name"]] = {
                "type": get_data_type(elm)
            }
        return None if not props else {"properties": props}


def fetch_metadata(namespace, serving_engine, inferenceservice_name):
    # Fetch metadata for a specific case
    print(namespace, serving_engine, inferenceservice_name)
    if namespace == "seldon" and serving_engine == "seldon" and inferenceservice_name == "income-classifier":
        return example_metadata.metadata
    else:
        return None


def get_data_type(elm):
    if elm["qdtype"] == "real":
        return elm["dtype"] if elm["dtype"] != None else "float"
    else:
        return "keyword"
