from seldon_core.user_model import SeldonResponse

class BinaryMetrics:
    def __init__(self):
        pass
    def transform(self, truth, response, request = None):

        response_class = int(response[0]) if isinstance(response, list) else int(response)
        truth_class = int(truth[0]) if isinstance(truth, list) else int(truth)

        correct = response_class == truth_class

        if truth_class:
            if correct:
                key = "seldon_metric_true_positive"
            else:
                key = "seldon_metric_false_negative"
        else:
            if correct:
                key = "seldon_metric_true_negative"
            else:
                key = "seldon_metric_false_positive"

        metrics = [{"key":key, "type": "COUNTER", "value": 1}]

        return SeldonResponse(None, None, metrics)
