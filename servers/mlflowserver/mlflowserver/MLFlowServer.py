from mlflow import pyfunc
import seldon_core
from seldon_core.user_model import SeldonComponent
from typing import Dict, List, Union, Iterable
import numpy as np
import os
import logging
import requests
import pandas as pd

log = logging.getLogger()

MLFLOW_SERVER = "model"

class MLFlowServer(SeldonComponent):

    def __init__(self, model_uri: str):
        super().__init__()
        log.info(f"Creating MLFLow server with URI: {model_uri}")
        self.model_uri = model_uri
        self.ready = False

    def load(self):
        log.info(f"Downloading model from {self.model_uri}")
        model_file = seldon_core.Storage.download(self.model_uri)
        self._model = pyfunc.load_model(model_file)
        self.ready = True

    def predict(
                self,
                X: np.ndarray,
                feature_names: Iterable[str] = [],
                meta: Dict = None
            ) -> Union[np.ndarray, List, Dict, str, bytes]:

        log.info(f"Requesting prediction with: {X}")
        if not self.ready:
            self.load()
            # TODO: Make sure this doesn't get called from here, but 
            #   from the actual python wrapper. Raise exception instead
            #raise requests.HTTPError("Model not loaded yet")
        #if not feature_names is None and len(feature_names)>0:
        #    df = pd.DataFrame(data=X, columns=feature_names)
        #else:
        #    df = pd.DataFrame(data=X)
        result = self._model.predict(X)
        log.info(f"Prediction result: {result}")
        return result

