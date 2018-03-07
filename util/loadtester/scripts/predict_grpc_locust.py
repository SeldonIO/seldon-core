from locust.stats import RequestStats
from locust import Locust, TaskSet, task, events
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
    print args
    if args.slave:
        print "Sleeping 10 secs hack"
        time.sleep(10)
        connect_to_master(args.master_host,args.master_port)
    return args.host, args.clients, args.hatch_rate

HOST, MAX_USERS_NUMBER, USERS_PER_SECOND = parse_arguments()

rsrc = resource.RLIMIT_NOFILE
soft, hard = resource.getrlimit(rsrc)
print 'RLIMIT_NOFILE soft limit starts as  :', soft

#resource.setrlimit(rsrc, (65535, hard)) #limit to one kilobyte

soft, hard = resource.getrlimit(rsrc)
print 'RLIMIT_NOFILE soft limit changed to :', soft

def getEnviron(key,default):
    if key in os.environ:
        return os.environ[key]
    else:
        return default
    
class GrpcLocust(Locust):
    def __init__(self, *args, **kwargs):
        super(GrpcLocust, self).__init__(*args, **kwargs)

class ApiUser(GrpcLocust):

    min_wait=int(getEnviron('MIN_WAIT',"900"))    # Min time between requests of each user
    max_wait=int(getEnviron('MAX_WAIT',"1100"))   # Max time between requests of each user
    stop_timeout= 1000000  # Stopping time
    

    class task_set(TaskSet):


        def get_token(self):
            print "Getting access token"
            r = requests.post("POST","/oauth/token",headers={"Accept":"application/json"},data={"grant_type":"client_credentials"},auth=(self.oauth_key,self.oauth_secret))
            if r.status_code == 200:
                j = json.loads(r.content)
                self.access_token =  j["access_token"]
                print "got access token "+self.access_token
            else:
                print "failed to get access token"
                sys.exit(1)


        def on_start(self):
            """
            get token
            :return:
            """
            print "on start"
            self.oauth_enabled = getEnviron('OAUTH_ENABLED',"false")
            self.oauth_key = getEnviron('OAUTH_KEY',"key")
            self.oauth_secret = getEnviron('OAUTH_SECRET',"secret")
            self.data_size = int(getEnviron('DATA_SIZE',"1"))
            self.send_feedback = int(getEnviron('SEND_FEEDBACK',"1"))
            self.oauth_endpoint = getEnviron('OAUTH_ENDPOINT',"http://127.0.0.1:30015")
            #self.grpc_endpoint = getEnviron('GRPC_ENDPOINT',"127.0.0.1:30017")
            if self.oauth_enabled == "true":
                self.get_token()
            else:
                self.access_token = "NONE"
            channel = grpc.insecure_channel(HOST)                
            self.stub = prediction_pb2_grpc.SeldonStub(channel)
            self.rewardProbas = [0.5,0.2,0.9,0.3,0.7]
            self.routeRewards = {}
            self.routesSeen = []


        @task
        def get_prediction(self):
            datadef = prediction_pb2.DefaultData(
                names = ["a","b"],
                tensor = prediction_pb2.Tensor(
                    shape = [3,2],
                    values = [1.0,1.0,2.0,3.0,4.0,5.0]
                )
            )
            request = prediction_pb2.SeldonMessage(data = datadef)
            start_time = time.time()
            try:
                if self.oauth_enabled:
                    metadata = [('oauth_token', self.access_token)]
                    response = self.stub.Predict(request=request,metadata=metadata)
                else:
                    response = self.stub.Predict(request=request)
            except Exception as e:
                total_time = int((time.time() - start_time) * 1000)
                print(e)
                events.request_failure.fire(request_type="grpc", name=HOST, response_time=total_time, exception=e)
            else:
                total_time = int((time.time() - start_time) * 1000)
                events.request_success.fire(request_type="grpc", name=HOST, response_time=total_time, response_length=0)
