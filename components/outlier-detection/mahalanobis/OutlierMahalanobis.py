import numpy as np
from scipy.linalg import eigh

from utils import flatten, performance, outlier_stats

class OutlierMahalanobis(object):
    """ Outlier detection using the Mahalanobis distance.
    
    Arguments:
        - threshold: (float): Mahalanobis distance threshold used to classify outliers
        - n_components (int): number of principal components used
        - n_stdev (float): stdev used for feature-wise clipping of observations
        - max_n (int): algorithm behaves as if it has seen at most max_n points
     
    Functions:
        - predict: detect and return outliers
        - send_feedback: add target labels as part of the feedback loop
        - metrics: return custom metrics
    """
    def __init__(self,threshold=25,n_components=3,n_stdev=3,start_clip=50,max_n=-1):
        
        self.threshold = threshold
        self.n_components = n_components
        self.max_n = max_n
        self.n_stdev = n_stdev
        self.start_clip = start_clip
        
        self.clip = None
        self.mean = 0
        self.C = 0
        self.n = 0
        
        self._predictions = []
        self._labels = []
        self._scores = []
        self.roll_window = 100
        self.metric = [float('nan') for i in range(18)]
    
    def predict(self,X,feature_names):
        """ Detect outliers using the Mahalanobis distance threshold. 
        
        Arguments:
            - X: input data
            - feature_names
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
        
        # update outlier scores and prediction list
        self._scores.append(self.score)
        self._scores = flatten(self._scores)
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
        
        if self.score.shape[0]>1:
            raise ValueError('Metrics can only handle single observations.')
        
        if self.n==1:
            pred = float('nan')
            err = float('nan')
            y_true = float('nan')
        else:
            pred = int(self._predictions[-2])
            err = self._scores[-2]
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
