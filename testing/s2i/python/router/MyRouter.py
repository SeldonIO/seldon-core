class MyRouter(object):
    def __init__(self, metrics_ok=True):
        print("Init called")

    def route(self, X, features_names):
        return 0

    def send_feedback(self, features, feature_names, routing, reward, truth):
        print("Feedback called")
