import dill
import logging


class Model:
    def __init__(self, *args, **kwargs):

        with open("preprocessor.dill", "rb") as prep_f:
            self.preprocessor = dill.load(prep_f)
        with open("model.dill", "rb") as model_f:
            self.clf = dill.load(model_f)

    def predict(self, X, feature_names=[]):
        logging.warn("Received: " + str(X))
        X_prep = self.preprocessor.transform(X)
        proba = self.clf.predict_proba(X_prep)
        logging.warn("Predicted: " + str(proba))
        return proba
