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
import os


logger = logging.getLogger(__name__)


COUNTER = "COUNTER"
GAUGE = "GAUGE"
TIMER = "TIMER"

# This sets the bins spread logarithmically between 0.001 and 30
BINS = [0] + list(np.logspace(-3, np.log10(30), 50)) + [np.inf]

# Development placeholder
LABELS = ["worker-id", "model", "image"]
my_labels = {"model": "latest", "image": "my-image"}


class SeldonMetrics:
    """Class to manage custom metrics stored in shared memory."""

    def __init__(self, worker_id_func=os.getpid):
        # We keep reference to Manager so it does not get garbage collected
        self._manager = Manager()
        self.data = self._manager.dict()
        self.worker_id_func = worker_id_func

    def __del__(self):
        self._manager.shutdown()

    def update(self, custom_metrics):
        data = self.data.get(self.worker_id_func(), {})

        for metrics in custom_metrics:
            key = metrics["type"], metrics["key"]
            if metrics["type"] == "COUNTER":
                value = data.get(key, 0)
                data[key] = value + metrics["value"]
            elif metrics["type"] == "TIMER":
                vals, sumv = data.get(key, (list(np.zeros(len(BINS) - 1)), 0))
                # Dividing by 1000 because unit is milliseconds
                data[key] = self._update_hist(metrics["value"] / 1000, vals, sumv)
            else:
                data[key] = metrics["value"]

        self.data[self.worker_id_func()] = data

    def collect(self):
        data = dict(self.data)

        for worker, metrics in data.items():
            labels = [str(worker), my_labels["model"], my_labels["image"]]
            for (item_type, item_name), item_value in metrics.items():
                if item_type == "GAUGE":
                    yield self._expose_gauge(item_name, item_value, labels)
                elif item_type == "COUNTER":
                    yield self._expose_counter(item_name, item_value, labels)
                elif item_type == "TIMER":
                    yield self._expose_histogram(item_name, item_value, labels)

    def generate_metrics(self):
        myregistry = CollectorRegistry()
        myregistry.register(self)
        return (
            exposition.generate_latest(myregistry).decode("utf-8"),
            exposition.CONTENT_TYPE_LATEST,
        )

    @staticmethod
    def _update_hist(x, vals, sumv):
        hist = np.histogram([x], BINS)[0]
        vals = list(np.array(vals) + hist)
        return vals, sumv + x

    @staticmethod
    def _expose_gauge(name, value, labels):
        metric = GaugeMetricFamily(name, "", labels=LABELS)
        metric.add_metric(labels, value)
        return metric

    @staticmethod
    def _expose_counter(name, value, labels):
        metric = CounterMetricFamily(name, "", labels=LABELS)
        metric.add_metric(labels, value)
        return metric

    @staticmethod
    def _expose_histogram(name, value, labels):
        vals, sumv = value
        buckets = [[floatToGoString(b), v] for v, b in zip(np.cumsum(vals), BINS[1:])]

        metric = HistogramMetricFamily(name, "", labels=LABELS)
        metric.add_metric(labels, buckets, sum_value=sumv)
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
