import numpy as np


class OutputTransformer(object):

    def __init__(self):
        print("init")

    def transform_output(self,X,names):
    	a = np.array([1])
    	return np.vstack((X, a))
