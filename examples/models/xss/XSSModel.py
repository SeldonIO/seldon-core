class XSSModel(object):
    """
    Dummy model which just returns its input back.
    """

    def predict(self, X, feature_names):
        return X
