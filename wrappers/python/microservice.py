from proto import prediction_pb2
import persistence

from flask import Flask, Blueprint, request
import argparse
import numpy as np
import os
import importlib
import json
import time

from google.protobuf.struct_pb2 import ListValue

PARAMETERS_ENV_NAME = "PREDICTIVE_UNIT_PARAMETERS"
SERVICE_PORT_ENV_NAME = "PREDICTIVE_UNIT_SERVICE_PORT"
DEFAULT_PORT = 5000

DEBUG_PARAMETER = "SELDON_DEBUG"
DEBUG = False

class SeldonMicroserviceException(Exception):
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

def sanity_check_request(req):
    if not type(req) == dict:
        raise SeldonMicroserviceException("Request must be a dictionary")
    data = req.get("data")
    if data is None:
        raise SeldonMicroserviceException("Request must contain Default Data")
    if not type(data) == dict:
        raise SeldonMicroserviceException("Data must be a dictionary")
    if data.get('ndarray') is None and data.get('tensor') is None:
        raise SeldonMicroserviceException("Data dictionary has no 'ndarray' or 'tensor' keyword.")
    # TODO: Should we check more things? Like shape not being None or empty for a tensor?

def extract_message():
    jStr = request.form.get("json")
    if jStr:
        message = json.loads(jStr)
    else:
        raise SeldonMicroserviceException("Empty json parameter in data")
    if message is None:
        raise SeldonMicroserviceException("Invalid Data Format")
    return message

def array_to_list_value(array,lv=None):
    if lv is None:
        lv = ListValue()
    if len(array.shape) == 1:
        lv.extend(array)
    else:
        for sub_array in array:
            sub_lv = lv.add_list()
            array_to_list_value(sub_array,sub_lv)
    return lv

def rest_datadef_to_array(datadef):
    if datadef.get("tensor") is not None:
        features = np.array(datadef.get("tensor").get("values")).reshape(datadef.get("tensor").get("shape"))
    elif datadef.get("ndarray") is not None:
        features = np.array(datadef.get("ndarray"))
    else:
        features = np.array([])
    return features

def array_to_rest_datadef(array,names,original_datadef):
    datadef = {"names":names}
    if original_datadef.get("tensor") is not None:
        datadef["tensor"] = {
            "shape":array.shape,
            "values":array.ravel().tolist()
        }
    elif original_datadef.get("ndarray") is not None:
        datadef["ndarray"] = array.tolist()
    else:
        datadef["ndarray"] = array.tolist()
    return datadef

def grpc_datadef_to_array(datadef):
    data_type = datadef.WhichOneof("data_oneof")
    if data_type == "tensor":
        features = np.array(datadef.tensor.values).reshape(datadef.tensor.shape)
    elif data_type == "ndarray":
        features = np.array(datadef.ndarray)
    else:
        features = np.array([])
    return features

def array_to_grpc_datadef(array,names,data_type):
    if data_type == "tensor":
        datadef = prediction_pb2.DefaultData(
            names = names,
            tensor = prediction_pb2.Tensor(
                shape = array.shape,
                values = array.ravel().tolist()
            )
        )
    elif data_type == "ndarray":
        datadef = prediction_pb2.DefaultData(
            names = names,
            ndarray = array_to_list_value(array)
        )
    else:
        datadef = prediction_pb2.DefaultData(
            names = names,
            ndarray = array_to_list_value(array)
        )

    return datadef

def parse_parameters(parameters):
    type_dict = {
        "INT":int,
        "FLOAT":float,
        "DOUBLE":float,
        "STRING":str,
        "BOOL":bool
    }
    parsed_parameters = {}
    for param in parameters:
        name = param.get("name")
        value = param.get("value")
        type_ = param.get("type")
        parsed_parameters[name] = type_dict[type_](value)
    return parsed_parameters
                          
if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("interface_name",type=str,help="Name of the user interface.")
    parser.add_argument("api_type",type=str,choices=["REST","GRPC"])

    parser.add_argument("--service-type",type=str,choices=["MODEL","ROUTER","TRANSFORMER","COMBINER","OUTLIER_DETECTOR"],default="MODEL")
    parser.add_argument("--persistence",nargs='?',default=0,const=1,type=int)
    parser.add_argument("--parameters",type=str,default=os.environ.get(PARAMETERS_ENV_NAME,"[]"))
    args = parser.parse_args()
    
    parameters = parse_parameters(json.loads(args.parameters))

    if parameters.get(DEBUG_PARAMETER):
        parameters.pop(DEBUG_PARAMETER)
        DEBUG = True
    
    interface_file = importlib.import_module(args.interface_name)
    user_class = getattr(interface_file,args.interface_name)

    if args.persistence:
        user_object = persistence.restore(user_class,parameters,debug=DEBUG)
        persistence.persist(user_object,parameters.get("push_frequency"))
    else:
        user_object = user_class(**parameters)

    if args.service_type == "MODEL":
        import model_microservice as seldon_microservice
    elif args.service_type == "ROUTER":
        import router_microservice as seldon_microservice
    elif args.service_type == "TRANSFORMER":
        import transformer_microservice as seldon_microservice
    elif args.service_type == "OUTLIER_DETECTOR":
        import outlier_detector_microservice as seldon_microservice

    port = int(os.environ.get(SERVICE_PORT_ENV_NAME,DEFAULT_PORT))
    
    if args.api_type == "REST":
        app = seldon_microservice.get_rest_microservice(user_object,debug=DEBUG)
        app.run(host='0.0.0.0', port=port)
        
    elif args.api_type=="GRPC":
        server = seldon_microservice.get_grpc_server(user_object,debug=DEBUG)
        server.add_insecure_port("0.0.0.0:{}".format(port))
        server.start()
        
        print("GRPC Microservice Running on port {}".format(port))
        while True:
            time.sleep(1000)
