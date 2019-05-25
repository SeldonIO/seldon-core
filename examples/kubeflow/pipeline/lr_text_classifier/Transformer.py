
import dill

class Transformer(object):
    def __init__(self):

        with open('/mnt/lr_text.model', 'rb') as model_file:
            self._lr_model = dill.load(model_file)

    def predict(self, X, feature_names):
        prediction = self._lr_model.predict_proba(X)
        return prediction


