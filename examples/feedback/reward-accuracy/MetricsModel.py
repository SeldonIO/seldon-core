import joblib


class Score:
    def __init__(self, TP, FP, TN, FN):
        self.TP = TP
        self.FP = FP
        self.TN = TN
        self.FN = FN


class MetricsModel:
    def __init__(self, model_name="binary-lr.joblib"):
        self.scores = Score(0, 0, 0, 0)
        self.model = joblib.load(model_name)
        self.model_name = model_name.split()[0]

    def predict(self, X, features_names=None, meta=None):
        return self.model.predict(X)

    def send_feedback(self, features, feature_names, reward, truth, routing=""):
        predicted = self.predict(features)
        print(f"Predicted: {predicted[0]}, Truth: {truth[0]}")
        if int(truth[0]) == 1:
            if int(predicted[0]) == int(truth[0]):
                self.scores.TP += 1
            else:
                self.scores.FN += 1
        else:
            if int(predicted[0]) == int(truth[0]):
                self.scores.TN += 1
            else:
                self.scores.FP += 1
        return []  # Ignore return statement as its not used

    def metrics(self):
        return [
            {"type": "GAUGE", "key": f"true_pos", "value": self.scores.TP},
            {"type": "GAUGE", "key": f"true_neg", "value": self.scores.FN},
            {"type": "GAUGE", "key": f"false_pos", "value": self.scores.TN},
            {"type": "GAUGE", "key": f"false_neg", "value": self.scores.FP},
        ]
