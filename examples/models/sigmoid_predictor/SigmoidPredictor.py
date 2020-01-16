from __future__ import print_function
import numpy as np
from sklearn.neural_network import MLPClassifier

def sigmoid(x):
    return 1 / (1 + np.exp(-x))

class SigmoidPredictor():
    
    def __init__(self):
        
        nb_samples = 10000
        X = np.random.normal(size=(nb_samples,10))
        y = (sigmoid(X[:,0]*X[:,1])>=0.5).astype(int)
        #y = (sigmoid(X[:,0]*X[:,1]+(X[:,2]*X[:,3]))>=0.5).astype(int)
    
        self.ffnn = MLPClassifier()
        self.ffnn.fit(X,y)
        print("Class", self, "variables", dir(self))
    def predict(self,X,features_names):
        return self.ffnn.predict_proba(X)
