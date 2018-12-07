from sklearn.externals import joblib
import logging

__version__ = "v1.0"
logger = logging.getLogger(__name__)


class RFModel(object):

    def __init__(self):
        logger.info('Starting %s Microservice, version %s',
                    __name__, __version__)
        self.model = joblib.load('RFModel.sav')
        self.cm = {'tp': 0, 'fp': 0, 'tn': 0, 'fn': 0}

    def predict(self, X, features_names):
        return self.model.predict_proba(X)

    def send_feedback(self, features, feature_names, reward, truth):
        logger.debug('RF model send-feedback entered')
        logger.debug(f"Truth: {truth}, Reward: {reward}")

        if reward == 1:
            if truth == 1:
                self.cm['tp'] += 1
            if truth == 0:
                self.cm['tn'] += 1
        if reward == 0:
            if truth == 1:
                self.cm['fn'] += 1
            if truth == 0:
                self.cm['fp'] += 1

        logger.debug(self.cm)

    def metrics(self):
        tp = {"type": "GAUGE", "key": "true_pos_total",
              "value": self.cm['tp'], "tags": {"branch_name": "rf"}}
        tn = {"type": "GAUGE", "key": "true_neg_total",
              "value": self.cm['tn'], "tags": {"branch_name": "rf"}}
        fp = {"type": "GAUGE", "key": "false_pos_total",
              "value": self.cm['fp'], "tags": {"branch_name": "rf"}}
        fn = {"type": "GAUGE", "key": "false_neg_total",
              "value": self.cm['fn'], "tags": {"branch_name": "rf"}}

        return [tp, tn, fp, fn]
