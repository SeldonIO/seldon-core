from __future__ import print_function
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
print('RLIMIT_NOFILE soft limit starts as  :', soft)

#resource.setrlimit(rsrc, (65535, hard)) #limit to one kilobyte

soft, hard = resource.getrlimit(rsrc)
print('RLIMIT_NOFILE soft limit changed to :', soft)

def getEnviron(key,default):
    if key in os.environ:
        return os.environ[key]
    else:
        return default


class SeldonJsLocust(TaskSet):


    def get_token(self):
        print("Getting access token")
        r = self.client.request("POST","/oauth/token",headers={"Accept":"application/json"},data={"grant_type":"client_credentials"},auth=(self.oauth_key,self.oauth_secret))
        if r.status_code == 200:
            j = json.loads(r.content)
            self.access_token =  j["access_token"]
            print("got access token "+self.access_token)
        else:
            print("failed to get access token")
            sys.exit(1)


    def on_start(self):
        print("on_start")
        self.oauth_enabled = getEnviron('OAUTH_ENABLED',"true")
        self.oauth_key = getEnviron('OAUTH_KEY',"key")
        self.oauth_secret = getEnviron('OAUTH_SECRET',"secret")
        self.data_size = int(getEnviron('DATA_SIZE',"1"))
        self.send_feedback = int(getEnviron('SEND_FEEDBACK',"1"))
        self.path_prefix = getEnviron('REST_PATH_PREFIX',"")        
        if self.oauth_enabled == "true":
            self.get_token()
        else:
            self.access_token = "NONE"
        self.rewardProbas = [0.5,0.2,0.9,0.3,0.7]
        self.routeRewards = {}
        self.routesSeen = []

    def sendFeedback(self,response):
        route = json.dumps(response["meta"]["routing"], sort_keys=True)
        if not route in self.routeRewards:
            if len(self.routesSeen) < len(self.rewardProbas):
                self.routesSeen.append(route)
                self.routesSeen.sort()
                self.routeRewards = dict(zip(self.routesSeen,self.rewardProbas))
            else:
                self.routeRewards[route] = 0.5
        rewardProba = self.routeRewards[route]
        print(route,rewardProba)
        if random()>rewardProba:
            j = {"response":response,"reward":1.0}
        else:
            j = {"response":response,"reward":0}
        jStr = json.dumps(j)
        r = self.client.request("POST",self.path_prefix+"/api/v0.1/feedback",headers={"Content-Type":"application/json","Accept":"application/json","Authorization":"Bearer "+self.access_token},name="feedback",data=jStr)
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
        fake_data = [[round(random(),2) for i in range(0,self.data_size)]]
        features = ["f"+str(i) for i in range (0,self.data_size)]
        j = {"data":{"names":features,"ndarray":fake_data}}
        jStr = json.dumps(j)
        r = self.client.request("POST",self.path_prefix+"/api/v0.1/predictions",headers={"Content-Type":"application/json","Accept":"application/json","Authorization":"Bearer "+self.access_token},name="predictions",data=jStr)
        if r.status_code == 200:
            if self.send_feedback == 1:
                j = json.loads(r.content)
                self.sendFeedback(j)
        else:
            print("Failed prediction request "+str(r.status_code))
            if r.status_code == 401:
                if self.oauth_enabled == "true":
                    self.get_token()
            else:
                print(r.headers)
                r.raise_for_status()

class WebsiteUser(HttpLocust):
    task_set = SeldonJsLocust
    min_wait=int(getEnviron('MIN_WAIT',"900"))    # Min time between requests of each user
    max_wait=int(getEnviron('MAX_WAIT',"1100"))   # Max time between requests of each user
    stop_timeout= 1000000  # Stopping time




