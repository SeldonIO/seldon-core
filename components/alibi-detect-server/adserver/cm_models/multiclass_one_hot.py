from seldon_core.user_model import SeldonResponse
import numpy as np


class MulticlassOneHot:
    """
    MultiClassOneHot Model

    Parameters
    -----------
    """

    def __init__(self):
        pass

    def transform(self, truth, response, request=None):
        """
        Perform a multiclass one-hot comparison between truth and response.

        Parameters
        -----------
        truth
            Actual data value as format of an array (or array of arrays) of the form of one-hot or probabilities.
        response
            Prediction data value as format of an array (or array of arrays) of the form of one-hot or probabilities.
        request
            Input data value as format of an array (or array of arrays) of the form of one-hot or probabilities.
        """

        metrics = []
        response = (
            response[0] if isinstance(response[0], (list, np.ndarray)) else response
        )
        truth = truth[0] if isinstance(truth[0], (list, np.ndarray)) else truth
        response_class = max(enumerate(response), key=lambda x: x[1])[0]
        truth_class = max(enumerate(truth), key=lambda x: x[1])[0]

        correct = response_class == truth_class

        if correct:
            metrics.append(
                {
                    "key": "seldon_metric_true_positive",
                    "type": "COUNTER",
                    "value": 1,
                    "tags": {"class": f"CLASS_{truth_class}"},
                }
            )
        else:
            metrics.append(
                {
                    "key": "seldon_metric_false_negative",
                    "type": "COUNTER",
                    "value": 1,
                    "tags": {"class": f"CLASS_{truth_class}"},
                }
            )
            metrics.append(
                {
                    "key": "seldon_metric_false_positive",
                    "type": "COUNTER",
                    "value": 1,
                    "tags": {"class": f"CLASS_{response_class}"},
                }
            )

        return SeldonResponse(None, None, metrics)
