import math

class OutlierZscore(object):
    def __init__(self):
        self.mean = None
        self.square_diff_sum = 0
        self.n = 0

    def score(self,features,feature_names):

        #TODO: rewrite all of this to handle batches
        
        self.n+=1
        
        # Online algorithm for keeping track of mean and variance        
        if self.mean is None:
            # Initialisation
            self.mean = features
            # This is the first point we observe, so we return 0 as the outlier score
            return 0
        else:
            # Updating mean and sum of squared differences (Welford method)
            new_mean = self.mean + 1/float(self.n)*(features - self.mean)
            self.square_diff_sum = self.square_diff_sum + (features-new_mean)*(features-self.mean)
            self.mean = new_mean
        

        # Computing mean and variance
        mean = self.mean
        std = math.sqrt(self.square_diff_sum/float(self.n-1))

        # Computing z-scores
        z = features - self.mean

        # We get one z-score per dimension. We now need to consolidate this into a single measurement
        
        
            
