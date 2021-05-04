import json
from typing import List, Dict, Optional, Union
import logging
import numpy as np
from .numpy_encoder import NumpyEncoder
from alibi_detect.utils.saving import Data
from adserver.base import AlibiDetectModel, ModelResponse


def _append_drift_metrcs(metrics, drift, name):
    metric_found = drift.get(name)

    # Assumes metric_found is always float/int or list/np.array when not none
    if metric_found is not None:
        if not isinstance(metric_found, (list, np.ndarray)):
            metric_found = [metric_found]

        for i, instance in enumerate(metric_found):
            metrics.append(
                {
                    "key": f"seldon_metric_drift_{name}",
                    "value": instance,
                    "type": "GAUGE",
                    "tags": {"index": str(i)},
                }
            )


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

    def process_event(self, inputs: Union[List, Dict], headers: Dict) -> Optional[ModelResponse]:
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
            self.batch = np.concatenate((self.batch, X))

        if self.batch.shape[0] >= self.drift_batch_size:
            logging.info(
                "Running drift detection. Batch size is %d. Needed %d",
                self.batch.shape[0],
                self.drift_batch_size,
            )

            cd_preds = self.model.predict(self.batch)

            logging.info("Ran drift test")
            self.batch = None

            output = json.loads(json.dumps(cd_preds, cls=NumpyEncoder))

            metrics: List[Dict] = []
            drift = output.get("data")

            if drift:
                _append_drift_metrcs(metrics, drift, "is_drift")
                _append_drift_metrcs(metrics, drift, "distance")
                _append_drift_metrcs(metrics, drift, "p_val")
                _append_drift_metrcs(metrics, drift, "threshold")

            return ModelResponse(data=output, metrics=metrics)
        else:
            logging.info(
                "Not running drift detection. Batch size is %d. Need %d",
                self.batch.shape[0],
                self.drift_batch_size,
            )
            return None
