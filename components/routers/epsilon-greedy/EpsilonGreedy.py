import random
import logging
import numpy as np

__version__ = "1.3"
logger = logging.getLogger(__name__)


class EpsilonGreedy(object):
    """ Multi-armed bandit routing using epsilon-greedy strategy.

    This class implements epsilon-greedy routing. The rewards are assumed to
    come from a Bernoulli distribution. The reward is assumed to be a single
    float between 0 and 1 indicating the mean reward for a batch of samples.

    Parameters
    ----------
    n_branches : int
        Number of child components/models the router will route requests to
    epsilon : float
        Epsilon parameter specifying the probability of choosing a random branch
    best_branch : int, optional
        Specify the starting branch
    verbose : bool
        Set the logger level
    seed : int, optional
        Set the random seed
    history : bool
        Set storing router history
    branch_names: str, optional
        A string specifying branch names separated by `:`

    """

    def __init__(
        self,
        n_branches=None,
        epsilon=0.1,
        best_branch=None,
        verbose=False,
        seed=None,
        history=False,
        branch_names=None,
    ):

        if verbose:
            logger.setLevel(10)
            logger.info("Enabling debug mode")

        logger.info("Starting %s Microservice, version %s", __name__, __version__)

        # for reproducibility
        if seed:
            logger.info("Setting random seed to %s", seed)
            random.seed(seed)
            np.random.seed(seed)

        try:
            n_branches = int(n_branches)
            assert n_branches > 0
        except (TypeError, ValueError, AssertionError) as e:
            logger.exception("n_branches parameter must be given")
            raise

        self.name = __name__ + __version__
        self.verbose = verbose
        self.n_branches = n_branches
        self.epsilon = epsilon
        self.history = history

        if best_branch:
            logger.info("Best branch supplied: %s", best_branch)
            self.best_branch = best_branch
        else:
            self.best_branch = random.choice(range(n_branches))
            logger.info("Starting with a random branch: %s", self.best_branch)

        self.branch_success = [0 for _ in range(n_branches)]
        self.branch_tries = [0 for _ in range(n_branches)]
        self.branch_values = [0 for _ in range(n_branches)]

        if self.history:
            logger.info("Enabling history")
            self.branch_history = []
            self.value_history = []

        if branch_names is not None:
            self.branch_names = branch_names.split(":")
            logger.info("Branch names: %s", self.branch_names)

        logger.info(
            "Router initialised, n_branches: %s, epsilon: %s",
            self.n_branches,
            self.epsilon,
        )

    def route(self, features, feature_names):
        logger.debug("Routing features %s", features)

        x = random.random()
        other_branches = [i for i in range(self.n_branches) if i != self.best_branch]
        selected_branch = (
            self.best_branch if x > self.epsilon else random.choice(other_branches)
        )

        if self.history:
            self.branch_history.append(selected_branch)
            self.value_history.append(self.branch_values.copy())

        logger.info("Routing to branch %s", selected_branch)
        logger.debug(
            "Current best branch %s has value %s",
            self.best_branch,
            self.branch_values[self.best_branch],
        )
        logger.debug(
            "Selected branch %s has value %s",
            selected_branch,
            self.branch_values[selected_branch],
        )

        logger.info(f"routing type: {type(selected_branch)}")
        return int(selected_branch)

    def send_feedback(self, features, feature_names, reward, truth, routing=None):
        logger.debug("Sending feedback with reward %s and truth %s", reward, truth)
        logger.debug("Prev success # %s", self.branch_success)
        logger.debug("Prev tries # %s", self.branch_tries)
        logger.debug("Prev best branch: %s", self.best_branch)

        n_success, n_failures = self.n_success_failures(features, reward)
        logger.debug("n_success: %s, n_failures: %s", n_success, n_failures)

        self.branch_success[routing] += n_success
        self.branch_tries[routing] += n_success + n_failures
        self.branch_values[routing] = (
            self.branch_success[routing] / self.branch_tries[routing]
        )

        # break ties randomly
        self.best_branch = np.random.choice(
            np.where(np.array(self.branch_values) == max(self.branch_values))[0]
        )

        logger.debug("New success # %s", self.branch_success)
        logger.debug("New tries # %s", self.branch_tries)
        logger.debug("Branch values %s", self.branch_values)
        logger.debug("New best branch: %s", self.best_branch)

    def n_success_failures(self, features, reward):
        n_predictions = features.shape[0]
        n_success = int(reward * n_predictions)
        n_failures = n_predictions - n_success
        return n_success, n_failures
