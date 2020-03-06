from prometheus_client.core import HistogramMetricFamily
from prometheus_client.core import GaugeMetricFamily, CounterMetricFamily
from prometheus_client.core import CollectorRegistry
from prometheus_client import exposition


from typing import List, Dict
import multiprocessing as mp
import logging
import os


logger = logging.getLogger(__name__)


manager = mp.Manager()
shared_dict = manager.dict()

COUNTER = "COUNTER"
GAUGE = "GAUGE"
TIMER = "TIMER"



def register_metrics(metrics_list):
    logger.info(f"Registering metrics_list: {metrics_list}")
    worker_pid = os.getpid()
    data = shared_dict.get(worker_pid, {})
    metrics = Metrics(worker_pid, data)

    for item in metrics_list:
        metrics.update(item)

    shared_dict[worker_pid] = metrics.data


def collect_metrics():
    myregistry = CollectorRegistry()

    for worker, data in shared_dict.items():
        metrics = Metrics(worker, data)
        myregistry.register(metrics)

    return exposition.generate_latest(myregistry).decode("utf-8")


class Metrics:

    def __init__(self, worker_pid, data):
        self.worker_pid = worker_pid
        self.data = data

    def update(self, item):
        key = item["type"], item["key"]
        if item["type"] == "COUNTER":
            value = self.data.get(key, 0)
            self.data[key] = value + item["value"]
        else:
            self.data[key] = item["value"]

    def collect(self):
        for (item_type, item_name), item_value in self.data.items():

            if item_type not in METRICS_MAP:
                print(f"Unknown metric type {item_type}")
                continue

            metric = new_metric(item_type, item_name)
            metric.add_metric(
                [str(self.worker_pid), labels["model"], labels["image"]],
                item_value
            )
            yield metric



METRICS_MAP = {
    "COUNTER": CounterMetricFamily,
    "GAUGE": GaugeMetricFamily,
}


labels = {"model": "latest", "image": "my-image"}

def new_metric(item_type, item_name):
    return METRICS_MAP[item_type](item_name, "", labels=["worker-pid", "model", "image"])



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
