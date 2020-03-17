import torch

from seldon_core.user_model import SeldonComponent
from seldon_core import Storage

import os
import logging

logger = logging.getLogger(__name__)


MODEL_FILE_NAME = "model.pt"


class TorchServer(SeldonComponent):
    def __init__(self, model_uri):
        super().__init__()

        self.model_uri = model_uri
        self.ready = False

        self.load()

    def load(self):
        model_file = os.path.join(Storage.download(self.model_uri), MODEL_FILE_NAME)
        self._model = torch.load(model_file)
        self.ready = True

    def predict(self, X, names, meta):
        if not self.ready:
            self.load()

        result = self._model(X)

        return result
