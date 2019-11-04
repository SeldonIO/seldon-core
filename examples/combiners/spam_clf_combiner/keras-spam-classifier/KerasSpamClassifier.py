from pathlib import Path
import numpy as np
import pickle
from sklearn.externals import joblib 
from keras.preprocessing import sequence
from keras.engine.saving import model_from_json

class KerasSpamClassifier():

    def __init__(self, model_path = Path('./model')):
        self._architecture_path = model_path.joinpath('architecture.json')
        self._weights_path = model_path.joinpath('weights.h5')
        tokenizer_path = model_path.joinpath('tokenizer.pkl')

        self.max_len = 150
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
        seq= self.tokenizer.texts_to_sequences([text])
        padded = sequence.pad_sequences(seq, maxlen=self.max_len)
        prob = self.model.predict(padded)

        return np.array([prob[0][0], "spam"])


if __name__== "__main__":

    text = 'please click this link to win the price'
    #vect_path=Path('/Users/sandhya.sandhya/Desktop/data/doc/data'),
    sc = KerasSpamClassifier(model_path=Path('./model'))
    res = sc.predict(text, feature_names=None)
    print(res)