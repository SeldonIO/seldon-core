import requests
from requests.auth import HTTPBasicAuth
from random import randint,random
import json
from matplotlib import pyplot as plt
import numpy as np
from tensorflow.examples.tutorials.mnist import input_data


AMBASSADOR_API_IP="localhost:8002"
API_HTTP="localhost:8003"

def get_token():
    payload = {'grant_type': 'client_credentials'}
    response = requests.post(
                "http://"+API_HTTP+"/oauth/token",
                auth=HTTPBasicAuth('oauth-key', 'oauth-secret'),
                data=payload)
    print(response.text)
    token =  response.json()["access_token"]
    return token

def rest_request_seldon(request):
    token = get_token()
    headers = {'Authorization': 'Bearer '+token}
    response = requests.post(
                "http://"+API_HTTP+"/api/v0.1/predictions",
                headers=headers,
                json=request)
    return response.json()


def rest_request(deploymentName,request):
    response = requests.post(
                "http://"+AMBASSADOR_API_IP+"/seldon/"+deploymentName+"/api/v0.1/predictions",
                json=request)
    return response.json()   
    
def rest_request_auth(deploymentName,data,username,password):
    payload = {"data":{"ndarray":data.tolist()}}
    response = requests.post(
                "http://"+AMBASSADOR_API_IP+"/seldon/"+deploymentName+"/api/v0.1/predictions",
                json=payload,
                auth=HTTPBasicAuth(username, password))
    print(response.status_code)
    return response.json()   

def send_feedback_rest(deploymentName,request,response,reward):
    feedback = {
        "request": request,
        "response": response,
        "reward": reward
    }
    ret = requests.post(
         "http://"+AMBASSADOR_API_IP+"/seldon/"+deploymentName+"/api/v0.1/feedback",
        json=feedback)
    return ret.text

def gen_image(arr):
    two_d = (np.reshape(arr, (28, 28)) * 255).astype(np.uint8)
    plt.imshow(two_d,cmap=plt.cm.gray_r, interpolation='nearest')
    return plt

def download_mnist():
    return input_data.read_data_sets("MNIST_data/", one_hot = True)

def predict_rest_mnist(mnist,deployment_name):
    batch_xs, batch_ys = mnist.train.next_batch(1)
    chosen=0
    gen_image(batch_xs[chosen]).show()
    data = batch_xs[chosen].reshape((1,784))
    features = ["X"+str(i+1) for i in range (0,784)]
    request = {"data":{"names":features,"ndarray":data.tolist()}}
    predictions = rest_request(deployment_name,request)
    #predictions = rest_request_seldon(request)    
    print("Route:"+json.dumps(predictions["meta"]["routing"],indent=2))
    fpreds = [ '%.2f' % elem for elem in predictions["data"]["ndarray"][0] ]
    m = dict(zip(predictions["data"]["names"],fpreds))
    print(json.dumps(m,indent=2))



    
