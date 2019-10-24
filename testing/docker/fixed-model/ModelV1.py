class ModelV1(object):
    def __init__(self):
        print("Initialising")

    def predict(self, X, features_names):
        print("Predict called")
        return [1, 2, 3, 4]

    def send_feedback(self, features, feature_names, reward, truth):
        print("Send feedback called")
        return []
