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
    #    Same model different versions
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
    { #schema from https://github.com/SeldonIO/ml-prediction-schema/blob/master/examples/income-classifier.json
        "uri": "gs://seldon-models/sklearn/income/model-0.23.2",
        "name": "income",
        "version": "v2.0.0",
        "artifact_type": "SKLEARN",
        "task_type": "classification",
        "tags": {"author": "Fred"},
        "prediction_schema": {"requests":[{"name":"Age","type":"REAL","data_type":"FLOAT"},{"name":"Workclass","type":"CATEGORICAL","data_type":"INT","n_categories":9,"category_map":{"0":"?","1":"Federal-gov","2":"Local-gov","3":"Never-worked","4":"Private","5":"Self-emp-inc","6":"Self-emp-not-inc","7":"State-gov","8":"Without-pay"}},{"name":"Education","type":"CATEGORICAL","data_type":"INT","n_categories":7,"category_map":{"0":"Associates","1":"Bachelors","2":"Doctorate","3":"Dropout","4":"High School grad","5":"Masters","6":"Prof-School"}},{"name":"Marital Status","type":"CATEGORICAL","data_type":"INT","n_categories":4,"category_map":{"0":"Married","1":"Never-Married","2":"Separated","3":"Widowed"}},{"name":"Occupation","type":"CATEGORICAL","data_type":"INT","n_categories":9,"category_map":{"0":"?","1":"Admin","2":"Blue-Collar","3":"Military","4":"Other","5":"Professional","6":"Sales","7":"Service","8":"White-Collar"}},{"name":"Relationship","type":"CATEGORICAL","data_type":"INT","n_categories":6,"category_map":{"0":"Husband","1":"Not-in-family","2":"Other-relative","3":"Own-child","4":"Unmarried","5":"Wife"}},{"name":"Race","type":"CATEGORICAL","data_type":"INT","n_categories":5,"category_map":{"0":"Amer-Indian-Eskimo","1":"Asian-Pac-Islander","2":"Black","3":"Other","4":"White"}},{"name":"Sex","type":"CATEGORICAL","data_type":"INT","n_categories":2,"category_map":{"0":"Female","1":"Male"}},{"name":"Capital Gain","type":"REAL","data_type":"FLOAT"},{"name":"Capital Loss","type":"REAL","data_type":"FLOAT"},{"name":"Hours per week","type":"REAL","data_type":"FLOAT"},{"name":"Country","type":"CATEGORICAL","data_type":"INT","n_categories":11,"category_map":{"0":"?","1":"British-Commonwealth","2":"China","3":"Euro_1","4":"Euro_2","5":"Latin-America","6":"Other","7":"SE-Asia","8":"South-America","9":"United-States","10":"Yugoslavia"}}],"responses":[{"name":"Income","type":"PROBA","data_type":"FLOAT","schema":[{"name":"<=$50K"},{"name":">$50K"}]}]}
    },
    {  # schema made up to test edge cases
        "uri": "gs://seldon-models/sklearn/iris2",
        "name": "dummy",
        "version": "v1.0.0",
        "artifact_type": "SKLEARN",
        "task_type": "classification",
        "tags": {"author": "Noname"},
        "prediction_schema": {"requests":[{"name":"dummy_one_hot","type":"ONE_HOT","data_type":"INT","schema":[{"name":"dummy_one_hot_1"},{"name":"dummy_one_hot_2"}]},{"name":"dummy_categorical","type":"CATEGORICAL","data_type":"INT","n_categories":2,"category_map":{"0":"dummy_cat_0","1":"dummy_cat_1"}},{"name":"dummy_float","type":"REAL","data_type":"FLOAT"}],"responses":[{"name":"dummy_proba","type":"PROBA","data_type":"FLOAT","schema":[{"name":"dummy_proba_0"},{"name":"dummy_proba_1"}]},{"name":"dummy_float","type":"REAL","data_type":"FLOAT"}]}
    }
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
        print(e)
        print(f"Couldn't create model: {json.loads(e.body)['message']}")