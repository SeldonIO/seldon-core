import requests
from requests.auth import HTTPBasicAuth
from proto import prediction_pb2
from proto import prediction_pb2_grpc
import grpc
import numpy as np

def create_random_data(data_size):
    shape = [2,data_size]
    arr = np.random.rand(2*data_size)
    return (shape,arr)

def get_token(oauth_key,oauth_secret,endpoint):
    payload = {'grant_type': 'client_credentials'}
    response = requests.post(
                "http://"+endpoint+"/oauth/token",
                auth=HTTPBasicAuth(oauth_key, oauth_secret),
                data=payload)
    print(response.text)
    token =  response.json()["access_token"]
    return token

    
def rest_request_api_gateway(oauth_key,oauth_secret,endpoint="localhost:8002",data_size=5):
    token = get_token(oauth_key,oauth_secret,endpoint)
    shape, arr = create_random_data(data_size)
    headers = {'Authorization': 'Bearer '+token}
    payload = {"data":{"tensor":{"shape":shape,"values":arr.tolist()}}}
    response = requests.post(
                "http://"+endpoint+"/api/v0.1/predictions",
                headers=headers,
                json=payload)
    print(response.text)

def grpc_request_api_gateway(oauth_key,oauth_secret,rest_endpoint="localhost:8002",grpc_endpoint="localhost:8003",data_size=5):
    token = get_token(oauth_key,oauth_secret,rest_endpoint)
    shape, arr = create_random_data(data_size)
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
    print(response)

def rest_request_ambassador(deploymentName,endpoint="localhost:8003",data_size=5):
        shape, arr = create_random_data(data_size)
        payload = {"data":{"names":["a","b"],"tensor":{"shape":shape,"values":arr.tolist()}}}
        response = requests.post(
            "http://"+endpoint+"/seldon/"+deploymentName+"/api/v0.1/predictions",
            json=payload)
        print(response.status_code)
        print(response.text)

def rest_request_ambassador_auth(deploymentName,username,password,endpoint="localhost:8003",data_size=5):
    shape, arr = create_random_data(data_size)    
    payload = {"data":{"names":["a","b"],"tensor":{"shape":shape,"values":arr.tolist()}}}
    response = requests.post(
        "http://"+endpoint+"/seldon/"+deploymentName+"/api/v0.1/predictions",
        json=payload,
        auth=HTTPBasicAuth(username, password))
    print(response.status_code)
    print(response.text)
                    
        
def grpc_request_ambassador(deploymentName,endpoint="localhost:8004",data_size=5):
    shape, arr = create_random_data(data_size)
    datadef = prediction_pb2.DefaultData(
            tensor = prediction_pb2.Tensor(
                shape = shape,
                values = arr
                )
            )
    request = prediction_pb2.SeldonMessage(data = datadef)
    channel = grpc.insecure_channel(endpoint)
    stub = prediction_pb2_grpc.SeldonStub(channel)
    metadata = [('seldon',deploymentName)]
    response = stub.Predict(request=request,metadata=metadata)
    print(response)


