#
# Derived from https://github.com/aws/sagemaker-chainer-container/blob/master/src/sagemaker_chainer_container/serving.py
#
from __future__ import print_function,absolute_import

import numpy as np
from sagemaker_containers.beta.framework import (content_types, encoders, env, modules, transformer,worker)
import logging

logging.basicConfig(format='%(asctime)s %(levelname)s - %(name)s - %(message)s', level=logging.INFO)

logging.getLogger('boto3').setLevel(logging.INFO)
logging.getLogger('s3transfer').setLevel(logging.INFO)
logging.getLogger('botocore').setLevel(logging.WARN)

logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)

class SagemakerSeldonError(Exception):
    def __init__(self, message):
        self.message = message

def default_input_fn(input_data, content_type):
    np_array = encoders.decode(input_data, content_type)
    return np_array.astype(np.float32) if content_type in content_types.UTF8_TYPES else np_array

def default_predict_fn(data, model):
    raise SagemakerSeldonError("You must provide a predict_fn")

def default_output_fn(prediction, accept):
    return worker.Response(response=encoders.encode(prediction, accept), mimetype=accept)

def default_model_fn(model_dir):
    return transformer.default_model_fn(model_dir)

def _user_module_transformer(user_module):
    model_fn = getattr(user_module, 'model_fn', default_model_fn)
    input_fn = getattr(user_module, 'input_fn', default_input_fn)
    predict_fn = getattr(user_module, 'predict_fn', default_predict_fn)
    output_fn = getattr(user_module, 'output_fn', default_output_fn)
    return transformer.Transformer(model_fn=model_fn, input_fn=input_fn, predict_fn=predict_fn,output_fn=output_fn)

app = None

def main(environ, start_response):
    global app
    if app is None:
        serving_env = env.ServingEnv()
        user_module = modules.import_module(serving_env.module_dir, serving_env.module_name)
        user_module_transformer = _user_module_transformer(user_module)
        user_module_transformer.initialize()
        app = worker.Worker(transform_fn=user_module_transformer.transform,module_name=serving_env.module_name)
    return app(environ, start_response)
