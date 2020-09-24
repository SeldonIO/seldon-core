from seldon_core.user_model import SeldonResponse

class MultiClassNumeric:

    def __init__(self):
        pass

    def transform(self, truth, response, request = None):

        metrics = []

        response_class = response if isinstance(response, list) else response[0]
        truth_class = truth if isinstance(truth, list) else truth[0]

        correct = response_class == truth_class

        if correct:
            metrics.append({"key":"seldon_metric_true_positive",
                             "type": "COUNTER", "value": 1,
                             "tags": { "class": f"CLASS_{truth_class}" }})
        else:
            metrics.append({"key":"seldon_metric_false_negative",
                             "type": "COUNTER", "value": 1,
                             "tags": { "class": f"CLASS_{truth_class}" }})
            metrics.append({"key":"seldon_metric_false_positive",
                             "type": "COUNTER", "value": 1,
                             "tags": { "class": f"CLASS_{response_class}" }})

            return SeldonResponse(None, None, metrics)

