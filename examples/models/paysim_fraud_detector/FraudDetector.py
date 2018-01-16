import pandas as pd
import numpy as np
from sklearn.externals import joblib
from transformer import transformer

class FraudDetector():
    
    def __init__(self):

        self.p = joblib.load('model_pipeline.sav')
    
    def predict(self,X,features_names):
        print X
        return self.p.predict_proba(X)
