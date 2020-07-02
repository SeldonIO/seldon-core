
import logging
import random
import os


NUMBER_OF_ROUTES = int(os.environ.get("NUMBER_OF_ROUTES", "2"))


class Node:
    def predict(self, features, names=[], meta=[]):
        logging.info(f"model features: {features}")
        logging.info(f"model names: {names}")
        logging.info(f"model meta: {meta}")
        return features.tolist()

    def transform_input(self, features, names=[], meta=[]):
        return self.predict(features, names, meta)

    def transform_output(self, features, names=[], meta=[]):
        return self.predict(features, names, meta)

    def aggregate(self, features, names=[], meta=[]):
        logging.info(f"model features: {features}")
        logging.info(f"model names: {names}")
        logging.info(f"model meta: {meta}")
        return [x.tolist() for x in features]

    def route(self, features, names=[], meta=[]):
        logging.info(f"model features: {features}")
        logging.info(f"model names: {names}")
        logging.info(f"model meta: {meta}")
        route = random.randint(0, NUMBER_OF_ROUTES)
        logging.info(f"routing to: {route}")
        return route
