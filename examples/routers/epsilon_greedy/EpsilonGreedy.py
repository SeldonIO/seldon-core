import random
import numpy as np

def n_success_failures(features,reward):
    n_predictions = features.shape[0]
    n_success = int(reward*n_predictions)
    n_failures = n_predictions - n_success
    return n_success, n_failures

class EpsilonGreedy(object):
    
    def __init__(self,n_branches=None,epsilon=0.1,verbose=False):
        if n_branches is None:
            raise Exception("n_branches parameter must be given")
        self.verbose = verbose
        self.epsilon = epsilon
        self.best_branch = 0
        self.branches_success = [0 for _ in range(n_branches)]
        self.branches_tries = [0 for _ in range(n_branches)]
        self.n_branches = n_branches
        if self.verbose:
            print "Router initialised"
            print "# branches:",self.n_branches
            print "Epsilon:",self.epsilon
            print
        
    def route(self,features,feature_names):
        x = random.random()
        best_branch = self.best_branch
        other_branches = [i for i in range(self.n_branches) if i!=best_branch]
        selected_branch = best_branch if x>self.epsilon else random.choice(other_branches)
        if self.verbose:
            print "Routing"
            print "Current best branch:",best_branch
            print "Selected branch:",selected_branch
            print
        return selected_branch
        
    def send_feedback(self,features,feature_names,routing,reward,truth):
        if self.verbose:
            print "Training"
            print "Prev success #", self.branches_success
            print "Prev tries #", self.branches_tries
            print "Prev best branch:", self.best_branch
        n_success, n_failures = n_success_failures(features,reward)
        self.branches_success[routing] += n_success
        self.branches_tries[routing] += n_success + n_failures
        perfs = [
            (self.branches_success[i]+1)/float(self.branches_tries[i]+1)
            for i
            in range(self.n_branches)
        ]
        self.best_branch = np.argmax(perfs)
        if self.verbose:
            print "New success #", self.branches_success
            print "New tries #", self.branches_tries
            print "New best branch:",self.best_branch
            print

