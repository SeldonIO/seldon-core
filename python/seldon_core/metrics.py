from prometheus_client.core import HistogramMetricFamily
from prometheus_client.core import GaugeMetricFamily, CounterMetricFamily
from prometheus_client.core import CollectorRegistry
from prometheus_client import exposition

from multiprocessing import Manager

from typing import List, Dict
import logging
import os


logger = logging.getLogger(__name__)


COUNTER = "COUNTER"
GAUGE = "GAUGE"
TIMER = "TIMER"


def generate_metrics(metrics):
    myregistry = CollectorRegistry()
    myregistry.register(metrics)
    return (
        exposition.generate_latest(myregistry).decode("utf-8"),
        exposition.CONTENT_TYPE_LATEST,
    )


class SeldonMetrics:
    """Class to manage custom metrics stored in shared memory."""

    def __init__(self, worker_id_func=os.getpid):
        self.manager = Manager()
        self.data = self.manager.dict()
        self.worker_id_func = worker_id_func

    def __del__(self):
        self.manager.shutdown()

    def update(self, custom_metrics):
        data = self.data.get(self.worker_id_func(), {})

        for metrics in custom_metrics:
            key = metrics["type"], metrics["key"]
            if metrics["type"] == "COUNTER":
                value = data.get(key, 0)
                data[key] = value + metrics["value"]
            else:
                data[key] = metrics["value"]

        self.data[self.worker_id_func()] = data

    def collect(self):
        data = dict(self.data)
        for worker, metrics in data.items():
            for (item_type, item_name), item_value in metrics.items():
                if item_type not in METRICS_MAP:
                    print(f"Unknown metric type {item_type}")
                    continue

                metric = METRICS_MAP[item_type](
                    item_name, "", labels=["worker-id", "model", "image"]
                )

                metric.add_metric(
                    [str(worker), labels["model"], labels["image"]], item_value
                )

                yield metric


METRICS_MAP = {
    "COUNTER": CounterMetricFamily,
    "GAUGE": GaugeMetricFamily,
}


labels = {"model": "latest", "image": "my-image"}


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
