import logging

logger = logging.getLogger(__name__)


class MyModel:
    """
    Model template. You can load your model parameters in __init__ from a location accessible at runtime
    """

    def __init__(self):
        """
        Add any initialization parameters. These will be passed at runtime from the graph definition parameters defined in your seldondeployment kubernetes resource manifest.
        """
        logger.info("Initializing model")

    def predict(self, X, features_names):
        """
        Return a prediction.

        Parameters
        ----------
        X : array-like
        feature_names : array of feature names (optional)
        """
        logger.info("Predict called - will run idenity function")
        return X

    def send_feedback(self, features, feature_names, reward, truth, routing=None):
        """
        Handle feedback

        Parameters
        ----------
        features : array - the features sent in the original predict request
        feature_names : array of feature names. May be None if not available.
        reward : float - the reward
        truth : array with correct value (optional)
        """
        logger.info("Send feedback called")
        return []
