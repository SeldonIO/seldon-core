import logging

import numpy as np


class Combiner(object):
    def aggregate(self, X, features_names=[]):
        logging.info(X)
        return np.array(X).tolist()

    def tags(self):
        return {"combiner": "yes"}
