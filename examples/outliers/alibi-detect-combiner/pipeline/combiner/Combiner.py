import logging
import json
import numpy as np

class Combiner(object):

    def aggregate(self, X, features_names=[]):
        logging.warning(X)
        output = {
            "loanclassifier": X[0],
            "outliersdetector": json.loads(X[1]),
        }
        return json.dumps(output, cls=NumpyEncoder)


class NumpyEncoder(json.JSONEncoder):
    def default(self, obj):
        if isinstance(obj, (
        np.int_, np.intc, np.intp, np.int8, np.int16, np.int32, np.int64, np.uint8, np.uint16, np.uint32, np.uint64)):
            return int(obj)
        elif isinstance(obj, (np.float_, np.float16, np.float32, np.float64)):
            return float(obj)
        elif isinstance(obj, (np.ndarray,)):
            return obj.tolist()
        return json.JSONEncoder.default(self, obj)
