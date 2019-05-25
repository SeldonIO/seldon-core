
import dill
import logging

class Transformer(object):
    def __init__(self):

        with open('/mnt/lr.model', 'rb') as model_file:
            self._lr_model = dill.load(model_file)

    def predict(self, X, feature_names):
        logging.warning(X)
        prediction = self._lr_model.predict_proba(X)
        logging.warning(prediction)
        return prediction


