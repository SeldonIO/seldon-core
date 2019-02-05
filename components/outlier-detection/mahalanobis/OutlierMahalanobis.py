import numpy as np

from CoreMahalanobis import CoreMahalanobis
from utils import flatten, performance, outlier_stats

class OutlierMahalanobis(CoreMahalanobis):
    """ Outlier detection using the Mahalanobis distance.
    
    Parameters
    ----------
        threshold (float) : Mahalanobis distance threshold used to classify outliers
        n_components (int) : number of principal components used
        n_stdev (float) : stdev used for feature-wise clipping of observations
        start_clip (int) : number of observations before clipping is applied
        max_n (int) : algorithm behaves as if it has seen at most max_n points
     
    Functions
    ----------
        send_feedback : add target labels as part of the feedback loop
        metrics : return custom metrics
    """
    def __init__(self,threshold=25,n_components=3,n_stdev=3,start_clip=50,max_n=-1):
        
        super().__init__(threshold=threshold,n_components=n_components,n_stdev=n_stdev,
                         start_clip=start_clip,max_n=max_n)
        
        self._predictions = []
        self._labels = []
        self._scores = []
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
        self._scores.append(self.score)
        self._scores = flatten(self._scores)
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
        
        if self.score.shape[0]>1:
            raise ValueError('Metrics can only handle single observations.')
        
        if self.n==1:
            pred = float('nan')
            err = float('nan')
            y_true = float('nan')
        else:
            pred = int(self._predictions[-1])
            err = self._scores[-1]
            y_true = int(self.label[0])
                         
        is_outlier = {"type":"GAUGE","key":"is_outlier","value":pred}
        outlier_score = {"type":"GAUGE","key":"outlier_score","value":err}
        obs = {"type":"GAUGE","key":"observation","value":self.n - 1}
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
        
        return [is_outlier,outlier_score,obs,threshold,label,
                accuracy_tot,precision_tot,recall_tot,f1_score_tot,f2_score_tot,
                accuracy_roll,precision_roll,recall_roll,f1_score_roll,f2_score_roll,
                true_negative,false_positive,false_negative,true_positive,
                nb_outliers_roll,nb_labels_roll,nb_outliers_tot,nb_labels_tot]
