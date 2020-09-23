import json
from typing import List, Dict, Optional, Union
import logging
import numpy as np
from adserver.constants import HEADER_RETURN_INSTANCE_SCORE
from .numpy_encoder import NumpyEncoder
from alibi_detect.utils.saving import load_detector, Data
from adserver.base import CEModel
from seldon_core.user_model import SeldonResponse
from seldon_core.flask_utils import SeldonMicroserviceException


class CustomMetricsModel(CEModel):  # pylint:disable=c-extension-no-member
    def __init__(self, name: str, storage_uri: str, model: Optional[Data] = None):
        """
        Outlier Detection / Concept Drift Model

        Parameters
        ----------
        name
             The name of the model
        storage_uri
             The URI location of the model
        """
        super().__init__(name)
        self.name = name
        self.storage_uri = storage_uri
        self.ready = False
        self.model: Data = model

    def load(self):
        """
        Load the model from storage

        """
        self.ready = True

    def process_event(self, inputs: Union[List, Dict], headers: Dict) -> Dict:
        """
        Process the event and return Alibi Detect score

        Parameters
        ----------
        inputs
             Input data
        headers
             Header options

        Returns
        -------
             Alibi Detect response

        """
        logging.info("PROCESSING EVENT.")
        logging.info(str(headers))
        logging.info("----")

        metrics = []
        output = {}

        if "truth" not in inputs:
            raise SeldonMicroserviceException(
                    f"No truth value provided in: {json.dumps(inputs)}",
                    status_code=400,
                    reason="NO_TRUTH_VALUE")

        # We automatically add any metrics provided in the truth
        if "metrics" in inputs:
            metrics.extend(inputs["metrics"])

        # If response is provided then we can perform a comparison
        # TODO: If Header UUID provided we could fetch from ELK to do the evaluation
        if "response" in inputs:
            # TODO: Add the extensions for the comparisons here
            response = inputs["response"]
            truth = inputs["truth"]

            if self.name == "BINARY_CLASSIFICATION":
                response_class = int(response) if isinstance(response, list) else int(response[0])
                truth_class = int(truth) if isinstance(truth, list) else int(truth[0])

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

                metrics.append({"key":key, "type": "COUNTER", "value": 1})

            else:
                if self.name == "MULTICLASS_CLASSIFICATION_ONE_HOT":
                    # TODO: Perform check that input is list
                    response = response if isinstance(response[0], list) else response[0]
                    truth = truth if isinstance(truth[0], list) else truth[0]
                    response_class = max(enumerate(response),key=lambda x: x[1])[0]
                    truth_class = max(enumerate(truth),key=lambda x: x[1])[0]

                elif self.name == "MULTICLASS_CLASSIFICATION_NUMERIC":
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


        # TODO: Allow for scores to be used to calculate metrics as well
        seldon_response = SeldonResponse(output or None, None, metrics)

        return seldon_response

