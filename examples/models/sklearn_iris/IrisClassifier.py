import joblib

class IrisClassifier(object):

    def __init__(self):
        self.model = joblib.load('IrisClassifier.sav')

    def predict(self,X,features_names):
        return self.model.predict_proba(X)
