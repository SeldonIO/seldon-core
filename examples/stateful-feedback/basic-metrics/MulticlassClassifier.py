import joblib


class Scores:
    def __init__(self, numclass):
        self.TP = [0] * numclass
        self.FP = [0] * numclass
        self.FN = [0] * numclass
        # self.TN = [0] * numclass # Most often there won't be true negatives in multiclass although w
        # We could explore using reward (or a binary param) to define if expected TN, but that seems more custom


class MulticlassClassifier:
    def __init__(self, model_name="multiclass-lr.joblib"):
        self.scores = Scores(3)
        self.model = joblib.load(model_name)

    def predict(self, X, features_names=None, meta=None):
        return self.model.predict(X)

    def send_feedback(self, features, feature_names, reward, truth, routing=""):
        predicted = self.predict(features)
        print(f"Predicted: {predicted}")
        print(f"Truth: {truth}")
        if int(predicted[0]) == int(truth[0]):
            self.scores.TP[int(truth[0])] += 1
        else:
            self.scores.FN[int(truth[0])] += 1
            self.scores.FP[int(predicted[0])] += 1
        return []  # Ignore return statement as its not used

    def metrics(self):
        metrics = []
        for score, arr in vars(self.scores).items():
            for i, val in enumerate(arr):
                metric = {
                    "type": "GAUGE",
                    "key": f"score_{score}",
                    "value": val,
                    "tags": {"class": f"class_{i}"},
                }
                print(metric)
                metrics.append(metric)
        return metrics
