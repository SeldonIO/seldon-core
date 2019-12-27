import dill
import os


dirname = os.path.dirname(__file__)


class Model:

    def __init__(self, *args, **kwargs):
        """Deserilize preprocessor and model."""
        with open(os.path.join(dirname, "preprocessor.dill"), "rb") as prep_f:
            self.preprocessor = dill.load(prep_f)
        with open(os.path.join(dirname, "model.dill"), "rb") as model_f:
            self.clf = dill.load(model_f)

    def predict(self, X, feature_names=[]):
        """Run input X through loanclassifier model."""
        X_prep = self.preprocessor.transform(X)
        return self.clf.predict_proba(X_prep)
