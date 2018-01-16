import numpy as np

class MeanTransformer(object):

    def __init__(self):
        pass

    def transform_input(self, X, feature_names):
        X = np.array(X)
        if X.max() == X.min():
            return np.zeros_like(X)
        return (X-X.min())/(X.max()-X.min())
