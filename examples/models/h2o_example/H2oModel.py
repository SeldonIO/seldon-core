import numpy as np
import pandas as pd
import h2o
from h2o.frame import H2OFrame

MODEL_PATH='/microservice/glm_fit1'

def _to_frame(X,features_names):
    """Create H2OFrame object from lists
    """
    return H2OFrame(X,column_names=features_names)

def _from_frame(frame):
    """Create numpy array with probabilities from H2OFrame object
    """
    preds = h2o.as_list(frame,use_pandas=False); preds.pop(0); [r.pop(0) for r in preds]
    return np.asarray(preds,dtype=np.float)

class H2oModel():
    
    def __init__(self):
        
        print 'Starting Java virtual machine'
        h2o.init(nthreads = -1, max_mem_size = 8)
        print 'Machine started!'

        print 'Loading model from %s...' % MODEL_PATH
        self.model = h2o.load_model(MODEL_PATH)
        print 'Model Loaded'
            
    def predict(self,X,features_names):
        return _from_frame(self.model.predict(_to_frame(X,features_names)))



