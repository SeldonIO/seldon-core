import os
import logging

import yaml
from typing import Dict, List, Union, Iterable

import numpy as np
import xgboost as xgb

import seldon_core
from seldon_core.user_model import SeldonComponent

logger = logging.getLogger(__name__)

BOOSTER_FILE = "model.json"
BOOSTER_FILE_DEPRECATED = "model.bst"


class XGBoostServer(SeldonComponent):
    def __init__(self, model_uri: str):
        super().__init__()
        self.model_uri = model_uri
        self.ready = False

    def load(self):
        model_file = os.path.join(
            seldon_core.Storage.download(self.model_uri), BOOSTER_FILE
        )
        if not os.path.exists(model_file):
            # Fallback to deprecated .bst format
            model_file = os.path.join(
                seldon_core.Storage.download(self.model_uri), BOOSTER_FILE_DEPRECATED
            )
            if os.path.exists(model_file):
                logger.warning(
                    "Using deprecated .bst format for XGBoost model. "
                    "Please update to the .json format in the future."
                )
            else:
                raise FileNotFoundError(f"Model file not found: {BOOSTER_FILE} or {BOOSTER_FILE_DEPRECATED}")

        # Load model using the new .json format for XGBoost >= 1.7.0
        self._booster = xgb.Booster()
        self._booster.load_model(model_file)

        # record expected feature count
        self._n_features = self._booster.num_features()

        logger.info(f"Loaded the model")
        self.ready = True

    def predict(
            self, X: np.ndarray, names: Iterable[str], meta: Dict = None
    ) -> Union[np.ndarray, List, str, bytes]:
        if not self.ready:
            self.load()

        if not isinstance(X, np.ndarray):
            raise ValueError("Input X must be a numpy array")

        if X.shape[1] != self._n_features:
            raise ValueError(
                f"Expected {self._n_features} features, got {X.shape[1]}"
            )

        dmatrix = xgb.DMatrix(X)
        result: np.ndarray = self._booster.predict(dmatrix)
        return result

    def init_metadata(self):
        file_path = os.path.join(self.model_uri, "metadata.yaml")

        try:
            with open(file_path, "r") as f:
                return yaml.safe_load(f.read())
        except FileNotFoundError:
            logger.debug(f"Metadata file {file_path} does not exist")
            return {}
        except yaml.YAMLError:
            logger.error(
                f"Metadata file {file_path} present but does not contain valid YAML"
            )
            return {}