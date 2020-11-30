import random
import logging
import numpy as np

__version__ = "0.1"
logger = logging.getLogger(__name__)


class ThompsonSampling(object):
    """ Multi-armed bandit routing using Thompson Sampling strategy.

    This class implements Thompson Sampling for the Beta-Binomial model, i.e.
    rewards are assumed to come from a Bernoulli distribution for which the
    conjugate prior is a Beta distribution.

    The reward is assumed to be a single float between 0 and 1 indicating the
    mean reward for a batch of samples. The prior is a Beta(1,1) distribution
    (Uniform over the child components).
#
    Parameters
    ----------
    n_branches : int
        Number of child components/models the router will route requests to
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
        except (TypeError, ValueError) as e:
            logger.exception("n_branches parameter must be given")
            raise

        self.name = __name__ + __version__
        self.n_branches = n_branches
        self.models_beta_params = [[1, 1] for _ in range(n_branches)]
        self.verbose = verbose
        self.history = history

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

        logger.info("Router initialised, n_branches: %s", self.n_branches)

    def route(self, features, feature_names):
        logger.debug("Routing features %s", features)

        self.branch_values = [np.random.beta(a, b) for a, b in self.models_beta_params]
        selected_branch = np.argmax(self.branch_values)
        logger.debug("Sampled branch values: %s", self.branch_values)

        if self.history:
            self.branch_history.append(selected_branch)
            self.value_history.append(self.branch_values)

        logger.info("Routing to branch %s", selected_branch)
        return selected_branch

    def send_feedback(self, features, feature_names, reward, truth, routing=None):
        logger.debug("Sending feedback with reward %s and truth %s", reward, truth)
        logger.debug("Prev success # %s", self.branch_success)
        logger.debug("Prev tries # %s", self.branch_tries)

        n_success, n_failures = self.n_success_failures(features, reward)
        logger.debug("n_success: %s, n_failures: %s", n_success, n_failures)

        self.models_beta_params[routing][0] += n_success
        self.models_beta_params[routing][1] += n_failures

        self.branch_success[routing] += n_success
        self.branch_tries[routing] += n_success + n_failures

        logger.debug("New success # %s", self.branch_success)
        logger.debug("New tries # %s", self.branch_tries)

    def n_success_failures(self, features, reward):
        n_predictions = features.shape[0]
        n_success = int(reward * n_predictions)
        n_failures = n_predictions - n_success
        return n_success, n_failures
