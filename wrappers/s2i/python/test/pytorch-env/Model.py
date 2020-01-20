import torch
from torch.nn import functional as F
from seldon_core.user_model import SeldonComponent


class Model(SeldonComponent):
    def predict(self, X, meta):
        tensor = torch.from_numpy(X)
        normalised = F.softmax(tensor)
        return normalised.numpy()
