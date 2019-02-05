import numpy as np

from CoreIsolationForest import CoreIsolationForest
from utils import flatten, performance, outlier_stats

class OutlierIsolationForest(CoreIsolationForest):
    """ Outlier detection using Isolation Forests.
    
    Parameters
    ----------
        threshold (float) : anomaly score threshold; scores below threshold are outliers
     
    Functions
    ----------
        send_feedback : add target labels as part of the feedback loop
        metrics : return custom metrics
    """
    def __init__(self,threshold=0.,model_name='if',load_path='./models/'):
        
        super().__init__(threshold=threshold, model_name=model_name, load_path=load_path)
        
        self._predictions = []
        self._labels = []
        self._anomaly_score = []
        self.roll_window = 100
        self.metric = [float('nan') for i in range(18)]
    
    
    def send_feedback(self,X,feature_names,reward,truth):
        """ Return outlier labels as part of the feedback loop.
        
        Parameters
        ----------
            X : array of the features sent in the original predict request
            feature_names : array of feature names. May be None if not available.
            reward (float): the reward
            truth : array with correct value (optional)
        """
        _ = super().send_feedback(X,feature_names,reward,truth)
        
        # historical reconstruction errors and predictions
        self._anomaly_score.append(self.decision_val)
        self._anomaly_score = flatten(self._anomaly_score)
        self._predictions.append(self.prediction)
        self._predictions = flatten(self._predictions)
        
        # target labels
        self.label = truth
        self._labels.append(self.label)
        self._labels = flatten(self._labels)
        
        # performance metrics
        scores = performance(self._labels,self._predictions,roll_window=self.roll_window)
        stats = outlier_stats(self._labels,self._predictions,roll_window=self.roll_window)
        
        convert = flatten([scores,stats])
        metric = []
        for c in convert: # convert from np to native python type to jsonify
            metric.append(np.asscalar(np.asarray(c)))
        self.metric = metric
        
        return []
    
    
    def metrics(self):
        """ Return custom metrics.
        Printed with a delay of 1 prediction because the labels are returned in the feedback step. 
        """
        
        if self.prediction.shape[0]>1:
            raise ValueError('Metrics can only handle single observations.')
        
        if self.N==1:
            pred = float('nan')
            dec_val = float('nan')
            y_true = float('nan')
        else:
            pred = int(self._predictions[-1])
            dec_val = self._anomaly_score[-1]
            y_true = int(self.label[0])
                         
        is_outlier = {"type":"GAUGE","key":"is_outlier","value":pred}
        anomaly_score = {"type":"GAUGE","key":"anomaly_score","value":dec_val}
        obs = {"type":"GAUGE","key":"observation","value":self.N - 1}
        threshold = {"type":"GAUGE","key":"threshold","value":self.threshold}
        
        label = {"type":"GAUGE","key":"label","value":y_true}
        
        accuracy_tot = {"type":"GAUGE","key":"accuracy_tot","value":self.metric[4]}
        precision_tot = {"type":"GAUGE","key":"precision_tot","value":self.metric[5]}
        recall_tot = {"type":"GAUGE","key":"recall_tot","value":self.metric[6]}
        f1_score_tot = {"type":"GAUGE","key":"f1_tot","value":self.metric[7]}
        f2_score_tot = {"type":"GAUGE","key":"f2_tot","value":self.metric[8]}
        
        accuracy_roll = {"type":"GAUGE","key":"accuracy_roll","value":self.metric[9]}
        precision_roll = {"type":"GAUGE","key":"precision_roll","value":self.metric[10]}
        recall_roll = {"type":"GAUGE","key":"recall_roll","value":self.metric[11]}
        f1_score_roll = {"type":"GAUGE","key":"f1_roll","value":self.metric[12]}
        f2_score_roll = {"type":"GAUGE","key":"f2_roll","value":self.metric[13]}
        
        true_negative = {"type":"GAUGE","key":"true_negative","value":self.metric[0]}
        false_positive = {"type":"GAUGE","key":"false_positive","value":self.metric[1]}
        false_negative = {"type":"GAUGE","key":"false_negative","value":self.metric[2]}
        true_positive = {"type":"GAUGE","key":"true_positive","value":self.metric[3]}
        
        nb_outliers_roll = {"type":"GAUGE","key":"nb_outliers_roll","value":self.metric[14]}
        nb_labels_roll = {"type":"GAUGE","key":"nb_labels_roll","value":self.metric[15]}
        nb_outliers_tot = {"type":"GAUGE","key":"nb_outliers_tot","value":self.metric[16]}
        nb_labels_tot = {"type":"GAUGE","key":"nb_labels_tot","value":self.metric[17]}
        
        return [is_outlier,anomaly_score,obs,threshold,label,
                accuracy_tot,precision_tot,recall_tot,f1_score_tot,f2_score_tot,
                accuracy_roll,precision_roll,recall_roll,f1_score_roll,f2_score_roll,
                true_negative,false_positive,false_negative,true_positive,
                nb_outliers_roll,nb_labels_roll,nb_outliers_tot,nb_labels_tot]