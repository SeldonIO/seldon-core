import numpy as np
import seldon_core
from seldon_core.user_model import SeldonComponent
from typing import Dict, List, Union, Iterable
import os
import xgboost as xgb

BOOSTER_FILE = "model.bst"


class XGBoostServer(SeldonComponent):
    def __init__(self, model_uri: str):
        super().__init__()
        self.model_uri = model_uri
        self.ready = False

    def load(self):
        model_file = os.path.join(
            seldon_core.Storage.download(self.model_uri), BOOSTER_FILE
        )
        self._booster = xgb.Booster(model_file=model_file)
        self.ready = True

    def predict(
        self, X: np.ndarray, names: Iterable[str], meta: Dict = None
    ) -> Union[np.ndarray, List, str, bytes]:
        if not self.ready:
            self.load()
        dmatrix = xgb.DMatrix(X)
        result: np.ndarray = self._booster.predict(dmatrix)
        return result
