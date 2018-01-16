from keras.models import load_model

class MnistClassifier(object):

    def __init__(self):
        self.model = load_model('MnistClassifier.h5')

    def predict(self,X,features_names):
        assert X.shape[0]>=1, 'wrong shape 0'
        if X.shape[0]==784:
            X = X.reshape(1,28,28,1)
        else:
            X = X.reshape(X.shape[0],28,28,1)
        return self.model.predict(X)
