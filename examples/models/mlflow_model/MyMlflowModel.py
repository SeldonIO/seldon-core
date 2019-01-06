from mlflow import pyfunc
import os
import pandas as pd

class MyMlflowModel(object):

    def __init__(self):
        self.pyfunc_model = pyfunc.load_pyfunc("mlruns/0/"+next(os.walk('mlruns/0'))[1][0]+"/artifacts/model")
        
    def predict(self,X,features_names):
        if not features_names is None and len(features_names)>0:
            df = pd.DataFrame(data=X,columns=features_names)
        else:
            df = pd.DataFrame(data=X)
        return self.pyfunc_model.predict(df)


