from flask import Blueprint, render_template, jsonify, current_app
from flask import request
import json
import numpy as np
import pandas as pd

predict_blueprint = Blueprint('predict',__name__)

class DataContractException(Exception):
    status_code = 400

    def __init__(self, message, status_code=None, payload=None):
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
    jStr = request.args.get("json")
    is_default = request.args.get("isDefault")
    if jStr:
        data = json.loads(jStr)
    else:
        raise DataContractException("Empty json parameter in request")
    return {
        'is_default':is_default,
        'request':data
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
    if data.get('values') is None and data.get('ndarray') is None:
        raise DataContractException("Data dictionary has no 'values' or 'ndarray' keyword.")
    if not data.get('values') is None:
        values = np.array(data.get("values"))
        shape = data.get('shape')
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

def create_response(names,preds):
    preds = np.array(preds)
    ret = {'response':
           {
               'features':names,
               'ndarray':preds.tolist()
           }
        }
    return ret

@predict_blueprint.route('/predict',methods=['GET','POST'])
def do_predict():
    """
    prediction endpoint
    """
    input_ = extract_input()
    pipeline =  current_app.config["seldon_pipeline"]
    if input_.get('request') is None or input_['request']['request'] is None:
        raise DataContractException("Request format invalid")
    if input_['is_default']:
        data = default_data_to_dataframe(input_['request']['request'])
    else:
        data = input_['request']['request']
    print 'DATA'
    print data
    preds = pipeline.predict_proba(data)
    print 'PREDICTIONS'
    print preds
    if hasattr(pipeline._final_estimator,"get_class_id_map"):
        names = pipeline._final_estimator.get_class_id_map()
    else:
        print "Final estimator has no class names defined. Consider implementing the get_class_id_map method."
        names = [str(i) for i in xrange(len(preds[0]))]

    ret = create_response(names,preds)
    json_ret = jsonify(ret)
    return json_ret
        
