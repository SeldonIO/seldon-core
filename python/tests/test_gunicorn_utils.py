import os
import signal
import time
from typing import Set

import pytest
import requests
from prometheus_client.parser import text_string_to_metric_families


def _get_workers(
    metrics_endpoint: str = "http://127.0.0.1:6005/metrics-endpoint",
) -> Set[str]:
    workers = set()

    res = requests.get("http://127.0.0.1:6005/metrics-endpoint")
    families = text_string_to_metric_families(res.content.decode())
    for fam in families:
        for sample in fam.samples:
            worker_id = sample.labels["worker_id"]
            workers.add(worker_id)

    return workers


@pytest.mark.parametrize(
    "microservice", [{"app_name": "custom-metrics-model"}], indirect=True
)
def test_worker_exit(microservice):
    # Warm up the custom metrics
    for _ in range(5):
        res = requests.post(
            "http://127.0.0.1:9000/api/v1.0/predictions",
            json={"data": {"ndarray": [[1, 2, 3]]}},
        )
        res.raise_for_status()

    # Fetch metrics and get current list of workers
    workers = _get_workers(metrics_endpoint="http://127.0.0.1:6005/metrics-endpoint")
    assert len(workers) != 0

    # Ask Gunicorn to restart all workers (through a HUP)
    server_pid = microservice.p.pid
    os.kill(server_pid, signal.SIGHUP)
    time.sleep(1)

    # Metrics should now be empty
    empty_workers = _get_workers(
        metrics_endpoint="http://127.0.0.1:6005/metrics-endpoint"
    )
    assert len(empty_workers) == 0
