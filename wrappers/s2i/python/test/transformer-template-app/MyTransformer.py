
class MyTransformer(object):
    """
    Router template. 
    """
    
    def __init__(self):
        """
        Add any initialization parameters. These will be passed at runtime from the graph definition parameters defined in your seldondeployment kubernetes resource manifest.
        """
        print("Initializing")


    def transform_input(self,features,feature_names):
        """
        transform input.

        Parameters
        ----------
        features : array-like
        feature_names : array of feature names (optional)
        """
        print("Running identity transform")
        return features

    def transform_output(self,features,feature_names):
        """
        transform output.

        Parameters
        ----------
        features : array-like
        feature_names : array of feature names (optional)
        """
        print("Running identity transform")
        return features
        

