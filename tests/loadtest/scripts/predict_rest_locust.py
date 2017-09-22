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


class SeldonJsLocust(TaskSet):

    def getEnviron(self,key,default):
        if key in os.environ:
            return os.environ[key]
        else:
            return default

    def get_token(self):
        print "Getting access token"
        r = self.client.request("POST","/oauth/token",headers={"Accept":"application/json"},data={"grant_type":"client_credentials"},auth=(self.oauth_key,self.oauth_secret))
        if r.status_code == 200:
            j = json.loads(r.content)
            self.access_token =  j["access_token"]
            print "got access token "+self.access_token
        else:
            print "failed to get access token"
            sys.exit(1)


    def on_start(self):
        print "on_start"
        self.oauth_key = self.getEnviron('OAUTH_KEY',"key")
        self.oauth_secret = self.getEnviron('OAUTH_SECRET',"secret")
        self.data_size = int(self.getEnviron('DATA_SIZE',"1"))
        self.get_token()

    def sendFeedback(self):
        if random()>0.5:
            j = {"reward":1.0}
        else:
            j = {"reward":0}
        jStr = json.dumps(j)
        print jStr
        r = self.client.request("POST","/api/v0.1/feedback",headers={"Content-Type":"application/json","Accept":"application/json","Authorization":"Bearer "+self.access_token},name="feedback",data=jStr)
        if r.status_code == 200:
            print "Successful feedback"
        else:
            print "Failed feedback request "+str(r.status_code)
            if r.status_code == 401:
                self.get_token()
            else:
                print r.headers
                r.raise_for_status()
        

    @task
    def getPrediction(self):
        fake_data = [[round(random(),2) for i in range(0,self.data_size)]]
        features = ["f"+str(i) for i in range (0,self.data_size)]
        j = {"request":{"features":features,"ndarray":fake_data}}
        jStr = json.dumps(j)
        print jStr
        r = self.client.request("POST","/api/v0.1/predictions",headers={"Content-Type":"application/json","Accept":"application/json","Authorization":"Bearer "+self.access_token},name="predictions",data=jStr)
        if r.status_code == 200:
            print r.content
            self.sendFeedback()
        else:
            print "Failed prediction request "+str(r.status_code)
            if r.status_code == 401:
                self.get_token()
            else:
                print r.headers
                r.raise_for_status()

class WebsiteUser(HttpLocust):
    task_set = SeldonJsLocust
    min_wait=900    # Min time between requests of each user
    max_wait=1100    # Max time between requests of each user
    stop_timeout= 1000000  # Stopping time




