import logging
import os
from urllib.parse import urlparse

import joblib
from minio import Minio

MODEL_URI = os.getenv("MODEL_URI", "")
MODEL_COMMIT_HASH = os.getenv("MODEL_COMMIT_HASH", "")
TEMP_DIR = os.getenv("TEMP_DIR", "/mnt/models")
_S3_PREFIX = "s3://"

logger = logging.getLogger(__name__)


class RSIModel(object):
    def __init__(self):
        super().__init__()
        self.ready = False
        logger.info(f"Model uri: {MODEL_URI} @ ${MODEL_COMMIT_HASH}")
        model_file = self._download_s3()
        logger.info(f"model file: {model_file}")
        self._model = joblib.load(model_file)

    def predict(self, X, features_names):
        """Predict!

        The return type is important. It has to fit into one of three types:
        https://docs.seldon.io/projects/seldon-core/en/v1.1.0/python/python_component.html

        The statsmodel model returns a pandas Series. I want to return the index, so I reset
        it to move it into a column, then use the to_numpy method to convert it to a valid
        seldon return type.

        Args:
            X (int): A scalar representing the number of periods to forecast.

        Returns:
            np.ndarray: The forecast [timestamps x predictions]
        """
        return self._model.forecast(X).reset_index().to_numpy()

    def init_metadata(self):
        meta = {
            "name": "RSIModel",
            "versions": [f"{MODEL_URI}@{MODEL_COMMIT_HASH}"],
            "platform": "seldon",
            "inputs": [
                {
                    "messagetype": "ndarray",
                    "schema": {"names": ["periods"], "shape": [1]},
                }
            ],
            "outputs": [{"messagetype": "ndarray", "schema": {"shape": [2]}}],
            "custom": {"MODEL_URI": MODEL_URI, "MODEL_COMMIT_HASH": MODEL_COMMIT_HASH},
        }
        return meta

    # Inspired by https://github.com/SeldonIO/kfserving/blob/master/python/kfserving/kfserving/storage.py
    def _download_s3(self) -> str:
        client = RSIModel._create_minio_client()
        bucket_args = MODEL_URI.replace(_S3_PREFIX, "", 1).split("/", 1)
        bucket_name = bucket_args[0]
        bucket_path = bucket_args[1] if len(bucket_args) > 1 else ""
        file_name = bucket_path.split("/")[-1]
        local_file = os.path.join(TEMP_DIR, file_name)
        logger.info(f"Downloading {bucket_path} from {bucket_name} to {local_file}.")
        try:
            client.fget_object(
                bucket_name, bucket_path, local_file, version_id=MODEL_COMMIT_HASH
            )
        except Exception as e:
            raise RuntimeError(f"Failed to fetch model {MODEL_URI}.\n{e}")
        return local_file

    @staticmethod
    def _create_minio_client():
        # Remove possible http scheme for Minio
        url = urlparse(os.getenv("AWS_ENDPOINT_URL", "s3.amazonaws.com"))
        use_ssl = (
            url.scheme == "https"
            if url.scheme
            else bool(os.getenv("S3_USE_HTTPS", "true"))
        )
        return Minio(
            url.netloc,
            access_key=os.getenv("AWS_ACCESS_KEY_ID", ""),
            secret_key=os.getenv("AWS_SECRET_ACCESS_KEY", ""),
            secure=use_ssl,
        )
