import pytest
from model_microservice import get_rest_microservice
import json

class UserObject(object):
    def __init__(self,metrics_ok=True):
        self.metrics_ok = metrics_ok

    def predict(self,X,features_names):
        """
        Return a prediction.

        Parameters
        ----------
        X : array-like
        feature_names : array of feature names (optional)
        """
        print("Predict called - will run identity function")
        print(X)
        return X

    def tags(self):
        return {"mytag":1}

    def metrics(self):
        if self.metrics_ok:
            return [{"type":"COUNTER","key":"mycounter","value":1}]
        else:
            return [{"type":"BAD","key":"mycounter","value":1}]



def test_model_ok():
    user_object = UserObject()
    app = get_rest_microservice(user_object,debug=True)
    client = app.test_client()
    rv = client.get('/predict?json={"data":{"ndarray":[]}}')
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 200
    assert j["meta"]["tags"] == {"mytag":1}
    assert j["meta"]["metrics"] == user_object.metrics()

def test_model_no_json():
    user_object = UserObject()
    app = get_rest_microservice(user_object,debug=True)
    client = app.test_client()
    uo = UserObject()
    rv = client.get('/predict?')
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 400

def test_model_bad_metrics():
    user_object = UserObject(metrics_ok=False)
    app = get_rest_microservice(user_object,debug=True)
    client = app.test_client()
    rv = client.get('/predict?json={"data":{"ndarray":[]}}')
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 400
    
