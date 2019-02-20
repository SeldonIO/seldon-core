import pytest
import json
import numpy as np
from google.protobuf import json_format
import base64
import tensorflow as tf

from seldon_core.wrapper import get_rest_microservice, SeldonModelGRPC, get_grpc_server
from seldon_core.proto import prediction_pb2

class UserObject(object):
    def __init__(self, metrics_ok=True):
        self.metrics_ok = metrics_ok

    def route(self, X, features_names):
        return 22

    def send_feedback(self,features, feature_names, reward, truth, routing=-1):
        print("Feedback called")
    
    def tags(self):
        return {"mytag": 1}

    def metrics(self):
        if self.metrics_ok:
            return [{"type": "COUNTER", "key": "mycounter", "value": 1}]
        else:
            return [{"type": "BAD", "key": "mycounter", "value": 1}]

class UserObjectLowLevel(object):
    def __init__(self, metrics_ok=True, ret_nparray=False):
        self.metrics_ok = metrics_ok
        self.ret_nparray = ret_nparray
        self.nparray = np.array([1, 2, 3])

    def route_rest(self, request):
        return {"data":{"ndarray":[[1]]}}

    def route_grpc(self, request):
        arr = np.array([1])
        datadef = prediction_pb2.DefaultData(
            tensor=prediction_pb2.Tensor(
                shape=(1, 1),
                values=arr
            )
        )
        request = prediction_pb2.SeldonMessage(data=datadef)
        return request


    def send_feedback_rest(self,request):
        print("Feedback called")

    def send_feedback_grpc(self,request):
        print("Feedback called")


class UserObjectLowLevelGrpc(object):
    def __init__(self, metrics_ok=True, ret_nparray=False):
        self.metrics_ok = metrics_ok
        self.ret_nparray = ret_nparray
        self.nparray = np.array([1, 2, 3])

    def route_grpc(self, request):
        arr = np.array([1])
        datadef = prediction_pb2.DefaultData(
            tensor=prediction_pb2.Tensor(
                shape=(1, 1),
                values=arr
            )
        )
        request = prediction_pb2.SeldonMessage(data=datadef)
        return request

    def send_feedback_grpc(self, request):
        print("Feedback called")

def test_router_ok():
    user_object = UserObject()
    app = get_rest_microservice(user_object)
    client = app.test_client()
    rv = client.get('/route?json={"data":{"ndarray":[2]}}')
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 200
    assert j["meta"]["tags"] == {"mytag": 1}
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]
    assert j["data"]["ndarray"] == [[22]]

def test_router_lowlevel_ok():
    user_object = UserObjectLowLevel()
    app = get_rest_microservice(user_object)
    client = app.test_client()
    rv = client.get('/route?json={"data":{"ndarray":[2]}}')
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 200
    assert j["data"]["ndarray"] == [[1]]


def test_router_no_json():
    user_object = UserObject()
    app = get_rest_microservice(user_object)
    client = app.test_client()
    uo = UserObject()
    rv = client.get('/route?')
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 400


def test_router_bad_metrics():
    user_object = UserObject(metrics_ok=False)
    app = get_rest_microservice(user_object)
    client = app.test_client()
    rv = client.get('/route?json={"data":{"ndarray":[]}}')
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 400

def test_router_feedback_ok():
    user_object = UserObject()
    app = get_rest_microservice(user_object)
    client = app.test_client()
    rv = client.get('/send-feedback?json={"request":{"data":{"ndarray":[]}},"response":{"meta":{"routing":{"1":1}}},"reward":1.0}')
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 200

def test_router_feedback_lowlevel_ok():
    user_object = UserObjectLowLevel()
    app = get_rest_microservice(user_object)
    client = app.test_client()
    rv = client.get('/send-feedback?json={"request":{"data":{"ndarray":[]}},"reward":1.0}')
    j = json.loads(rv.data)
    print(j)
    assert rv.status_code == 200
 
    

def test_router_proto_ok():
    user_object = UserObject()
    app = SeldonModelGRPC(user_object)
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(
            shape=(2, 1),
            values=arr
        )
    )
    request = prediction_pb2.SeldonMessage(data=datadef)
    resp = app.Route(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    print(j)
    assert j["meta"]["tags"] == {"mytag": 1}
    # add default type
    assert j["meta"]["metrics"][0]["key"] == user_object.metrics()[0]["key"]
    assert j["meta"]["metrics"][0]["value"] == user_object.metrics()[0]["value"]
    assert j["data"]["tensor"]["shape"] == [1, 1]    
    assert j["data"]["tensor"]["values"] == [22]

def test_router_proto_lowlevel_ok():
    user_object = UserObjectLowLevelGrpc()
    app = SeldonModelGRPC(user_object)
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(
            shape=(2, 1),
            values=arr
        )
    )
    request = prediction_pb2.SeldonMessage(data=datadef)
    resp = app.Route(request, None)
    jStr = json_format.MessageToJson(resp)
    j = json.loads(jStr)
    print(j)
    assert j["data"]["tensor"]["shape"] == [1, 1]    
    assert j["data"]["tensor"]["values"] == [1]
    
    
def test_proto_feedback():
    user_object = UserObject()
    app = SeldonModelGRPC(user_object)
    arr = np.array([1, 2])
    datadef = prediction_pb2.DefaultData(
        tensor=prediction_pb2.Tensor(
            shape=(2, 1),
            values=arr
        )
    )
    meta = prediction_pb2.Meta()
    metaJson = {}
    routing = {"1":1}
    metaJson["routing"] = routing
    json_format.ParseDict(metaJson, meta)

    request = prediction_pb2.SeldonMessage(data=datadef)
    response = prediction_pb2.SeldonMessage(meta=meta,data=datadef)    
    feedback = prediction_pb2.Feedback(request=request,response=response,reward=1.0)
    resp = app.SendFeedback(feedback, None)

def test_get_grpc_server():
    user_object = UserObject()
    server = get_grpc_server(user_object)
