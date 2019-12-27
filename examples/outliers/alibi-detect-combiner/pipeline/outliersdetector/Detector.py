import dill
import json
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
        od_preds = self.od.predict(X_prep, return_instance_score=True)
        return json.dumps(od_preds, cls=NumpyEncoder)


class NumpyEncoder(json.JSONEncoder):
    def default(self, obj):
        if isinstance(obj, (
            np.int_, np.intc, np.intp, np.int8, np.int16, np.int32, np.int64,
            np.uint8, np.uint16, np.uint32, np.uint64)
        ):
            return int(obj)
        elif isinstance(obj, (np.float_, np.float16, np.float32, np.float64)):
            return float(obj)
        elif isinstance(obj, (np.ndarray,)):
            return obj.tolist()
        return json.JSONEncoder.default(self, obj)
