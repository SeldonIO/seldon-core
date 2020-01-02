import logging
import numpy as np

class Combiner(object):

    def aggregate(self, X, features_names=[]):
        logging.info("Input: " + str(X))
        output = {
            "loanclassifier": X[0].tolist(),
            "outliersdetector": X[1].tolist(),
        }
        logging.info("Output: " + str(output))
        return output
