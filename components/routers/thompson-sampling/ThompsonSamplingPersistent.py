import random
import logging
import numpy as np
import redis
import os

PRED_UNIT_ID = os.environ.get("PREDICTIVE_UNIT_ID", "0")
PREDICTOR_ID = os.environ.get("PREDICTOR_ID", "0")
DEPLOYMENT_ID = os.environ.get("SELDON_DEPLOYMENT_ID", "0")

REDIS_HOST = os.environ.get("REDIS_SERVICE_HOST", "localhost")
REDIS_PORT = os.environ.get("REDIS_SERVICE_PORT", 6379)

KEY_PREFIX = f"seldon_{DEPLOYMENT_ID}_{PREDICTOR_ID}_{PRED_UNIT_ID}"
KEY_BETA_PARAMS = "-beta-params"

logger = logging.getLogger(__name__)
__version__ = "0.1"


class ThompsonSamplingPersistent(object):
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

        logger.info(f"Starting {__name__} Microservice")

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

        self.rc = redis.Redis(host=REDIS_HOST, port=REDIS_PORT)

        self.key = self.key + __name__ + __version__
        self.n_branches = n_branches
        self.verbose = verbose

        if not self.rc.exists(self.key):
            models_beta_params = [1 for _ in range(n_branches) * 2]
            self.rc.lpush(self.key, *models_beta_params)

        if branch_names is not None:
            self.branch_names = branch_names.split(":")
            logger.info("Branch names: %s", self.branch_names)

        logger.info("Router initialised, n_branches: %s", self.n_branches)

    def route(self, features, feature_names):
        logger.debug("Routing features %s", features)

        models_beta_params = [int(i) for i in self.rc.lrange(self.key, 0, -1)]

        # Use zip iter to iterate across each pair of numbers in the list
        branch_values = [np.random.beta(a, b) for a, b in zip(*[iter(models_beta_params)] * 2)]

        selected_branch = np.argmax(branch_values)
        logger.debug("Sampled branch values: %s", branch_values)

        logger.info("Routing to branch %s", selected_branch)
        return int(selected_branch)

    def send_feedback(self, features, feature_names, reward, truth, routing=None):
        logger.debug(f"Sending feedback with reward {reward} and truth {truth}")

        n_success, n_failures = self.n_success_failures(features, reward)
        logger.debug(f"n_success: {n_success}, n_failures: {n_failures}")

        # TODO: Non atomic / non-thread-safe operation which will get overridden by other replicas/threads
        self.rc.lset(self.key, routing*2, self.rc.lindex(self.key, routing*2) + n_success)
        self.rc.lset(self.key, routing*2 + 1, self.rc.lindex(self.key, routing*2 + 1) + n_failures)

    def n_success_failures(self, features, reward):
        n_predictions = features.shape[0]
        n_success = int(reward * n_predictions)
        n_failures = n_predictions - n_success
        return n_success, n_failures
