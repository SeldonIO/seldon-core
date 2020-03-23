class ModelWithMetrics(object):

    def __init__(self):
        print("Initialising")

    def predict(self,X,features_names):
        print("Predict called")
        return X

    def metrics(self):
        print("metrics called")
        return [
            {"type": "COUNTER", "key": "mycounter", "value": 1}, # a counter which will increase by the given value
            {"type": "GAUGE", "key": "mygauge", "value": 100},   # a gauge which will be set to given value
            {"type": "TIMER", "key": "mytimer", "value": 20.2},  # a timer which will add sum and count metrics - assumed millisecs
        ]
