
import dill
import logging

class Transformer(object):
    def __init__(self):

        with open('/mnt/tfidf.model', 'rb') as model_file:
            self._tfidf_vectorizer = dill.load(model_file)

    def predict(self, X, feature_names):
        logging.warning(X)
        tfidf_sparse = self._tfidf_vectorizer.transform(X)
        tfidf_array = tfidf_sparse.toarray()
        logging.warning(tfidf_array)
        return tfidf_array


