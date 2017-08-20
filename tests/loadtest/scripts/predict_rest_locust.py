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

def parse_arguments():
    parser = argparse.ArgumentParser(prog='locust')
    parser.add_argument('--host')
    parser.add_argument('--clients',default=1, type=int)
    parser.add_argument('--hatch-rate',default=1, type=int)
    parser.add_argument('--master', action='store_true')
    args, unknown = parser.parse_known_args() 
    #args = parser.parse_args()
    opts = vars(args)
    if not args.master:
        time.sleep(5)
    print args
    return args.host, args.clients, args.hatch_rate

HOST, MAX_USERS_NUMBER, USERS_PER_SECOND = parse_arguments()

slaves_connect = []
slave_report = EventHook()
ALL_SLAVES_CONNECTED = False
SLAVES_NUMBER = 1
def on_my_event(client_id,data):
    """
    Waits for all slaves to be connected and launches the swarm
    :param client_id:
    :param data:
    :return:
    """
    global ALL_SLAVES_CONNECTED
    if not ALL_SLAVES_CONNECTED:
        print "Event was fired with arguments"
        if client_id not in slaves_connect:
            slaves_connect.append(client_id)
        if len(slaves_connect) == SLAVES_NUMBER:
            print "All Slaves Connected"
            ALL_SLAVES_CONNECTED = True
            print events.slave_report._handlers
            header = {'Content-Type': 'application/x-www-form-urlencoded'}
            r = requests.post('http://127.0.0.1:8089/swarm',data={'hatch_rate':USERS_PER_SECOND,'locust_count':MAX_USERS_NUMBER},headers=header)
import resource

rsrc = resource.RLIMIT_NOFILE
soft, hard = resource.getrlimit(rsrc)
print 'RLIMIT_NOFILE soft limit starts as  :', soft

#resource.setrlimit(rsrc, (65535, hard)) #limit to one kilobyte

soft, hard = resource.getrlimit(rsrc)
print 'RLIMIT_NOFILE soft limit changed to :', soft

events.slave_report += on_my_event # Register method in slaves report event



class SeldonJsLocust(TaskSet):

    def getEnviron(self,key,default):
        if key in os.environ:
            return os.environ[key]
        else:
            return default

    def on_start(self):
        print "on_start"
        self.oauth_key = self.getEnviron('OAUTH_KEY',"key")
        self.oauth_secret = self.getEnviron('OAUTH_SECRET',"secret")
        self.data_size = int(self.getEnviron('DATA_SIZE',"784"))
        print "Getting access token"
        r = self.client.request("POST","/oauth/token",headers={"Accept":"application/json"},data={"grant_type":"client_credentials"},auth=(self.oauth_key,self.oauth_secret))
        if r.status_code == 200:
            j = json.loads(r.content)
            self.access_token =  j["access_token"]
            print "got access token "+self.access_token
        else:
            print "failed to get access token"
            sys.exit(1)

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
        else:
            print "Failed request "+str(r.status_code)
            print r.headers
            r.raise_for_status()

class WebsiteUser(HttpLocust):
    task_set = SeldonJsLocust
    min_wait=900    # Min time between requests of each user
    max_wait=1100    # Max time between requests of each user
    stop_timeout= 1000000  # Stopping time




