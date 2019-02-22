import logging
import numpy as np
logger = logging.getLogger(__name__)

class ImageNetCombiner(object):

    def aggregate(self, Xs, features_names):
        print("ImageNet Combiner aggregate called")
        logger.debug(Xs)
        return (np.reshape(Xs[0],(1,-1)) + np.reshape(Xs[1], (1,-1)))/2.0

