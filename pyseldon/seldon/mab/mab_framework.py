import redis
import pickle
import os
import json
import time, logging, threading
import grpc
from concurrent import futures

import seldon.proto.prediction_pb2 as prediction_pb2
import seldon.proto.prediction_pb2_grpc as prediction_pb2_grpc

DEFAULT_REDIS_HOST = '127.0.0.1'
DEFAULT_REDIS_PORT = 6379
DEFAULT_GRPC_PORT = 50051
DEFAULT_GRPC_HOST = '0.0.0.0'

class Repeater(threading.Thread):
    def __init__(self,function,frequency):
        self.function = function
        self.frequency = frequency
        self._stopped = False
        super(Repeater,self).__init__()
        
    def stop(self):
        print "Stopping Repeater"
        self._stopped = True
        
    def run(self):
        while not self._stopped:
            time.sleep(self.frequency)
            self.function()
            
class SavedDict(dict):
    """Dictionary that saves itself in a redis node at a regular time interval."""

    def __init__(self, redis_client, key, push_frequency=300):
        self._client = redis_client
        self._key = key
        self._saver = Repeater(self._get_saver(),push_frequency)
        self._saver.start()
        super(SavedDict,self).__init__()

    def _get_saver(self):
        def inner():
            """Overwrites redis data with local data"""
            print "Writing to redis"
            binary_data = pickle.dumps(dict(self))
            self._client.set(self._key,binary_data)
        return inner


class MAB(prediction_pb2_grpc.MABServicer):
    def __init__(self,id,deployment_key,n_branches,redis_host,push_frequency=300,**kwargs):

        self.id = id
        self.deployment_key = deployment_key
        self.n_branches = n_branches
        self.redis_client = redis.StrictRedis(host=redis_host)

        # mab_state stores the mab model up to date parameters
        self.mab_state = SavedDict(self.redis_client, self.deployment_key + '_' + self.id, push_frequency)
        
        saved_mab_state_binary = self.redis_client.get(self.deployment_key + '_' + self.id)
        if saved_mab_state_binary is not None:
            # Saved parameters already exist in redis for this MAB, use them as initial values
            self.mab_state.update(pickle.loads(saved_mab_state_binary))
        else:
            # Otherwise run setup
            self.setup(self.mab_state,self.n_branches,**kwargs)
        
    def Route(self, request, context):
        return prediction_pb2.RouteResponseDef(branch=self.route(request,self.mab_state))

    def Train(self,request,context):
        self.train(request,self.mab_state)
        return prediction_pb2.PredictionFeedbackResponseDef(success=True)
    
    def setup(self):
        pass
    
    def route(self,data,parameters):
        raise NotImplemented
        
    def train(self,data,parameters):
        raise NotImplemented


class MABServiceFactory(object):

    @staticmethod
    def create_MAB_microservice(mab_class,mab_id,deployment_key,n_branches,push_frequency=6,redis_host=DEFAULT_REDIS_HOST,grpc_port=DEFAULT_GRPC_PORT,grpc_host=DEFAULT_GRPC_HOST,**kwargs):
        """
        Create a Multi Armed Bandits microservice app

        Parameters
        ----------

        mab_class : python class
           The MAB class to use
        mab_id : str or int
           unique identifier for this mab within the predictor
        deployment_key :
           unique identifier of the deployment
        n_branches :
           number of children for this mab
        push_frequency :
           How often in seconds the mab model is pushed to redis
        """

        server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
        mab_servicer = mab_class(mab_id, deployment_key, n_branches, redis_host, push_frequency, **kwargs)

        prediction_pb2_grpc.add_MABServicer_to_server(mab_servicer, server)

        #TODO: make secure
        server.add_insecure_port('{}:{}'.format(grpc_host,grpc_port))

        return server
