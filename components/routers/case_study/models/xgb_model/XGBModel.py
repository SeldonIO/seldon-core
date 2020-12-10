import joblib
import logging

__version__ = "0.1"
logger = logging.getLogger(__name__)


class XGBModel(object):
    def __init__(self):
        logger.info("Starting %s Microservice, version %s", __name__, __version__)
        self.model = joblib.load("XGBModel.sav")
        self.cm = {"tp": 0, "fp": 0, "tn": 0, "fn": 0}

        self.tries = 0
        self.success = 0
        self.value = 0

    def predict(self, X, features_names):
        return self.model.predict_proba(X)

    def send_feedback(self, features, feature_names, reward, truth, routing=None):
        logger.debug("RF model send-feedback entered")
        logger.debug(f"Truth: {truth}, Reward: {reward}")

        if reward == 1:
            if truth == 1:
                self.cm["tp"] += 1
            if truth == 0:
                self.cm["tn"] += 1
        if reward == 0:
            if truth == 1:
                self.cm["fn"] += 1
            if truth == 0:
                self.cm["fp"] += 1

        self.tries += 1
        self.success = self.success + 1 if reward else self.success
        self.value = self.success / self.tries

        logger.debug(self.cm)
        logger.debug(
            "Tries: %s, successes: %s, values: %s", self.tries, self.success, self.value
        )

    def metrics(self):
        tp = {
            "type": "GAUGE",
            "key": "true_pos_total",
            "value": self.cm["tp"],
            "tags": {"branch_name": "xgb"},
        }
        tn = {
            "type": "GAUGE",
            "key": "true_neg_total",
            "value": self.cm["tn"],
            "tags": {"branch_name": "xgb"},
        }
        fp = {
            "type": "GAUGE",
            "key": "false_pos_total",
            "value": self.cm["fp"],
            "tags": {"branch_name": "xgb"},
        }
        fn = {
            "type": "GAUGE",
            "key": "false_neg_total",
            "value": self.cm["fn"],
            "tags": {"branch_name": "xgb"},
        }

        value = {
            "type": "GAUGE",
            "key": "branch_value",
            "value": self.value,
            "tags": {"branch_name": "xgb"},
        }
        success = {
            "type": "GAUGE",
            "key": "n_success_total",
            "value": self.success,
            "tags": {"branch_name": "xgb"},
        }
        tries = {
            "type": "GAUGE",
            "key": "n_tries_total",
            "value": self.tries,
            "tags": {"branch_name": "xgb"},
        }

        return [tp, tn, fp, fn, value, success, tries]
