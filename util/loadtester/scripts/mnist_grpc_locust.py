from locust.stats import RequestStats
from locust import HttpLocust, TaskSet, task, events
import os
import sys, getopt, argparse
from random import randint,random
import json
from locust.events import EventHook
import requests
import re
import time
import resource
import socket
import signal
from socket import error as socket_error
import errno
from tensorflow.examples.tutorials.mnist import input_data
import numpy as np
from google.protobuf.json_format import MessageToJson
from proto import prediction_pb2
from proto import prediction_pb2_grpc
import grpc

def connect_to_master(host,port):
    success = False
    while not success:
        s = socket.socket(
            socket.AF_INET, socket.SOCK_STREAM)
        s.settimeout(1)
        try:
            s.connect((host, int(port)))
            s.shutdown(socket.SHUT_RD)
            print("Connected to master")
            success = True
        except socket.error as serr:
            print("Connection failed - sleeping")
            time.sleep(1)



def parse_arguments():
    parser = argparse.ArgumentParser(prog='locust')
    parser.add_argument('--host')
    parser.add_argument('--master-host',default="127.0.0.1")
    parser.add_argument('--master-port',default="5557")
    parser.add_argument('--clients',default=1, type=int)
    parser.add_argument('--hatch-rate',default=1, type=int)
    parser.add_argument('--master', action='store_true')
    parser.add_argument('--slave', action='store_true')
    args, unknown = parser.parse_known_args() 
    #args = parser.parse_args()
    opts = vars(args)
    print(args)
    if args.slave:
        print("Sleeping 10 secs hack")
        time.sleep(10)
        connect_to_master(args.master_host,args.master_port)
    return args.host, args.clients, args.hatch_rate

HOST, MAX_USERS_NUMBER, USERS_PER_SECOND = parse_arguments()

rsrc = resource.RLIMIT_NOFILE
soft, hard = resource.getrlimit(rsrc)

#resource.setrlimit(rsrc, (65535, hard)) #limit to one kilobyte

soft, hard = resource.getrlimit(rsrc)

def getEnviron(key,default):
    if key in os.environ:
        return os.environ[key]
    else:
        return default


class SeldonJsLocust(TaskSet):


    def get_token(self):
        print("Getting access token with key "+self.oauth_key+" and secret "+self.oauth_secret)
        r = self.client.request("POST","/oauth/token",headers={"Accept":"application/json"},data={"grant_type":"client_credentials"},auth=(self.oauth_key,self.oauth_secret))
        if r.status_code == 200:
            j = json.loads(r.content)
            self.access_token =  j["access_token"]
            print("got access token "+self.access_token)
        else:
            print("failed to get access token")
            print(r.status_code)
            sys.exit(1)


    def on_start(self):
        print("on_start")
        self.oauth_enabled = getEnviron('OAUTH_ENABLED',"false")
        self.oauth_key = getEnviron('OAUTH_KEY',"key")
        self.oauth_secret = getEnviron('OAUTH_SECRET',"secret")
        self.data_size = int(getEnviron('DATA_SIZE',"1"))
        self.send_feedback = int(getEnviron('SEND_FEEDBACK',"0"))
        self.endpoint = getEnviron('API_ENDPOINT',"external")        
        if self.oauth_enabled == "true":
            self.get_token()
        else:
            self.access_token = "NONE"
        channel = grpc.insecure_channel(HOST)
        if self.endpoint == "external":
            self.stub = prediction_pb2_grpc.SeldonStub(channel)
        else:
            self.stub = prediction_pb2_grpc.ModelStub(channel)
        self.mnist = input_data.read_data_sets("MNIST_data/", one_hot = True)

    def sendFeedback(self,request,response,reward):
        j = {"request":request,"response":response,"reward":reward}
        jStr = json.dumps(j)
        r = self.client.request("POST","/api/v0.1/feedback",headers={"Content-Type":"application/json","Accept":"application/json","Authorization":"Bearer "+self.access_token},name="feedback",data=jStr)
        if not r.status_code == 200:
            print("Failed feedback request "+str(r.status_code))
            if r.status_code == 401:
                if self.oauth_enabled == "true":
                    self.get_token()
            else:
                print(r.headers)
                r.raise_for_status()
                
    @task
    def getPrediction(self):
        batch_xs, batch_ys = self.mnist.train.next_batch(1)
        data = batch_xs[0].reshape((1,784))
        data = np.around(data,decimals=2)
        features = ["X"+str(i+1) for i in range (0,self.data_size)]
        #request = {"data":{"names":features,"ndarray":data.tolist()}}
        datadef = prediction_pb2.DefaultData(
            names = features,
            tensor = prediction_pb2.Tensor(
                shape = [1,784],
                values = data.flatten()
            )
        )
        request = prediction_pb2.SeldonMessage(data = datadef)
        start_time = time.time()
        try:
            response = self.stub.Predict(request=request)
            if self.send_feedback == 1:
                response = MessageToJson(response)
                response = json.loads(response)
                route = response.get("meta").get("routing").get("eg-router")
                proba = response["data"]["tensor"]["values"]
                predicted = proba.index(max(proba))
                correct = np.argmax(batch_ys[0])
                j = json.loads(response.content)
                if predicted == correct:
                    self.sendFeedback(request,j,1.0)
                    print("Correct!")
                else:
                    self.sendFeedback(request,j,0.0)
                    print("Incorrect!")                
        except Exception as e:
            total_time = int((time.time() - start_time) * 1000)            
            print(e)
            events.request_failure.fire(request_type="grpc", name=HOST, response_time=total_time, exception=e)
        else:
            total_time = int((time.time() - start_time) * 1000)
            events.request_success.fire(request_type="grpc", name=HOST, response_time=total_time, response_length=0)
                                        
class WebsiteUser(HttpLocust):
    task_set = SeldonJsLocust
    min_wait=int(getEnviron('MIN_WAIT',"900"))    # Min time between requests of each user
    max_wait=int(getEnviron('MAX_WAIT',"1100"))   # Max time between requests of each user
    stop_timeout= 1000000  # Stopping time




