from flask import Flask, Blueprint, render_template, jsonify, current_app
from flask import request
import argparse
import os
import importlib
import json
import numpy as np
import pandas as pd
from concurrent import futures
import grpc
import time

from proto import prediction_pb2
from seldon_model import SeldonModel

predict_blueprint = Blueprint('predict',__name__)

class DataContractException(Exception):
    status_code = 400

    def __init__(self, message, status_code= None, payload=None):
        Exception.__init__(self)
        self.message = message
        if status_code is not None:
            self.status_code = status_code
        self.payload = payload

    def to_dict(self):
        rv = {"status":{"status":1,"info":self.message,"code":-1,"reason":"MICROSERVICE_BAD_DATA"}}
        return rv

@predict_blueprint.errorhandler(DataContractException)
def handle_invalid_usage(error):
    response = jsonify(error.to_dict())
    response.status_code = 400
    return response

def sanity_check(data):
    if not type(data) == dict:
        raise DataContractException("Data must be a dictionary")
    if data.get('ndarray') is None and data.get('tensor') is None:
        raise DataContractException("Data dictionary has no 'ndarray' or 'tensor' keyword.")
    # TODO: Should we check more things? Like shape not being None or empty for a tensor?

def extract_input():
    jStr = request.form.get("json")
    is_default = request.form.get("isDefault")
    if jStr:
        data = json.loads(jStr)
    else:
        raise DataContractException("Empty json parameter in data")
    if data is None or data.get('request') is None:
        raise DataContractException("Data format invalid")
    return data['request'], is_default

@predict_blueprint.route('/predict',methods=['GET','POST'])
def do_predict():
    """Prediction endpoint"""
    data,is_default = extract_input()
    model =  current_app.config["seldon_model"]
    if is_default:
        sanity_check(data)
    response = model.predict_rest(data,is_default)
    json_ret = jsonify(response)
    return json_ret
        

def get_rest_microservice(seldon_model,model_name):
    app = Flask(__name__)
    
    app.config['seldon_model'] = seldon_model
    app.config['seldon_model_name'] = model_name
    app.config['seldon_ready'] = True

    app.register_blueprint(predict_blueprint)

    return app

def get_grpc_server(seldon_model):
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    prediction_pb2.add_ModelServicer_to_server(seldon_model, server)

    return server

type_dict = {
    "INT":int,
    "FLOAT":float,
    "DOUBLE":float,
    "STRING":str
    }

PARAMETERS_ENV_NAME = "PREDICTIVE_UNIT_PARAMETERS"
SERVICE_PORT_ENV_NAME = "PREDICTIVE_UNIT_SERVICE_PORT"
DEFAULT_PORT = 5000
                          
def parse_parameters(parameters):
    parsed_parameters = {}
    for param in parameters:
        name = param.get("name")
        value = param.get("value")
        type_ = param.get("type")
        parsed_parameters[name] = type_dict[type_](value)
    return parsed_parameters

                          
if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("model_name",type=str,help="Name of the model.")
    parser.add_argument("api_type",type=str,choices=["REST","GRPC"])
    parser.add_argument("--parameters",type=str,default=os.environ.get(PARAMETERS_ENV_NAME,"[]"))
    args = parser.parse_args()
    
    parameters = parse_parameters(json.loads(args.parameters))
    
    model_file = importlib.import_module(args.model_name)
    model_class = getattr(model_file,args.model_name)

    seldon_model = SeldonModel(model_class,parameters)

    port = os.environ.get(SERVICE_PORT_ENV_NAME,DEFAULT_PORT)
    
    if args.api_type == "REST":
        app = get_rest_microservice(seldon_model,args.model_name)
        app.run(host='0.0.0.0', port=port)
        
    elif args.api_type=="GRPC":
        server = get_grpc_server(seldon_model)
        server.add_insecure_port("0.0.0.0:{}".format(port))
        server.start()
        
        print "GRPC Microservice Running on port {}".format(port)
        while True:
            time.sleep(1000)


