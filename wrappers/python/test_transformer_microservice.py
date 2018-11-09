import pytest
from transformer_microservice import get_rest_microservice
import json

class UserObject(object):
    def __init__(self,metrics_ok=True):
        self.metrics_ok = metrics_ok

    def transform_input(self,X,features_names):
        print("Transform input called - will run identity function")
        print(X)
        return X

    def transform_output(self,X,features_names):
        print("Transform output called - will run identity function")
        print(X)
        return X

    def tags(self):
        return {"mytag":1}

    def metrics(self):
        if self.metrics_ok:
            return [{"type":"COUNTER","key":"mycounter","value":1}]
        else:
            return [{"type":"BAD","key":"mycounter","value":1}]



def test_transformer_input_ok():
    user_object = UserObject()
    app = get_rest_microservice(user_object,debug=True)
    client = app.test_client()
    rv = client.get('/transform-input?json={"data":{"ndarray":[1]}}')
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 200
    assert j["meta"]["tags"] == {"mytag":1}
    assert j["meta"]["metrics"] == user_object.metrics()
    assert j["data"]["ndarray"] == [1]

def test_tranform_input_no_json():
    user_object = UserObject()
    app = get_rest_microservice(user_object,debug=True)
    client = app.test_client()
    uo = UserObject()
    rv = client.get('/transform-input?')
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 400

def test_transform_input_bad_metrics():
    user_object = UserObject(metrics_ok=False)
    app = get_rest_microservice(user_object,debug=True)
    client = app.test_client()
    rv = client.get('/transform-input?json={"data":{"ndarray":[]}}')
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 400

def test_transformer_output_ok():
    user_object = UserObject()
    app = get_rest_microservice(user_object,debug=True)
    client = app.test_client()
    rv = client.get('/transform-output?json={"data":{"ndarray":[1]}}')
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 200
    assert j["meta"]["tags"] == {"mytag":1}
    assert j["meta"]["metrics"] == user_object.metrics()
    assert j["data"]["ndarray"] == [1]

def test_tranform_output_no_json():
    user_object = UserObject()
    app = get_rest_microservice(user_object,debug=True)
    client = app.test_client()
    uo = UserObject()
    rv = client.get('/transform-output?')
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 400

def test_transform_output_bad_metrics():
    user_object = UserObject(metrics_ok=False)
    app = get_rest_microservice(user_object,debug=True)
    client = app.test_client()
    rv = client.get('/transform-output?json={"data":{"ndarray":[]}}')
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 400
    
