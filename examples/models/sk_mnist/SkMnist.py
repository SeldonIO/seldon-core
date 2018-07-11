from sklearn.externals import joblib

class SkMnist(object):
    def __init__(self):
        self.class_names = ["class:{}".format(str(i)) for i in range(10)]
        self.clf = joblib.load('sk.pkl') 

    def predict(self,X,feature_names):
        predictions = self.clf.predict_proba(X)
        return predictions

    
