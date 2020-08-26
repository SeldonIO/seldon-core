import kfserving
import numpy as np
from typing import Dict, List, Union, Iterable
import os
import logging
import joblib

JOBLIB_FILE = "model.joblib"
logger = logging.getLogger(__name__)

class SKLearnServer():
    def __init__(self, model_uri: str = None, method: str = "predict_proba"):
        super().__init__()
        self.model_uri = model_uri
        self.method = method
        self.ready = False
        logger.info(f"Model uri: {self.model_uri}")
        logger.info(f"method: {self.method}")
        self.load()

    def load(self):
        logger.info("load")
        model_file = os.path.join(
            kfserving.Storage.download(self.model_uri), JOBLIB_FILE
        )
        logger.info(f"model file: {model_file}")
        self._joblib = joblib.load(model_file)
        self.ready = True

    def predict(self, X: np.ndarray) -> Union[np.ndarray, List, str, bytes]:
        if not isinstance(X, np.ndarray):
            if isinstance(X,list):
                X = np.array(X)
            else:
                X = np.array([X])
        try:
            if not self.ready:
                self.load()
            if self.method == "predict_proba":
                logger.info("Calling predict_proba")
                result = self._joblib.predict_proba(X)
            else:
                logger.info("Calling predict")
                result = self._joblib.predict(X)
            return result
        except Exception as ex:
            logging.exception("Exception during predict")
