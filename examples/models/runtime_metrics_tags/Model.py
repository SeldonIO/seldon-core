import logging

from seldon_core.user_model import SeldonResponse


def reshape(x):
    if len(x.shape) < 2:
        return x.reshape(1, -1)
    else:
        return x


class Model:
    def predict(self, features, names=[], meta={}):
        X = reshape(features)

        logging.info(f"model features: {features}")
        logging.info(f"model names: {names}")
        logging.info(f"model meta: {meta}")

        logging.info(f"model X: {X}")

        runtime_metrics = [{"type": "COUNTER", "key": "instance_counter", "value": len(X)}]
        runtime_tags = {"runtime": "tag", "shared": "right one"}
        return SeldonResponse(data=X, metrics=runtime_metrics, tags=runtime_tags)

    def metrics(self):
        return [{"type": "COUNTER", "key": "requests_counter", "value": 1}]

    def tags(self):
        return {"static": "tag", "shared": "not right one"}      
