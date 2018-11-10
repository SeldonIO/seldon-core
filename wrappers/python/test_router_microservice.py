import pytest
from router_microservice import get_rest_microservice
import json

class UserObject(object):
    def __init__(self,metrics_ok=True):
        self.metrics_ok = metrics_ok

    def route(self,X,features_names):
        return 22
    
    def tags(self):
        return {"mytag":1}

    def metrics(self):
        if self.metrics_ok:
            return [{"type":"COUNTER","key":"mycounter","value":1}]
        else:
            return [{"type":"BAD","key":"mycounter","value":1}]



def test_router_ok():
    user_object = UserObject()
    app = get_rest_microservice(user_object,debug=True)
    client = app.test_client()
    rv = client.get('/route?json={"data":{"ndarray":[2]}}')
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 200
    assert j["meta"]["tags"] == {"mytag":1}
    assert j["meta"]["metrics"] == user_object.metrics()
    assert j["data"]["ndarray"] == [[22]]    

def test_router_no_json():
    user_object = UserObject()
    app = get_rest_microservice(user_object,debug=True)
    client = app.test_client()
    uo = UserObject()
    rv = client.get('/route?')
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 400

def test_router_bad_metrics():
    user_object = UserObject(metrics_ok=False)
    app = get_rest_microservice(user_object,debug=True)
    client = app.test_client()
    rv = client.get('/route?json={"data":{"ndarray":[]}}')
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 400
    
