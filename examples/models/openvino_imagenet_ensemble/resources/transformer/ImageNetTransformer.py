import numpy as np
from keras.applications.imagenet_utils import preprocess_input, decode_predictions
from keras.preprocessing import image
from seldon_core.proto import prediction_pb2
import tensorflow as tf
import logging

logger = logging.getLogger(__name__)

class ImageNetTransformer(object):
    def __init__(self, metrics_ok=True):
        print("Init called")
        f = open('imagenet_classes.json')
        self.cnames = eval(f.read())
        
    def transform_input_grpc(self, request):
        logger.info("Transform called")
        X = tf.make_ndarray(request.data.tftensor)
        X = np.expand_dims(X, axis=0)
        X = preprocess_input(X)
        X = X.transpose((0,3,1,2))
        datadef = prediction_pb2.DefaultData(
            names = 'x',
            tftensor = tf.make_tensor_proto(X)
        )
        request = prediction_pb2.SeldonMessage(data = datadef)
        return request

    def transform_output_grpc(self, request):
        print("Transform output called")
        result = tf.make_ndarray(request.data.tftensor)
        result = result.reshape(1,1000)

        single_result = result[[0],...]
        ma = np.argmax(single_result)
        name = self.cnames[ma]

        response = prediction_pb2.SeldonMessage(strData = name)
        
        return response
