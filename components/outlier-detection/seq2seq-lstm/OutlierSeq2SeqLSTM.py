import numpy as np
import pickle
import random
from model import model
from utils import flatten, performance, outlier_stats

class OutlierSeq2SeqLSTM(object):
    """ Outlier detection using a sequence-to-sequence (seq2seq) LSTM model.
    
    Arguments:
        - threshold: (float): reconstruction error (mse) threshold used to classify outliers
        - reservoir_size (int): number of observations kept in memory using reservoir sampling
     
    Functions:
        - reservoir_sampling: applies reservoir sampling to incoming data
        - predict: detect and return outliers
        - send_feedback: add target labels as part of the feedback loop
        - metrics: return custom metrics
    """
    def __init__(self,threshold=0.003,reservoir_size=50000,model_name='seq2seq',load_path='./models/'):
        
        self.threshold = threshold
        self.reservoir_size = reservoir_size
        self.batch = []
        self.N = 0 # total sample count up until now for reservoir sampling
        
        # load model architecture parameters
        with open(load_path + model_name + '.pickle', 'rb') as f:
            self.timesteps, self.n_features, encoder_dim, decoder_dim, output_activation = pickle.load(f)
            
        # instantiate model
        self.s2s, self.enc, self.dec = model(self.n_features,encoder_dim=encoder_dim,
                                             decoder_dim=decoder_dim,output_activation=output_activation)
        self.s2s.load_weights(load_path + model_name + '_weights.h5') # load pretrained model weights
        self.s2s._make_predict_function()
        self.enc._make_predict_function()
        self.dec._make_predict_function()
        
        # load data preprocessing info
        with open(load_path + 'preprocess_' + model_name + '.pickle', 'rb') as f:
            preprocess = pickle.load(f)
        self.preprocess, self.clip, self.axis = preprocess[:3]
        if self.preprocess=='minmax':
            self.xmin, self.xmax = preprocess[3:5]
            self.min, self.max = preprocess[5:]
        elif self.preprocess=='standardized':
            self.mu, self.sigma = preprocess[3:]
        
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
        
        if update_stand:
            if self.preprocess=='minmax':
                self.xmin = np.array(self.batch).min(axis=self.axis)
                self.xmax = np.array(self.batch).max(axis=self.axis)
            elif self.preprocess=='standardized':
                self.mu = np.array(self.batch).mean(axis=self.axis)
                self.sigma = np.array(self.batch).std(axis=self.axis)
        return
    
    
    def decode_sequence(self,input_seq):
        """ Feed output of encoder to decoder and make sequential predictions. """
        
        # use encoder the get state vectors
        states_value = self.enc.predict(input_seq)
        
        # generate initial target sequence
        target_seq = input_seq[0,0,:].reshape((1,1,self.n_features))
        
        # sequential prediction of time series
        decoded_seq = np.zeros((1, self.timesteps, self.n_features))
        decoded_seq[0,0,:] = target_seq[0,0,:]
        i = 1
        while i < self.timesteps:
            
            decoder_output = self.dec.predict([target_seq] + states_value)

            # update the target sequence
            target_seq = np.zeros((1, 1, self.n_features))
            target_seq[0, 0, :] = decoder_output[0]

            # update output
            decoded_seq[0, i, :] = decoder_output[0]

            # update states
            states_value = decoder_output[1:]

            i+=1

        return decoded_seq
    
    
    def predict(self,X,feature_names):
        """ Detect outliers from mse using the threshold. 
        
        Arguments:
            - X: input data
            - feature_names
        """
        
        # clip data per feature
        for col,clip in enumerate(self.clip):
            X[:,:,col] = np.clip(X[:,:,col],-clip,clip)
        
        # update reservoir
        if self.N < self.reservoir_size:
            update_stand = False
        else:
            update_stand = True
        
        self.reservoir_sampling(X,update_stand=update_stand)
        
        # apply scaling
        if self.preprocess=='minmax':
            X = ((X - self.xmin) / (self.xmax - self.xmin)) * (self.max - self.min) + self.min
        elif self.preprocess=='standardized':
            X = (X - self.mu) / (self.sigma + 1e-10)
        
        # make predictions
        n_obs = X.shape[0]
        self.mse = np.zeros(n_obs)
        for obs in range(n_obs):
            input_seq = X[obs:obs+1,:,:]
            decoded_seq = self.decode_sequence(input_seq)
            self.mse[obs] = np.mean(np.power(input_seq[0,:,:] - decoded_seq[0,:,:], 2))
        self.prediction = np.array([1 if e > self.threshold else 0 for e in self.mse]).astype(int)
        
        # update mse and prediction list
        self._mse.append(self.mse)
        self._mse = flatten(self._mse)
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