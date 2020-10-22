import json
from typing import List, Dict, Optional
import logging
import numpy as np
from .numpy_encoder import NumpyEncoder
from alibi_detect.utils.saving import load_detector, Data
from adserver.base import AlibiDetectModel
from seldon_core.user_model import SeldonResponse


def _drift_to_metrics(drift):
    metrics = []

    batch_score = drift.get("batch_score")
    if batch_score:
        metrics.append(
            {
                "key": "seldon_metric_drift_batch_score",
                "value": batch_score,
                "type": "GAUGE",
            }
        )

    feature_score = drift.get("feature_score")
    if feature_score:
        metrics.append(
            {
                "key": "seldon_metric_drift_feature_score",
                "value": feature_score,
                "type": "COUNTER",
            }
        )

    is_drift = drift.get("is_drift")
    if is_drift:
        metrics.append(
            {
                "key": "seldon_metric_drift_is_drift",
                "value": is_drift,
                "type": "COUNTER",
            }
        )

    p_val = drift.get("p_val")
    if p_val and isinstance(p_val, list):
        for i, p in enumerate(p_val):
            metrics.append(
                {
                    "key": "seldon_metric_drift_p_val",
                    "value": p,
                    "type": "COUNTER",
                    "tags": {"p_val_index": str(i)},
                }
            )

    return metrics


class AlibiDetectConceptDriftModel(
    AlibiDetectModel
):  # pylint:disable=c-extension-no-member
    def __init__(
        self,
        name: str,
        storage_uri: str,
        model: Optional[Data] = None,
        drift_batch_size: int = 1000,
    ):
        """
        Outlier Detection / Concept Drift Model

        Parameters
        ----------
        name
             The name of the model
        storage_uri
             The URI location of the model
        drift_batch_size
             The batch size to fill before checking for drift
        model
             Alibi detect model
        """
        super().__init__(name, storage_uri, model)
        self.drift_batch_size = drift_batch_size
        self.batch: np.array = None
        self.model: Data = model

    def process_event(self, inputs: List, headers: Dict) -> Optional[Dict]:
        """
        Process the event and return Alibi Detect score

        Parameters
        ----------
        inputs
             Input data
        headers
             Header options

        Returns
        -------
             Alibi Detect response

        """
        logging.info("PROCESSING EVENT.")
        logging.info(str(headers))
        logging.info("----")
        try:
            X = np.array(inputs)
        except Exception as e:
            raise Exception(
                "Failed to initialize NumPy array from inputs: %s, %s" % (e, inputs)
            )

        if self.batch is None:
            self.batch = X
        else:
            self.batch = np.vstack((self.batch, X))

        if self.batch.shape[0] >= self.drift_batch_size:
            logging.info(
                "Running drift detection. Batch size is %d. Needed %d",
                self.batch.shape[0],
                self.drift_batch_size,
            )
            cd_preds = self.model.predict(self.batch)
            self.batch = None

            output = json.loads(json.dumps(cd_preds, cls=NumpyEncoder))

            metrics = _drift_to_metrics(output.get("data", {}))

            seldon_response = SeldonResponse(output, None, metrics)

            return seldon_response
        else:
            logging.info(
                "Not running drift detection. Batch size is %d. Need %d",
                self.batch.shape[0],
                self.drift_batch_size,
            )
            return None
