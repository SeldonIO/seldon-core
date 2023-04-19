import json
import logging
import time
import uuid
from subprocess import run

import pytest
import requests
from tenacity import RetryError, Retrying, stop_after_attempt, wait_fixed

from seldon_core.batch_processor import start_multithreaded_batch_worker
from seldon_e2e_utils import (
    API_ISTIO_GATEWAY,
    create_random_data,
    initial_rest_request,
    rest_request,
    rest_request_ambassador,
    retry_run,
    wait_for_deployment,
    wait_for_rollout,
    wait_for_status,
)

AFTER_WAIT_SLEEP = 20
TENACITY_WAIT = 10
TENACITY_STOP_AFTER_ATTEMPT = 10


def print_logs(app, namespace):
    cmd = f"kubectl logs -lapp={app} -n {namespace}"
    run(cmd, shell=True)


class TestADServer:
    truck_json = "../../components/alibi-detect-server/cifar10-v2.json"
    truck_json_outlier = "../../components/alibi-detect-server/cifar10-v2-outlier.json"
    HEADERS = {
        "ce-namespace": "default",
        "ce-modelid": "cifar10",
        "ce-type": "io.seldon.serving.inference.request",
        "ce-id": "1234",
        "ce-source": "localhost",
        "ce-specversion": "1.0",
    }

    @pytest.mark.sequential
    def test_alibi_detect_cifar10(self, namespace):
        spec = "../resources/adserver-cifar10-od.yaml"
        name = "cifar10-od-server"
        vs_prefix = name

        retry_run(f"kubectl apply -f {spec} -n {namespace}")

        wait_for_deployment(name, namespace)

        time.sleep(AFTER_WAIT_SLEEP)

        with open(self.truck_json) as f:
            data = json.load(f)

        print(f"==== AT START ======")
        print_logs(app="cifar10-od-server", namespace=namespace)
        idx = 0
        for attempt in Retrying(
            wait=wait_fixed(TENACITY_WAIT),
            stop=stop_after_attempt(TENACITY_STOP_AFTER_ATTEMPT),
        ):
            with attempt:
                try:
                    r = requests.post(
                        f"http://localhost:8004/{vs_prefix}/",
                        json=data,
                        headers=self.HEADERS,
                    )
                    j = r.json()
                finally:
                    print(f"==== AFTER ATTEMPT {idx} ======")
                    print_logs(app="cifar10-od-server", namespace=namespace)
                    idx += 1

        assert j["data"]["is_outlier"][0] == 0
        assert j["meta"]["name"] == "OutlierVAE"
        assert j["meta"]["detector_type"] == "offline"
        assert j["meta"]["data_type"] == "image"

        with open(self.truck_json_outlier) as f:
            data = json.load(f)

        for attempt in Retrying(
            wait=wait_fixed(TENACITY_WAIT),
            stop=stop_after_attempt(TENACITY_STOP_AFTER_ATTEMPT),
        ):
            with attempt:
                r = requests.post(
                    f"http://localhost:8004/{vs_prefix}/",
                    json=data,
                    headers=self.HEADERS,
                )
                j = r.json()

        assert j["data"]["is_outlier"][0] == 1
        assert j["meta"]["name"] == "OutlierVAE"
        assert j["meta"]["detector_type"] == "offline"
        assert j["meta"]["data_type"] == "image"

        run(f"kubectl delete -f {spec} -n {namespace}", shell=True)

    @pytest.mark.sequential
    def test_alibi_detect_cifar10_rclone(self, namespace):
        spec = "../resources/adserver-cifar10-od-rclone.yaml"
        name = "cifar10-od-server-rclone"
        vs_prefix = name

        retry_run(f"kubectl apply -f {spec} -n {namespace}")

        wait_for_deployment(name, namespace)

        time.sleep(AFTER_WAIT_SLEEP)

        with open(self.truck_json) as f:
            data = json.load(f)

        for attempt in Retrying(
            wait=wait_fixed(TENACITY_WAIT),
            stop=stop_after_attempt(TENACITY_STOP_AFTER_ATTEMPT),
        ):
            with attempt:
                r = requests.post(
                    f"http://localhost:8004/{vs_prefix}/",
                    json=data,
                    headers=self.HEADERS,
                )
                j = r.json()

        assert j["data"]["is_outlier"][0] == 0
        assert j["meta"]["name"] == "OutlierVAE"
        assert j["meta"]["detector_type"] == "offline"
        assert j["meta"]["data_type"] == "image"

        with open(self.truck_json_outlier) as f:
            data = json.load(f)

        for attempt in Retrying(
            wait=wait_fixed(TENACITY_WAIT),
            stop=stop_after_attempt(TENACITY_STOP_AFTER_ATTEMPT),
        ):
            with attempt:
                r = requests.post(
                    f"http://localhost:8004/{vs_prefix}/",
                    json=data,
                    headers=self.HEADERS,
                )
                j = r.json()

        assert j["data"]["is_outlier"][0] == 1
        assert j["meta"]["name"] == "OutlierVAE"
        assert j["meta"]["detector_type"] == "offline"
        assert j["meta"]["data_type"] == "image"

        run(f"kubectl delete -f {spec} -n {namespace}", shell=True)
