import threading
import os
import time
import logging
try:
    # python 2
    import cPickle as pickle
except ImportError:
    # python 3
    import pickle
import redis

logger = logging.getLogger(__name__)

PRED_UNIT_ID = os.environ.get("PREDICTIVE_UNIT_ID", "0")
PREDICTOR_ID = os.environ.get("PREDICTOR_ID", "0")
DEPLOYMENT_ID = os.environ.get("SELDON_DEPLOYMENT_ID", "0")
REDIS_KEY = "persistence_{}_{}_{}".format(
    DEPLOYMENT_ID, PREDICTOR_ID, PRED_UNIT_ID)

REDIS_HOST = os.environ.get('REDIS_SERVICE_HOST', 'localhost')
REDIS_PORT = os.environ.get("REDIS_SERVICE_PORT", 6379)
DEFAULT_PUSH_FREQUENCY = 60


def restore(user_class, parameters, debug=False):
    logger.info("Restoring saved model from redis")

    redis_client = redis.StrictRedis(host=REDIS_HOST, port=REDIS_PORT)
    saved_state_binary = redis_client.get(REDIS_KEY)
    if saved_state_binary is None:
        logger.info("Saved state is empty, restoration aborted")
        return user_class(**parameters)
    else:
        return pickle.loads(saved_state_binary)


def persist(user_object, push_frequency=None, debug=False):

    if push_frequency is None:
        push_frequency = DEFAULT_PUSH_FREQUENCY
    logger.info("Creating persistence thread, with frequency %s", push_frequency)
    persistence_thread = PersistenceThread(user_object, push_frequency)
    persistence_thread.start()


class PersistenceThread(threading.Thread):

    def __init__(self, user_object, push_frequency):
        self.user_object = user_object
        self.push_frequency = push_frequency
        self._stopped = False
        self.redis_client = redis.StrictRedis(host=REDIS_HOST, port=REDIS_PORT)
        super(PersistenceThread, self).__init__()

    def stop(self):
        logger.info("Stopping Persistence Thread")
        self._stopped = True

    def run(self):
        while not self._stopped:
            time.sleep(self.push_frequency)
            binary_data = pickle.dumps(self.user_object)
            self.redis_client.set(REDIS_KEY, binary_data)
