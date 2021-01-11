import kfserving
from typing import Optional
from adserver.base.model import CEModel
from alibi_detect.utils.saving import load_detector, Data


class AlibiDetectModel(CEModel):  # pylint:disable=c-extension-no-member
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
        self.model: Optional[Data] = model

    def load(self):
        """
        Load the model from storage

        """
        model_folder = kfserving.Storage.download(self.storage_uri)
        self.model: Data = load_detector(model_folder)
        self.ready = True

