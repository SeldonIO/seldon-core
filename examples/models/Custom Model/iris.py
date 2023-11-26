import joblib
import numpy as np

class iris(object):
    def __init__(self):
        self._lr_model = joblib.load('model_hinge.pkl')

    def predict(self, X):
        to_predict_list = np.array(list(map(float, X))).reshape(1, -1)
        predictions = self._lr_model.predict(to_predict_list)
        print(predictions)
        return predictions