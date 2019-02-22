import logging
import numpy as np
import pickle
from sklearn.ensemble import IsolationForest

logger = logging.getLogger(__name__)

class CoreIsolationForest(object):
    """ Outlier detection using Isolation Forests.
    
    Parameters
    ----------
        threshold (float) : anomaly score threshold; scores below threshold are outliers
     
    Functions
    ----------
        predict : detect and return outliers
        transform_input : detect outliers and return input features
        send_feedback : add target labels as part of the feedback loop
        tags : add metadata for input transformer
        metrics : return custom metrics
    """
    
    def __init__(self,threshold=0.,model_name='if',load_path='./models/'):
        
        logger.info("Initializing model")
        self.threshold = threshold
        self.N = 0 # total sample count up until now
        self.nb_outliers = 0
        
        # load pre-trained model
        with open(load_path + model_name + '.pickle', 'rb') as f:
            self.clf = pickle.load(f)
            
    
    def predict(self, X, feature_names):
        """ Return outlier predictions.

        Parameters
        ----------
        X : array-like
        feature_names : array of feature names (optional)
        """
        logger.info("Using component as a model")
        return self._get_preds(X)
    
    
    def transform_input(self, X, feature_names):
        """ Transform the input. 
        Used when the outlier detector sits on top of another model.

        Parameters
        ----------
        X : array-like
        feature_names : array of feature names (optional)
        """
        logger.info("Using component as an outlier-detector transformer")
        self.prediction_meta = self._get_preds(X)
        return X
    
    
    def _get_preds(self,X):
        """ Detect outliers below the anomaly score threshold.  
        
        Parameters
        ----------
        X : array-like
        """
        self.decision_val = self.clf.decision_function(X) # anomaly scores
        
        # make prediction
        self.prediction = (self.decision_val < self.threshold).astype(int) # scores below threshold are outliers
        
        self.N+=self.prediction.shape[0] # update counter
        
        return self.prediction
    
    
    def send_feedback(self,X,feature_names,reward,truth):
        """ Return additional data as part of the feedback loop.
        
        Parameters
        ----------
            X : array of the features sent in the original predict request
            feature_names : array of feature names. May be None if not available.
            reward (float): the reward
            truth : array with correct value (optional)
        """
        logger.info("Send feedback called")
        return []
    
    
    def tags(self):
        """
        Use predictions made within transform to add these as metadata
        to the response. Tags will only be collected if the component is
        used as an input-transformer.
        """
        try:
            return {"outlier-predictions": self.prediction_meta.tolist()}
        except AttributeError:
            logger.info("No metadata about outliers")
            
    
    def metrics(self):
        """ Return custom metrics averaged over the prediction batch.
        """
        self.nb_outliers += np.sum(self.prediction)
        
        is_outlier = {"type":"GAUGE","key":"is_outlier","value":np.mean(self.prediction)}
        anomaly_score = {"type":"GAUGE","key":"anomaly_score","value":np.mean(self.decision_val)}
        nb_outliers = {"type":"GAUGE","key":"nb_outliers","value":int(self.nb_outliers)}
        fraction_outliers = {"type":"GAUGE","key":"fraction_outliers","value":int(self.nb_outliers)/self.N}
        obs = {"type":"GAUGE","key":"observation","value":self.N}
        threshold = {"type":"GAUGE","key":"threshold","value":self.threshold}

        return [is_outlier,anomaly_score,nb_outliers,fraction_outliers,obs,threshold]