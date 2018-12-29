import grpc
import numpy

import requests
import json
import numpy as np

from sagemaker_containers.beta.framework import (content_types, encoders)

class SagemakerServerError(Exception):

    def __init__(self, message):
        self.message = message

'''
A basic Sagemaker serving proxy
'''
class SagemakerProxy(object):

    def __init__(self,endpoint=None):
        print("endpoint:",endpoint)
        self.endpoint=endpoint
        
    def predict(self,X,features_names):
        print("predict")
        r = requests.post(
            self.endpoint+"/invocations",
            json = X.tolist())
        if r.status_code == 200:
            result = encoders.decode(r.content,r.headers.get('content-type'))
            if len(result.shape) == 1:
                result = numpy.expand_dims(result, axis=0)
                return result
        else:
            print("Error from server:",r)
            raise SagemakerServerError(r.json())

    
