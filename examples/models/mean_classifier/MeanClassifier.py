import numpy as np
import math

def f(x):
    return 1/(1+math.exp(-x))

class MeanClassifier(object):

    def __init__(self, intValue=0):
        self.class_names = ["proba"]
        assert type(intValue) == int, "intValue parameters must be an integer"
        self.int_value = intValue

        print("Loading model here")

        X = np.load(open("model.npy",'rb'), encoding='latin1')
        self.threshold_ = X.mean() + self.int_value

    def _meaning(self, x):
        return f(x.mean()-self.threshold_)

    def predict(self, X, feature_names):
        print(X)
        X = np.array(X)
        assert len(X.shape) == 2, "Incorrect shape"

        return [[self._meaning(x)] for x in X]

    def health_status(self):
        return {"status":"ok"}

    def metadata(self):
        return {"metadata":{"modelName":"mean_classifier"}}


