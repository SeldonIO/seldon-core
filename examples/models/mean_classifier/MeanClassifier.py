import numpy as np
import math
import time
import logging

def f(x):
    return 1/(1+math.exp(-x))

class MeanClassifier(object):

    def __init__(self, intValue=0, delaySecs=0):
        self.class_names = ["proba"]
        assert type(intValue) == int, "intValue parameter must be an integer"
        self.int_value = intValue
        logging.info("intValue set to %d",intValue)

        assert type(delaySecs) == int, "delaySecs parameter must be an integer"
        self.delay_secs = delaySecs
        logging.info("Delay secs set to %d",delaySecs)

        logging.info("loading model here")

        X = np.load(open("model.npy",'rb'), encoding='latin1')
        self.threshold_ = X.mean() + self.int_value

    def _meaning(self, x):
        return f(x.mean()-self.threshold_)

    def predict(self, X, feature_names):
        logging.info("Input %s",X)
        X = np.array(X)
        assert len(X.shape) == 2, "Incorrect shape"

        if self.delay_secs > 0:
            logging.info("Delaying %d secs",self.delay_secs)
            time.sleep(self.delay_secs)

        return [[self._meaning(x)] for x in X]

    def health_status(self):
        return {"status":"ok"}

    def metadata(self):
        return {"metadata":{"modelName":"mean_classifier"}}


