from pathlib import Path
import numpy as np
import pickle
from sklearn.externals import joblib 


class SpamClassifier():

    def __init__(self, model_path = Path('./model')):

        self.models_path = model_path
        self.clf = joblib.load(model_path.joinpath('model.pkl'))
        self.vectorizer = joblib.load(model_path.joinpath('vectorizer.pkl'))


    def predict(self, text, feature_names): #List[Tuple[float, float]]:
        """
        Predict on a english text you got from translator service. The output returns the probability of text being spam
        """

        clean_text  = self.preprocess(text)
        data = self.vectorizer.transform([clean_text]).todense()
        probas = self.clf.predict_proba(data)
        prob = probas[0][1]

        return np.array([prob, "spam"])


