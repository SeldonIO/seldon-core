import logging
import os



NODE = os.environ.get("NODE_NAME", "default")

ALL_MDOELS_META = {
    "default": {
        "name": "model-name",
        "versions": ["model-version"],
        "platform": "platform-name",
        "inputs": [{"name": "input", "datatype": "BYTES", "shape": [1]}],
        "outputs": [{"name": "output", "datatype": "BYTES", "shape": [5]}],
    },
    "Model 1": {
        "name": "Model 1",
        "versions": ["model-version"],
        "platform": "platform-name",
        "inputs": [{"name": "input", "datatype": "BYTES", "shape": [1, 5]}],
        "outputs": [{"name": "output", "datatype": "BYTES", "shape": [1, 3]}],
    },
    "Model 2": {
        "name": "Model 2",
        "versions": ["model-version"],
        "platform": "platform-name",
        "inputs": [{"name": "input", "datatype": "BYTES", "shape": [1, 3]}],
        "outputs": [{"name": "output", "datatype": "BYTES", "shape": [3]}],
    },
    "Model A1": {
        "name": "Model A1",
        "versions": ["model-version"],
        "platform": "platform-name",
        "inputs": [{"name": "input", "datatype": "BYTES", "shape": [1, 5]}],
        "outputs": [{"name": "output", "datatype": "BYTES", "shape": [1, 10]}],
    },
    "Model A2": {
        "name": "Model A2",
        "versions": ["model-version"],
        "platform": "platform-name",
        "inputs": [{"name": "input", "datatype": "BYTES", "shape": [1, 5]}],
        "outputs": [{"name": "output", "datatype": "BYTES", "shape": [1, 20]}],
    },
    "Model Combiner": {
        "name": "Model Combiner",
        "versions": ["model-version"],
        "platform": "platform-name",
        "inputs": [
            {"name": "input-1", "datatype": "BYTES", "shape": [1, 10]},
            {"name": "input-2", "datatype": "BYTES", "shape": [1, 20]},
        ],
        "outputs": [{"name": "combined output", "datatype": "BYTES", "shape": [3]}],
    },
}

METADATA = ALL_MDOELS_META[NODE]


class Model:
    def predict(self, features, names=[], meta=[]):
        logging.info(f"model features: {features}")
        logging.info(f"model names: {names}")
        logging.info(f"model meta: {meta}")
        return features.tolist()

    def init_metadata(self):
        return METADATA
