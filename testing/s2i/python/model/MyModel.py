class MyModel(object):
    def __init__(self, metrics_ok=True):
        print("Init called")

    def predict(self, X, features_names):
        return X + 1

    def send_feedback(self, features, feature_names, routing, reward, truth):
        print("Feedback called")
