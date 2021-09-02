import log_helper
import os
import seldon_deploy_sdk
from seldon_deploy_sdk import Configuration, ApiClient,EnvironmentApi,ModelMetadataServiceApi
from seldon_deploy_sdk.auth import OIDCAuthenticator

oidc_server = os.getenv("OIDC_PROVIDER")
oidc_client_id = os.getenv("CLIENT_ID")
oidc_client_secret = os.getenv("CLIENT_SECRET")
oidc_scopes = os.getenv("OIDC_SCOPES")
oidc_username = os.getenv("OIDC_USERNAME")
oidc_password = os.getenv("OIDC_PASSWORD")
oidc_auth_method = os.getenv("OIDC_AUTH_METHOD")
deploy_api_host = os.getenv("DEPLOY_API_HOST")
verify_ssl = os.getenv("VERIFY_SSL", "False")
env_api = None
metadata_api = None

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
    endpoint_name = upsert_body[log_helper.ENDPOINT_HEADER_NAME] if log_helper.ENDPOINT_HEADER_NAME in upsert_body else ""
    serving_engine = upsert_body["ServingEngine"] if "ServingEngine" in upsert_body else "seldon"

    metadata = fetch_metadata(
        namespace_name, serving_engine, inferenceservice_name, endpoint_name)
    if not metadata or metadata is None:
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

def init_api():
    config = Configuration()
    config.host = deploy_api_host
    config.oidc_client_id = oidc_client_id
    config.oidc_server = oidc_server
    config.username = oidc_username
    config.password = oidc_password
    config.oidc_client_secret = oidc_client_secret
    config.auth_method = oidc_auth_method
    if verify_ssl.lower() != "true":
        config.verify_ssl = False
        os.environ["CURL_CA_BUNDLE"] = ""

    if not config.auth_method:
        config.auth_method = 'password_grant'

    if not config.host:
        print('No DEPLOY_API_HOST - will not look up metadata from Deploy')
        return

    if not config.oidc_server:
        print('No OIDC_PROVIDER - auth will not be used in connecting to metadata')
        return

    auth = None
    if config.oidc_server:
        auth = OIDCAuthenticator(config)
        config.access_token = auth.authenticate()

    api_client = ApiClient(configuration=config, authenticator=auth)

    env_api = EnvironmentApi(api_client)
    print('connected to deploy')
    print(env_api.read_user())
    global metadata_api
    metadata_api = ModelMetadataServiceApi(api_client)


def fetch_user():
    user = env_api.read_user()
    return user


def fetch_metadata(namespace, serving_engine, inferenceservice_name, predictor_name):

    deployment_type = None
    if serving_engine == 'seldon':
        deployment_type = 'SeldonDeployment'
    if serving_engine == 'inferenceservice':
        deployment_type = 'InferenceService'

    if not deployment_type:
        print('unknown deployment type for '+namespace+' / '+inferenceservice_name)
        print(deployment_type)

    if metadata_api is None:
        print('metadata service not configured')
        return None

    # TODO: in next iteration will only need one lookup straight to model metadata
    # was expcting to set deployment_type=serving_engine but deployment_type does not seem to be a param
    runtime_metadata = metadata_api.model_metadata_service_list_runtime_metadata_for_model(
        deployment_name=inferenceservice_name,deployment_namespace=namespace,
        predictor_name=predictor_name,deployment_type=deployment_type,
        deployment_status="Running")

    if runtime_metadata is not None and runtime_metadata and \
            runtime_metadata.runtime_metadata is not None and runtime_metadata.runtime_metadata:
        if len(runtime_metadata.runtime_metadata) == 0:
            print('no runtime metadata for '+namespace+'/'+inferenceservice_name)
            return None
        if len(runtime_metadata.runtime_metadata) > 1:
            print('multiple models for '+namespace+'/'+inferenceservice_name+'/'+predictor_name+
                  ' - runtime metadata will not be used')
            return None
        model_uri = runtime_metadata.runtime_metadata[0].model_uri
        print('model is '+model_uri)
        model_metadata = metadata_api.model_metadata_service_list_model_metadata(uri=model_uri)
        if model_metadata is None or len(model_metadata.models) == 0:
            print('no model corresponding to runtime metadata '+namespace+'/'+inferenceservice_name)
            return None

        print('prediction schema for '+namespace+'/'+inferenceservice_name)
        if model_metadata.models[0].prediction_schema:
            return model_metadata.models[0].prediction_schema.to_dict()
        else:
            return None
    else:
        print('no metadata found for '+namespace+' / '+inferenceservice_name+' / '+predictor_name)
    return None


def update_props_element(props, elm):
    if not ("type" in elm):
        props[elm["name"]] = {
            "type": "float"  # TODO: Use data type if available
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
                "type": "float"  # TODO: Use data type if available
            }
            return props
