class ModelWithMetrics(object):
    def __init__(self):
        print("Initialising")

    def predict(self, X, features_names):
        print("Predict called")
        return X

    def metrics(self):
        print("metrics called")
        return [
            # a counter which will increase by the given value
            {
                "type": "COUNTER",
                "key": "mycounter",
                "value": 1,
            },
            # a gauge which will be set to given value
            {
                "type": "GAUGE",
                "key": "mygauge",
                "value": 100,
            },
            # a histogram
            {
                "type": "HISTOGRAM",
                "key": "myhistogram",
                "value": 0.5,
                "bins": [0, 1, 2],
            },
            # a timer, which is also an histogram, but would assume the unit to
            # be seconds and convert it to milliseconds (i.e. divide the value
            # by 1000)
            {
                "type": "TIMER",
                "key": "mytimer",
                "value": 20.2,
            },
        ]
