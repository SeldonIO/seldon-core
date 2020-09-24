from seldon_core.user_model import SeldonResponse

class MulticlassOneHot:
    def __init__(self):
        pass

    def transform(self, truth, response, request = None):

        metrics = []
        response = response[0] if isinstance(response[0], list) else response
        truth = truth[0] if isinstance(truth[0], list) else truth
        response_class = max(enumerate(response),key=lambda x: x[1])[0]
        truth_class = max(enumerate(truth),key=lambda x: x[1])[0]

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

