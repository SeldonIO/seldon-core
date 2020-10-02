import logging
import time
import os

import ray

import numpy as np
from seldon_core.utils import getenv_as_bool


RAY_PROXY = getenv_as_bool("RAY_PROXY", default=False)
MODEL_FILE = "/microservice/pytorch_model.bin"

BATCH_SIZE = int(os.environ.get("BATCH_SIZE", "100"))
NUM_ACTORS = int(os.environ.get("NUM_ACTORS", "10"))


class RobertaModel:
    def __init__(self, load_on_init=False):
        if load_on_init:
            self.load()

    def load(self):
        import torch
        from simpletransformers.model import TransformerModel

        logging.info("starting RobertaModel...")
        model = TransformerModel(
            "roberta",
            "roberta-base",
            args=({"fp16": False, "use_multiprocessing": False}),
            use_cuda=False,
        )
        model.model.load_state_dict(torch.load(MODEL_FILE))
        self.model = model
        logging.info("... started RobertaModel")

    def predict(self, data, names=[], meta={}):
        logging.info(f"received inference request: {data}")
        data = data.astype("U")
        output = self.model.predict(data)[1].argmax(axis=1)
        logging.info("finished calculating prediction")
        return output


class ProxyModel:
    def load(self):
        ray.init(address="auto")

        self.actors = [
            ray.remote(RobertaModel).remote(load_on_init=True)
            for _ in range(NUM_ACTORS)
        ]

        self.pool = ray.util.ActorPool(self.actors)

    def predict(self, data, names=[], meta=[]):
        logging.info(f"data received: {data}")
        batches = np.array_split(data, max(data.shape[0] // BATCH_SIZE, 1))
        logging.info(f"spllited into {len(batches)} batches")

        t1 = time.perf_counter()
        results = list(self.pool.map(lambda a, v: a.predict.remote(v), batches))
        results = np.concatenate(results).tolist()
        t2 = time.perf_counter()

        return {"time-taken": t2 - t1, "results": results}


if not RAY_PROXY:
    logging.info("Model = RobertaModel")
    Model = RobertaModel
else:
    logging.info("Model = ProxyModel")
    Model = ProxyModel
