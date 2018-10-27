class Test(object):
    """
    Model template. You can load your model parameters in __init__ from a location accessible at runtime
    """

    def __init__(self):
        """
        Add any initialization parameters. These will be passed at runtime from the graph definition parameters defined in your seldondeployment kubernetes resource manifest.
        """
        print("Initializing")

    def predict(self,X,features_names):
        """
        Return a prediction.

        Parameters
        ----------
        X : array-like
        feature_names : array of feature names (optional)
        """
        print("Predict called - will run identity function")
        print(X)
        return X

    #
    # OPTIONAL
    #
    # This is an optional method that can added to provide custom service endpoints.
    #
    def custom_service(self):
        from flask import Flask, jsonify, request, json

        app = Flask(__name__)

        @app.route("/prometheus_metrics",methods=["GET"])
        def prometheus_metrics():
            return "somemetric 10\n"

        @app.route("/data",methods=["POST"])
        def data():
            data_str = request.data
            message = json.loads(data_str)
            print(data_str)
            return jsonify(message)

        print("Starting custom service")
        app.run(host='0.0.0.0', port=5055)

