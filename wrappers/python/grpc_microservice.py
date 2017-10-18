import argparse
import os
import grpc
import json
import importlib
import time
from concurrent import futures
from rest_microservice import parse_parameters

from proto import prediction_pb2

def get_grpc_server(model_class,parameters = None):
    if parameters is None:
        parameters = {}

    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    servicer = model_class(**parameters)
    
    prediction_pb2.add_ModelServicer_to_server(servicer, server)

    return server


PARAMETERS_ENV_NAME = "PREDICTIVE_UNIT_PARAMETERS"
SERVICE_PORT_ENV_NAME = "PREDICTIVE_UNIT_SERVICE_PORT"
DEFAULT_PORT = 5000

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("model_name",type=str,help="Name of the model.")
    parser.add_argument("--parameters",type=str,default=os.environ.get(PARAMETERS_ENV_NAME,"[]"))

    args = parser.parse_args()

    parameters = parse_parameters(json.loads(args.parameters))

    model_file = importlib.import_module(args.model_name)
    model_class = getattr(model_file,args.model_name)

    server = get_grpc_server(model_class,parameters=parameters)
    server.add_insecure_port("0.0.0.0:{}".format(os.environ.get(SERVICE_PORT_ENV_NAME,DEFAULT_PORT)))
    server.start()
    print "GRPC Microservice Running"

    while True:
        time.sleep(1000)
