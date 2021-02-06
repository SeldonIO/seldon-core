import json
import logging
import os
from multiprocessing import Manager
from typing import Dict, List, Tuple

import numpy as np
from prometheus_client import exposition
from prometheus_client.core import (
    CollectorRegistry,
    CounterMetricFamily,
    GaugeMetricFamily,
    HistogramMetricFamily,
)
from prometheus_client.utils import floatToGoString

logger = logging.getLogger(__name__)


NONIMPLEMENTED_MSG = "NOT_IMPLEMENTED"

ENV_SELDON_DEPLOYMENT_NAME = "SELDON_DEPLOYMENT_ID"
ENV_MODEL_NAME = "PREDICTIVE_UNIT_ID"
ENV_MODEL_IMAGE = "PREDICTIVE_UNIT_IMAGE"
ENV_PREDICTOR_NAME = "PREDICTOR_ID"
ENV_PREDICTOR_LABELS = "PREDICTOR_LABELS"

FEEDBACK_KEY = "seldon_api_model_feedback"
FEEDBACK_REWARD_KEY = "seldon_api_model_feedback_reward"

COUNTER = "COUNTER"
GAUGE = "GAUGE"
TIMER = "TIMER"
HISTOGRAM = "HISTOGRAM"
_ALLOWED_METRIC_TYPES = {COUNTER, GAUGE, TIMER, HISTOGRAM}

# This sets the bins spread logarithmically between 0.001 and 30
BINS = [0] + list(np.logspace(-3, np.log10(30), 50)) + [np.inf]


def split_image_tag(tag: str) -> Tuple[str]:
    """
    Extract image name and version from an image tag.

    Parameters
    ----------
    tag
        Fully qualified docker image tag. Eg. seldonio/sklearn-iris:0.1

    Returns
    -------
        Image name, image version tuple
    """
    *name_parts, version = tag.split(":")
    return ":".join(name_parts), version


# Development placeholder
image = os.environ.get(ENV_MODEL_IMAGE, f"{NONIMPLEMENTED_MSG}:{NONIMPLEMENTED_MSG}")
model_image, model_version = split_image_tag(image)
predictor_version = json.loads(os.environ.get(ENV_PREDICTOR_LABELS, "{}")).get(
    "version", f"{NONIMPLEMENTED_MSG}"
)

legacy_mode = os.environ.get("SELDON_EXECUTOR_ENABLED", "true").lower() == "false"

DEFAULT_LABELS = {
    "deployment_name": os.environ.get(
        ENV_SELDON_DEPLOYMENT_NAME, f"{NONIMPLEMENTED_MSG}"
    ),
    "model_name": os.environ.get(ENV_MODEL_NAME, f"{NONIMPLEMENTED_MSG}"),
    "model_image": model_image,
    "model_version": model_version,
    "predictor_name": os.environ.get(ENV_PREDICTOR_NAME, f"{NONIMPLEMENTED_MSG}"),
    "predictor_version": predictor_version,
}

# Compatibility layer of tags until Seldon-Core 1.3
DEFAULT_LABELS["seldon_deployment_name"] = DEFAULT_LABELS["deployment_name"]
DEFAULT_LABELS["image_name"] = DEFAULT_LABELS["model_image"]
DEFAULT_LABELS["image_version"] = DEFAULT_LABELS["model_version"]

FEEDBACK_METRIC_METHOD_TAG = "feedback"
PREDICT_METRIC_METHOD_TAG = "predict"
INPUT_TRANSFORM_METRIC_METHOD_TAG = "inputtransform"
OUTPUT_TRANSFORM_METRIC_METHOD_TAG = "outputtransform"
ROUTER_METRIC_METHOD_TAG = "router"
AGGREGATE_METRIC_METHOD_TAG = "aggregate"
HEALTH_METRIC_METHOD_TAG = "health"


class SeldonMetrics:
    """Class to manage custom metrics stored in shared memory."""

    def __init__(self, worker_id_func=os.getpid, extra_default_labels={}):
        # We keep reference to Manager so it does not get garbage collected
        self._manager = Manager()
        self._lock = self._manager.Lock()
        self.data = self._manager.dict()
        self.worker_id_func = worker_id_func
        self._extra_default_labels = extra_default_labels

    def __del__(self):
        self._manager.shutdown()

    def update_reward(self, reward: float):
        """"Update metrics key corresponding to feedback reward counter."""
        if not reward or legacy_mode:
            return
        self.update(
            [{"type": "COUNTER", "key": FEEDBACK_KEY, "value": 1}],
            FEEDBACK_METRIC_METHOD_TAG,
        )
        self.update(
            [{"type": "COUNTER", "key": FEEDBACK_REWARD_KEY, "value": reward}],
            FEEDBACK_METRIC_METHOD_TAG,
        )

    def update(self, custom_metrics: List[Dict], method: str):
        # Read a corresponding worker's metric data with lock as Proxy objects
        # are not thread-safe, see "Thread safety of proxies" here
        # https://docs.python.org/3.7/library/multiprocessing.html#programming-guidelines
        logger.debug("Updating metrics: {}".format(custom_metrics))
        with self._lock:
            worker_data = self.data.get(self.worker_id_func(), {})
        logger.debug("Read current metrics data from shared memory")

        for metric in custom_metrics:
            metric_type = metric.get("type", "COUNTER")

            tags = metric.get("tags", {})

            tags["method"] = method

            worker_data_key = (
                metric_type,
                metric["key"],
                SeldonMetrics._generate_tags_key(tags),
            )

            if metric_type == "COUNTER":
                value = worker_data.get(worker_data_key, {}).get("value", 0)
                worker_data[worker_data_key] = {
                    "value": value + metric["value"],
                    "tags": tags,
                }
            elif metric_type in {"HISTOGRAM", "TIMER"}:
                bins = metric.get("bins", BINS)

                current_values, current_sum = worker_data.get(worker_data_key, {}).get(
                    "value",
                    (np.zeros(len(bins) - 1).tolist(), 0),
                )

                new_value = (
                    metric["value"] / 1000
                    if metric_type == "TIMER"
                    else metric["value"]
                )

                worker_data[worker_data_key] = {
                    "value": self._update_hist(
                        new_value, current_values, current_sum, bins
                    ),
                    "tags": tags,
                }
            elif metric_type == "GAUGE":
                worker_data[worker_data_key] = {
                    "value": metric["value"],
                    "tags": tags,
                }
            else:
                logger.error(f"Unkown metrics type: {metric_type}")

        # Write worker's data with lock (again - Proxy objects are not thread-safe)
        with self._lock:
            self.data[self.worker_id_func()] = worker_data
        logger.debug("Updated metrics in the shared memory.")

    def collect(self):
        # Read all workers metrics with lock to avoid other processes / threads
        # writing to it at the same time. Casting to `dict` works like reading of data.
        logger.debug("SeldonMetrics.collect called")
        with self._lock:
            data = dict(self.data)
        logger.debug("Read current metrics data from shared memory")

        for worker_id, worker_data in data.items():
            for (item_type, item_name, item_tags), item in worker_data.items():
                labels_keys, labels_values = self._merge_labels(
                    str(worker_id), item["tags"]
                )
                if item_type == "GAUGE":
                    yield self._expose_gauge(
                        item_name, item["value"], labels_keys, labels_values
                    )
                elif item_type == "COUNTER":
                    yield self._expose_counter(
                        item_name, item["value"], labels_keys, labels_values
                    )
                elif item_type in ["HISTOGRAM", "TIMER"]:
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

    def _merge_labels(self, worker, tags):
        labels = {
            **tags,
            **DEFAULT_LABELS,
            **self._extra_default_labels,
            "worker_id": str(worker),
        }
        return list(labels.keys()), list(labels.values())

    @staticmethod
    def _generate_tags_key(tags):
        return "_".join(["-".join(i) for i in tags.items()])

    @staticmethod
    def _update_hist(
        x: float, vals: List[float], sumv: float, bins: List[float]
    ) -> Tuple[List[float], float]:
        """Updated vals with x according to which bin it belongs, and add x to
        sumv.

        Args:
            x: the new value to be added to the historgram
            vals: current values of each bin in the histogram
            sumv: the sum of all x vals in history
            bins: bins for the histogram.

        Returns:
            the tuple of updated (vals, sumv)
        """
        hist = np.histogram([x], bins)[0]
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
            if not all(i in metric for i in ["key", "value", "type"]):
                return False
            if not metric["type"] in _ALLOWED_METRIC_TYPES:
                return False
            try:
                metric["value"] + 1
            except TypeError:
                return False
    else:
        return False
    return True
