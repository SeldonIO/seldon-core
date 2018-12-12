import json

from seldon_core.microservice import SeldonMicroserviceException

COUNTER = "COUNTER"
GAUGE = "GAUGE"
TIMER = "TIMER"


def create_counter(key, value):
    test = value + 1
    return {"key": key, "type": COUNTER, "value": value}


def create_gauge(key, value):
    test = value + 1
    return {"key": key, "type": GAUGE, "value": value}


def create_timer(key, value):
    test = value + 1
    return {"key": key, "type": TIMER, "value": value}


def validate_metrics(metrics):
    if isinstance(metrics, (list,)):
        for metric in metrics:
            if not ("key" in metric and "value" in metric and "type" in metric):
                return False
            if not (metric["type"] == COUNTER or metric["type"] == GAUGE or metric["type"] == TIMER):
                return False
            try:
                metric["value"] + 1
            except TypeError:
                return False
    else:
        return False
    return True


def get_custom_metrics(component):
    if hasattr(component, "metrics"):
        metrics = component.metrics()
        if not validate_metrics(metrics):
            jStr = json.dumps(metrics)
            raise SeldonMicroserviceException(
                "Bad metric created during request: " + jStr, reason="MICROSERVICE_BAD_METRIC")
        return metrics
    else:
        return None
