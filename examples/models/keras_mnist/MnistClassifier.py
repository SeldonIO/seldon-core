from keras.models import load_model

class MnistClassifier(object):

    def __init__(self):
        self.model = load_model('MnistClassifier.h5')
        self.model._make_predict_function()

    def predict(self,X,features_names):
        return self.model.predict(X)
