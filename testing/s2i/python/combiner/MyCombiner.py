import logging

logger = logging.getLogger(__name__)


class MyCombiner(object):
    def __init__(self, metrics_ok=True):
        print("Combiner Init called")

    def aggregate(self, Xs, features_names):
        print("Combiner aggregate called")
        logger.info(Xs)
        return Xs[0] + 1
