import joblib


class MulticlassClassifier:
    def __init__(self, model_name="multiclass-lr.joblib"):
        self.scores = {"correct": 0, "incorrect": 0}
        self.model = joblib.load(model_name)

    def predict(self, X, features_names=None, meta=None):
        return self.model.predict(X)

    def send_feedback(self, features, feature_names, reward, truth, routing=""):
        predicted = self.predict(features)
        if int(predicted[0]) == int(truth[0]):
            self.scores["correct"] += 1
        else:
            self.scores["incorrect"] += 1
        return []  # Ignore return statement as its not used

    def metrics(self):
        return [
            {"type": "GAUGE", "key": "correct", "value": self.scores["correct"]},
            {"type": "GAUGE", "key": "incorrect", "value": self.scores["incorrect"]},
        ]
