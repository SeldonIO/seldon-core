import logging
logger = logging.getLogger(__name__)

class ImageNetCombiner(object):

    def aggregate(self, Xs, features_names):
        print("ImageNet Combiner aggregate called")
        logger.info(Xs)
        return (Xs[0]+Xs[1])/2.0

