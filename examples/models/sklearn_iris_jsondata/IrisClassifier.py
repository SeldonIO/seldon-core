import joblib
import sys


def eprint(*args, **kwargs):
    print(*args, file=sys.stderr, **kwargs)


class IrisClassifier(object):
    def __init__(self):
        self.model = joblib.load("IrisClassifier.sav")

    def predict(self, X, features_names):
        eprint("--------------------")
        eprint("Input dict")
        eprint(X)
        eprint("--------------------")
        ndarray = X["some_data"]["some_ndarray"]
        return self.model.predict_proba(ndarray)
