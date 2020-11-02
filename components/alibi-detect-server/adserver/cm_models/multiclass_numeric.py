from seldon_core.user_model import SeldonResponse
import numpy as np


class MultiClassNumeric:
    """
    MultiClassNumeric Model

    Parameters
    -----------
    """

    def __init__(self):
        pass

    def transform(self, truth, response, request=None):
        """
        Perform a multiclass numeric comparison between truth and response.

        Parameters
        -----------
        truth
            Actual data value as format of <number> or [<number>]
        response
            Prediction data value as format of <number> or [<number>]
        request
            Input data value as format of <number> or [<number>]
        """

        metrics = []

        response_class = (
            response[0] if isinstance(response, (list, np.ndarray)) else response
        )
        truth_class = truth[0] if isinstance(truth, (list, np.ndarray)) else truth

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
