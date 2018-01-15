from keras.utils import to_categorical
from sklearn.preprocessing import LabelEncoder,MinMaxScaler
import numpy as np

class transformer():
    
    def __init__(self,categorical=True):
        self.le = LabelEncoder()
        self.scaler = MinMaxScaler()
        self.categorical = categorical
        
    def fit(self,X,y=None):
        
        self.le.fit(X[:,0])
        X1 = X[:,1:]
        self.scaler.fit(X1)
        return self
    
    def transform(self,X):
        X0 = self.le.transform(X[:,0])
        if self.categorical:
            X0 = to_categorical(X0, num_classes=5)
        else:
            X0 = X0.reshape(-1,1)
        X1 = X[:,1:]
        X1_scaled = self.scaler.transform(X1)
        self.X = np.hstack((X0,X1_scaled))
        print self.X.shape
        return self.X
    
    def inverse_transform(self,X):
        
        X0 = self.le.inverse_transform(X[:,0].astype(int)).reshape(-1,1)
        X1 = X[:,1:]
        X1 = self.scaler.inverse_transform(X1)
        
        return np.hstack((X0,X1))
