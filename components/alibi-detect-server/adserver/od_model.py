import json
from typing import List, Dict, Optional, Union
import logging
import kfserving
import numpy as np
from .numpy_encoder import NumpyEncoder
from adserver.base import CEModel
from alibi_detect.utils.saving import load_detector, Data
from adserver.constants import (
    HEADER_RETURN_INSTANCE_SCORE,
    HEADER_RETURN_FEATURE_SCORE,
    HEADER_OUTLIER_TYPE,
)


class AlibiDetectOutlierModel(CEModel):  # pylint:disable=c-extension-no-member
    def __init__(self, name: str, storage_uri: str, model: Optional[Data] = None):
        """
        Outlier Detection Model

        Parameters
        ----------
        name
             The name of the model
        storage_uri
             The URI location of the model
        """
        super().__init__(name)
        self.name = name
        self.storage_uri = storage_uri
        self.ready = False
        self.model: Data = model

    def load(self):
        """
        Load the model from storage

        """
        model_folder = kfserving.Storage.download(self.storage_uri)
        self.model: Data = load_detector(model_folder)
        self.ready = True

    def process_event(self, inputs: Union[List, Dict], headers: Dict) -> Dict:
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

        ret_instance_score = False
        if (
            HEADER_RETURN_INSTANCE_SCORE in headers
            and headers[HEADER_RETURN_INSTANCE_SCORE] == "true"
        ):
            ret_instance_score = True

        outlier_type = "instance"
        if HEADER_OUTLIER_TYPE in headers and headers[HEADER_OUTLIER_TYPE]:
            outlier_type = headers[HEADER_OUTLIER_TYPE]
        ret_feature_score = False
        if (
            HEADER_RETURN_FEATURE_SCORE in headers
            and headers[HEADER_RETURN_FEATURE_SCORE] == "true"
        ):
            ret_feature_score = True
        od_preds = {}
        name = self.model.meta["name"]
        if (
            name == "IForest"
            or name == "OutlierAEGMM"
            or name == "Mahalanobis"
            or name == "SpectralResidual"
            or name == "OutlierVAEGMM"
        ):
            od_preds = self.model.predict(
                X,
                # scores used to determine outliers
                return_instance_score=ret_instance_score,
            )
        else:
            od_preds = self.model.predict(
                X,
                outlier_type=outlier_type,
                # use 'feature' or 'instance' level
                return_feature_score=ret_feature_score,
                # scores used to determine outliers
                return_instance_score=ret_instance_score,
            )
        # clean result
        if (
            "data" in od_preds
            and "instance_score" in od_preds["data"]
            and od_preds["data"]["instance_score"] is None
        ):
            del od_preds["data"]["instance_score"]
        if (
            "data" in od_preds
            and "feature_score" in od_preds["data"]
            and od_preds["data"]["feature_score"] is None
        ):
            del od_preds["data"]["feature_score"]

        return json.loads(json.dumps(od_preds, cls=NumpyEncoder))
