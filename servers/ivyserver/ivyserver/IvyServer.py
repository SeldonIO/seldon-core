import yaml
import os
import logging
import requests
import ivy

from seldon_core import Storage
from seldon_core.user_model import SeldonComponent, SeldonNotImplementedError
from typing import Dict, List, Union

logger = logging.getLogger()

IVY_SERVER = "model"

class IvyServer(SeldonComponent):
    def __init__(self, model_uri: str, xtype: str = "ivy.array"):
        super().__init__()
        logger.info(f"Creating Ivy server with URI {model_uri}")
        logger.info(f"xtype: {xtype}")
        self.model_uri = model_uri
        self.xtype = xtype
        self.ready = False
        self.column_names = None
    
    def load(self):
        logger.info(f"Downloading model from {self.model_uri}")
        #specify a local folder to get model with file://
        model_folder = Storage.download(self.model_uri)
        #need a way to load the model from the model folder
        self._model = ivy.load(model_folder)
        self.ready = True

    def predict(
            self,
            X,
            meta: Dict = None
    ) -> Union[ivy.array, List, Dict, str, bytes]:
        
        logger.debug(f"Requesting prediction with: {X}")

        if not self.ready:
            raise requests.HTTPError("Model not loaded yet")
        
        if self.xtype == "ivy.array":
            result = self._model.predict(X)
        else:
            X = ivy.array(X)
            result = self._model.predict(X)
        
        logger.debug(f"Prediction result: {result}")
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