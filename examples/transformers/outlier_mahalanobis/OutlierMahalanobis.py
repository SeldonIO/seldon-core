import numpy as np
from scipy.linalg import eigh

_EPSILON = 1e-8

class OutlierMahalanobis(object):
    def __init__(self,n_components=3):
        self.mean = 0
        self.complete_mean = 0
        self.C = 0
        self.complete_C = 0
        self.n = 0
        self.n_components = n_components

    def score(self,features,feature_names):

        nb = features.shape[0] # batch size
        p = features.shape[1] # number of features
        n_components = min(self.n_components,p)

        # PCA
        roll_partial_means = features.cumsum(axis=0)/(np.arange(nb)+1).reshape((nb,1))
        coefs = (np.arange(nb)+1.)/(np.arange(nb)+self.n+1.)
        new_means = self.complete_mean + coefs.reshape((nb,1))*(roll_partial_means-self.complete_mean)
        new_means_offset = np.empty_like(new_means)
        new_means_offset[0] = self.complete_mean
        new_means_offset[1:] = new_means[:-1]
        B = np.matmul((features - new_means)[:,:,None],(features - new_means_offset)[:,None,:])
        cov_complete_batch = (self.complete_C + B.sum(axis=0))/(self.n+nb-1.)

        eigvals, eigvects = eigh(cov_complete_batch,eigvals=(p-n_components,p-1))

        self.complete_mean = new_means[-1]
        self.complete_C = cov_complete_batch

        # Projecting features on the principal components subspace
        proj_features = np.matmul(features,eigvects)

        # Outlier detection is the PC subspace
        roll_partial_means = proj_features.cumsum(axis=0)/(np.arange(nb)+1).reshape((nb,1))
        coefs = (np.arange(nb)+1.)/(np.arange(nb)+self.n+1.)
        new_means = self.mean + coefs.reshape((nb,1))*(roll_partial_means-self.mean)
        new_means_offset = np.empty_like(new_means)
        new_means_offset[0] = self.mean
        new_means_offset[1:] = new_means[:-1]

        B = np.matmul((proj_features - new_means)[:,:,None],(proj_features - new_means_offset)[:,None,:])
        all_C = self.C + B.cumsum(axis=0)
        all_C_inv = np.zeros_like(B)
        all_C = np.roll(all_C,1,axis=0)
        all_C[0] = self.C
        c_inv = None

        for i, b in enumerate(B):
            if c_inv is None:
                if abs(np.linalg.det(all_C[i])) > _EPSILON:
                    c_inv = np.linalg.inv(all_C[i])
                else:
                    continue
            else:
                c_inv = all_C_inv[i-1]
            BC1 = np.matmul(b,c_inv)
            all_C_inv[i] = c_inv - 1./(1.+np.trace(BC1))*np.matmul(c_inv,BC1)

        all_C_inv *= np.arange(nb).reshape((nb,1,1)) + self.n

        self.C += B.sum(axis=0)

        self.mean = new_means[-1]
        
        self.n += nb

        feat_diff = proj_features-new_means_offset
        outlier_scores = np.matmul(feat_diff[:,None,:],np.matmul(all_C_inv,feat_diff[:,:,None])).reshape(nb)

        return list(outlier_scores)
