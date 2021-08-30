import numpy as np
import logging
from tensorflow.keras.models import load_model
import tensorflow as tf

def get_logger(name):
    logger = logging.getLogger(name)
    log_formatter = logging.Formatter("%(asctime)s - %(name)s - "
                                      "%(levelname)s - %(message)s")
    logger.setLevel('DEBUG')

    console_handler = logging.StreamHandler()
    console_handler.setFormatter(log_formatter)
    logger.addHandler(console_handler)

    return logger

logger = get_logger(__name__)

class Cifar10(object):
    def load(self):
        tf.config.threading.set_inter_op_parallelism_threads(1)
        tf.config.threading.set_intra_op_parallelism_threads(1)
        self.model = load_model('model.h5')

    def predict(self,X,feature_names):
        yhat = self.model.predict(X)
        return yhat


def main():
    c = Cifar10()
    c.load()
    data = np.random.randn(1,32,32,3)
    res = c.predict(data, None)
    print(res)

if __name__ == "__main__":
    main()

