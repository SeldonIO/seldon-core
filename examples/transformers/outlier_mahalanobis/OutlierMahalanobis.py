import numpy as np
from scipy.linalg import eigh

_EPSILON = 1e-8

class OutlierMahalanobis(object):
    def __init__(self,n_components=3):
        self.mean = 0
        self.C = 0
        self.n = 0
        self.n_components = n_components

    def score(self,features,feature_names):

        nb = features.shape[0] # batch size
        p = features.shape[1] # number of features
        n_components = min(self.n_components,p)

        # PCA
        roll_partial_means = features.cumsum(axis=0)/(np.arange(nb)+1).reshape((nb,1))
        coefs = (np.arange(nb)+1.)/(np.arange(nb)+self.n+1.)
        new_means = self.mean + coefs.reshape((nb,1))*(roll_partial_means-self.mean)
        new_means_offset = np.empty_like(new_means)
        new_means_offset[0] = self.mean
        new_means_offset[1:] = new_means[:-1]

        coefs = ((self.n+np.arange(nb))/(self.n+np.arange(nb)+1.)).reshape((nb,1,1))
        B = coefs*np.matmul((features - new_means_offset)[:,:,None],(features - new_means_offset)[:,None,:])
        cov_batch = (self.n-1.)/(self.n+nb-1.)*self.C + 1./(self.n+nb-1.)*B.sum(axis=0)

        eigvals, eigvects = eigh(cov_batch,eigvals=(p-n_components,p-1))
    
        # Projections
        proj_features = np.matmul(features,eigvects)
        proj_means = np.matmul(new_means_offset,eigvects)
        if self.C == 0:
            proj_cov = np.diag(np.zeros(p))
        else:
            proj_cov = np.matmul(eigvects.transpose(),np.matmul(self.C,eigvects))

        # Outlier detection in the PC subspace
        coefs = (1./(self.n+np.arange(nb)+1.)).reshape((nb,1,1))
        B = coefs*np.matmul((proj_features - proj_means)[:,:,None],(proj_features - proj_means)[:,None,:])

        all_C_inv = np.zeros_like(B)
        c_inv = None
        _EPSILON = 1e-8

        for i, b in enumerate(B):
            #print i
            if c_inv is None:
                if abs(np.linalg.det(proj_cov)) > _EPSILON:
                    c_inv = np.linalg.inv(proj_cov)
                    all_C_inv[i] = c_inv
                    continue
                else:
                    if self.n + i == 0:
                        continue
                    proj_cov = (self.n + i -1. )/(self.n + i)*proj_cov + b
                    continue
            else:
                c_inv = (self.n + i - 1.)/float(self.n + i - 2.)*all_C_inv[i-1]
            BC1 = np.matmul(B[i-1],c_inv)
            all_C_inv[i] = c_inv - 1./(1.+np.trace(BC1))*np.matmul(c_inv,BC1)

        # Updates
        self.mean = new_means[-1]
        self.C = cov_batch
        self.n += nb

        feat_diff = proj_features-proj_means
        outlier_scores = np.matmul(feat_diff[:,None,:],np.matmul(all_C_inv,feat_diff[:,:,None])).reshape(nb)

        return list(outlier_scores)
