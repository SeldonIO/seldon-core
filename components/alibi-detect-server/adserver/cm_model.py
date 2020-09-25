import json
from typing import List, Dict, Optional, Union
import logging
import numpy as np
from enum import Enum
import kfserving
import importlib
import pickle
import os
from adserver.constants import HEADER_RETURN_INSTANCE_SCORE
from .numpy_encoder import NumpyEncoder
from adserver.base import CEModel
from seldon_core.user_model import SeldonResponse
from seldon_core.flask_utils import SeldonMicroserviceException

def _load_class_module(module_path: str) -> str:
    components = module_path.split(".")
    mod = __import__(".".join(components[:-1]))
    for comp in components[1:]:
        print(mod, comp)
        mod = getattr(mod, comp)
    return mod


class CustomMetricsModel(CEModel):  # pylint:disable=c-extension-no-member
    def __init__(self, name: str, storage_uri: str, model = None):
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
        self.model = model
        self.ready = False

    def load(self):
        """
        Load the model from storage

        """
        if "/" in self.storage_uri:
            model_folder = kfserving.Storage.download(self.storage_uri)
            self.model = pickle.load(open(os.path.join(model_folder, 'meta.pickle'), 'rb'))
        else:
            # Load from locally available models
            MetricsClass = _load_class_module(self.storage_uri)
            self.model = MetricsClass()

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
        logging.info("PROCESSING Feedback Event.")
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
            response = inputs["response"]
            truth = inputs["truth"]
            r = self.model.transform(truth, response)
            metrics.extend(r.metrics)

        seldon_response = SeldonResponse(output or None, None, metrics)

        return seldon_response

