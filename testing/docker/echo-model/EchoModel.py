from seldon_core.user_model import SeldonComponent


class EchoModel(SeldonComponent):
    def __init__(self):
        print("Initialising")

    def predict(self, X, features_names):
        print("Predict called")
        return X

    def metrics(self):
        print("metrics called")
        return [
            # a counter which will increase by the given value
            {"type": "COUNTER", "key": "mycounter", "value": 1},

            # a gauge which will be set to given value
            {"type": "GAUGE", "key": "mygauge", "value": 100},

            # a timer (in msecs) which  will be aggregated into HISTOGRAM
            {"type": "TIMER", "key": "mytimer", "value": 20.2},
        ]
