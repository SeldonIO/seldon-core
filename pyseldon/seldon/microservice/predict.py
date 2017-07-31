from flask import Blueprint, render_template, jsonify, current_app
from flask import request
import json
import numpy as np
import pandas as pd

predict_blueprint = Blueprint('predict',__name__)

class DataContractException(Exception):
    pass

def extract_input():
    jStr = request.args.get("json")
    is_default = request.args.get("isDefault")
    data = json.loads(jStr)
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
    if data.get('values') is None:
        raise DataContractException("Data dictionary has no 'values' keyword.")
    values = np.array(data.get("values"))
    shape = data.get('shape')
    if shape is None or len(shape)==0:
        if len(values.shape) == 1:
            values = np.expand_dims(values, axis=0)
    else:
        values = values.reshape(shape)
    if not len(values.shape) == 2:
        raise DataContractException("Data values must be a 2-dimensional array.")
    if data.get('keys') is not None:
        features = data.get('keys')
        if len(features) != values.shape[1]:
            raise DataContractException("Length of features vector different from length of values vectors")
        return pd.DataFrame(values,columns=features)
    return pd.DataFrame(values)

@predict_blueprint.route('/ping',methods=['GET'])
def ping():
    return "pong"

def create_response(names,preds):
    preds = np.array(preds)
    ret = {'response':
           {
               'keys':names,
               'shape':preds.shape,
               'values':preds.flatten().tolist()
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
            
        
