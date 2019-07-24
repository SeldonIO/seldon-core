import joblib
import numpy as np
import seldon_core
from seldon_core.user_model import SeldonComponent
from typing import Dict, List, Union, Iterable
import os

JOBLIB_FILE = "model.joblib"

class SKLearnServer(SeldonComponent):
    def __init__(self, model_uri: str):
        super().__init__()
        self.model_uri = model_uri
        self.ready = False

    def load(self):
        print("load")
        model_file = os.path.join(seldon_core.Storage.download(self.model_uri), JOBLIB_FILE)
        print("model file",model_file)
        self._joblib = joblib.load(model_file)
        self.ready = True

    def predict(self, X: np.ndarray, names: Iterable[str], meta: Dict = None) -> Union[np.ndarray, List, str, bytes]:
        print("predict")
        if not self.ready:
            self.load()
        result = self._joblib.predict(X)
        return result
