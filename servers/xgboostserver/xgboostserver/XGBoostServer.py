import numpy as np
import seldon_core
from seldon_core.user_model import SeldonComponent
from typing import Dict, List, Union, Iterable
import os
import yaml
import logging
import xgboost as xgb

logger = logging.getLogger(__name__)

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

    def init_metadata(self):
        file_path = os.path.join(self.model_uri, "metadata.yaml")

        try:
            with open(file_path, "r") as f:
                return yaml.safe_load(f.read())
        except FileNotFoundError:
            logger.debug(f"metadata file {file_path} does not exist")
            return {}
        except yaml.YAMLError:
            logger.error(
                f"metadata file {file_path} present but does not contain valid yaml"
            )
            return {}
