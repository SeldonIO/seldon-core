import random
import numpy as np

__version__ = "v1.2"

def n_success_failures(features,reward):
    n_predictions = features.shape[0]
    n_success = int(reward*n_predictions)
    n_failures = n_predictions - n_success
    return n_success, n_failures

class EpsilonGreedy(object):

    def __init__(self,n_branches=None,epsilon=0.1,best_branch=None,verbose=False,seed=None,history=False):
        print("Starting Epsilon Greedy Microservice, version {}".format(__version__))

        # for reproducibility
        if seed:
            random.seed(seed)
            np.random.seed(seed)

        if n_branches is None:
            raise Exception("n_branches parameter must be given")

        self.name = 'epsilon_greedy_' + __version__
        self.verbose = verbose
        self.epsilon = epsilon
        self.history = history

        if best_branch:
            self.best_branch = best_branch
        else:
            self.best_branch = random.choice(range(n_branches))

        self.branch_success = [0 for _ in range(n_branches)]
        self.branch_tries = [0 for _ in range(n_branches)]
        self.branch_values = [0 for _ in range(n_branches)]

        if self.history:
            self.branch_history = []
            self.value_history = []

        if self.verbose:
            print("Router initialised")
            print("# branches:",self.n_branches)
            print("Epsilon:",self.epsilon)
            print()

    def route(self,features,feature_names):
        x = random.random()
        other_branches = [i for i in range(self.n_branches) if i != self.best_branch]
        selected_branch = self.best_branch if x > self.epsilon else random.choice(other_branches)

        if self.history:
            self.branch_history.append(selected_branch)
            self.value_history.append(self.branch_values.copy())

        if self.verbose:
            print("Routing")
            print("Current best branch:",best_branch)
            print("Selected branch:",selected_branch)
            print()
        return selected_branch

    def send_feedback(self,features,feature_names,routing,reward,truth):
        if self.verbose:
            print("Training")
            print("Prev success #", self.branches_success)
            print("Prev tries #", self.branches_tries)
            print("Prev best branch:", self.best_branch)

        n_success, n_failures = self.n_success_failures(features,reward)
        self.branch_success[routing] += n_success
        self.branch_tries[routing] += n_success + n_failures
        self.branch_values[routing] = self.branch_success[routing]/self.branch_tries[routing]

        # break ties randomly
        self.best_branch = np.random.choice(np.where(np.array(self.branch_values) == max(self.branch_values))[0])

        if self.verbose:
            print("New success #", self.branches_success)
            print("New tries #", self.branches_tries)
            print("New best branch:",self.best_branch)
            print()

    def metrics(self):
        return [{"type":"GAUGE", "key":"branch_value_{}".format(b), "value":self.branch_values[b]} for b in range(self.n_branches)]

    def n_success_failures(self,features,reward):
        n_predictions = features.shape[0]
        n_success = int(reward*n_predictions)
        n_failures = n_predictions - n_success
        return n_success, n_failures
