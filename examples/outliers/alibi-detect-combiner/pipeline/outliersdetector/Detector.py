import dill
import os

import numpy as np


dirname = os.path.dirname(__file__)


class Detector:
    def __init__(self, *args, **kwargs):

        with open(os.path.join(dirname, "preprocessor.dill"), "rb") as prep_f:
            self.preprocessor = dill.load(prep_f)
        with open(os.path.join(dirname, "model.dill"), "rb") as model_f:
            self.od = dill.load(model_f)

    def predict(self, X, feature_names=[]):
        X_prep = self.preprocessor.transform(X)
        od_preds = self.od.predict(X_prep)
        return od_preds['data']['is_outlier']
