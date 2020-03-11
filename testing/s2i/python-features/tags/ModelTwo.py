import logging


class ModelTwo:
    def predict(self, X, feature_names, meta):
        logging.info(X)
        logging.info(feature_names)
        logging.info(meta)
        return ["model-2"]

    def tags(self):
        return {"model-2": "yes", "common": 2}
