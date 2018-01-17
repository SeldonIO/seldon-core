from keras.models import load_model

class MnistClassifier(object):

    def __init__(self):
        self.model = load_model('MnistClassifier.h5')

    def predict(self,X,features_names):
        return self.model.predict(X)
