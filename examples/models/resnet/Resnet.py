import tensorflow as tf
import numpy as np
import logging
import datetime

def get_logger(name):
    logger = logging.getLogger(name)
    log_formatter = logging.Formatter("%(asctime)s - %(name)s - "
                                      "%(levelname)s - %(message)s")
    logger.setLevel('DEBUG')

    console_handler = logging.StreamHandler()
    console_handler.setFormatter(log_formatter)
    logger.addHandler(console_handler)

    return logger

logger = get_logger(__name__)

class Resnet(object):
    def __init__(self):
        self.class_names = ["class:{}".format(str(i)) for i in range(1000)]
        self.sess = tf.Session()
        tf.saved_model.loader.load(self.sess, ["serve"], "model", clear_devices=True)

        graph = tf.get_default_graph()
        self.x = graph.get_tensor_by_name("input:0")
        self.y = graph.get_tensor_by_name("resnet_v1_50/predictions/Reshape_1:0")

    def predict(self,X,feature_names):
        start_time = datetime.datetime.now()
        predictions = self.sess.run(self.y,feed_dict={self.x:X})
        end_time = datetime.datetime.now()
        duration = (end_time - start_time).total_seconds() * 1000
        logger.debug("Processing time: {:.2f} ms".format(duration))
        return predictions.astype(np.float64)

    
