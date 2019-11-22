import numpy as np
import logging
import requests
from mlflow import pyfunc
from seldon_core import Storage
from seldon_core.user_model import SeldonComponent
from typing import Dict, List, Union, Iterable

log = logging.getLogger()

MLFLOW_SERVER = "model"


class MLFlowServer(SeldonComponent):
    def __init__(self, model_uri: str):
        super().__init__()
        log.info(f"Creating MLFLow server with URI {model_uri}")
        self.model_uri = model_uri
        self.ready = False

    def load(self):
        log.info(f"Downloading model from {self.model_uri}")
        model_folder = Storage.download(self.model_uri)
        self._model = pyfunc.load_model(model_folder)
        self.ready = True

    def predict(
        self, X: np.ndarray, feature_names: Iterable[str] = [], meta: Dict = None
    ) -> Union[np.ndarray, List, Dict, str, bytes]:
        log.info(f"Requesting prediction with: {X}")

        if not self.ready:
            raise requests.HTTPError("Model not loaded yet")

        result = self._model.predict(X)
        log.info(f"Prediction result: {result}")
        return result
