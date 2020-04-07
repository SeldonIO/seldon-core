import logging
import os


class Model:

    def predict(self, features, names=[], meta=[]):
        logging.info(f"My id is {id(self)}")
        logging.info(f"OS pid is {os.getpid()}")
        logging.info(f"model features: {features}")
        logging.info(f"model names: {names}")
        logging.info(f"model meta: {meta}")
        return features.tolist()

    def tags(self):
        return {"a": 1, "b": 2}

    def metrics(self):
        return [
            {"type": "COUNTER", "key": "mycounter", "value": 1, "tags": {"mytag": "mytagvalue"}},
            {"type": "GAUGE", "key": "mygauge", "value": 100},
            {"type": "TIMER", "key": "mytimer", "value": 20.2},
        ]
