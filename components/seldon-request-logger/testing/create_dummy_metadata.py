import time
import seldon_deploy_sdk
import json
from seldon_deploy_sdk import V1Model, ModelMetadataServiceApi, Configuration, ApiClient, EnvironmentApi
from seldon_deploy_sdk.auth import OIDCAuthenticator
from seldon_deploy_sdk.rest import ApiException
import os

config = Configuration()
config.host = os.getenv('DEPLOY_API_HOST')
config.oidc_client_id = os.getenv('CLIENT_ID')
config.oidc_server = os.getenv('OIDC_PROVIDER')
config.username = os.getenv('OIDC_USERNAME')
config.password = os.getenv('OIDC_PASSWORD')
config.auth_method = 'password_grant'
config.scope = os.getenv('OIDC_SCOPES')
#to use client credential set above to client_credentials and uncomment and set config.oidc_client_secret
#config.oidc_client_secret = 'xxxxx'
#note client has to be configured in identity provider for client_credentials

auth = OIDCAuthenticator(config)

config.access_token = auth.authenticate()
print(config.access_token)
api_client = ApiClient(configuration=config,authenticator=auth)
api_instance = ModelMetadataServiceApi(api_client)


models = [
    #     Same model different versions
    {
        "uri": "gs://test-model-beta-v2.0.0",
        "name": "iris",
        "version": "v1.0.0",
        "artifact_type": "SKLEARN",
        "task_type": "classification",
        "tags": {"author": "Jon"},
    },
    {
        "uri": "gs://seldon-models/sklearn/iris",
        "name": "iris",
        "version": "v2.0.0",
        "artifact_type": "SKLEARN",
        "task_type": "classification",
        "tags": {"author": "Bob"},
    },
]

for model in models:
    body = V1Model(**model)
    try:
        env_api = EnvironmentApi(api_client)
        print(env_api.read_user())

        api_response = api_instance.model_metadata_service_create_model_metadata(body)
        print(str(api_response))

        metadata_list = api_instance.model_metadata_service_list_model_metadata()
        print(str(metadata_list))
    except ApiException as e:
        print(f"Couldn't create model: {json.loads(e.body)['message']}")