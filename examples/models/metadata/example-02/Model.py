import logging
import json
import os


class Model:
    def predict(self, features, names=[], meta=[]):
        logging.info(f"model features: {features}")
        logging.info(f"model names: {names}")
        logging.info(f"model meta: {meta}")
        return features.tolist()
