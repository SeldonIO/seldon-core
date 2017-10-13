from flask import Flask, Blueprint, render_template, jsonify, current_app
from flask import request
import argparse
import os
import importlib
import json
import numpy as np
import pandas as pd

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

def extract_input():
    jStr = request.form.get("json")
    is_default = request.form.get("isDefault")
    if jStr:
        data = json.loads(jStr)
    else:
        raise DataContractException("Empty json parameter in request")
    return {
        "is_default":is_default,
        "request":data
        }

def default_data_to_dataframe(data):
    """
    The default format is the following:
    data = {
    'features': ['a','b','c'] # Name of the features, optional
    'values': [[0,1,2],[3,1,2]] # Values, 2 dimensional array. First dimension corresponds to batch elements, there will be one prediction made per element
        # Second dimension corresponds to values for each feature
    """
    if not type(data) == dict:
        raise DataContractException("Data must be a dictionary")
    if data.get('ndarray') is None and data.get('tensor') is None:
        raise DataContractException("Data dictionary has no 'ndarray' or 'tensor' keyword.")
    if not data.get('tensor') is None:
        values = np.array(data["tensor"].get("values"))
        shape = data["tensor"].get('shape')
        if shape is None or len(shape)==0:
            if len(values.shape) == 1:
                values = np.expand_dims(values, axis=0)
        else:
            values = values.reshape(shape)
    else:
        values = np.array(data.get("ndarray"))
    if not len(values.shape) == 2:
        raise DataContractException("Data values must be a 2-dimensional array.")
    if data.get('features') is not None:
        features = data.get('features')
        if len(features) != values.shape[1]:
            raise DataContractException("Length of features vector different from length of values vectors")
        return pd.DataFrame(values,columns=features)
    return pd.DataFrame(values)


@predict_blueprint.route('/pausez',methods=['POST'])
def pause():
    current_app.config["seldon_ready"] = False
    ret = {"ready": current_app.config["seldon_ready"] }
    json_ret = jsonify(ret)
    return json_ret

@predict_blueprint.route('/restartz',methods=['POST'])
def restart():
    current_app.config["seldon_ready"] = True
    ret = {"ready": current_app.config["seldon_ready"] }
    json_ret = jsonify(ret)
    return json_ret

@predict_blueprint.route('/readyz',methods=['GET'])
def ready():
    ret = {"ready": current_app.config["seldon_ready"] }
    response = jsonify(ret)
    if not current_app.config["seldon_ready"]:
        response.status_code = 503
    return response

@predict_blueprint.route('/healthz',methods=['GET'])
def health():
    return "healthy"

def create_response_tensor(preds):
    preds = np.array(preds)
    return {
        'shape':preds.shape,
        'values':preds.ravel().tolist()
    }

def create_response_ndarray(preds):
    preds = np.array(preds)
    return preds.tolist()

def create_response(names,preds,tensor=True):
    response = {'features':names}
    if tensor:
        response['tensor'] = create_response_tensor(preds)
    else:
        response['ndarray'] = create_response_ndarray(preds)
    ret = {'response':response}
    return ret

@predict_blueprint.route('/predict',methods=['GET','POST'])
def do_predict():
    """
    prediction endpoint
    """
    input_ = extract_input()
    model =  current_app.config["seldon_model"]
    if input_.get('request') is None or input_['request']['request'] is None:
        raise DataContractException("Request format invalid")
    if input_['is_default']:
        data = default_data_to_dataframe(input_['request']['request'])
    else:
        data = input_['request']['request']
    print 'DATA'
    print data
    preds = model.predict(data)
    print 'PREDICTIONS'
    print preds
    
    names = model.class_names

    ret = create_response(names,preds,tensor=input_['request']['request'].has_key('tensor'))
    json_ret = jsonify(ret)
    return json_ret
        

def get_rest_microservice(model_class,model_name,parameters = None):
    if parameters is None:
        parameters = {}
    
    app = Flask(__name__)
    
    app.config['seldon_model'] = model_class(**parameters)
    app.config['seldon_model_name'] = model_name
    app.config['seldon_ready'] = True

    app.register_blueprint(predict_blueprint)

    return app

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
    parser.add_argument("--parameters",type=str,default=os.environ.get(PARAMETERS_ENV_NAME,"[]"))
    args = parser.parse_args()
    
    parameters = parse_parameters(json.loads(args.parameters))
    
    model_file = importlib.import_module(args.model_name)
    model_class = getattr(model_file,args.model_name)

    app = get_rest_microservice(model_class,args.model_name,parameters=parameters)
    app.run(host='0.0.0.0', port=os.environ.get(SERVICE_PORT_ENV_NAME,DEFAULT_PORT))


