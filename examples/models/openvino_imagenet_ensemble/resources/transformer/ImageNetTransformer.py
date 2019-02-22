import numpy as np
from seldon_core.proto import prediction_pb2
import tensorflow as tf
import logging
import datetime
import cv2
import os

logger = logging.getLogger(__name__)

class ImageNetTransformer(object):
    def __init__(self, metrics_ok=True):
        print("Init called")
        f = open('imagenet_classes.json')
        self.cnames = eval(f.read())
        self.size = os.getenv('SIZE', 224)
        self.dtype = os.getenv('DTYPE', 'float')
        self.classes = os.getenv('CLASSES', 1000)

    def crop_resize(self, img,cropx,cropy):
        y,x,c = img.shape
        if y < cropy:
            img = cv2.resize(img, (x, cropy))
            y = cropy
        if x < cropx:
            img = cv2.resize(img, (cropx,y))
            x = cropx
        startx = x//2-(cropx//2)
        starty = y//2-(cropy//2)
        return img[starty:starty+cropy,startx:startx+cropx,:]

    def transform_input_grpc(self, request):
        logger.info("Transform called")
        start_time = datetime.datetime.now()
        X = np.frombuffer(request.binData, dtype=np.uint8)
        X = cv2.imdecode(X, cv2.IMREAD_COLOR)  # BGR format
        X = self.crop_resize(X, self.size, self.size)
        X = X.astype(self.dtype)
        X = X.transpose(2,0,1).reshape(1,3,self.size,self.size)
        logger.info("Shape: %s; Dtype: %s; Min: %s; Max: %s",X.shape,X.dtype,np.amin(X),np.amax(X))
        jpeg_time = datetime.datetime.now()
        jpeg_duration = (jpeg_time - start_time).total_seconds() * 1000
        logger.info("jpeg preprocessing: %s ms", jpeg_duration)
        datadef = prediction_pb2.DefaultData(
            names = 'x',
            tftensor = tf.make_tensor_proto(X)
        )
        end_time = datetime.datetime.now()
        duration = (end_time - start_time).total_seconds() * 1000
        logger.info("Total transformation: %s ms", duration)
        request = prediction_pb2.SeldonMessage(data = datadef)
        return request

    def transform_output_grpc(self, request):
        logger.info("Transform output called")
        result = tf.make_ndarray(request.data.tftensor)
        result = result.reshape(1,self.classes)

        single_result = result[[0],...]
        ma = np.argmax(single_result)
        name = self.cnames[ma]

        response = prediction_pb2.SeldonMessage(strData = name)
        
        return response
