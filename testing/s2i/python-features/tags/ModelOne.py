import logging


class ModelOne:
    def predict(self, X, feature_names, meta):
        logging.info(X)
        logging.info(feature_names)
        logging.info(meta)
        return ["model-1"]

    def tags(self):
        return {"model-1": "yes", "common": 1}
