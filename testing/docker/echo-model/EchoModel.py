from seldon_core.user_model import SeldonComponent


class EchoModel(SeldonComponent):
    def __init__(self):
        print("Initialising")

    def predict(self, X, features_names):
        print("Predict called")
        return X
