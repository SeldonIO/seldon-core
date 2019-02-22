import logging
import numpy as np
import pickle
import random

from model import model

logger = logging.getLogger(__name__)

class CoreSeq2SeqLSTM(object):
    """ Outlier detection using a sequence-to-sequence (seq2seq) LSTM model.
    
    Parameters
    ----------
        threshold (float):  reconstruction error (mse) threshold used to classify outliers
        reservoir_size (int) : number of observations kept in memory using reservoir sampling
     
    Functions
    ----------
        reservoir_sampling : applies reservoir sampling to incoming data
        predict : detect and return outliers
        transform_input : detect outliers and return input features
        send_feedback : add target labels as part of the feedback loop
        tags : add metadata for input transformer
        metrics : return custom metrics
    """
    
    def __init__(self,threshold=0.003,reservoir_size=50000,model_name='seq2seq',load_path='./models/'):
        
        logger.info("Initializing model")
        self.threshold = threshold
        self.reservoir_size = reservoir_size
        self.batch = []
        self.N = 0 # total sample count up until now for reservoir sampling
        self.nb_outliers = 0
        
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
    
    
    def _get_preds(self,X):
        """ Detect outliers if the reconstruction error is above the threshold. 
        
        Parameters
        ----------
        X : array-like
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
        mse = {"type":"GAUGE","key":"mse","value":np.mean(self.mse)}
        nb_outliers = {"type":"GAUGE","key":"nb_outliers","value":int(self.nb_outliers)}
        fraction_outliers = {"type":"GAUGE","key":"fraction_outliers","value":int(self.nb_outliers)/self.N}
        obs = {"type":"GAUGE","key":"observation","value":self.N}
        threshold = {"type":"GAUGE","key":"threshold","value":self.threshold}

        return [is_outlier,mse,nb_outliers,fraction_outliers,obs,threshold]