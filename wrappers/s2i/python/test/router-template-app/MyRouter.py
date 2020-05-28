
class MyRouter(object):
    """
    Router template. 
    """
    
    def __init__(self):
        """
        Add any initialization parameters. These will be passed at runtime from the graph definition parameters defined in your seldondeployment kubernetes resource manifest.
        """
        print("Initializing")


    def route(self,features,feature_names):
        """
        Route a request.

        Parameters
        ----------
        features : array-like
        feature_names : array of feature names (optional)
        """
        return 0
        
    def send_feedback(self,features,feature_names,reward,truth,routing):
        """
        Handle feedback for your routings. Optional.

        Parameters
        ----------
        features : array - the features sent in the original predict request
        feature_names : array of feature names. May be None if not available.
        routing: int - the route chosen by this router for this request
        reward : float - the reward
        truth : array with correct value (optional)

        """
        print("Received feedback")

