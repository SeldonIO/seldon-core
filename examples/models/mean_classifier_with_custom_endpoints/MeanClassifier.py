import numpy as np
import math

# use the multiprocessing module to work with shared data
import multiprocessing as mp

def f(x):
    return 1/(1+math.exp(-x))

class MeanClassifier(object):

    def __init__(self, intValue=0):
        self.class_names = ["proba"]
        assert type(intValue) == int, "intValue parameters must be an integer"
        self.int_value = intValue

        print("Loading model here")

        X = np.load(open("model.npy",'rb'), encoding='latin1')
        self.threshold_ = X.mean() + self.int_value

        # Initialize a counter as a shared variable between the /predict endpoint
        # and other custom endpoints.
        self.predict_call_count_sv=mp.Value('i',0)

    def _meaning(self, x):
        return f(x.mean()-self.threshold_)

    def predict(self, X, feature_names):

        # Increment the call counter.
        with self.predict_call_count_sv.get_lock():
            self.predict_call_count_sv.value += 1
        print("predict, predict_call_count_sv[{}]".format(self.predict_call_count_sv.value))

        print(X)
        X = np.array(X)
        assert len(X.shape) == 2, "Incorrect shape"

        return [[self._meaning(x)] for x in X]

    #
    # Define a custom service that will run in addition to the main service.
    #
    # Example here defines a /prometheus_metrics endpoint that runs on port "5055"
    # It uses data shared with the /predict endpoint to deliver metrics that could
    # be scraped by prometheus.
    #
    def custom_service(self):
        from flask import Flask, jsonify, request, json

        app = Flask(__name__)

        @app.route("/prometheus_metrics",methods=["GET"])
        def prometheus_metrics():
            print("prometheus_metrics, predict_call_count_sv[{}]".format(self.predict_call_count_sv.value))
            return "predict_call_count {}\n".format(self.predict_call_count_sv.value)

        print("Starting custom service")
        app.run(host='0.0.0.0', port=5055)

