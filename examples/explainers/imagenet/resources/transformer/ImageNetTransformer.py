import logging
from tensorflow.keras.applications.inception_v3 import preprocess_input
import numpy as np
logger = logging.getLogger(__name__)

class ImageNetTransformer(object):
    def __init__(self, metrics_ok=True):
        print("Init called")

    def transform_input(self, X, names, meta):
        logger.info("Transform called")
        t = preprocess_input(X)
        logger.info("Data type is %s", t.dtype)
        t = np.float32(t)
        logger.info("Data type is %s", t.dtype)
        return t


