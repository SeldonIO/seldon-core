from prometheus_client.core import (
    HistogramMetricFamily,
    GaugeMetricFamily,
    CounterMetricFamily,
    CollectorRegistry,
)
from prometheus_client import exposition
from prometheus_client.utils import floatToGoString

from multiprocessing import Manager

import numpy as np

from typing import List, Dict
import logging
import json
import os


logger = logging.getLogger(__name__)

NONIMPLEMENTED_MSG = "NOT_IMPLEMENTED"

ENV_SELDON_DEPLOYMENT_NAME = "SELDON_DEPLOYMENT_ID"
ENV_MODEL_NAME = "PREDICTIVE_UNIT_ID"
ENV_MODEL_IMAGE = "PREDICTIVE_UNIT_IMAGE"
ENV_PREDICTOR_NAME = "PREDICTOR_ID"
ENV_PREDICTOR_LABELS = "PREDICTOR_LABELS"

COUNTER = "COUNTER"
GAUGE = "GAUGE"
TIMER = "TIMER"

# This sets the bins spread logarithmically between 0.001 and 30
BINS = [0] + list(np.logspace(-3, np.log10(30), 50)) + [np.inf]

# Development placeholder
image = os.environ.get(ENV_MODEL_IMAGE, f"{NONIMPLEMENTED_MSG}:{NONIMPLEMENTED_MSG}")
image_name, image_version = image.split(":")
predictor_version = json.loads(os.environ.get(ENV_PREDICTOR_LABELS, "{}")).get(
    "version", f"{NONIMPLEMENTED_MSG}"
)

DEFAULT_LABELS = {
    "seldon_deployment_name": os.environ.get(
        ENV_SELDON_DEPLOYMENT_NAME, f"{NONIMPLEMENTED_MSG}"
    ),
    "model_name": os.environ.get(ENV_MODEL_NAME, f"{NONIMPLEMENTED_MSG}"),
    "image_name": image_name,
    "image_version": image_version,
    "predictor_name": os.environ.get(ENV_PREDICTOR_NAME, f"{NONIMPLEMENTED_MSG}"),
    "predictor_version": predictor_version,
}


class SeldonMetrics:
    """Class to manage custom metrics stored in shared memory."""

    def __init__(self, worker_id_func=os.getpid):
        # We keep reference to Manager so it does not get garbage collected
        self._manager = Manager()
        self._lock = self._manager.Lock()
        self.data = self._manager.dict()
        self.worker_id_func = worker_id_func

    def __del__(self):
        self._manager.shutdown()

    def update(self, custom_metrics):
        # Read a corresponding worker's metric data with lock as Proxy objects
        # are not thread-safe, see "Thread safety of proxies" here
        # https://docs.python.org/3.7/library/multiprocessing.html#programming-guidelines
        logger.debug("Updating metrics: {}".format(custom_metrics))
        with self._lock:
            data = self.data.get(self.worker_id_func(), {})
        logger.debug("Read current metrics data from shared memory")

        for metrics in custom_metrics:
            metrics_type = metrics.get("type", "COUNTER")
            key = metrics_type, metrics["key"]
            tags = metrics.get("tags", {})
            if metrics_type == "COUNTER":
                value = data.get(key, {}).get("value", 0)
                data[key] = {"value": value + metrics["value"], "tags": tags}
            elif metrics_type == "TIMER":
                vals, sumv = data.get(key, {}).get(
                    "value", (list(np.zeros(len(BINS) - 1)), 0)
                )
                # Dividing by 1000 because unit is milliseconds
                data[key] = {
                    "value": self._update_hist(metrics["value"] / 1000, vals, sumv),
                    "tags": tags,
                }
            elif metrics_type == "GAUGE":
                data[key] = {"value": metrics["value"], "tags": tags}
            else:
                logger.error(f"Unkown metrics type: {metrics_type}")

        # Write worker's data with lock (again - Proxy objects are not thread-safe)
        with self._lock:
            self.data[self.worker_id_func()] = data
        logger.debug("Updated metrics in the shared memory.")

    def collect(self):
        # Read all workers metrics with lock to avoid other processes / threads
        # writing to it at the same time. Casting to `dict` works like reading of data.
        logger.debug("SeldonMetrics.collect called")
        with self._lock:
            data = dict(self.data)
        logger.debug("Read current metrics data from shared memory")

        for worker, metrics in data.items():
            for (item_type, item_name), item in metrics.items():
                labels_keys, labels_values = self._merge_labels(
                    str(worker), item["tags"]
                )
                if item_type == "GAUGE":
                    yield self._expose_gauge(
                        item_name, item["value"], labels_keys, labels_values
                    )
                elif item_type == "COUNTER":
                    yield self._expose_counter(
                        item_name, item["value"], labels_keys, labels_values
                    )
                elif item_type == "TIMER":
                    yield self._expose_histogram(
                        item_name, item["value"], labels_keys, labels_values
                    )

    def generate_metrics(self):
        myregistry = CollectorRegistry()
        myregistry.register(self)
        return (
            exposition.generate_latest(myregistry).decode("utf-8"),
            exposition.CONTENT_TYPE_LATEST,
        )

    @staticmethod
    def _merge_labels(worker, tags):
        labels = {**tags, **DEFAULT_LABELS, "worker_id": str(worker)}
        return list(labels.keys()), list(labels.values())

    @staticmethod
    def _update_hist(x, vals, sumv):
        hist = np.histogram([x], BINS)[0]
        vals = list(np.array(vals) + hist)
        return vals, sumv + x

    @staticmethod
    def _expose_gauge(name, value, labels_keys, labels_values):
        metric = GaugeMetricFamily(name, "", labels=labels_keys)
        metric.add_metric(labels_values, value)
        return metric

    @staticmethod
    def _expose_counter(name, value, labels_keys, labels_values):
        metric = CounterMetricFamily(name, "", labels=labels_keys)
        metric.add_metric(labels_values, value)
        return metric

    @staticmethod
    def _expose_histogram(name, value, labels_keys, labels_values):
        vals, sumv = value
        buckets = [[floatToGoString(b), v] for v, b in zip(np.cumsum(vals), BINS[1:])]

        metric = HistogramMetricFamily(name, "", labels=labels_keys)
        metric.add_metric(labels_values, buckets, sum_value=sumv)
        return metric


def create_counter(key: str, value: float):
    """
    Utility method to create a counter metric
    Parameters
    ----------
    key
       Counter name
    value
       Counter value

    Returns
    -------
       Valid counter metric dict

    """
    test = value + 1
    return {"key": key, "type": COUNTER, "value": value}


def create_gauge(key: str, value: float) -> Dict:
    """
    Utility method to create a guage metric
    Parameters
    ----------
    key
      Guage name
    value
      Guage value

    Returns
    -------
       Valid Guage metric dict

    """
    test = value + 1
    return {"key": key, "type": GAUGE, "value": value}


def create_timer(key: str, value: float) -> Dict:
    """
    Utility mehtod to create a timer metric
    Parameters
    ----------
    key
      Name of metric
    value
      Value for metric

    Returns
    -------
       Valid timer metric dict

    """
    test = value + 1
    return {"key": key, "type": TIMER, "value": value}


def validate_metrics(metrics: List[Dict]) -> bool:
    """
    Validate a list of metrics
    Parameters
    ----------
    metrics
       List of metrics

    Returns
    -------

    """
    if isinstance(metrics, (list,)):
        for metric in metrics:
            if not ("key" in metric and "value" in metric and "type" in metric):
                return False
            if not (
                metric["type"] == COUNTER
                or metric["type"] == GAUGE
                or metric["type"] == TIMER
            ):
                return False
            try:
                metric["value"] + 1
            except TypeError:
                return False
    else:
        return False
    return True
