from pathlib import Path
import numpy as np
import pickle
from sklearn.externals import joblib 


class KerasSpamClassifier():

    def __init__(self, parent_path = Path('/model')):
        self._architecture_path = model_path.joinpath('architecture.json')
        self._weights_path = model_path.joinpath('weights.h5')
        tokenizer_path = model_path.joinpath('tokenizer.pkl')

        # loading of trained tokenizer
        with tokenizer_path.open('rb') as handle:
            self.tokenizer = joblib.load(handle)

        with self._architecture_path.open() as f:
            self.model = model_from_json(f.read())
            self.model.load_weights(self._weights_path.as_posix())
            self.model._make_predict_function()



    def predict(self, text, feature_names): #List[Tuple[float, float]]:
        """
        Predict on a english text you got from translator service. The output returns the probability of text being spam
        """

        probas = self.model.predict([x, np.array(x_char)])[0]

        prob = probas[0][1]

        return np.array([prob, "spam"])


