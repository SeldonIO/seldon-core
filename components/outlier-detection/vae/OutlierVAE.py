import numpy as np
import pickle
import random

from model import model
from utils import flatten, performance, outlier_stats


class OutlierVAE(object):
    """ Outlier detection using variational autoencoders (VAE).
    
    Arguments:
        - threshold: (float): reconstruction error (mse) threshold used to classify outliers
        - reservoir_size (int): number of observations kept in memory using reservoir sampling used for mean and stdev
     
    Functions:
        - reservoir_sampling: applies reservoir sampling to incoming data
        - predict: detect and return outliers
        - send_feedback: add target labels as part of the feedback loop
        - metrics: return custom metrics
    """
    def __init__(self,threshold=10,reservoir_size=50000,load_path='./models/'):
        
        self.threshold = threshold
        self.reservoir_size = reservoir_size
        self.batch = []
        self.N = 0 # total sample count up until now for reservoir sampling
        
        # load model architecture parameters
        with open(load_path + 'model.pickle', 'rb') as f:
            n_features, hidden_layers, latent_dim, hidden_dim = pickle.load(f)
            
        # instantiate model
        self.vae = model(n_features,hidden_layers=hidden_layers,latent_dim=latent_dim,hidden_dim=hidden_dim)
        self.vae.load_weights(load_path + 'vae_weights.h5') # load pretrained model weights
        self.vae._make_predict_function()
        
        # load mu and sigma vectors for each feature
        with open(load_path + 'mu_sigma.pickle', 'rb') as f:
            self.mu, self.sigma = pickle.load(f)
        
        self._predictions = []
        self._labels = []
        self._mse = []
        self.roll_window = 100
        self.metric = [float('nan') for i in range(18)]
        
    
    def reservoir_sampling(self,X,update_stand=False):
        """ Keep batch of data in memory using reservoir sampling. """
        for item in X:
            self.N+=1
            if len(self.batch) < self.reservoir_size:
                self.batch.append(item)
            else:
                s = int(random.random() * self.N)
                if s < self.reservoir_size:
                    self.batch[s] = item
        
        if update_stand: # update mu and sigma
            self.mu = np.mean(self.batch,axis=0)
            self.sigma = np.std(self.batch,axis=0)
        return

    
    def predict(self,X,feature_names):
        """ Detect outliers from mse using the threshold. 
        
        Arguments:
            - X: input data
            - feature_names
        """
        
        if self.N < self.reservoir_size:
            update_stand = False
        else:
            update_stand = True
            
        self.reservoir_sampling(X,update_stand=update_stand)
        
        X_scaled = (X - self.mu) / (self.sigma + 1e-10) # standardize input variables
        
        # sample latent variables and calculate reconstruction errors
        N = 10
        mse = np.zeros([X.shape[0],N])
        for i in range(N):
            preds = self.vae.predict(X_scaled)
            mse[:,i] = np.mean(np.power(X_scaled - preds, 2), axis=1)
        self.mse = np.mean(mse, axis=1)
        self._mse.append(self.mse)
        self._mse = flatten(self._mse)
        
        # make prediction
        self.prediction = np.array([1 if e > self.threshold else 0 for e in self.mse]).astype(int)
        self._predictions.append(self.prediction)
        self._predictions = flatten(self._predictions)
        
        return self.prediction
    
    
    def send_feedback(self,X,feature_names,reward,truth):
        """ Return outlier labels as part of the feedback loop.
        
        Arguments:
            - X: input data
            - feature_names
            - reward
            - truth: outlier labels
        """
        self.label = truth
        self._labels.append(self.label)
        self._labels = flatten(self._labels)
        
        scores = performance(self._labels,self._predictions,roll_window=self.roll_window)
        stats = outlier_stats(self._labels,self._predictions,roll_window=self.roll_window)
        
        convert = flatten([scores,stats])
        metric = []
        for c in convert: # convert from np to native python type to jsonify
            metric.append(np.asscalar(np.asarray(c)))
        self.metric = metric
            
        return

    
    def metrics(self):
        """ Return custom metrics.
        Printed with a delay of 1 prediction because the labels are returned in the feedback step. 
        """
        
        if self.mse.shape[0]>1:
            raise ValueError('Metrics can only handle single observations.')
        
        if self.N==1:
            pred = float('nan')
            err = float('nan')
            y_true = float('nan')
        else:
            pred = int(self._predictions[-2])
            err = self._mse[-2]
            y_true = int(self.label[0])
                         
        is_outlier = {"type":"GAUGE","key":"is_outlier","value":pred}
        mse = {"type":"GAUGE","key":"mse","value":err}
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
        
        return [is_outlier,mse,obs,threshold,label,
                accuracy_tot,precision_tot,recall_tot,f1_score_tot,f2_score_tot,
                accuracy_roll,precision_roll,recall_roll,f1_score_roll,f2_score_roll,
                true_negative,false_positive,false_negative,true_positive,
                nb_outliers_roll,nb_labels_roll,nb_outliers_tot,nb_labels_tot]