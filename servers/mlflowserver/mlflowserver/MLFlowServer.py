import numpy as np
import logging
import requests
from mlflow import pyfunc
from seldon_core import Storage
from seldon_core.user_model import SeldonComponent
from typing import Dict, List, Union, Iterable
import yaml
import os

logger = logging.getLogger()

MLFLOW_SERVER = "model"


class MLFlowServer(SeldonComponent):
    def __init__(self, model_uri: str):
        super().__init__()
        logger.info(f"Creating MLFLow server with URI {model_uri}")
        self.model_uri = model_uri
        self.ready = False

    def load(self):
        logger.info(f"Downloading model from {self.model_uri}")
        model_folder = Storage.download(self.model_uri)
        self._model = pyfunc.load_model(model_folder)
        self.ready = True

    def predict(
        self, X: np.ndarray, feature_names: Iterable[str] = [], meta: Dict = None
    ) -> Union[np.ndarray, List, Dict, str, bytes]:
        logger.info(f"Requesting prediction with: {X}")

        if not self.ready:
            raise requests.HTTPError("Model not loaded yet")

        result = self._model.predict(X)
        logger.info(f"Prediction result: {result}")
        return result

    def init_metadata(self):
        file_path = os.path.join(self.model_uri, "metadata.yaml")

        try:
            with open(file_path, "r") as f:
                return yaml.load(f.read())
        except FileNotFoundError:
            logger.debug(f"metadata file {file_path} does not exist")
            return {}
        except yaml.YAMLError:
            logger.error(
                f"metadata file {file_path} present but does not contain valid yaml"
            )
            return {}
