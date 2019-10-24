class MyTransformer(object):
    def __init__(self, metrics_ok=True):
        print("Init called")

    def transform_input(self, X, features_names):
        return X + 1

    def transform_output(self, X, features_names):
        return X + 1
