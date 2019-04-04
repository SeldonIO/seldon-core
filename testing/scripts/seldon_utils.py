import requests
from requests.auth import HTTPBasicAuth
from seldon_core.proto import prediction_pb2
from seldon_core.proto import prediction_pb2_grpc
import grpc
import numpy as np
import random
from retrying import retry

def create_random_data(data_size,rows=1):
    shape = [rows,data_size]
    arr = np.random.rand(rows*data_size)
    return (shape,arr)

def get_token(oauth_key,oauth_secret,namespace,endpoint):
    payload = {'grant_type': 'client_credentials'}
    if namespace is None:
        key = oauth_key
    else:
        key = oauth_key+namespace
    response = requests.post(
                "http://"+endpoint+"/oauth/token",
                auth=HTTPBasicAuth(key, oauth_secret),
                data=payload)
    print(response.text)
    token =  response.json()["access_token"]
    return token

@retry(wait_exponential_multiplier=1000, wait_exponential_max=10000, stop_max_attempt_number=7)
def rest_request_api_gateway(oauth_key,oauth_secret,namespace,endpoint="localhost:8002",data_size=5,rows=1,data=None):
    token = get_token(oauth_key,oauth_secret,namespace,endpoint)
    if data is None:
        shape, arr = create_random_data(data_size,rows)
    else:
        shape = data.shape
        arr = data.flatten()
    headers = {'Authorization': 'Bearer '+token}
    payload = {"data":{"tensor":{"shape":shape,"values":arr.tolist()}}}
    response = requests.post(
                "http://"+endpoint+"/api/v0.1/predictions",
                headers=headers,
                json=payload)
    return response

@retry(wait_exponential_multiplier=1000, wait_exponential_max=10000, stop_max_attempt_number=7)
def grpc_request_api_gateway(oauth_key,oauth_secret,namespace,rest_endpoint="localhost:8002",grpc_endpoint="localhost:8003",data_size=5,rows=1,data=None):
    token = get_token(oauth_key,oauth_secret,namespace,rest_endpoint)
    if data is None:
        shape, arr = create_random_data(data_size,rows)
    else:
        shape = data.shape
        arr = data.flatten()
    datadef = prediction_pb2.DefaultData(
            tensor = prediction_pb2.Tensor(
                shape = shape,
                values = arr
                )
            )
    request = prediction_pb2.SeldonMessage(data = datadef)
    channel = grpc.insecure_channel(grpc_endpoint)
    stub = prediction_pb2_grpc.SeldonStub(channel)
    metadata = [('oauth_token', token)]
    response = stub.Predict(request=request,metadata=metadata)
    return response

@retry(wait_exponential_multiplier=1000, wait_exponential_max=10000, stop_max_attempt_number=7)
def rest_request_ambassador(deploymentName,namespace,endpoint="localhost:8003",data_size=5,rows=1,data=None):
    if data is None:
        shape, arr = create_random_data(data_size,rows)
    else:
        shape = data.shape
        arr = data.flatten()
    payload = {"data":{"names":["a","b"],"tensor":{"shape":shape,"values":arr.tolist()}}}
    if namespace is None:
        response = requests.post(
            "http://"+endpoint+"/seldon/"+deploymentName+"/api/v0.1/predictions",
            json=payload)
    else:
        response = requests.post(
            "http://"+endpoint+"/seldon/"+namespace+"/"+deploymentName+"/api/v0.1/predictions",
            json=payload)
    return response

@retry(wait_exponential_multiplier=1000, wait_exponential_max=10000, stop_max_attempt_number=7)
def rest_request_ambassador_auth(deploymentName,namespace,username,password,endpoint="localhost:8003",data_size=5,rows=1,data=None):
    if data is None:
        shape, arr = create_random_data(data_size,rows)
    else:
        shape = data.shape
        arr = data.flatten()
    payload = {"data":{"names":["a","b"],"tensor":{"shape":shape,"values":arr.tolist()}}}
    if namespace is None:
        response = requests.post(
            "http://"+endpoint+"/seldon/"+deploymentName+"/api/v0.1/predictions",
            json=payload,
            auth=HTTPBasicAuth(username, password))
    else:
        response = requests.post(
            "http://"+endpoint+"/seldon/"+namespace+"/"+deploymentName+"/api/v0.1/predictions",
            json=payload,
            auth=HTTPBasicAuth(username, password))
    return response

@retry(wait_exponential_multiplier=1000, wait_exponential_max=100000, stop_max_attempt_number=9)
def grpc_request_ambassador(deploymentName,namespace,endpoint="localhost:8004",data_size=5,rows=1,data=None):
    if data is None:
        shape, arr = create_random_data(data_size,rows)
    else:
        shape = data.shape
        arr = data.flatten()
    datadef = prediction_pb2.DefaultData(
            tensor = prediction_pb2.Tensor(
                shape = shape,
                values = arr
                )
            )
    request = prediction_pb2.SeldonMessage(data = datadef)
    channel = grpc.insecure_channel(endpoint)
    stub = prediction_pb2_grpc.SeldonStub(channel)
    if namespace is None:
        metadata = [('seldon',deploymentName)]
    else:
        metadata = [('seldon',deploymentName),('namespace',namespace)]
    response = stub.Predict(request=request,metadata=metadata)
    return response

def grpc_request_ambassador2(deploymentName,namespace,endpoint="localhost:8004",data_size=5,rows=1,data=None):
    try:
        grpc_request_ambassador(deploymentName,namespace,endpoint=endpoint,data_size=data_size,rows=rows,data=data)
    except:
        print("Warning - caught exception")
        grpc_request_ambassador(deploymentName,namespace,endpoint=endpoint,data_size=data_size,rows=rows,data=data)

def grpc_request_api_gateway2(oauth_key,oauth_secret,namespace,rest_endpoint="localhost:8002",grpc_endpoint="localhost:8003",data_size=5,rows=1,data=None):
    try:
        grpc_request_api_gateway(oauth_key,oauth_secret,namespace,rest_endpoint=rest_endpoint,grpc_endpoint=grpc_endpoint,data_size=data_size,rows=rows,data=data)
    except:
        print("Warning - caught exception")        
        grpc_request_api_gateway(oauth_key,oauth_secret,namespace,rest_endpoint=rest_endpoint,grpc_endpoint=grpc_endpoint,data_size=data_size,rows=rows,data=data)        
    
