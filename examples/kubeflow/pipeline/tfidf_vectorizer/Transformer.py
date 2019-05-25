
import dill

class Transformer(object):
    def __init__(self):

        with open('/mnt/tfidf.model', 'rb') as model_file:
            self._tfidf_vectorizer = dill.load(model_file)

    def predict(self, X, feature_names):
        tfidf_features = self._tfidf_vectorizer.transform(X)
        return tfidf_features 


