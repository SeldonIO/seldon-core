import json
from typing import List, Dict, Optional, Union
import logging
import numpy as np
from adserver.constants import HEADER_RETURN_INSTANCE_SCORE
from .numpy_encoder import NumpyEncoder
from alibi_detect.utils.saving import load_detector, Data
from adserver.base import CEModel, ModelResponse
from adserver.base.storage import download_model


class AlibiDetectAdversarialDetectionModel(
    CEModel
):  # pylint:disable=c-extension-no-member
    def __init__(self, name: str, storage_uri: str, model: Optional[Data] = None):
        """
        Outlier Detection / Concept Drift Model

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
        model_folder = download_model(self.storage_uri)
        self.model: Data = load_detector(model_folder)
        self.ready = True

        # or create

    def process_event(self, inputs: Union[List, Dict], headers: Dict) -> ModelResponse:
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

        ret_instance_score = True
        if (
            HEADER_RETURN_INSTANCE_SCORE in headers
            and headers[HEADER_RETURN_INSTANCE_SCORE] == "false"
        ):
            ret_instance_score = False

        ad_preds = self.model.predict(X, return_instance_score=ret_instance_score)

        data =  json.loads(json.dumps(ad_preds, cls=NumpyEncoder))
        return ModelResponse(data=data, metrics=None)
