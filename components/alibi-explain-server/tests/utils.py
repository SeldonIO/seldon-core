import logging
import os
import tempfile
from typing import List, Union

import gcsfs
import joblib
import numpy as np

JOBLIB_FILE = "model.joblib"
logger = logging.getLogger(__name__)


def download_from_gcs(gcs_uri: str, dirname: str) -> None:
    # Parse model_uri to extract the prefix (folder path)
    # e.g. gs://seldon-models/sklearn/iris-0.23.2/lr_model/*
    bucket, prefix = gcs_uri[5:].split("/", 1)
    prefix = prefix.rstrip("*").rstrip("/")  # strip trailing /* if present

    fs = gcsfs.GCSFileSystem(token="anon")  # public / anonymous access

    files = fs.find(f"{bucket}/{prefix}")
    for path in files:
        if path.endswith("/"):
            continue  # skip directory placeholders

        # Compute local destination preserving relative folder structure
        rel_path = os.path.relpath(path, f"{bucket}/{prefix}")
        local_path = os.path.join(dirname, rel_path)

        os.makedirs(os.path.dirname(local_path), exist_ok=True)

        with fs.open(path, "rb") as src, open(local_path, "wb") as dst:
            dst.write(src.read())


class SKLearnServer:
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
        with tempfile.TemporaryDirectory() as model_dir:
            download_from_gcs(self.model_uri, model_dir)
            model_file = os.path.join(model_dir, JOBLIB_FILE)
            logger.info(f"model file: {model_file}")
            self._joblib = joblib.load(model_file)
        self.ready = True

    def predict(self, X: np.ndarray) -> Union[np.ndarray, List, str, bytes]:
        if not isinstance(X, np.ndarray):
            if isinstance(X, list):
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
