import logging
logger = logging.getLogger(__name__)

class MnistCombiner(object):
    def __init__(self, metrics_ok=True):
        print("MNIST Combiner Init called")

    def aggregate(self, Xs, features_names):
        print("MNIST Combiner aggregate called")
        logger.info(Xs)
        return (Xs[0]+Xs[1])/2.0

