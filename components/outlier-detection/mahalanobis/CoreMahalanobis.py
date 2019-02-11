import logging
import numpy as np
from scipy.linalg import eigh

logger = logging.getLogger(__name__)

class CoreMahalanobis(object):
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
        predict : detect and return outliers
        transform_input : detect outliers and return input features
        send_feedback : add target labels as part of the feedback loop
        tags : add metadata for input transformer
        metrics : return custom metrics
    """
    def __init__(self,threshold=25,n_components=3,n_stdev=3,start_clip=50,max_n=-1):
        
        logger.info("Initializing model")
        self.threshold = threshold
        self.n_components = n_components
        self.max_n = max_n
        self.n_stdev = n_stdev
        self.start_clip = start_clip
        
        self.clip = None
        self.mean = 0
        self.C = 0
        self.n = 0
        self.nb_outliers = 0
    
    
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
        """ Detect outliers using the Mahalanobis distance threshold.  
        
        Parameters
        ----------
        X : array-like
        """

        nb = X.shape[0] # batch size
        p = X.shape[1] # number of features
        n_components = min(self.n_components,p)
        if self.max_n>0:
            n = min(self.n,self.max_n) # n can never be above max_n
        else:
            n = self.n
        
        # Clip X
        if self.n > self.start_clip:
            Xclip = np.clip(X,self.clip[0],self.clip[1])
        else:
            Xclip = X
            
        # Tracking the mean and covariance matrix
        roll_partial_means = Xclip.cumsum(axis=0)/(np.arange(nb)+1).reshape((nb,1))
        coefs = (np.arange(nb)+1.)/(np.arange(nb)+n+1.)
        new_means = self.mean + coefs.reshape((nb,1))*(roll_partial_means-self.mean)
        new_means_offset = np.empty_like(new_means)
        new_means_offset[0] = self.mean
        new_means_offset[1:] = new_means[:-1]

        coefs = ((n+np.arange(nb))/(n+np.arange(nb)+1.)).reshape((nb,1,1))
        B = coefs*np.matmul((Xclip - new_means_offset)[:,:,None],(Xclip - new_means_offset)[:,None,:])
        cov_batch = (n-1.)/(n+max(1,nb-1.))*self.C + 1./(n+max(1,nb-1.))*B.sum(axis=0)

        # PCA
        eigvals, eigvects = eigh(cov_batch,eigvals=(p-n_components,p-1))
    
        # Projections
        proj_x = np.matmul(X,eigvects)
        proj_x_clip = np.matmul(Xclip,eigvects)
        proj_means = np.matmul(new_means_offset,eigvects)
        if type(self.C) == int and self.C == 0:
            proj_cov = np.diag(np.zeros(n_components))
        else:
            proj_cov = np.matmul(eigvects.transpose(),np.matmul(self.C,eigvects))

        # Outlier detection in the PC subspace
        coefs = (1./(n+np.arange(nb)+1.)).reshape((nb,1,1))
        B = coefs*np.matmul((proj_x_clip - proj_means)[:,:,None],(proj_x_clip - proj_means)[:,None,:])

        all_C_inv = np.zeros_like(B)
        c_inv = None
        _EPSILON = 1e-8

        for i, b in enumerate(B):
            if c_inv is None:
                if abs(np.linalg.det(proj_cov)) > _EPSILON:
                    c_inv = np.linalg.inv(proj_cov)
                    all_C_inv[i] = c_inv
                    continue
                else:
                    if n + i == 0:
                        continue
                    proj_cov = (n + i -1. )/(n + i)*proj_cov + b
                    continue
            else:
                c_inv = (n + i - 1.)/float(n + i - 2.)*all_C_inv[i-1]
            BC1 = np.matmul(B[i-1],c_inv)
            all_C_inv[i] = c_inv - 1./(1.+np.trace(BC1))*np.matmul(c_inv,BC1)

        # Updates
        self.mean = new_means[-1]
        self.C = cov_batch
        stdev = np.sqrt(np.diag(cov_batch))
        self.n += nb
        if self.n > self.start_clip:
            self.clip = [self.mean-self.n_stdev*stdev,self.mean+self.n_stdev*stdev]
        
        # Outlier scores and predictions
        x_diff = proj_x-proj_means
        self.score = np.matmul(x_diff[:,None,:],np.matmul(all_C_inv,x_diff[:,:,None])).reshape(nb)
        self.prediction = np.array([1 if s > self.threshold else 0 for s in self.score]).astype(int)

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
        outlier_score = {"type":"GAUGE","key":"outlier_score","value":np.mean(self.score)}
        nb_outliers = {"type":"GAUGE","key":"nb_outliers","value":int(self.nb_outliers)}
        fraction_outliers = {"type":"GAUGE","key":"fraction_outliers","value":int(self.nb_outliers)/self.n}
        obs = {"type":"GAUGE","key":"observation","value":self.n}
        threshold = {"type":"GAUGE","key":"threshold","value":self.threshold}

        return [is_outlier,outlier_score,nb_outliers,fraction_outliers,obs,threshold]